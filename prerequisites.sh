#!/usr/bin/env bash
# ----------------------------------------------------------------------------
# Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
# Licensed under the Apache License, Version 2.0.
# ----------------------------------------------------------------------------
#
# Provisions a fresh WSO2 IS tenant for the demo: creates the "insurance.org" classic
# multi-tenant, its SCIM2 birthday/age custom schema extension, the portal:* custom OAuth2
# scopes (as an API Resource), two OAuth2/OIDC applications (portal-backend BFF +
# insurance-portal, both JWT access tokens), two named users (admin@gmail.com / dpo@gmail.com)
# with role-scoped permissions (portaladmin / dpo), a selfsignup role for future self-registered
# users, and self user registration with auto-login enabled — then writes the resulting client
# id/secret + tenant-qualified URLs to demo-artifacts/generated/{portal-backend,insurance-portal}.env
# for docker-compose to pick up.
#
# Every IS user this script creates (tenant admin, admin@gmail.com, dpo@gmail.com) shares the
# same demo-only password — see DEMO_USER_PASSWORD below. Never do this against a real deployment.
#
# Run by `demo.sh start` after WSO2 IS's container reports healthy — see docker-compose.yml.
#
# ***********************************************************************************
# * KNOWN GAPS (grep "KNOWN GAP" below for details, best-effort, does not abort):        *
# * 1. Self-registered users don't automatically get the 'selfsignup' role — no IS       *
# *    mechanism for this was found; needs either a manual admin step per user or        *
# *    app-side role assignment on first login.                                          *
# * 2. Role/scope enforcement: portal-backend's own API resource scopes ARE now bound     *
# *    via Authorized APIs with policy "RBAC" (see section 10) — this is the real         *
# *    mechanism (the `scopeValidators` field tried earlier was the wrong/legacy one),    *
# *    but whether IS actually restricts issued scopes per-caller-role based on this      *
# *    binding hasn't been confirmed end-to-end with a real token. Worth a live check     *
# *    if this needs to be relied on for access control, not just IS-console visibility.  *
# ***********************************************************************************

set -euo pipefail

IS_BASE_URL="${IS_BASE_URL:-https://wso2is:9443}"
SUPER_ADMIN_USER="${SUPER_ADMIN_USER:-admin}"
SUPER_ADMIN_PASSWORD="${SUPER_ADMIN_PASSWORD:-admin}"
TENANT_DOMAIN="${TENANT_DOMAIN:-insurance.org}"
TENANT_ADMIN_USER="${TENANT_ADMIN_USER:-admin}"
# Final desired tenant admin password. Deliberately weak (demo only) — "admin" fails IS's
# default password policy (min length 8 + upper/lower/digit), which also blocks using it as
# the tenant-creation password directly (that tenant's OWN policy doesn't exist to relax yet
# until the tenant itself exists). So tenant creation below bootstraps with a policy-compliant
# temp password, then this script relaxes the policy and resets to this value right after.
TENANT_ADMIN_PASSWORD="${TENANT_ADMIN_PASSWORD:-admin}"
TENANT_ADMIN_BOOTSTRAP_PASSWORD="${TENANT_ADMIN_BOOTSTRAP_PASSWORD:-Abc@1234}"
TENANT_ADMIN_EMAIL="${TENANT_ADMIN_EMAIL:-admin@insurance.org}"
OUT_DIR="${OUT_DIR:-demo-artifacts/generated}"

# Demo-only shared password for every IS user this script creates (tenant admin + the two
# named portal users below) — never do this against a real deployment.
DEMO_USER_PASSWORD="${DEMO_USER_PASSWORD:-Abc@1234}"

# BFF_AUTH__SCOPES list, mirrored from portal/backend/.env.example — kept in sync by hand.
PORTAL_SCOPES="openid email profile internal_login portal:consents:read:any portal:consents:write:any portal:consents:read:self portal:consents:write:self portal:elements:read portal:elements:write portal:purposes:read portal:purposes:write portal:grievances:manage portal:grievances:read:self portal:grievances:write:self portal:event-subscriptions:manage"
ADMIN_SCOPE="internal_user_mgt_delete"
API_RESOURCE_IDENTIFIER="openfgc-portal"
SELF_SIGNUP_SCOPES="portal:consents:read:self portal:consents:write:self portal:grievances:read:self portal:grievances:write:self"
DPO_SCOPES="portal:grievances:manage portal:consents:read:self portal:consents:write:self"

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
  --arg pass "$TENANT_ADMIN_BOOTSTRAP_PASSWORD" \
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

# The tenant admin's password may already be at its final value (a prior run of this script
# already completed the reset below) or still at the bootstrap value (freshly created tenant,
# or a prior run crashed before resetting it) — probe with the final password first since
# that's the terminal state, falling back to bootstrap if that probe fails.
CURL_AUTH="$TENANT_ADMIN_QUALIFIED:$TENANT_ADMIN_PASSWORD"
probe_status=$(curl -sS -k -o /dev/null -w '%{http_code}' -u "$CURL_AUTH" "$TENANT_BASE/api/server/v1/applications?limit=1")
if [ "$probe_status" != "200" ]; then
  log "Tenant admin password not yet at its final value — relaxing password policy and resetting it ..."
  CURL_AUTH="$TENANT_ADMIN_QUALIFIED:$TENANT_ADMIN_BOOTSTRAP_PASSWORD"
  # Relax this tenant's password policy so "$TENANT_ADMIN_PASSWORD" (demo-weak, e.g. "admin")
  # passes validation — confirmed live that IS's default policy (min length 8 + upper/lower/
  # digit) rejects it otherwise, and that this validation-rules endpoint only accepts a body
  # WITHOUT the "field" key (the GET response includes one, but PUTting it back verbatim 400s
  # with a generic "not in the expected format" — same class of asymmetry as the authorized-
  # apis policyIdentifier bug found earlier).
  relax_body='{"rules":[{"validator":"LengthValidator","properties":[{"key":"min.length","value":"1"},{"key":"max.length","value":"64"}]},{"validator":"NumeralValidator","properties":[{"key":"min.length","value":"0"}]},{"validator":"UpperCaseValidator","properties":[{"key":"min.length","value":"0"}]},{"validator":"LowerCaseValidator","properties":[{"key":"min.length","value":"0"}]},{"validator":"SpecialCharacterValidator","properties":[{"key":"min.length","value":"0"}]}]}'
  resp=$(curl_json PUT "$TENANT_BASE/api/server/v1/validation-rules/password" "$relax_body")
  status=$(http_status "$resp")
  [ "$status" = "200" ] || warn "Relaxing the password policy returned HTTP $status (continuing): $(http_body "$resp")"

  admin_lookup_resp=$(curl_json GET "$TENANT_BASE/scim2/Users?filter=userName+eq+$TENANT_ADMIN_USER")
  admin_user_id=$(http_body "$admin_lookup_resp" | jq -r '.Resources[0].id // empty')
  if [ -n "$admin_user_id" ]; then
    password_patch_body=$(jq -n --arg pass "$TENANT_ADMIN_PASSWORD" \
      '{schemas:["urn:ietf:params:scim:api:messages:2.0:PatchOp"], Operations:[{op:"replace", value:{password:$pass}}]}')
    resp=$(curl_json PATCH "$TENANT_BASE/scim2/Users/$admin_user_id" "$password_patch_body")
    status=$(http_status "$resp")
    if [ "$status" = "200" ]; then
      log "Tenant admin password reset to its final value."
      CURL_AUTH="$TENANT_ADMIN_QUALIFIED:$TENANT_ADMIN_PASSWORD"
    else
      die "Resetting the tenant admin's password failed (HTTP $status): $(http_body "$resp")"
    fi
  else
    die "Could not resolve the tenant admin user id to reset its password"
  fi
fi

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

# --- 5. Custom OAuth2 scopes (portal:*), as an API Resource ------------------------------
# Verified live: the standalone OAuth2 Scope Management API (`/api/server/v1/scopes`) 403s
# even for the tenant's own super-admin-equivalent owner user — IS 7.x replaced it with an
# "API Resources" model instead. Creating the resource WITH all its scopes in a single POST
# works cleanly (201, all scopes attached) — no separate per-scope call needed.
log "Creating API resource '$API_RESOURCE_IDENTIFIER' with portal:* scopes ..."
scopes_json=$(
  for scope in $PORTAL_SCOPES; do
    # NOTE: deliberately `if [[ ]]`, not `case`/esac — macOS's bash 3.2 has a confirmed
    # parser bug where a `case` pattern's closing `)` inside a `$(...)` command substitution
    # is miscounted against the substitution's own closing paren, breaking the whole script.
    if [[ "$scope" != portal:* ]]; then
      continue
    fi
    jq -n --arg name "$scope" --arg desc "OpenFGC portal scope: $scope" \
      '{name:$name, displayName:$name, description:$desc}'
  done | jq -s '.'
)
api_resource_body=$(jq -n --argjson scopes "$scopes_json" --arg id "$API_RESOURCE_IDENTIFIER" \
  '{name:"OpenFGC Portal API", identifier:$id, description:"Custom scopes for the OpenFGC portal (consents, elements, purposes, grievances, event-subscriptions)", requiresAuthorization:true, scopes:$scopes}')
resp=$(curl_json POST "$TENANT_BASE/api/server/v1/api-resources" "$api_resource_body")
status=$(http_status "$resp")
if [ "$status" = "201" ] || [ "$status" = "409" ]; then
  log "API resource '$API_RESOURCE_IDENTIFIER' ready (HTTP $status)."
else
  warn "API resource creation returned HTTP $status (continuing — role permission bindings below may fail to resolve scopes): $(http_body "$resp")"
fi

# --- 6. OAuth2/OIDC applications ---------------------------------------------------------
create_app() {
  local name="$1" callback="$2" grants_json="$3" extra_redirect_uris="${4:-}"
  local body app_id client_id client_secret existing_id create_status create_id create_line callback_value
  # WSO2 IS validates post_logout_redirect_uri against the SAME callbackURLs field as the
  # login redirect — confirmed live (error message literally says "...registered callback
  # URI"), there is no separate post-logout-uri field. When there's more than one valid
  # target (login callback + one-or-more post-logout landing pages), IS's callbackURLs
  # accepts a `regexp=(a|b|c)` value covering all of them.
  if [ -n "$extra_redirect_uris" ]; then
    callback_value="regexp=($callback|$(echo "$extra_redirect_uris" | tr ' ' '|'))"
  else
    callback_value="$callback"
  fi
  # NOTE: accessToken.type is deliberately NOT set here at create time — verified live that
  # including `accessToken:{type:"JWT"}` in the application-creation POST body makes IS
  # return a bare 500 (APP-65006, "Unexpected Processing Error"), reproducible even for a
  # brand new app name. Setting it via a follow-up PUT on inbound-protocols/oidc (below,
  # after creation) works fine (200) — so JWT is applied as a second step instead.
  body=$(jq -n --arg name "$name" --arg cb "$callback_value" --argjson grants "$grants_json" \
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

  # PUT (full replace) the same config back with accessToken.type switched to JWT — the BFF
  # and insurance-portal validate access tokens locally (JWT signature/claims), not via IS
  # token introspection, so the default opaque "Default" token type doesn't work for them.
  jwt_put_body=$(http_body "$oidc_resp" | jq 'del(.clientSecret) | .accessToken.type = "JWT"')
  jwt_resp=$(curl_json PUT "$TENANT_BASE/api/server/v1/applications/$app_id/inbound-protocols/oidc" "$jwt_put_body")
  jwt_status=$(http_status "$jwt_resp")
  [ "$jwt_status" = "200" ] || warn "Setting accessToken.type=JWT for '$name' returned HTTP $jwt_status (continuing with the default opaque token type): $(http_body "$jwt_resp")"

  # Role audience = ORGANIZATION, subject = username (not the internal user UUID) — both
  # verified live via PATCH on the application resource itself (not inbound-protocols/oidc).
  # allowedAudience already defaults to ORGANIZATION on a freshly created app in this IS
  # version, but it's set explicitly here rather than relied on, since a default isn't a
  # guarantee across versions.
  #
  # NOTE: setting subject.claim.uri alone is NOT enough — confirmed via IS Console that the
  # "Assign alternate subject identifier" toggle shows as checked but flags itself as
  # ineffective ("can be assigned only if user attributes are selected") unless the username
  # claim is ALSO added to claimConfiguration.requestedClaims. Adding it here makes IS
  # auto-populate a matching claimMappings entry too (verified via GET after this PATCH).
  #
  # skipLogoutConsent: true — skips IS's "are you sure you want to log out" consent screen
  # on the way out, matching the login flow (which already runs consentless for this demo).
  app_patch_body='{"associatedRoles":{"allowedAudience":"ORGANIZATION"},"claimConfiguration":{"subject":{"claim":{"uri":"http://wso2.org/claims/username"}},"requestedClaims":[{"claim":{"uri":"http://wso2.org/claims/username"},"mandatory":true}]},"advancedConfigurations":{"skipLogoutConsent":true}}'
  app_patch_resp=$(curl_json PATCH "$TENANT_BASE/api/server/v1/applications/$app_id" "$app_patch_body")
  app_patch_status=$(http_status "$app_patch_resp")
  [ "$app_patch_status" = "200" ] || warn "Setting role audience/subject claim for '$name' returned HTTP $app_patch_status (continuing): $(http_body "$app_patch_resp")"

  printf '%s %s %s\n' "$app_id" "$client_id" "$client_secret"
}

log "Creating OAuth2 application for portal-backend ..."
bff_app_line="$(create_app "portal-backend" \
  "http://localhost:8081/auth/callback" '["authorization_code","client_credentials","refresh_token"]' \
  "http://localhost:5173/")"
read -r bff_app_id bff_client_id bff_client_secret <<<"$bff_app_line"

log "Creating OAuth2 application for insurance-portal ..."
ip_app_line="$(create_app "insurance-portal" \
  "http://localhost:3020/auth/callback" '["authorization_code","refresh_token"]' \
  "http://localhost:3020/home.html http://localhost:3020/delegate.html")"
read -r ip_app_id ip_client_id ip_client_secret <<<"$ip_app_line"

# --- 7. Portal users (admin@gmail.com, dpo@gmail.com) -----------------------------------
# Delete+recreate on 409 (same rationale as create_app: idempotent across re-runs without a
# full `demo.sh stop`), so the password always ends up as $DEMO_USER_PASSWORD even if a prior
# run used something else.
create_user() {
  local username="$1" given_name="$2" family_name="$3"
  local body user_id create_status
  body=$(jq -n --arg u "$username" --arg p "$DEMO_USER_PASSWORD" --arg given "$given_name" --arg family "$family_name" \
    '{schemas:["urn:ietf:params:scim:schemas:core:2.0:User"], userName:$u, password:$p, emails:[{value:$u, primary:true}], name:{givenName:$given, familyName:$family}}')
  resp=$(curl_json POST "$TENANT_BASE/scim2/Users" "$body")
  create_status=$(http_status "$resp")
  if [ "$create_status" = "409" ]; then
    log "User '$username' already exists — deleting and recreating with the current demo password ..."
    list_resp=$(curl_json GET "$TENANT_BASE/scim2/Users?filter=userName+eq+$username")
    existing_id=$(http_body "$list_resp" | jq -r '.Resources[0].id // empty')
    [ -n "$existing_id" ] || die "User '$username' exists but its id could not be resolved for deletion"
    curl_json DELETE "$TENANT_BASE/scim2/Users/$existing_id" >/dev/null
    resp=$(curl_json POST "$TENANT_BASE/scim2/Users" "$body")
    create_status=$(http_status "$resp")
  fi
  [ "$create_status" = "201" ] || die "User '$username' creation failed (HTTP $create_status): $(http_body "$resp")"
  user_id=$(http_body "$resp" | jq -r '.id')
  log "User '$username' ready (id=$user_id)."
  printf '%s\n' "$user_id"
}

log "Creating portal user 'admin@gmail.com' ..."
admin_user_id="$(create_user "admin@gmail.com" "Portal" "Admin")"
log "Creating portal user 'dpo@gmail.com' ..."
dpo_user_id="$(create_user "dpo@gmail.com" "Data Protection" "Officer")"

# --- 8. Roles: portaladmin, dpo, selfsignup ----------------------------------------------
# Every role/permission binding below was verified live: role creation accepts `permissions`
# (referencing our API resource's scope NAMEs directly, e.g. {"value":"portal:consents:read:any"})
# and `users` (by SCIM user id) in the SAME create call — no separate binding step needed.
# Delete+recreate on 409, same idempotency rationale as apps/users above.
#
# org_id is derived from the tenant's own always-present "system" role rather than guessed or
# hardcoded — verified live to be the same organization id every role's `audience` needs.
system_role_resp=$(curl_json GET "$TENANT_BASE/scim2/v2/Roles?filter=displayName+eq+system")
org_id=$(http_body "$system_role_resp" | jq -r '.Resources[0].audience.value // empty')
[ -n "$org_id" ] || die "Could not resolve the tenant's organization id (needed for role audience) from the default 'system' role"

create_role() {
  local role_name="$1" scopes="$2" user_id="${3:-}"
  local permissions_json users_json body create_status existing_id
  permissions_json=$(for s in $scopes; do jq -n --arg s "$s" '{value:$s}'; done | jq -s '.')
  if [ -n "$user_id" ]; then
    users_json=$(jq -n --arg u "$user_id" '[{value:$u}]')
  else
    users_json='[]'
  fi
  body=$(jq -n --arg name "$role_name" --arg org "$org_id" --argjson perms "$permissions_json" --argjson users "$users_json" \
    '{displayName:$name, audience:{value:$org, type:"organization"}, permissions:$perms, users:$users}')
  resp=$(curl_json POST "$TENANT_BASE/scim2/v2/Roles" "$body")
  create_status=$(http_status "$resp")
  if [ "$create_status" = "409" ]; then
    log "Role '$role_name' already exists — deleting and recreating with the current scope/user list ..."
    list_resp=$(curl_json GET "$TENANT_BASE/scim2/v2/Roles?filter=displayName+eq+$role_name")
    existing_id=$(http_body "$list_resp" | jq -r '.Resources[0].id // empty')
    [ -n "$existing_id" ] || die "Role '$role_name' exists but its id could not be resolved for deletion"
    curl_json DELETE "$TENANT_BASE/scim2/v2/Roles/$existing_id" >/dev/null
    resp=$(curl_json POST "$TENANT_BASE/scim2/v2/Roles" "$body")
    create_status=$(http_status "$resp")
  fi
  [ "$create_status" = "201" ] || die "Role '$role_name' creation failed (HTTP $create_status): $(http_body "$resp")"
  log "Role '$role_name' ready."
}

log "Creating role 'portaladmin' (all portal:* scopes) for admin@gmail.com ..."
# NOTE: `if [[ ]]`, not `case`/esac — see the bash-3.2-inside-$()-parser-bug note above.
all_portal_scopes=$(for s in $PORTAL_SCOPES; do if [[ "$s" == portal:* ]]; then echo "$s"; fi; done)
create_role "portaladmin" "$all_portal_scopes" "$admin_user_id"

log "Creating role 'dpo' ($DPO_SCOPES) for dpo@gmail.com ..."
create_role "dpo" "$DPO_SCOPES" "$dpo_user_id"

log "Creating role 'selfsignup' ($SELF_SIGNUP_SCOPES) — no user assigned; this is the role" \
  "self-registered users should get, see the known gap noted below."
create_role "selfsignup" "$SELF_SIGNUP_SCOPES" ""

# --- 9. Self user registration, with auto-login -- KNOWN GAP: default role not auto-assigned
# Verified live via the Identity Governance API (User Onboarding category, self-sign-up
# connector): enabling self-registration, disabling the account lock so new accounts are
# immediately usable, and SelfRegistration.AutoLogin.Enable together produce exactly "sign up
# -> automatically logged in" with no email-verification step in the way. ReCaptcha is
# disabled too since it needs its own site-key configuration this demo doesn't set up.
#
# UNRESOLVED: could not find any IS mechanism (governance connector property, userstore
# setting, or role attribute) that auto-assigns a role to newly self-registered users —
# searched User Onboarding, Account Management, and Other Settings connector categories, and
# the userstore config, with no match. The 'selfsignup' role above exists with the right
# scopes but nothing assigns it automatically; a self-registered user currently gets no
# portal:* scopes until an admin manually adds them to it (IS Console > Roles > selfsignup >
# Users, or SCIM2). If this needs to be automatic, it likely has to be done in application
# code (e.g. portal-backend's or insurance-portal's OIDC callback handler assigning the role
# via SCIM2 on a user's first login) rather than IS config.
log "Enabling self user registration with auto-login (known gap: default role isn't auto-assigned, see comment above) ..."
governance_body='{
  "operation": "UPDATE",
  "properties": [
    {"name": "SelfRegistration.Enable", "value": "true"},
    {"name": "SelfRegistration.LockOnCreation", "value": "false"},
    {"name": "SelfRegistration.ReCaptcha", "value": "false"},
    {"name": "SelfRegistration.AutoLogin.Enable", "value": "true"}
  ]
}'
resp=$(curl_json PATCH "$TENANT_BASE/api/server/v1/identity-governance/VXNlciBPbmJvYXJkaW5n/connectors/c2VsZi1zaWduLXVw" "$governance_body")
status=$(http_status "$resp")
if [ "$status" = "200" ]; then
  log "Self user registration + auto-login enabled."
else
  warn "Enabling self-registration returned HTTP $status (continuing): $(http_body "$resp")"
fi

# IS 7.x also has a NEWER, separate "Registration Flow" visual builder (IS Console >
# Flows > Registration) with its OWN isEnabled/isAutoLoginEnabled config — distinct from
# the classic governance connector above (this only became apparent from IS Console's own
# Network tab: it PUTs .../api/server/v1/flow and PATCHes .../api/server/v1/flow/config).
# The default template ships with an OTP-email-verification step and asks for email as well
# as username/password; this replaces it with a plain username+password Sign Up step going
# straight to End, and sets this flow's own isEnabled/isAutoLoginEnabled explicitly rather
# than relying on the classic connector properties above to also govern it.
#
# The username field's Label is deliberately "Email" while its identifier/placeholder stay
# on http://wso2.org/claims/username — display-only, not a real change of attribute (this
# demo's users are self-registered with an email address AS their username, so the field is
# genuinely asking for what looks like an email even though IS still treats it as username).
log "Simplifying the registration flow (username+password only, no OTP step) ..."
registration_flow_body='{
  "flowType": "REGISTRATION",
  "steps": [
    {
      "id": "view_zs8h",
      "type": "VIEW",
      "data": {
        "components": [
          {"id":"typography_aoj2","category":"DISPLAY","type":"TYPOGRAPHY","variant":"H3","config":{"text":"Sign Up"}},
          {
            "id": "form_qx79",
            "category": "BLOCK",
            "type": "FORM",
            "components": [
              {"id":"input_4a8c","category":"FIELD","type":"INPUT","variant":"TEXT","config":{"type":"text","hint":"","label":"Email","required":true,"placeholder":"Enter your username","identifier":"http://wso2.org/claims/username"}},
              {"id":"input_o5yj","category":"FIELD","type":"INPUT","variant":"PASSWORD","config":{"identifier":"password","type":"password","hint":"","label":"Password","required":true,"placeholder":"Enter your password"}},
              {"id":"rich_text_yves","category":"DISPLAY","type":"RICH_TEXT","config":{"text":"<p class=\"rich-text-paragraph\"><br></p><p class=\"rich-text-paragraph\"><span class=\"rich-text-pre-wrap\">When you click Sign Up, you are agreeing to our </span><a href=\"{{branding.termsOfUseUrl}}\" target=\"_blank\" rel=\"noopener noreferrer\" class=\"rich-text-link\"><span class=\"rich-text-pre-wrap\">Terms of Service</span></a><span class=\"rich-text-pre-wrap\"> and </span><a href=\"{{branding.privacyPolicyUrl}}\" target=\"_blank\" rel=\"noopener noreferrer\" class=\"rich-text-link\"><span class=\"rich-text-pre-wrap\">Privacy Policy</span></a></p>"}},
              {"id":"button_sldf","category":"ACTION","type":"BUTTON","variant":"PRIMARY","action":{"type":"EXECUTOR","executor":{"name":"PasswordProvisioningExecutor"},"next":"END"},"config":{"type":"submit","text":"Sign Up"}},
              {"id":"rich_text_2723","category":"DISPLAY","type":"RICH_TEXT","config":{"text":"<p class=\"rich-text-paragraph\"><br></p><p class=\"rich-text-paragraph\"><span class=\"rich-text-pre-wrap\">Already have an account? </span><a href=\"{{application.callbackOrAccessUrl}}\" target=\"_blank\" rel=\"noopener noreferrer\" class=\"rich-text-link\"><span class=\"rich-text-pre-wrap\">Sign in</span></a></p>"}}
            ],
            "config": {}
          }
        ]
      },
      "size": {"height": 700.0, "width": 350.0},
      "position": {"x": 300.0, "y": 200.0}
    },
    {
      "id": "END",
      "type": "END",
      "data": {
        "components": [
          {"id":"typography_u2oz","category":"DISPLAY","type":"TYPOGRAPHY","variant":"H3","config":{"text":"Registration Successful!"}},
          {"id":"display_hgvm","category":"DISPLAY","type":"RICH_TEXT","config":{"text":"<p class=\"rich-text-paragraph rich-text-align-center\"><span class=\"rich-text-pre-wrap\">You can now sign in with your new account.</span></p>"}}
        ],
        "action": {"type":"EXECUTOR","executor":{"name":"UserProvisioningExecutor"}}
      },
      "size": {"height": 257.0, "width": 350.0},
      "position": {"x": 800.0, "y": 300.0}
    }
  ]
}'
resp=$(curl_json PUT "$TENANT_BASE/api/server/v1/flow?flowType=REGISTRATION" "$registration_flow_body")
status=$(http_status "$resp")
if [ "$status" = "200" ]; then
  log "Registration flow simplified (username+password only, no OTP)."
else
  warn "Simplifying the registration flow returned HTTP $status (continuing): $(http_body "$resp")"
fi

log "Enabling the registration flow itself, with auto-login ..."
flow_config_body='{"flowType":"REGISTRATION","isEnabled":true,"flowCompletionConfigs":{"isAutoLoginEnabled":"true","isEmailVerificationEnabled":"false","isAccountLockOnCreationEnabled":"false","isFlowCompletionNotificationEnabled":"false"}}'
resp=$(curl_json PATCH "$TENANT_BASE/api/server/v1/flow/config" "$flow_config_body")
status=$(http_status "$resp")
if [ "$status" = "200" ]; then
  log "Registration flow enabled with auto-login."
else
  warn "Enabling the registration flow returned HTTP $status (continuing): $(http_body "$resp")"
fi

# --- 10. Authorize the BFF app for its own API resource scopes + the admin scope ---------
# Two bindings, both via the "Authorized APIs" endpoint: (a) all portal:* scopes of our own
# "openfgc-portal" API resource, and (b) the built-in internal_user_mgt_delete scope guarding
# the admin SCIM2 Users API.
#
# RESOLVED (previously a long-standing "known gap" — every attempt 400'd with a generic
# "UE-10000 Invalid Request. Provided request body content is not in the expected format"):
# the request field is `policyIdentifier`, NOT `policyId` — confirmed by capturing the real
# request IS Console itself sends when you click "Authorize resource" in the UI. The API
# asymmetrically returns the bound value back as `policyId` on GET, which is what every
# previous attempt was (reasonably, but wrongly) modeled on.
log "Authorizing portal-backend for its own API resource scopes ..."
own_resource_resp=$(curl_json GET "$TENANT_BASE/api/server/v1/api-resources?filter=identifier+eq+$API_RESOURCE_IDENTIFIER") || true
own_resource_id=$(http_body "$own_resource_resp" | jq -r '.apiResources[0].id // empty' 2>/dev/null || true)
if [ -n "${own_resource_id:-}" ]; then
  own_scopes_resp=$(curl_json GET "$TENANT_BASE/api/server/v1/api-resources/$own_resource_id/scopes")
  own_scopes=$(http_body "$own_scopes_resp" | jq -c '[.[].name]' 2>/dev/null || echo '[]')
  authz_body=$(jq -n --arg id "$own_resource_id" --argjson scopes "$own_scopes" '{id:$id, policyIdentifier:"RBAC", scopes:$scopes}')
  resp=$(curl_json POST "$TENANT_BASE/api/server/v1/applications/$bff_app_id/authorized-apis" "$authz_body") || true
  status=$(http_status "$resp")
  if [ "$status" = "200" ] || [ "$status" = "201" ] || [ "$status" = "409" ]; then
    log "'$API_RESOURCE_IDENTIFIER' authorized for portal-backend (HTTP $status)."
  else
    warn "Authorized-APIs binding for '$API_RESOURCE_IDENTIFIER' returned HTTP $status: $(http_body "$resp")"
  fi
else
  warn "Could not resolve the '$API_RESOURCE_IDENTIFIER' API resource id — skipping."
fi

log "Authorizing portal-backend for scope '$ADMIN_SCOPE' ..."
api_resources_resp=$(curl_json GET "$TENANT_BASE/api/server/v1/api-resources?filter=name+eq+%22SCIM2+Users+API%22") || true
scim_api_id=$(http_body "$api_resources_resp" | jq -r '.apiResources[] | select(.type=="TENANT") | .id' 2>/dev/null | head -n1 || true)
if [ -n "${scim_api_id:-}" ]; then
  authz_body=$(jq -n --arg id "$scim_api_id" --arg scope "$ADMIN_SCOPE" '{id:$id, policyIdentifier:"RBAC", scopes:[$scope]}')
  resp=$(curl_json POST "$TENANT_BASE/api/server/v1/applications/$bff_app_id/authorized-apis" "$authz_body") || true
  status=$(http_status "$resp")
  if [ "$status" = "200" ] || [ "$status" = "201" ] || [ "$status" = "409" ]; then
    log "'$ADMIN_SCOPE' authorized for portal-backend (HTTP $status)."
  else
    warn "Authorized-APIs binding for '$ADMIN_SCOPE' returned HTTP $status: $(http_body "$resp")"
  fi
else
  warn "Could not resolve the SCIM2 Users API resource id — skipping admin scope authorization."
fi

# --- 11. Write generated env files -------------------------------------------------------
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
