# Portal BFF Implementation Plan
**Confidential OIDC BFF · Split-Token Cookies · JWKS Validation · OpenFGC Integration**

---

## 1. Purpose and Architecture

The portal backend is both an OIDC confidential client and the protected API boundary. It performs login, authorization-code exchange, token refresh, logout, and JWT validation without a frontend IdP SDK. Access and refresh tokens use split-token transport: JavaScript sends the readable first half while the browser supplies the secure HTTP-only second half. The BFF reconstructs and validates the complete access token on every protected request before deriving identity or authorization context.

```
┌────────────────────────────────────────────────────┐
│ React Portal                                       │
│ - No IdP SDK or client secret                      │
│ - Starts login through BFF auth endpoints          │
│ - Sends Authorization: Bearer <access-token-part-1>│
└───────────────────────────┬────────────────────────┘
                            │ HTTPS
                            ▼
┌──────────────────────────────────────────────────────┐
│ Portal Backend / BFF (Go, port 8080)                 │
│ - OIDC confidential client and OAuth resource server │
│ - Reconstructs tokens from request + HTTP-only cookie│
│ - Validates ID/access tokens using IdP JWKS          │
│ - Enforces route, scope, organization, and ownership │
│ - Maps approved requests to OpenFGC                  │
└───────────────────────────┬──────────────────────────┘
                            │ private network
                            ▼
┌──────────────────────────────────────────────────────┐
│ OpenFGC server (port 8060)                           │
│ - Internal downstream API                            │
│ - Receives only BFF-derived trusted context headers  │
└──────────────────────────────────────────────────────┘

          ┌──────────────────────────────────────┐
          │ IdP                                  │
          │ Authorization server / OIDC provider │
          └──────────────────────────────────────┘
```

### Responsibilities

| Component | Responsibilities | Must not do |
|---|---|---|
| React Portal | Start BFF login/logout, read token part 1, attach it as a partial Bearer value, call refresh, and decode the ID token for profile display | Use an IdP SDK, hold a client secret, reconstruct complete access/refresh tokens, or make authorization decisions from UI data |
| Portal BFF | Act as the OIDC confidential client; exchange codes; issue, reconstruct, refresh, and clear split-token cookies; validate JWTs; derive principal; authorize requests; protect OpenFGC boundary | Expose complete access/refresh tokens to JavaScript or trust token claims before JWKS validation |
| IdP | Authenticate users and issue tokens | Be bypassed by the portal or BFF |
| OpenFGC | Execute internal consent operations | Be reachable directly from the browser/public network |

### Assumptions

- The selected IdP issues signed JWT access tokens for the portal API audience and supports OIDC discovery/JWKS.
- The BFF is registered as an OAuth confidential client using Authorization Code flow. The client secret is available only to the BFF.
- The confidential-client flow uses OAuth `state` and S256 PKCE with one-time, short-lived HTTP-only transaction cookies. OIDC `nonce` remains deferred to a separate hardening change.
- The frontend and BFF are deployed on the same host in production, normally behind one ingress, so the host-only part-1 cookies are readable by the portal. If they use different origins on that host, CORS credentials and exact origin allowlisting are required.
- The validated token `sub` is the same identifier OpenFGC uses in `userIds`.
- The BFF and OpenFGC communicate on a private network. OpenFGC is available at `http://localhost:8060` in local development.
- OpenFGC trusts BFF-derived identity/context headers. If OpenFGC later validates user access tokens itself, that is a separate downstream-authentication design decision.

---

## 2. Core Design Decisions

| Decision | Rationale |
|---|---|
| BFF performs OIDC as a confidential client | Keeps the client secret and token endpoint interaction outside the browser and removes the frontend IdP SDK dependency. |
| Stateless login correlation and S256 PKCE | Binds each callback to one initiating browser using one-time HTTP-only state/verifier cookies without a server session, shared store, encryption key, or signing key. |
| Split-token transport | A complete access or refresh token is not JavaScript-readable; the readable half alone is unusable away from the browser session. |
| Partial Bearer header plus HTTP-only cookie | The frontend sends access-token part 1 as `Authorization: Bearer <part-1>` and the browser supplies part 2 as a secure HTTP-only cookie. |
| BFF remains a JWT resource server | Reconstructing a token is transport handling only; the BFF still validates it through issuer discovery/JWKS on every protected request. |
| Refresh token split across request body and cookie | Refresh-token part 1 is posted to `/auth/refresh`; part 2 remains HTTP-only and is never returned in a response body. |
| JavaScript-readable ID token | The frontend may decode the ID token for display/profile data, but neither the frontend nor BFF uses those browser-decoded claims for authorization. |
| JWT validation through issuer discovery/JWKS | Validates ID and access tokens from the configured IdP while supporting signing-key rotation. |
| Explicit claim-to-principal mapping | Prevents business authorization from depending on unstructured token payloads. |
| Canonical scope registry | `internal/system/auth/scopes.go` is the single runtime source of truth for portal scope names; route policies, configuration validation, and tests reference its typed constants. |
| Allowlisted BFF routes | Prevents an unrestricted proxy to OpenFGC. |
| BFF-derived upstream headers only | Browser-provided identity, organization, and client headers are never trusted. |
| OpenFGC private-only exposure | Makes the BFF the mandatory policy-enforcement point. |

---

## 3. Authentication and Request Flows

### 3.1 Frontend Authentication Contract

The frontend uses a small portal-owned auth client, not an IdP SDK. It knows only the BFF auth routes, the readable cookie names, and the partial-token transport contract.

```
PortalAuthClient
  - isAuthenticated()
  - getUserProfile()          // reconstruct and decode ID-token parts locally
  - login()                   // navigate to /auth/login
  - refresh()                 // POST refresh-token part 1
  - logout(returnTo?)         // POST /auth/logout
  - attachAccessTokenPart1(request)
```

Portal components must not parse token claims for authorization. The API client reads access-token part 1 from its secure, JavaScript-readable cookie and sends `Authorization: Bearer <part-1>`. It must use credentials-enabled requests so the browser also sends the HTTP-only part-2 cookie.

### 3.2 Login and Callback

```
1. The frontend navigates to `GET /auth/login`.
2. The BFF generates cryptographically random OAuth state and PKCE verifier values, stores them in short-lived HTTP-only cookies, and redirects to the IdP with the state and S256 code challenge.
3. The user authenticates at the IdP.
4. The IdP redirects to `GET /auth/callback?code=...&state=...`.
5. The BFF consumes the transaction cookies, validates the returned state using constant-time comparison, and exchanges the code with the matching PKCE verifier and its confidential-client authentication method.
6. The BFF validates the ID token through discovery/JWKS, including signature, allowed algorithm,
   issuer, client audience, and expiry. OIDC nonce validation remains deferred.
7. The BFF validates the access token through discovery/JWKS, splits the returned tokens, sets the
   token cookies, and redirects to the configured portal URL.
```

The BFF must reject missing or duplicate codes/state values, missing or duplicate transaction cookies, state mismatch, replayed callbacks, token-exchange failures, and invalid returned tokens. Transaction cookies are consumed before code exchange and never reused. Authorization codes, state, PKCE verifiers, and tokens must never be logged. State and PKCE provide login-CSRF and authorization-code interception protection; OIDC nonce-based ID-token replay correlation remains deferred.

### 3.3 Token Cookie and Transport Contract

Tokens are split at a deterministic midpoint and reconstructed only in BFF auth middleware/services. Cookie names are configurable; the logical defaults below describe the contract.

| Token | Part 1 | Part 2 | Request transport |
|---|---|---|---|
| Access token | JavaScript-readable secure cookie `__Host-portal-at-p1` | Secure HTTP-only cookie `__Host-portal-at-p2` | Part 1 in `Authorization: Bearer <part-1>`; part 2 automatically attached as a cookie |
| Refresh token | JavaScript-readable secure cookie `__Host-portal-rt-p1` | Secure HTTP-only cookie `__Host-portal-rt-p2` | Part 1 in the `POST /auth/refresh` body; part 2 automatically attached as a cookie |
| ID token | JavaScript-readable secure cookie `__Host-portal-id-p1` | JavaScript-readable secure cookie `__Host-portal-id-p2` | Frontend may reconstruct and decode it for display/profile data; the BFF may reconstruct it as an end-session hint only after appropriate validation; never accepted as API credentials |

Login transactions use two additional configurable, secure HTTP-only cookies: OAuth state and the PKCE verifier. They contain independent cryptographically random values, expire within at most ten minutes, use `SameSite=Lax` by default, and are cleared before callback validation or token exchange. They are not encrypted because they are ephemeral correlation secrets, and tampering can only invalidate the login transaction. No server-side session or BFF-owned cryptographic key is introduced. Production `__Host-` names require `Secure`, no `Domain`, and `Path=/`; local HTTP development may use non-`__Host-` names. A new login in the same browser replaces the previous outstanding transaction, so only the most recently initiated login can complete.

All production cookies use `Secure`, `Path=/`, no `Domain`, and an explicitly configured `SameSite` value. Access/refresh part-2 cookies use `HttpOnly`; part-1 and both ID-token cookies are JavaScript-readable. Splitting the ID token is for cookie-size handling only—the complete ID token remains available to JavaScript. Cookie expiry must not exceed the corresponding token expiry. Development cookie names may omit the `__Host-` prefix only when local HTTP operation makes secure cookies impossible.

Each cookie must remain below a conservative browser-compatible size limit. If any token cannot fit within its defined cookie parts, authentication fails with a generic error; the BFF must not silently truncate it. Minimize unnecessary ID-token claims rather than adding unbounded cookie fragments.

The split limits complete-token exfiltration but does not neutralize XSS: injected same-origin JavaScript can still issue requests while the browser supplies part 2. A restrictive frontend CSP and normal XSS controls remain mandatory.

### 3.4 Protected BFF Request

```
React Portal
  Authorization: Bearer <access-token-part-1>
  Cookie: __Host-portal-at-p2=<access-token-part-2>
             │
             ▼
Portal BFF
  1. Parse exactly one partial Bearer Authorization header.
  2. Read exactly one configured access-token part-2 cookie.
  3. Enforce per-part and reconstructed-token size limits and concatenate part 1 + part 2.
  4. Obtain verification keys from the pinned issuer discovery/JWKS cache.
  5. Validate the reconstructed JWT signature and registered claims.
  6. Map validated claims to an authenticated principal.
  7. Enforce route-level scope, organization, and ownership policy.
  8. Strip client-supplied trusted headers and inject BFF-derived context.
  9. Map the allowlisted route to OpenFGC.
             │
             ▼
OpenFGC server :8060
```

Missing either half, duplicate cookies/authorization values, malformed parts, or a failed reconstructed-token validation returns `401`. Neither partial nor reconstructed token is logged or placed in request context.

### 3.5 Token Validation Requirements

For every protected request, the BFF must validate the reconstructed access token:

- JWT signature using the configured IdP JWKS and an explicitly allowed signing algorithm.
- `iss` equals `BFF_AUTH__ISSUER_URL`.
- `aud` contains the configured portal API audience.
- `exp`, `nbf` (when present), and `iat` (when required by policy), allowing only configured clock skew.
- A non-empty `sub` claim.
- Token type/purpose only when `BFF_AUTH__REQUIRE_ACCESS_TOKEN_TYPE=true`, so an ID token cannot be accepted as an API access token when the IdP provides a reliable distinguishing claim.
- Required scopes and organization claims for the target route.

The callback separately validates the ID token against the OIDC client ID; nonce validation remains deferred. The BFF must reject complete browser-supplied tokens, query-string tokens, bearer values without the matching part-2 cookie, unsigned tokens, unsupported algorithms, unknown issuers/audiences, and expired tokens.

### 3.6 Refresh

`POST /auth/refresh` accepts refresh-token part 1 in an `application/x-www-form-urlencoded` body and obtains part 2 only from the configured HTTP-only cookie. It reconstructs the refresh token, authenticates the confidential client to the token endpoint, and never returns tokens in its response body.

On success, the BFF atomically replaces both access-token cookies, both refresh-token cookies when the IdP rotates the refresh token, and both JavaScript-readable ID-token cookies when a new ID token is returned. It clears stale token cookies if refresh fails because the refresh token is invalid or expired. The frontend uses one module-level refresh promise to deduplicate concurrent `401` recovery within one browser runtime. Matching the reference server behavior, this stage adds no BFF-local or distributed refresh lock; cross-tab and cross-replica rotation races remain a documented limitation.

`/auth/refresh` must reject cookie-only requests. Possession of refresh-token part 1 in the explicit request body, combined with the matching HTTP-only part-2 cookie, is the CSRF defense used by the reference split-cookie design. A cross-site origin cannot read part 1 through the same-origin policy and cannot reconstruct the refresh token from automatically attached cookies alone.

### 3.7 Profile, Logout, and Expiry

- The frontend may reconstruct and decode the JavaScript-readable ID-token parts for profile display only. Browser-decoded claims are never sent back as trusted identity or used for authorization.
- `POST /auth/logout` requires the same partial Bearer header and matching HTTP-only access-token part-2 cookie as other protected requests. It reconstructs/revokes tokens when supported, clears every token cookie using the exact issuance attributes, and initiates the IdP end-session flow when configured.
- On a protected-route `401`, the frontend may attempt one refresh and retry the original request once. If refresh fails, it starts login; it must not create retry loops.
- On `403`, the frontend shows an authorization error and must not refresh or repeatedly start login.
- Expired or invalid split cookies are cleared at the authentication boundary where safe to do so.

---

## 4. API Routes and Authorization

Every path and method currently defined in `portal/backend/openapi/portal-backend.yaml` is listed below. Routes require a successfully reconstructed and validated access token unless explicitly marked public. Authorization is evaluated from the validated principal, never from browser-supplied user or organization values. The BFF authorizes capabilities through scopes; IdP roles are used only to grant scope bundles and are not evaluated by the BFF.

### 4.1 Public System Endpoints

| Endpoint | Method | Description | Authentication |
|---|---|---|---|
| `/health` | GET | Service health. | Public |
| `/health/liveness` | GET | Process liveness check. | Public |
| `/health/readiness` | GET | Dependency/readiness check. | Public |

Health endpoints expose only minimal operational status. They must not return configuration, identity, token, dependency credentials, stack traces, or other sensitive diagnostics.

### 4.2 User Endpoints

| Endpoint | Method | Description | Required scope | Additional authorization |
|---|---|---|---|---|
| `/me/consents` | GET | List the authenticated user's consents. | `portal:consents:read:self` | Self-scoped to `UserIdentity.UserID` |
| `/me/consents/{consentId}` | GET | Get enriched details for an owned consent. | `portal:consents:read:self` | Ownership validation |
| `/me/consents/{consentId}/approve` | POST | Approve selected optional elements for an owned consent. | `portal:consents:write:self` | Ownership validation |
| `/me/consents/{consentId}/revoke` | POST | Revoke an owned consent. | `portal:consents:write:self` | Ownership validation |

Rules:

- `GET /me/consents` maps to `GET /api/v1/consents?userIds=<UserIdentity.UserID>` on OpenFGC.
- Ignore or overwrite every browser-supplied `userIds` value on `/me/*` routes.
- Before reading, approving, or revoking a specific consent, verify ownership against the authenticated `sub`.
- Ignore any request body `userId` that differs from `UserIdentity.UserID`.
- Purpose data fetched internally to enrich `/me/consents` responses is covered by the parent consent scope; it does not grant independent purpose browsing or management.

### 4.3 Controlled `/api/*` Passthrough Endpoints

`/api/{path...}` is the existing controlled passthrough route. It is split-token authenticated and restricted by the explicit method/path allowlist; it is not an unrestricted proxy.

| BFF route | Methods | Required scope |
|---|---|---|
| `/api/consents` | GET | `portal:consents:read:any` |
| `/api/consents` | POST | `portal:consents:write:any` |
| `/api/consents/attributes` | GET | `portal:consents:read:any` |
| `/api/consents/validate` | POST | `portal:consents:read:any` |
| `/api/consents/{consentId}` | GET | `portal:consents:read:any` |
| `/api/consents/{consentId}` | PUT | `portal:consents:write:any` |
| `/api/consents/{consentId}/history` | GET | `portal:consents:read:any` |
| `/api/consents/{consentId}/revoke` | POST | `portal:consents:write:any` |
| `/api/consents/{consentId}/authorizations` | GET, POST | `portal:consents:read:any` for GET; `portal:consents:write:any` for POST |
| `/api/consents/{consentId}/authorizations/{authorizationId}` | GET, PUT | `portal:consents:read:any` for GET; `portal:consents:write:any` for PUT |

| `/api/consent-elements` | GET, POST | `portal:elements:read` for GET; `portal:elements:write` for POST |
| `/api/consent-elements/{elementId}` | GET | `portal:elements:read` |
| `/api/consent-elements/{elementId}/versions` | GET, POST | `portal:elements:read` for GET; `portal:elements:write` for POST |
| `/api/consent-elements/{elementId}/versions/{version}` | GET, DELETE | `portal:elements:read` for GET; `portal:elements:write` for DELETE |
| `/api/consent-purposes` | GET, POST | `portal:purposes:read` for GET; `portal:purposes:write` for POST |
| `/api/consent-purposes/{purposeId}` | GET | `portal:purposes:read` |
| `/api/consent-purposes/{purposeId}/versions` | GET, POST | `portal:purposes:read` for GET; `portal:purposes:write` for POST |
| `/api/consent-purposes/{purposeId}/versions/{version}` | GET, DELETE | `portal:purposes:read` for GET; `portal:purposes:write` for DELETE |

Rules:

- `:any` permits cross-user consent access only within `UserIdentity.OrgID`; the BFF preserves authorized `userIds` filters and always injects `org-id`.
- `:self` is intentionally available only through `/me/*`; the BFF overwrites the target user with `UserIdentity.UserID`.
- IdP role bundles are: end user receives `:self` scopes; DPO receives consent `:any` scopes; admin receives consent `:any` scopes plus element and purpose scopes.
- Consent history is a consent read operation. Listing or retrieving element/purpose versions requires the corresponding read scope; creating or deleting a version requires the corresponding write scope.
- Deny unknown `/api/*` paths and methods before proxying.

### 4.4 Scope Enforcement Rules

- Define every supported portal scope once as a typed constant in `internal/system/auth/scopes.go`. Production code must not repeat scope-name string literals in routers, middleware, handlers, services, or proxy policy definitions.
- `scopes.go` also exposes the immutable known-scope set used to validate configured requested scopes at startup. Unknown configured portal scopes cause startup failure; standard OIDC scopes such as `openid` and `profile` are validated separately.
- Route policy tables reference the typed scope constants. The frontend does not duplicate or enforce the scope registry because browser checks are not an authorization boundary.
- OpenAPI cannot import Go constants directly. Generate its scope descriptions from the canonical registry or add a contract test that fails when OpenAPI scope names differ from `scopes.go`.
- The BFF reads the space-delimited `scope` claim only after JWT signature and registered-claim validation.
- Missing, invalid, expired, or wrong-audience tokens return `401`; a valid token missing a required route scope returns `403`.
- Scope values use `:` within a scope name and spaces only to separate distinct scopes, for example: `portal:consents:read:self portal:consents:write:self`.
- The BFF never accepts scopes, user IDs, or organization context from request headers, paths, query parameters, or request bodies.

---

## 5. Middleware and Router Structure

### 5.1 Middleware Order

```
SecurityHeaders
  → CORS
    → RequestID / Logger
      → SplitTokenAuth
        → UserIdentity
          → AuthorizationPolicy
            → Route Handler / OpenFGC Proxy
```

`SplitTokenAuth` must run before `UserIdentity` and all protected route handlers. It reconstructs and validates the access token before storing the normalized principal. `UserIdentity` is the only resolver of effective user and organization identifiers; it stores both values in request context. `AuthorizationPolicy` uses the validated principal and route metadata to enforce scopes, organization rules, and ownership requirements.

### 5.2 Split-Token Authentication Middleware

1. Require exactly one `Authorization` header matching `Bearer <access-token-part-1>` and reject comma-joined or duplicate authorization values.
2. Require exactly one configured HTTP-only access-token part-2 cookie and reject ambiguous duplicates.
3. Enforce conservative size and character limits on each half, concatenate part 1 followed by part 2, and retain the result only for validation.
4. Fetch provider metadata from the configured issuer discovery document at startup and use its `jwks_uri`.
5. Cache JWKS keys with bounded TTL and refresh the cache when an otherwise valid token references an unknown `kid`.
6. Validate signature, algorithm, issuer, audience, time claims, subject, and configured token-purpose claims.
7. Map claims into `Principal` only after successful validation.
8. Store `Principal` in request context; never store either token half or the reconstructed access token in context.
9. Return `401` with a generic error response and `WWW-Authenticate: Bearer` for failed authentication.

### 5.3 User Identity Middleware

`internal/system/middleware/identity.go` is the single, central resolver of effective request identity for `/me/*` handlers, services, and proxy transforms.

- It writes one `UserIdentity` value into `internal/system/context`:

  ```
  UserIdentity {
    UserID string // OpenFGC userIds; derived from token sub
    OrgID  string // OpenFGC org-id; derived from token org_id claim
  }
  ```

- In normal operation, it requires the `Principal` placed in request context by `SplitTokenAuth`, derives `UserIdentity` from the validated `Principal.Subject` and `Principal.OrgID`, and writes it through `system/context.WithUserIdentity`.
- In explicit placeholder mode, it derives the same `UserIdentity` from `PlaceholderUserID` and `PlaceholderOrgID`; handlers and services therefore use the identical contract in tests and local development.
- Placeholder mode is permitted only in local/test environments and is rejected during startup when `BFF_ENV=production`; this production safeguard already exists in `internal/system/config/config.go`.
- Placeholder mode must be an explicit configuration choice; it must never activate because authentication failed or because either token half is missing.
- Production routes always require split-token authentication; test/local route wiring may use `UserIdentity` with placeholder mode instead of `SplitTokenAuth` only when the explicit non-production configuration is enabled.

Application handlers, services, and proxy transforms read `UserIdentityFromContext` only. `UserIDFromContext` may remain as a convenience wrapper over `UserIdentity.UserID` while existing code is migrated. They must not read JWT claims or placeholder configuration directly.

### 5.4 Authorization Policy

The principal model must contain only normalized, validated information:

```
Principal {
  Subject string
  Scopes  []string
  OrgID   string
}
```

Claim names and mappings are configuration-driven because IdP claim conventions vary. Route policy must distinguish authentication failures (`401`) from authenticated-but-forbidden requests (`403`).

### 5.5 CORS

- Allow only origins in `BFF_CORS__ALLOWED_ORIGINS`; never use `*`.
- Allow `Authorization`, `Content-Type`, and `X-Correlation-ID` request headers.
- Allow only the BFF's supported HTTP methods.
- Return `Vary: Origin`.
- Set `Access-Control-Allow-Credentials: true` when the frontend and BFF are cross-origin, because protected requests require the part-2 cookie. Never combine credentials with a wildcard origin.
- Handle preflight `OPTIONS` before split-token authentication.

---

## 6. OpenFGC Proxy Boundary

### 6.1 Request Transformation

For an authorized route, the proxy must:

- Map only approved BFF routes to their corresponding OpenFGC `/api/v1/*` route.
- Preserve approved query parameters and request bodies while applying user-scoping rules.
- Strip hop-by-hop headers: `Connection`, `Keep-Alive`, `Proxy-Authenticate`, `Proxy-Authorization`, `TE`, `Trailer`, `Transfer-Encoding`, and `Upgrade`.
- Remove any browser-supplied `org-id`, `TPP-client-id`, user, or identity headers.
- Inject only the trusted context explicitly defined in §6.2 from the validated principal and route policy.
- Propagate a validated `X-Correlation-ID`, or generate one when absent.
- Enforce request-size limits, downstream timeouts, and a conservative retry policy appropriate for idempotency.

### 6.2 Trusted Context Mapping

The BFF strips any browser-supplied values for the trusted headers below, then derives them only from validated request context.

| BFF source | OpenFGC target | Rule |
|---|---|---|
| `UserIdentity.UserID` | `userIds` query parameter | Set only for `/me/*` routes; overwrite all browser-supplied `userIds`. It is also used for user-owned authorization and action payloads. |
| `UserIdentity.OrgID` | `org-id` header | Set on every OpenFGC request. Reject the request if the required organization identity is absent. |
| Group ID read from the fetched consent resource | `group-id` header | Set only when a self-service update/revoke operation first reads the owned consent and must send that consent-bound group back upstream. Never derive a generic group ID from token claims or placeholder configuration. |
| No user/IdP-derived source | `TPP-client-id` header | Strip any browser value and do not set this header in the current design. If OpenFGC requires it, define a separate service-level trusted source before enabling that route. |
| Validated incoming correlation ID, or generated ID | `X-Correlation-ID` header | Propagate only after format/length validation; otherwise generate a new ID. |

The BFF never forwards the browser `Authorization` header or a browser-supplied trusted-context header to OpenFGC.

### 6.3 Downstream Authentication

This plan uses a header-trust boundary: OpenFGC trusts the BFF's injected context, not client-provided headers. The BFF does not forward the browser access token to OpenFGC by default.

If OpenFGC must validate access tokens in the future, explicitly add one of these designs before forwarding credentials:

- token forwarding with OpenFGC validation of the same audience/token contract; or
- OAuth token exchange / service-to-service credentials for the OpenFGC audience.

Do not forward a token merely because it is available; its audience and downstream authorization meaning must be correct.

---

## 7. Security Requirements

### 7.1 Browser and Token Handling

- The frontend reads only part-1 cookies. It sends access-token part 1 only to the trusted BFF origin and sends refresh-token part 1 only to `/auth/refresh`.
- The BFF accepts access-token part 1 only in the partial Bearer header and refresh-token part 1 only in the refresh request body. Complete browser-supplied tokens are rejected.
- Access/refresh part-2 cookies are `HttpOnly`; both ID-token parts are intentionally JavaScript-readable for profile display. All production token cookies are `Secure`, host-only, and use an explicit `SameSite` policy.
- Do not put token halves, reconstructed tokens, authorization codes, or client secrets in URLs, application logs, analytics events, error messages, or telemetry attributes.
- Apply strict maximum sizes to headers, cookies, individual token parts, and reconstructed tokens.
- Use HTTPS in staging and production for both portal-to-BFF and BFF-to-IdP traffic.
- Keep access-token, refresh-token, ID-token, and cookie lifetimes aligned with IdP policy; cookies must not extend token validity.

### 7.2 JWT/JWKS Safety

- Pin the expected issuer and audience; do not infer either from untrusted token claims.
- Permit only explicitly configured asymmetric signing algorithms.
- Use the issuer's discovered JWKS URI; do not accept a JWKS URL from the token.
- Bound JWKS cache lifetime and apply HTTP timeouts. Retain and use the last known-good keys during transient discovery/JWKS retrieval failures, provided they still validate the token; do not discard a usable cache merely because refresh failed.
- Refresh JWKS on an unknown `kid` before rejecting a token, supporting normal IdP key rotation.
- Fail closed if discovery/JWKS integrity cannot be established and no valid cached key can validate the token.

### 7.3 Cross-Site Request and XSS Protection

The design does not issue a separate API CSRF cookie or token. Protected API calls and logout require a JavaScript-created partial Bearer header plus the matching HTTP-only access-token half. Refresh requires refresh-token part 1 in the explicit request body plus the matching HTTP-only part-2 cookie. Because another origin cannot read either JavaScript-readable part through the same-origin policy, automatically attached cookies alone cannot authenticate these operations. Login/callback correlation uses a separate one-time OAuth state cookie and S256 PKCE verifier cookie; OIDC nonce remains deferred.

The split-token model reduces complete-token exfiltration but does not prevent same-origin action by injected scripts. The React portal must maintain a restrictive CSP, avoid unsafe HTML injection, review third-party scripts, and apply normal output-encoding and dependency controls.

### 7.4 Security Headers and Caching

Apply to BFF responses:

```
Content-Security-Policy: default-src 'none'
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
Referrer-Policy: strict-origin-when-cross-origin
Cache-Control: no-store
Pragma: no-cache
```

The BFF returns JSON and redirects only; the React portal has its own CSP.

### 7.5 Frontend XSS Hardening

The static-compatible hardening baseline is implemented as part of the authentication boundary. The production build emits an enforcing CSP meta policy plus a `_headers` deployment artifact; the static host must apply the emitted HTTP headers so header-only directives such as `frame-ancestors` are enforced.

The split-token design limits complete-token exfiltration but JavaScript executing in the portal origin can still read every part-1 cookie and issue authenticated requests while the browser supplies part 2. Frontend XSS prevention is therefore part of the authentication boundary, not only a UI concern.

Serve the React application with a restrictive, deployment-specific CSP. All application JavaScript is emitted as external same-origin build assets, and inline scripts, inline event handlers, string-to-code execution, and `'unsafe-eval'` remain prohibited. Oxygen UI currently uses Emotion-generated runtime style elements, so styles retain a narrowly documented `style-src 'unsafe-inline'` exception until per-response style nonces or a build-time styling migration is implemented. A production baseline is:

```text
Content-Security-Policy:
  default-src 'self';
  script-src 'self';
  style-src 'self' 'unsafe-inline'; # temporary Oxygen UI/Emotion exception
  connect-src 'self' <exact-bff-origin-if-not-self>;
  img-src 'self' data:;
  font-src 'self';
  object-src 'none';
  base-uri 'none';
  frame-ancestors 'none';
  form-action 'self' <exact-idp-origin-if-required>;
  manifest-src 'self';
  upgrade-insecure-requests
```

The production HTML must not contain inline scripts, inline event handlers, inline style attributes, or inline style elements. Runtime Emotion style elements are the only current exception and are covered by the temporary style policy above; no script nonce/hash handling is introduced. Keep development-only script allowances out of production configuration. Introduce future CSP changes in report-only mode when necessary, review violations, and then enforce; reporting payloads must not include token or user data.

Additionally:

- Do not use `dangerouslySetInnerHTML`, direct `innerHTML`, `document.write`, string-to-code APIs, or dynamic script construction. When rendering trusted rich-text is an explicit requirement, sanitize it with an allowlist-based, centrally configured sanitizer and test bypass payloads.
- Prefer text rendering and framework escaping for all IdP claims, API values, URLs, and error messages. Treat ID-token profile claims as untrusted display input even though the token signature was validated.
- Do not copy token parts into `localStorage`, `sessionStorage`, application state snapshots, Redux/devtools persistence, analytics, or error-reporting context.
- Self-host scripts where practical. Pin third-party dependencies and use Subresource Integrity plus `crossorigin` for any explicitly approved external static script or stylesheet.
- Restrict URL/navigation sinks to validated relative paths or explicit origin allowlists. Never insert token claims or API data into executable URLs.
- Consider enforcing Trusted Types where supported after removing incompatible sinks.
- Run dependency vulnerability checks and targeted DOM-XSS tests in CI, and verify production CSP/security headers through integration tests.

---

## 8. Project Structure

```
portal/backend/
├── cmd/server/
│   ├── main.go
│   └── servicemanager.go
├── internal/
│   ├── me/
│   │   ├── init.go             # module initialization and /me route registration
│   │   ├── handler.go          # /me request handlers
│   │   └── service.go          # self-service consent operations
│   ├── proxy/
│   │   ├── init.go             # module initialization and /api route registration
│   │   ├── handler.go          # proxy HTTP handler
│   │   ├── service.go          # OpenFGC request transformation/proxying
│   │   └── service_test.go
│   └── system/
│       ├── auth/
│       │   ├── oidc_client.go  # authorization URL and confidential-client token exchange
│       │   ├── cookies.go      # split, issue, reconstruct, rotate, and clear token cookies
│       │   ├── handlers.go     # login, callback, refresh, and logout endpoints
│       │   ├── validator.go    # issuer discovery, JWKS cache, ID/access JWT validation
│       │   ├── principal.go    # validated claim-to-Principal mapping
│       │   ├── scopes.go       # canonical typed portal scope names and known-scope set
│       │   └── policy.go       # scope, organization, and ownership helpers
│       ├── config/
│       │   ├── config.go       # configuration loading, defaults, validation
│       │   └── config_test.go
│       ├── context/
│       │   └── identity.go     # request-scoped Principal and UserIdentity accessors
│       ├── correlation/
│       │   └── id.go           # correlation-ID helpers
│       ├── healthcheck/
│       │   └── handler.go
│       ├── middleware/
│       │   ├── split_token_auth.go # reconstructs access token, validates it, stores Principal
│       │   ├── correlationid.go
│       │   ├── correlationid_test.go
│       │   ├── cors.go
│       │   ├── cors_test.go
│       │   ├── identity.go      # central UserIdentity resolver: token claims or test placeholders
│       │   └── identity_test.go
│       └── log/
│           ├── log.go
│           └── log_test.go
├── tests/
│   └── integration/
│       ├── health_test.go
│       ├── proxy_invalid_consent_id_test.go
│       ├── proxy_phase2_test.go
│       └── server_test.go
├── openapi/
│   └── bff.yaml
├── README.md
├── Taskfile.yml
├── go.mod
├── Dockerfile
├── docker-compose.yml
└── go.sum
```

Authentication is added within the existing `internal/system` foundation. `internal/system/auth` owns confidential-client token exchange, split-token cookies, refresh/logout flows, provider discovery, JWKS/JWT validation, normalized principal creation, the canonical scope registry, and authorization middleware. It reconstructs the access token, invokes validation, discards the raw token, and stores only the validated principal (`UserID`, `OrgID`, and scopes) in `internal/system/context`. The same middleware creates that principal from explicit non-production placeholders when placeholder mode is enabled.

Add unit tests beside the package under test and extend `tests/integration` for end-to-end OIDC, split-token authentication, and proxy scenarios.

---

## 9. Configuration

Use Koanf with environment variables as the primary runtime source and optional file overlays for local development.

```bash
# Server
BFF_SERVER__PORT=8080
BFF_ENV=development
BFF_LOG__LEVEL=info

# IdP OIDC confidential client and resource-server validation
BFF_AUTH__ISSUER_URL=https://localhost:9443
BFF_AUTH__CLIENT_ID=portal-bff
BFF_AUTH__CLIENT_SECRET=replace-with-environment-secret
BFF_AUTH__PORTAL_URL=http://localhost:3000/consents
BFF_AUTH__REDIRECT_URI=http://localhost:8080/auth/callback
BFF_AUTH__POST_LOGOUT_REDIRECT_URI=http://localhost:3000
BFF_AUTH__SCOPES=openid profile portal:consents:read:self portal:consents:write:self
BFF_AUTH__RESOURCE_AUDIENCE=portal-api
BFF_AUTH__ALLOWED_SIGNING_ALGORITHMS=RS256
BFF_AUTH__HTTP_TIMEOUT=5s
BFF_AUTH__CLOCK_SKEW=30s
# Disabled unless the IdP emits a reliable access-token-type claim.
BFF_AUTH__REQUIRE_ACCESS_TOKEN_TYPE=false
BFF_AUTH__ACCESS_TOKEN_TYPE_CLAIM=token_type
BFF_AUTH__ACCESS_TOKEN_TYPE_VALUE=access_token

# Split-token cookies
BFF_AUTH__ACCESS_TOKEN_PART1_COOKIE=portal-at-p1
BFF_AUTH__ACCESS_TOKEN_PART2_COOKIE=portal-at-p2
BFF_AUTH__REFRESH_TOKEN_PART1_COOKIE=portal-rt-p1
BFF_AUTH__REFRESH_TOKEN_PART2_COOKIE=portal-rt-p2
BFF_AUTH__ID_TOKEN_PART1_COOKIE=portal-id-p1
BFF_AUTH__ID_TOKEN_PART2_COOKIE=portal-id-p2
BFF_AUTH__OAUTH_STATE_COOKIE=portal-oauth-state
BFF_AUTH__PKCE_VERIFIER_COOKIE=portal-pkce-verifier
BFF_AUTH__COOKIE_SECURE=false
BFF_AUTH__COOKIE_SAME_SITE=Lax
BFF_AUTH__LOGIN_TRANSACTION_MAX_AGE_SECONDS=600
BFF_AUTH__MAX_TOKEN_PART_BYTES=3800
BFF_AUTH__MAX_RECONSTRUCTED_TOKEN_BYTES=7600
BFF_AUTH__REFRESH_TIMEOUT=10s

# Claim mapping
BFF_AUTH__SCOPE_CLAIM=scope
BFF_AUTH__ORG_ID_CLAIM=org_id

# OpenFGC server
BFF_PROXY__OPENFGC_API_URL=http://localhost:8060
BFF_PROXY__OPENFGC_API_TIMEOUT=10s

# CORS
BFF_CORS__ALLOWED_ORIGINS=http://localhost:3000
BFF_CORS__ALLOWED_HEADERS=Authorization,Content-Type,X-Correlation-ID
BFF_CORS__ALLOW_CREDENTIALS=true

# Proxy safety
BFF_PROXY__MAX_REQUEST_BYTES=1048576

# Test/local identity only — startup must reject this in production
BFF_PROXY__PLACEHOLDER_MODE_ENABLED=false
BFF_PROXY__PLACEHOLDER_USER_ID=
BFF_PROXY__PLACEHOLDER_ORG_ID=
```

When placeholder mode is enabled in an allowed non-production environment, both `BFF_PROXY__PLACEHOLDER_USER_ID` and `BFF_PROXY__PLACEHOLDER_ORG_ID` are required and must be non-empty. Placeholder mode remains invalid in production.

The OIDC client secret must come from a secret manager or mounted secret file, not committed configuration or ordinary environment dumps. Validate redirect URIs, cookie settings, allowed origins, issuer, audience, and placeholder-mode restrictions at startup. The BFF owns no transaction-encryption or JWT-signing keys; the IdP remains the token issuer.

The React application has no IdP client configuration or IdP SDK. Its portal auth client contains only BFF-relative auth routes, readable part-1 cookie names, split-token request transport, and the one-refresh/one-retry behavior.

---

## 10. Deployment Notes

### 10.1 Network Boundary

- Expose the React portal and BFF through the intended ingress only.
- Do not expose OpenFGC port `8060` to browsers or public networks.
- Permit inbound OpenFGC traffic only from the BFF/private network identity.
- Strip trusted context headers at ingress so only the BFF can set them.
- Prefer mTLS or equivalent service-to-service authentication between BFF and OpenFGC.

### 10.2 Availability and Rate Limiting

- Rate-limit protected API routes and malformed/failed-authentication requests at the gateway or reverse proxy.
- Apply stricter limits to expensive mutation endpoints.
- Configure timeouts for IdP discovery/JWKS and OpenFGC calls.
- Monitor JWKS refresh failures, authentication failure reason codes, authorization denials, and OpenFGC latency without recording token material.

### 10.3 Key Rotation

The IdP owns signing-key rotation. The BFF supports it by respecting JWKS cache lifetime and refreshing JWKS when it encounters an unknown `kid`. The BFF requires an operational rotation procedure for its OIDC client secret but owns no transaction-encryption or JWT-signing keys.

### 10.4 Local Development

- Portal: `http://localhost:3000`
- BFF: `http://localhost:8080`
- IdP: `https://localhost:9443`
- OpenFGC: `http://localhost:8060`

Trust the local IdP certificate through the container/host trust store. Do not use `InsecureSkipVerify` outside an explicitly isolated local-development setup.

Local HTTP development may use non-`__Host-` cookie names with `Secure=false` through explicit development-only configuration. Startup must reject insecure authentication cookies in staging/production.

---

## 11. Implementation Checklist

### React Portal

- [x] Implement the portal-owned auth client with login, logout, refresh, local ID-token profile decoding, and signed-in status helpers; do not add an IdP SDK.
- [x] Add an API interceptor that reads access-token part 1 and sets `Authorization: Bearer <part-1>` with credentials enabled.
- [x] Send refresh-token part 1 only in the `/auth/refresh` request body; never make cookie-only refresh requests.
- [x] Handle `401` with at most one refresh and one request retry before starting login.
- [x] Handle `403` as an authorization failure.
- [x] Reconstruct and decode the ID-token cookies only for profile display; never send browser-decoded claims back as trusted identity or use them for authorization.
- [x] Apply the static-compatible restrictive frontend CSP, prohibit unsafe script execution, and verify that frontend storage and production source sinks do not capture token parts. The temporary Emotion style exception remains tracked below.

### BFF Authentication and Authorization

- [x] Implement OIDC confidential-client discovery, authorization URL construction, code exchange, and refresh-token exchange.
- [x] Generate one-time OAuth state and S256 PKCE transactions, store state/verifier only in short-lived HTTP-only cookies, validate state in constant time, consume transactions before exchange, and reject callback replay without server-side session or key storage.
- [x] Implement split-token cookie issuance, reconstruction, rotation, exact-attribute clearing, and size limits.
- [x] Implement strict partial Bearer-header and part-2-cookie parsing.
- [x] Implement issuer discovery and library-managed remote JWKS caching through `go-oidc`, with bounded HTTP timeouts, cached-key reuse, and unknown-`kid` refresh for normal key rotation.
- [x] Validate ID tokens at callback, including signature, allowed algorithm, issuer, client audience, and time claims; nonce validation is deferred.
- [x] Validate reconstructed access tokens, including signature, allowed algorithm, issuer, resource audience, time claims, and subject; validate access-token type only when explicitly enabled and supported by the IdP.
- [x] Refresh JWKS on unknown `kid` and support normal key rotation.
- [x] Map validated claims to the normalized `Principal` model.
- [ ] Define all portal scope names as typed constants and an immutable known-scope set in `internal/system/auth/scopes.go`; remove duplicated production scope literals.
- [x] Validate configured requested portal scopes against the canonical scope registry during startup.
- [x] Implement `/auth/login`, `/auth/callback`, `/auth/refresh`, and `/auth/logout` with generic errors and token redaction.
- [ ] Serialize or otherwise make concurrent refresh requests safe for refresh-token rotation.
- [x] Reject cookie-only refresh/logout requests: refresh requires refresh-token part 1 in the body, while logout requires the partial Bearer header and matching access-token part-2 cookie.
- [x] Store the validated `Principal` containing user ID, organization ID, and scopes in request context, including equivalent placeholder-mode identity.
- [x] Retain explicit placeholder identity mode only for test/local use; reject it in production and never use it as an authentication fallback.
- [x] Configure and enforce the §4 scope, organization, and ownership policy for both `/me/*` and `/api/*` routes.
- [x] Return `401` with `WWW-Authenticate: Bearer` for authentication failures and `403` without a bearer challenge for authorization failures; cover both behaviors in unit and integration tests.

### BFF Routes and Proxy

- [x] Enforce self-scoping and object ownership on `/me/*` routes; reuse only a fetched, owned consent's group ID for the corresponding upstream mutation.
- [x] Enforce the §4 scope-to-route policy for the existing controlled `/api/*` passthrough route.
- [x] Update `openapi/portal-backend.yaml`: document auth endpoints and define the documentation-only `PortalOAuthDocumentation` scheme plus `AccessTokenPart2` as an AND security requirement. Document required route scopes explicitly, explain the non-standard browser-facing split-token contract, and contract-test documented scopes against `scopes.go` route policies.
- [x] Maintain explicit method/path mappings to OpenFGC `/api/v1/*`.
- [x] Point `BFF_PROXY__OPENFGC_API_URL` to port `8060`.
- [x] Strip client-supplied trusted headers; inject user and organization context only from `UserIdentity`.
- [x] Remove generic placeholder/token-derived `group-id` injection; preserve only consent-bound group reuse for self-service mutations.
- [x] Apply body limits, timeouts, correlation IDs, and hop-by-hop header stripping.

### Security and Operations

- [x] Configure exact CORS origins, permit the partial `Authorization` header, and enable credentials only for those origins.
- [x] Enforce secure host-only production cookies with explicit `SameSite`, correct expiry, HTTP-only access/refresh part-2 cookies, and intentionally JavaScript-readable ID-token parts.
- [ ] Store the OIDC client secret in the deployment secret manager and document rotation; do not introduce BFF-owned transaction or JWT-signing keys.
- [ ] Apply security and no-store cache headers.
- [x] Ensure token halves, reconstructed tokens, codes, and client secrets are redacted from application, proxy, and observability logs.
- [ ] Make OpenFGC port `8060` private-only and enforce BFF-only access.
- [ ] Add infrastructure rate limiting and monitoring for failed authentication/JWKS availability.

### Frontend XSS Hardening

- [x] Define a production React CSP with an exact BFF `connect-src`, prohibit inline/eval scripts, embed the enforceable meta subset, and emit the complete policy for the static host as `dist/_headers`.
- [x] Emit JavaScript and CSS build assets from the same origin and verify that production HTML contains no inline scripts, event handlers, style attributes, or style elements.
- [x] Set `object-src 'none'`, `base-uri 'none'`, `frame-ancestors 'none'`, constrained `form-action`, and `upgrade-insecure-requests` for HTTPS production builds.
- [x] Reject unsafe DOM/code sinks and web-storage use during production build verification; no trusted rich-text rendering currently requires a sanitizer.
- [x] Treat decoded ID-token claims and API content as untrusted display input, rely on React escaping, and test markup-shaped profile claims.
- [x] Verify that token parts never enter web storage or persisted frontend state. No analytics or frontend error-reporting integration is present.
- [x] Self-host frontend scripts and styles and pin direct dependencies. No external static script or stylesheet currently requires Subresource Integrity.
- [x] Add CSP policy tests, DOM-XSS rendering tests, production HTML/source-sink verification, and high-severity production dependency auditing to CI.
- [ ] Replace the temporary `style-src 'unsafe-inline'` exception required by Oxygen UI/Emotion with per-response style nonces or a reviewed build-time styling migration. Test dynamic MUI/Oxygen components before enforcing the replacement.
- [ ] Add sanitized CSP violation reporting when an approved reporting endpoint and data-retention policy are available; reports must exclude token and user data.

### Testing

- [x] Unit tests: token splitting/reconstruction, missing or duplicate halves, wrong half order, malformed partial Bearer header, cookie attributes, exact clearing, and token/cookie size limits.
- [x] Unit tests: callback errors and token-endpoint error mapping.
- [x] Unit tests: OAuth state generation and constant-time matching, RFC 7636 S256 challenge generation, PKCE verifier validation, transaction-cookie attributes/consumption, missing/mismatched/duplicate state, and replay rejection.
- [x] Unit tests: valid reconstructed token, invalid signature, unsupported algorithm, expired/not-yet-valid token, wrong issuer, wrong audience, missing subject, and wrong token type.
- [x] Unit tests: ID-token signature, issuer, audience, and time validation; nonce tests are deferred.
- [x] Unit tests: JWKS cache hits, unknown-`kid` refresh, rotation, and transient JWKS failure behavior.
- [x] Unit tests: canonical scope uniqueness, configured-scope validation, route-policy references, and OpenAPI scope parity.
- [x] Unit tests: claim mapping, scope policy, organization policy, and ownership policy.
- [x] Integration tests: login/callback cookie issuance and portal-style partial Bearer plus HTTP-only-cookie authentication across every protected route policy.
- [x] Integration tests: missing part 1, missing part 2, complete-token injection, duplicate cookies, and tampered reconstruction all return `401`.
- [x] Integration tests: successful refresh, rotated and retained refresh tokens, expired/invalid refresh, concurrent non-rotating refresh, and atomic cookie replacement. Refresh-token rotation race serialization remains explicitly deferred.
- [x] Integration tests: cookie-only refresh/logout rejection, logout cookie clearing, and IdP end-session and fallback behavior.
- [x] Integration tests: `401` versus `403` behavior and `WWW-Authenticate` response.
- [x] Integration tests: credentialed CORS preflight and exact `Authorization` header allowance.
- [x] Integration tests: OpenFGC path mapping, query preservation, port-8060 target, and header override prevention.
- [x] Integration tests: `/me/consents` always injects `Principal.UserID` and ignores client user filters; `/api/*` rejects missing `:any` scopes and preserves authorized filters only within `Principal.OrgID`.

### Deferred Final Hardening

- [ ] Add OIDC `nonce` generation, one-time cookie storage, ID-token validation, replay tests, and cleanup using the existing stateless login-transaction mechanism.

---

## 12. Standards

- HTTP stack: Go `net/http` with `ServeMux` and composable middleware.
- Configuration: centralized Koanf loading and startup validation.
- API design: contract-first route and payload definitions; deny-by-default proxy routing.
- Security model: confidential-client OIDC, split-token cross-site request protection, standards-based JWKS/JWT validation, least-privilege scopes, private downstream boundary, and externalized secrets/configuration.
- Testing model: unit tests for split cookies, reconstruction, state/PKCE correlation, JWT validation, and policy logic plus integration tests for the browser-to-BFF auth lifecycle and BFF-to-OpenFGC transformations. OIDC nonce and the remaining CSP items remain deferred hardening work.
