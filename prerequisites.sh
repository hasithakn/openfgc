#!/usr/bin/env bash
# ----------------------------------------------------------------------------
# Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
# Licensed under the Apache License, Version 2.0.
# ----------------------------------------------------------------------------
#
# Provisions a fresh WSO2 IS tenant for the demo: creates the "insurance.org" classic
# multi-tenant, its SCIM2 birthday/age custom schema extension, the portal:* custom OAuth2
# scopes, and two OAuth2/OIDC applications (portal-backend BFF + insurance-portal), then
# writes the resulting client id/secret + tenant-qualified URLs to
# demo-artifacts/generated/{portal-backend,insurance-portal}.env for docker-compose to pick up.
#
# Run by `demo.sh start` after WSO2 IS's container reports healthy — see docker-compose.yml.
#
# ***********************************************************************************
# * KNOWN RISK: tenant creation, user creation, and OAuth2 application creation below *
# * are well-documented, stable WSO2 IS REST APIs. The local-claim / SCIM2 extension  *
# * mapping and custom-scope registration steps are marked NEEDS LIVE VERIFICATION —  *
# * their exact request/response shape can vary by IS patch version and has not been  *
# * confirmed against a running wso2is:7.3.0 container. Those steps are best-effort:  *
# * they log their raw response and continue rather than aborting the whole script.   *
# ***********************************************************************************

set -euo pipefail

IS_BASE_URL="${IS_BASE_URL:-https://wso2is:9443}"
SUPER_ADMIN_USER="${SUPER_ADMIN_USER:-admin}"
SUPER_ADMIN_PASSWORD="${SUPER_ADMIN_PASSWORD:-admin}"
TENANT_DOMAIN="${TENANT_DOMAIN:-insurance.org}"
TENANT_ADMIN_USER="${TENANT_ADMIN_USER:-admin}"
TENANT_ADMIN_PASSWORD="${TENANT_ADMIN_PASSWORD:-Admin123!}"
TENANT_ADMIN_EMAIL="${TENANT_ADMIN_EMAIL:-admin@insurance.org}"
OUT_DIR="${OUT_DIR:-demo-artifacts/generated}"

# BFF_AUTH__SCOPES list, mirrored from portal/backend/.env.example — kept in sync by hand.
PORTAL_SCOPES="openid email profile internal_login portal:consents:read:any portal:consents:write:any portal:consents:read:self portal:consents:write:self portal:elements:read portal:elements:write portal:purposes:read portal:purposes:write portal:grievances:manage portal:grievances:read:self portal:grievances:write:self portal:event-subscriptions:manage"
ADMIN_SCOPE="internal_user_mgt_delete"

curl_json() {
  # curl_json <method> <url> [json-body] — always -k (self-signed demo cert), always sends
  # and expects JSON, prints "<body>\n<http-status>" so callers can split the two.
  local method="$1" url="$2" body="${3:-}"
  if [ -n "$body" ]; then
    curl -sS -k -w '\n%{http_code}' -u "${CURL_AUTH:-$SUPER_ADMIN_USER:$SUPER_ADMIN_PASSWORD}" \
      -X "$method" "$url" -H 'Content-Type: application/json' -d "$body"
  else
    curl -sS -k -w '\n%{http_code}' -u "${CURL_AUTH:-$SUPER_ADMIN_USER:$SUPER_ADMIN_PASSWORD}" \
      -X "$method" "$url" -H 'Content-Type: application/json'
  fi
}

http_status() { tail -n1 <<<"$1"; }
http_body() { sed '$d' <<<"$1"; }

# post_json_location <url> <json-body> — POST that returns 201 with an EMPTY body and only
# the created resource's id in the Location header (confirmed live: application creation,
# like claim-dialect creation, behaves this way). Echoes "<status> <id-or-empty>".
post_json_location() {
  local url="$1" body="$2" header_file status location id
  header_file=$(mktemp)
  status=$(curl -sS -k -o /dev/null -D "$header_file" -w '%{http_code}' \
    -u "${CURL_AUTH:-$SUPER_ADMIN_USER:$SUPER_ADMIN_PASSWORD}" \
    -X POST "$url" -H 'Content-Type: application/json' -d "$body")
  location=$(grep -i '^location:' "$header_file" | tail -n1 | tr -d '\r' | sed -E 's/^[Ll]ocation:[[:space:]]*//')
  rm -f "$header_file"
  id="${location##*/}"
  printf '%s %s\n' "$status" "$id"
}

log() { echo "[prerequisites] $*" >&2; }
warn() { echo "[prerequisites] WARNING: $*" >&2; }
die() { echo "[prerequisites] ERROR: $*" >&2; exit 1; }

for cmd in curl jq; do
  command -v "$cmd" >/dev/null 2>&1 || die "'$cmd' is required but not found on PATH"
done

mkdir -p "$OUT_DIR"

# --- 1. Wait for IS readiness ------------------------------------------------------------
log "Waiting for WSO2 IS at $IS_BASE_URL ..."
is_up=false
for _ in $(seq 1 60); do
  status=$(curl -sS -k -o /dev/null -w '%{http_code}' "$IS_BASE_URL/api/health-check/v1.0/health" 2>/dev/null || echo "000")
  if [[ "$status" =~ ^2 ]]; then
    is_up=true
    break
  fi
  sleep 5
done
$is_up || die "WSO2 IS did not report healthy at $IS_BASE_URL/api/health-check/v1.0/health in time"
log "IS is up."

# --- 2. Create tenant (idempotent) -------------------------------------------------------
log "Creating tenant '$TENANT_DOMAIN' ..."
tenant_body=$(jq -n \
  --arg domain "$TENANT_DOMAIN" \
  --arg user "$TENANT_ADMIN_USER" \
  --arg pass "$TENANT_ADMIN_PASSWORD" \
  --arg email "$TENANT_ADMIN_EMAIL" \
  '{domain:$domain, owners:[{username:$user, password:$pass, email:$email, firstname:"Insurance", lastname:"Admin", provisioningMethod:"inline-password"}]}')
resp=$(curl_json POST "$IS_BASE_URL/api/server/v1/tenants" "$tenant_body")
status=$(http_status "$resp")
tenant_error_code=$(http_body "$resp" | jq -r '.code // empty' 2>/dev/null || true)
case "$status" in
  201) log "Tenant created." ;;
  409) log "Tenant already exists — continuing." ;;
  400)
    if [ "$tenant_error_code" = "TM-60009" ]; then
      log "Tenant already exists — continuing."
    else
      die "Tenant creation failed (HTTP $status): $(http_body "$resp")"
    fi
    ;;
  *) die "Tenant creation failed (HTTP $status): $(http_body "$resp")" ;;
esac

TENANT_BASE="$IS_BASE_URL/t/$TENANT_DOMAIN"
TENANT_ADMIN_QUALIFIED="${TENANT_ADMIN_USER}@${TENANT_DOMAIN}"
CURL_AUTH="$TENANT_ADMIN_QUALIFIED:$TENANT_ADMIN_PASSWORD"

# --- 3. Local claims: birthday, age ------------------------------------------------------
# Verified against a live wso2is:7.3.0 container: the local dialect's path param is the
# lowercase literal "local" (not "LOCAL"), and claim creation requires an explicit
# attributeMapping to a userstore attribute — both confirmed via this IS instance's own
# existing local claims (GET .../claim-dialects/local/claims).
create_local_claim() {
  local claim_uri="$1" display_name="$2" mapped_attribute="$3"
  local body
  body=$(jq -n --arg uri "$claim_uri" --arg name "$display_name" --arg attr "$mapped_attribute" \
    '{claimURI:$uri, displayName:$name, dataType:"string", supportedByDefault:true, required:false, readOnly:false, attributeMapping:[{mappedAttribute:$attr, userstore:"PRIMARY"}]}')
  resp=$(curl_json POST "$TENANT_BASE/api/server/v1/claim-dialects/local/claims" "$body")
  status=$(http_status "$resp")
  if [ "$status" = "201" ] || [ "$status" = "409" ]; then
    log "Local claim '$claim_uri' ready (HTTP $status)."
  else
    warn "Local claim '$claim_uri' creation returned HTTP $status: $(http_body "$resp")"
  fi
}
log "Creating local claims (birthday, age) ..."
create_local_claim "http://wso2.org/claims/birthday" "Birthday" "birthday"
create_local_claim "http://wso2.org/claims/age" "Age" "age"

# --- 4. SCIM2 custom extension dialect + external claim mapping --------------------------
# Verified against a live wso2is:7.3.0 container: dialect/claim creation responses are 201
# with an EMPTY body (the id is only in the Location header) — but the dialect id is simply
# base64url(dialectURI) with no padding, so it's simplest to just compute it directly rather
# than parse a Location header or do a lookup round-trip.
SCIM_EXT_DIALECT="urn:scim:schemas:extension:custom:User"
dialect_id=$(printf '%s' "$SCIM_EXT_DIALECT" | base64 | tr -d '\n' | tr '+/' '-_' | tr -d '=')
log "Creating SCIM2 extension dialect '$SCIM_EXT_DIALECT' (id=$dialect_id) ..."
dialect_body=$(jq -n --arg uri "$SCIM_EXT_DIALECT" '{dialectURI:$uri}')
resp=$(curl_json POST "$TENANT_BASE/api/server/v1/claim-dialects" "$dialect_body")
status=$(http_status "$resp")
if [ "$status" = "201" ] || [ "$status" = "409" ]; then
  log "SCIM2 extension dialect ready (HTTP $status)."
else
  warn "SCIM2 extension dialect creation returned HTTP $status: $(http_body "$resp")"
  dialect_id=""
fi

if [ -n "$dialect_id" ]; then
  map_external_claim() {
    local local_uri="$1" ext_uri="$2"
    local body
    body=$(jq -n --arg uri "$ext_uri" --arg local "$local_uri" '{claimURI:$uri, mappedLocalClaimURI:$local}')
    resp=$(curl_json POST "$TENANT_BASE/api/server/v1/claim-dialects/$dialect_id/claims" "$body")
    status=$(http_status "$resp")
    if [ "$status" = "201" ] || [ "$status" = "409" ]; then
      log "SCIM2 external claim '$ext_uri' ready (HTTP $status)."
    else
      warn "SCIM2 external claim '$ext_uri' creation returned HTTP $status: $(http_body "$resp")"
    fi
  }
  map_external_claim "http://wso2.org/claims/birthday" "$SCIM_EXT_DIALECT:birthday"
  map_external_claim "http://wso2.org/claims/age" "$SCIM_EXT_DIALECT:age"
else
  warn "Could not determine SCIM2 extension dialect id — skipping external claim mapping. Birthday/age will not appear over SCIM2 until this is fixed up manually (IS Console > Attributes > SCIM2 > Custom Schema)."
fi

# --- 5. Custom OAuth2 scopes (portal:*) -- STILL NEEDS LIVE VERIFICATION -----------------
# WSO2 IS does not require pre-registering arbitrary (non "internal_"-prefixed) OAuth2 scopes
# for them to be requestable/issued — this step is a best-effort attempt in case scope
# validation is enforced in this environment; failures here do not block the rest of the demo.
#
# Verified live: every call below 403s even for the tenant's own super-admin-equivalent owner
# user (confirmed via SCIM2 to hold both the tenant "admin" org role and "Administrator"
# console role) — so this isn't a permissions gap on our end. The likely explanation: IS 7.x
# replaced the standalone OAuth2 Scope Management API with an "API Resources" model (each API
# resource owns its own /scopes sub-collection, discovered while debugging the admin-scope
# authorization step below) — i.e. registering portal:* scopes here may need to go through
# creating a dedicated API Resource for the OpenFGC portal API first, not this endpoint. Not
# pursued further: a client_credentials token request for portal-backend including an
# unregistered portal:* scope alongside internal_user_mgt_delete was accepted by the token
# endpoint (200, no invalid_scope error) — evidence, not proof (couldn't introspect the
# resulting opaque token to confirm the scope actually landed in it; introspection needs its
# own auth setup not covered here). Likely a nice-to-have IS-console-visibility gap rather
# than a functional blocker, but treat as unconfirmed.
log "Registering portal:* custom scopes (best-effort) ..."
for scope in $PORTAL_SCOPES; do
  case "$scope" in
    portal:*) ;;
    *) continue ;;
  esac
  body=$(jq -n --arg name "$scope" --arg desc "OpenFGC portal scope: $scope" '{name:$name, displayName:$name, description:$desc}')
  resp=$(curl_json POST "$TENANT_BASE/api/server/v1/scopes" "$body") || true
  status=$(http_status "$resp")
  if [ "$status" != "201" ] && [ "$status" != "409" ]; then
    warn "Scope '$scope' registration returned HTTP $status (continuing): $(http_body "$resp")"
  fi
done

# --- 6. OAuth2/OIDC applications ---------------------------------------------------------
create_app() {
  local name="$1" callback="$2" grants_json="$3"
  local body app_id client_id client_secret existing_id create_status create_id create_line
  body=$(jq -n --arg name "$name" --arg cb "$callback" --argjson grants "$grants_json" \
    '{name:$name, inboundProtocolConfiguration:{oidc:{grantTypes:$grants, callbackURLs:[$cb], publicClient:false, pkce:{mandatory:true, supportPlainTransformAlgorithm:false}}}}')
  # Verified live: like claim-dialect creation, this POST returns 201 with an EMPTY body —
  # the new application's id is only in the Location header.
  #
  # NOTE: deliberately using command substitution + a here-string for `read`, NOT
  # `read ... < <(...)` process substitution — macOS's default /bin/bash is the ancient
  # bash 3.2, which has a confirmed bug where `set -e` inside a process-substitution
  # subshell silently kills the whole script. Here-strings avoid that subshell entirely.
  create_line="$(post_json_location "$TENANT_BASE/api/server/v1/applications" "$body")"
  read -r create_status create_id <<<"$create_line"
  if [ "$create_status" = "409" ]; then
    # Re-running prerequisites.sh against an already-provisioned tenant (e.g. iterating
    # without a full `demo.sh stop`) — the app already exists. IS does not return a
    # previously-created client secret on lookup, so delete and recreate for a usable one.
    log "Application '$name' already exists — deleting and recreating for a fresh client secret ..."
    list_resp=$(curl_json GET "$TENANT_BASE/api/server/v1/applications?filter=name+eq+$name")
    existing_id=$(http_body "$list_resp" | jq -r '.applications[0].id // empty')
    [ -n "$existing_id" ] || die "Application '$name' exists but its id could not be resolved for deletion"
    curl_json DELETE "$TENANT_BASE/api/server/v1/applications/$existing_id" >/dev/null
    create_line="$(post_json_location "$TENANT_BASE/api/server/v1/applications" "$body")"
    read -r create_status create_id <<<"$create_line"
  fi
  [ "$create_status" = "201" ] || die "Application '$name' creation failed (HTTP $create_status)"
  app_id="$create_id"
  [ -n "$app_id" ] || die "Application '$name' created but no id could be parsed from its Location header"
  log "Application '$name' created (id=$app_id)."

  # Client id/secret can be returned inline on create in some IS versions; fall back to a
  # dedicated GET for the inbound OIDC config to be safe across versions.
  oidc_resp=$(curl_json GET "$TENANT_BASE/api/server/v1/applications/$app_id/inbound-protocols/oidc")
  oidc_status=$(http_status "$oidc_resp")
  [ "$oidc_status" = "200" ] || die "Fetching inbound OIDC config for '$name' failed (HTTP $oidc_status): $(http_body "$oidc_resp")"
  client_id=$(http_body "$oidc_resp" | jq -r '.clientId')
  client_secret=$(http_body "$oidc_resp" | jq -r '.clientSecret')
  [ -n "$client_id" ] && [ "$client_id" != "null" ] || die "Application '$name' has no clientId in its inbound OIDC config"

  printf '%s %s %s\n' "$app_id" "$client_id" "$client_secret"
}

log "Creating OAuth2 application for portal-backend ..."
bff_app_line="$(create_app "portal-backend" \
  "http://localhost:8081/auth/callback" '["authorization_code","client_credentials","refresh_token"]')"
read -r bff_app_id bff_client_id bff_client_secret <<<"$bff_app_line"

log "Creating OAuth2 application for insurance-portal ..."
ip_app_line="$(create_app "insurance-portal" \
  "http://localhost:3020/auth/callback" '["authorization_code","refresh_token"]')"
read -r ip_app_id ip_client_id ip_client_secret <<<"$ip_app_line"

# --- 7. Authorize the BFF app for the admin scope -- STILL NEEDS LIVE VERIFICATION -------
# internal_user_mgt_delete is a built-in IS scope guarding the admin SCIM2 Users API; binding
# an application to it goes through the "Authorized APIs" endpoint. Best-effort: log and
# continue on any failure — account deletion via the BFF's client_credentials flow may need
# manual fixup in IS Console.
#
# Verified live: the API resource is named "SCIM2 Users API" (not "SCIM2 Users"), and there
# are TWO resources with that name — an "ORGANIZATION"-type one (/o/scim2/Users, for IS's B2B
# org feature) and the "TENANT"-type one (/scim2/Users) that the BFF's admin-scoped SCIM2
# calls actually hit; only the latter is correct here. What's still unresolved: every
# authorized-apis POST body shape tried below returns "UE-10000 Invalid Request" even with
# the correct resource id and a real scope name from that resource's own /scopes list — the
# right field names/structure for this endpoint were not found. Left as a known gap; bind
# the scope manually in IS Console (Applications > portal-backend > API Authorization) if
# account deletion needs to work.
log "Attempting to authorize portal-backend for scope '$ADMIN_SCOPE' (best-effort) ..."
api_resources_resp=$(curl_json GET "$TENANT_BASE/api/server/v1/api-resources?filter=name+eq+%22SCIM2+Users+API%22") || true
scim_api_id=$(http_body "$api_resources_resp" | jq -r '.apiResources[] | select(.type=="TENANT") | .id' 2>/dev/null | head -n1 || true)
if [ -n "${scim_api_id:-}" ]; then
  authz_body=$(jq -n --arg id "$scim_api_id" --arg scope "$ADMIN_SCOPE" '{id:$id, policyId:"RBAC", scopes:[$scope]}')
  resp=$(curl_json POST "$TENANT_BASE/api/server/v1/applications/$bff_app_id/authorized-apis" "$authz_body") || true
  warn "Authorized-APIs binding response (HTTP $(http_status "$resp")) — known unresolved gap, see comment above: $(http_body "$resp")"
else
  warn "Could not resolve the SCIM2 Users API resource id — skipping admin scope authorization. Bind '$ADMIN_SCOPE' to the portal-backend app manually in IS Console if account deletion needs to work."
fi

# --- 8. Write generated env files --------------------------------------------------------
log "Writing generated env files to $OUT_DIR/ ..."

cat > "$OUT_DIR/portal-backend.env" <<EOF
BFF_AUTH__ISSUER_URL=$TENANT_BASE/oauth2/token
BFF_AUTH__CLIENT_ID=$bff_client_id
BFF_AUTH__CLIENT_SECRET=$bff_client_secret
BFF_AUTH__RESOURCE_AUDIENCE=$bff_client_id
BFF_AUTH__SCOPES=$PORTAL_SCOPES
BFF_AUTH__ADMIN_SCOPE=$ADMIN_SCOPE
BFF_AUTH__ORG_ID_CLAIM=org_handle
BFF_IDENTITY_SERVER__SCIM_BASE_URL=$TENANT_BASE
BFF_IDENTITY_SERVER__SCIM_AGE_ATTRIBUTE_PATH=$SCIM_EXT_DIALECT:age
BFF_IDENTITY_SERVER__SCIM_BIRTHDAY_ATTRIBUTE_PATH=$SCIM_EXT_DIALECT:birthday
EOF

cat > "$OUT_DIR/insurance-portal.env" <<EOF
IS_ISSUER_URL=$TENANT_BASE/oauth2/token
IS_CLIENT_ID=$ip_client_id
IS_CLIENT_SECRET=$ip_client_secret
SCIM_BIRTHDAY_ATTRIBUTE_PATH=$SCIM_EXT_DIALECT:birthday
EOF

log "Done. Tenant '$TENANT_DOMAIN' provisioned; tenant admin is '$TENANT_ADMIN_QUALIFIED'."
