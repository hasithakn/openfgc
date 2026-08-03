#!/usr/bin/env bash
# ----------------------------------------------------------------------------
# Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
# Licensed under the Apache License, Version 2.0.
# ----------------------------------------------------------------------------
#
# Orchestrates the dockerized demo stack: build (native per-module builds -> demo-artifacts/
# -> docker compose build), start (bring the stack up, provision WSO2 IS, start the rest),
# stop (tear everything down, including volumes, for a clean next start).
#
# See docker-compose.yml for the service topology and prerequisites.sh for what gets
# provisioned in WSO2 IS.

set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")"

ROOT_DIR="$(pwd)"
ARTIFACTS_DIR="$ROOT_DIR/demo-artifacts"

log() { echo "[demo] $*"; }
die() { echo "[demo] ERROR: $*" >&2; exit 1; }
require_cmd() { command -v "$1" >/dev/null 2>&1 || die "'$1' is required but not found on PATH"; }

usage() {
  cat <<'EOF'
Usage: ./demo.sh <build|start|stop>

  build   Build all 6 modules with their own native toolchain, assemble demo-artifacts/,
          then `docker compose build` every image from that staged output.
  start   Bring up mysql/wso2is/openfgc/event-framework/webhook-listener, wait for them to
          report healthy, run prerequisites.sh to provision the WSO2 IS tenant + apps, then
          start portal-backend/portal-frontend/insurance-portal.
  stop    `docker compose down -v` — removes every container AND volume (MySQL + WSO2 IS
          data), so the next `start` provisions a genuinely fresh tenant.
EOF
}

# ---------------------------------------------------------------------------------------
# build
# ---------------------------------------------------------------------------------------
cmd_build() {
  require_cmd go
  require_cmd task
  require_cmd mvn
  require_cmd docker
  require_cmd npm

  local npm_client=npm
  command -v pnpm >/dev/null 2>&1 && npm_client=pnpm

  log "Resetting $ARTIFACTS_DIR ..."
  rm -rf "$ARTIFACTS_DIR"
  mkdir -p "$ARTIFACTS_DIR"/{openfgc,portal-backend,portal-frontend,event-framework,webhook-listener,insurance-portal,dbscripts,generated}

  # Docker images run linux regardless of the host OS this script runs on — both native Go
  # builds below must cross-compile for linux (keeping the host's native GOARCH, so this
  # works unmodified on both amd64 and arm64 dev machines building for a same-arch daemon).
  local go_arch
  go_arch="$(go env GOARCH)"

  log "Building openfgc (consent-server) for linux/$go_arch ..."
  (cd "$ROOT_DIR" && ./build.sh build linux "$go_arch")
  cp -R "$ROOT_DIR/target/server/." "$ARTIFACTS_DIR/openfgc/"
  cp "$ROOT_DIR/consent-server/cmd/server/repository/conf/deployment.docker.yaml" \
    "$ARTIFACTS_DIR/openfgc/repository/conf/deployment.yaml"

  log "Building portal-backend for linux/$go_arch ..."
  (cd "$ROOT_DIR/portal/backend" && GOOS=linux GOARCH="$go_arch" task build)
  cp "$ROOT_DIR/portal/backend/bin/portal-backend" "$ARTIFACTS_DIR/portal-backend/bff"

  log "Building portal-frontend (VITE_API_BASE_URL=http://localhost:8081) ..."
  # VITE_AUTH_LOGOUT_ALLOWED_ORIGINS must include the WSO2 IS origin — the frontend's own
  # logout() only navigate()s to a logoutUrl whose origin is in this build-time allowlist,
  # otherwise it throws "navigation URL origin is not allowed" and the UI shows a generic
  # "Unable to sign out" toast even though the backend's own /auth/logout call succeeded.
  (cd "$ROOT_DIR/portal/frontend" && \
    VITE_API_BASE_URL=http://localhost:8081 VITE_ORG_ID=insurance.org \
    VITE_AUTH_LOGOUT_ALLOWED_ORIGINS=https://wso2is:9443 "$npm_client" run build)
  cp -R "$ROOT_DIR/portal/frontend/dist" "$ARTIFACTS_DIR/portal-frontend/dist"
  cp "$ROOT_DIR/docker/portal-frontend.nginx.conf" "$ARTIFACTS_DIR/portal-frontend/nginx.conf"

  log "Building event-framework ..."
  (cd "$ROOT_DIR/event-framework" && mvn -q clean package -DskipTests)
  cp "$ROOT_DIR/event-framework/event-notification-runner/target/event-notification-runner-1.0.0-SNAPSHOT.jar" \
    "$ARTIFACTS_DIR/event-framework/event-notification-runner.jar"

  log "Staging webhook-listener ..."
  cp "$ROOT_DIR/event-framework/webhook-listener.js" "$ARTIFACTS_DIR/webhook-listener/webhook-listener.js"

  log "Building insurance-portal (npm ci --omit=dev) ..."
  (cd "$ROOT_DIR/dpdp-insuarance-portal" && npm ci --omit=dev)
  for f in server.js auth.js config.json package.json package-lock.json; do
    cp "$ROOT_DIR/dpdp-insuarance-portal/$f" "$ARTIFACTS_DIR/insurance-portal/$f"
  done
  cp -R "$ROOT_DIR/dpdp-insuarance-portal/public" "$ARTIFACTS_DIR/insurance-portal/public"
  cp -R "$ROOT_DIR/dpdp-insuarance-portal/node_modules" "$ARTIFACTS_DIR/insurance-portal/node_modules"

  log "Assembling dbscripts (consent_mgt, enf_db) ..."
  {
    echo "CREATE DATABASE IF NOT EXISTS consent_mgt;"
    echo "USE consent_mgt;"
    echo
    cat "$ROOT_DIR/consent-server/dbscripts/db_schema_mysql.sql"
  } > "$ARTIFACTS_DIR/dbscripts/01-consent-mgt.sql"
  {
    echo "CREATE DATABASE IF NOT EXISTS enf_db;"
    echo "USE enf_db;"
    echo
    cat "$ROOT_DIR/event-framework/event-notification-dao/src/main/resources/dbscripts/mysql.sql"
  } > "$ARTIFACTS_DIR/dbscripts/02-enf.sql"

  # docker compose parses env_file references for every subcommand, including `build` — these
  # are placeholders until prerequisites.sh overwrites them with real content during `start`.
  touch "$ARTIFACTS_DIR/generated/portal-backend.env" "$ARTIFACTS_DIR/generated/insurance-portal.env"

  log "docker compose build ..."
  (cd "$ROOT_DIR" && docker compose build)

  log "Build complete. Artifacts staged in $ARTIFACTS_DIR/"
}

# ---------------------------------------------------------------------------------------
# start
# ---------------------------------------------------------------------------------------
wait_healthy() {
  local service="$1" timeout="${2:-120}" waited=0 cid status
  log "Waiting for '$service' to become healthy (timeout ${timeout}s) ..."
  while true; do
    cid=$(docker compose ps -q "$service" 2>/dev/null || true)
    if [ -z "$cid" ]; then
      status="starting"
    else
      status=$(docker inspect --format '{{.State.Health.Status}}' "$cid" 2>/dev/null || echo "starting")
    fi
    if [ "$status" = "healthy" ]; then
      log "'$service' is healthy."
      return 0
    fi
    if [ "$waited" -ge "$timeout" ]; then
      die "'$service' did not become healthy within ${timeout}s (last status: $status). Check: docker compose logs $service"
    fi
    sleep 5
    waited=$((waited + 5))
  done
}

check_hosts_entry() {
  if ! grep -qE '(^|[[:space:]])wso2is([[:space:]]|$)' /etc/hosts 2>/dev/null; then
    die "Missing /etc/hosts entry — the browser needs to resolve 'wso2is' the same way containers on the compose network do. Add this one-time entry, then re-run './demo.sh start':
  echo '127.0.0.1  wso2is' | sudo tee -a /etc/hosts"
  fi
}

cmd_start() {
  require_cmd docker
  [ -d "$ARTIFACTS_DIR" ] || die "demo-artifacts/ not found — run './demo.sh build' first."
  check_hosts_entry

  log "Starting wave 1: mysql, wso2is, openfgc, event-framework, webhook-listener ..."
  docker compose up -d mysql wso2is openfgc event-framework webhook-listener

  wait_healthy mysql 120
  wait_healthy wso2is 300
  wait_healthy openfgc 120
  wait_healthy event-framework 60

  log "Running prerequisites.sh (provisioning the insurance.org tenant + OAuth apps in IS) ..."
  IS_BASE_URL="https://wso2is:9443" "$ROOT_DIR/prerequisites.sh"

  log "Starting wave 2: portal-backend, portal-frontend, insurance-portal ..."
  docker compose up -d portal-backend portal-frontend insurance-portal
  wait_healthy portal-backend 60

  cat <<'EOF'

============================================================
  Demo stack is up. Go here:

    http://localhost:3020

============================================================
  Other endpoints (admin/debugging): consent portal http://localhost:5173, WSO2 IS console
  https://wso2is:9443/carbon (tenant: insurance.org), OpenFGC API http://localhost:8060,
  webhook listener http://localhost:9091.
============================================================
EOF
}

# ---------------------------------------------------------------------------------------
# stop
# ---------------------------------------------------------------------------------------
cmd_stop() {
  require_cmd docker
  log "Stopping and removing all containers + volumes (mysql, wso2is data) ..."
  docker compose down -v
  log "Stopped. The next './demo.sh start' will provision a fresh tenant from scratch."
}

case "${1:-}" in
  build) cmd_build ;;
  start) cmd_start ;;
  stop) cmd_stop ;;
  *) usage; exit 1 ;;
esac
