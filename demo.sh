#!/usr/bin/env bash
# ----------------------------------------------------------------------------
# Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
# Licensed under the Apache License, Version 2.0.
# ----------------------------------------------------------------------------
#
# Orchestrates the dockerized demo stack: build (docker compose build — every image builds
# from source inside its own Dockerfile, so this needs nothing but Docker on the host), start
# (bring the stack up, provision WSO2 IS, start the rest, auto-building first if needed),
# stop (stop containers, keep images), clean (remove everything — containers, images,
# volumes, generated files — for a truly from-scratch rebuild).
#
# See docker-compose.yml for the service topology and prerequisites.sh for what gets
# provisioned in WSO2 IS.

set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")"

ROOT_DIR="$(pwd)"
ARTIFACTS_DIR="$ROOT_DIR/demo-artifacts"
BUILT_SERVICES="wso2is openfgc event-framework webhook-listener portal-backend portal-frontend insurance-portal"

log() { echo "[demo] $*"; }
die() { echo "[demo] ERROR: $*" >&2; exit 1; }
require_cmd() { command -v "$1" >/dev/null 2>&1 || die "'$1' is required but not found on PATH"; }

usage() {
  cat <<'EOF'
Usage: ./demo.sh <build|start|stop|clean>

  build   `docker compose build` — every image builds from source inside its own
          Dockerfile (Go/Node/Maven toolchains only exist inside the build containers).
          Requires nothing but Docker on the host.
  start   Builds first if any image is missing, then brings up mysql/wso2is/openfgc/
          event-framework/webhook-listener, waits for them to report healthy, runs
          prerequisites.sh to provision the WSO2 IS tenant + apps, then starts
          portal-backend/portal-frontend/insurance-portal.
  stop    `docker compose down -v` — removes every container AND volume (MySQL + WSO2 IS
          data), so the next `start` provisions a genuinely fresh tenant. Keeps built images.
  clean   Removes everything: containers, volumes, built images, and demo-artifacts/ — the
          next `build`/`start` compiles completely from scratch.
EOF
}

# ---------------------------------------------------------------------------------------
# build
# ---------------------------------------------------------------------------------------
cmd_build() {
  require_cmd docker

  # docker compose parses env_file references for every subcommand, including `build` — these
  # are placeholders until prerequisites.sh overwrites them with real content during `start`.
  mkdir -p "$ARTIFACTS_DIR/generated"
  touch "$ARTIFACTS_DIR/generated/portal-backend.env" "$ARTIFACTS_DIR/generated/insurance-portal.env"

  log "docker compose build (this compiles every module from source — Go/Node/Maven only run inside the build containers, not on this host) ..."
  docker compose build

  log "Build complete."
}

# Returns success (0) if every build-based service already has an image; failure otherwise —
# used by `start` to decide whether to build automatically first.
all_images_built() {
  local service
  for service in $BUILT_SERVICES; do
    [ -n "$(docker compose images -q "$service" 2>/dev/null)" ] || return 1
  done
  return 0
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
  check_hosts_entry

  # Belt-and-suspenders: docker compose parses env_file references for every subcommand, and
  # these placeholders might be missing even when images are already built (e.g. after a
  # manual `docker compose down -v` outside demo.sh, or a partially-cleaned state).
  mkdir -p "$ARTIFACTS_DIR/generated"
  touch "$ARTIFACTS_DIR/generated/portal-backend.env" "$ARTIFACTS_DIR/generated/insurance-portal.env"

  if ! all_images_built; then
    log "One or more images aren't built yet — building first ..."
    cmd_build
  fi

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
  log "Stopped. Built images are kept — './demo.sh start' will reuse them. Use './demo.sh clean' to also remove images and force a from-scratch rebuild."
}

# ---------------------------------------------------------------------------------------
# clean
# ---------------------------------------------------------------------------------------
cmd_clean() {
  require_cmd docker
  log "Removing all containers, volumes, and built images ..."
  docker compose down -v --rmi local --remove-orphans

  log "Removing $ARTIFACTS_DIR/ ..."
  rm -rf "$ARTIFACTS_DIR"

  # Leftover host-side build output from the old (pre-docker-only) workflow, when demo.sh
  # build ran go/task/npm/mvn directly on the host — harmless to leave, but cleaning them up
  # is exactly what "clean" implies, and none of them are needed anymore.
  local leftover_dirs=(
    "$ROOT_DIR/target"
    "$ROOT_DIR/portal/backend/bin"
    "$ROOT_DIR/portal/frontend/dist"
    "$ROOT_DIR/event-framework/event-notification-runner/target"
    "$ROOT_DIR/event-framework/event-notification-dao/target"
    "$ROOT_DIR/event-framework/event-notification-service/target"
    "$ROOT_DIR/event-framework/event-notification-endpoint/target"
    "$ROOT_DIR/dpdp-insuarance-portal/node_modules"
  )
  local dir
  for dir in "${leftover_dirs[@]}"; do
    if [ -e "$dir" ]; then
      log "Removing leftover $dir ..."
      rm -rf "$dir"
    fi
  done

  log "Clean. The next './demo.sh build' (or './demo.sh start') compiles everything from scratch."
}

case "${1:-}" in
  build) cmd_build ;;
  start) cmd_start ;;
  stop) cmd_stop ;;
  clean) cmd_clean ;;
  *) usage; exit 1 ;;
esac
