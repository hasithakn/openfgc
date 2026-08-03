/*
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
 * Licensed under the Apache License, Version 2.0.
 */

// Package auth implements the BFF's OIDC client and split-token authentication boundary.
package auth

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"

	"github.com/wso2/openfgc/portal/backend/internal/system/config"
	systemcontext "github.com/wso2/openfgc/portal/backend/internal/system/context"
)

var errInvalidCredentials = errors.New("invalid credentials")

// Manager owns OIDC discovery, token validation, auth routes, and authorization middleware.
type Manager struct {
	cfg                config.AuthConfig
	proxyCfg           config.ProxyConfig
	log                *slog.Logger
	httpClient         *http.Client
	oauthConfig        oauth2.Config
	accessVerifier     *oidc.IDTokenVerifier
	idVerifier         *oidc.IDTokenVerifier
	endSessionEndpoint string
	// adminTokenSource mints app-level (not tied to any end user) access tokens via the
	// client_credentials grant, for admin-scoped calls like SCIM2 user deletion. Nil when
	// cfg.AdminScope is blank — AdminAccessToken reports that case as an error.
	adminTokenSource oauth2.TokenSource
}

// NewManager initializes OIDC discovery when auth is enabled. insecureSkipTLSVerify disables
// certificate verification for every call to the IS (discovery, token exchange, admin
// client_credentials) — only ever needed against a self-signed demo/dev IS instance.
func NewManager(
	ctx context.Context,
	cfg config.AuthConfig,
	proxyCfg config.ProxyConfig,
	insecureSkipTLSVerify bool,
	log *slog.Logger,
) (*Manager, error) {
	m := &Manager{cfg: cfg, proxyCfg: proxyCfg, log: log}
	if !cfg.Enabled {
		return m, nil
	}
	if err := validateConfiguredScopes(cfg.Scopes); err != nil {
		return nil, err
	}
	m.httpClient = &http.Client{Timeout: cfg.HTTPTimeout}
	if insecureSkipTLSVerify {
		m.httpClient.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec // opt-in demo-only flag
		}
	}
	discoveryContext := oidc.ClientContext(ctx, m.httpClient)
	provider, err := oidc.NewProvider(discoveryContext, cfg.IssuerURL)
	if err != nil {
		return nil, err
	}
	m.oauthConfig = oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  cfg.RedirectURI,
		Scopes:       append([]string(nil), cfg.Scopes...),
	}
	m.accessVerifier = provider.Verifier(&oidc.Config{
		ClientID:             cfg.ResourceAudience,
		SupportedSigningAlgs: append([]string(nil), cfg.AllowedSigningAlgorithms...),
	})
	m.idVerifier = provider.Verifier(&oidc.Config{
		ClientID:             cfg.ClientID,
		SupportedSigningAlgs: append([]string(nil), cfg.AllowedSigningAlgorithms...),
	})
	var metadata struct {
		EndSessionEndpoint string `json:"end_session_endpoint"`
	}
	if err := provider.Claims(&metadata); err == nil {
		m.endSessionEndpoint = strings.TrimSpace(metadata.EndSessionEndpoint)
	}
	if adminScope := strings.TrimSpace(cfg.AdminScope); adminScope != "" {
		adminConfig := &clientcredentials.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			TokenURL:     provider.Endpoint().TokenURL,
			Scopes:       []string{adminScope},
		}
		m.adminTokenSource = adminConfig.TokenSource(discoveryContext)
	}
	return m, nil
}

// AdminAccessToken returns an app-level access token (not tied to any end user) suitable for
// admin-scoped upstream calls, minted via the client_credentials grant. The returned
// oauth2.TokenSource caches and auto-refreshes the token, so this is cheap to call per request.
func (m *Manager) AdminAccessToken(ctx context.Context) (string, error) {
	if m.adminTokenSource == nil {
		return "", errors.New("admin client credentials are not configured (auth.admin_scope is blank)")
	}
	token, err := m.adminTokenSource.Token()
	if err != nil {
		return "", err
	}
	return token.AccessToken, nil
}

func validateConfiguredScopes(configured []string) error {
	return validateConfiguredScopesWithPrefix(configured, ScopePrefix, AllPortalScopes)
}

func validateConfiguredScopesWithPrefix(configured []string, prefix string, portalScopes []string) error {
	canonical := make(map[string]struct{}, len(portalScopes))
	for _, scope := range portalScopes {
		canonical[scope] = struct{}{}
	}
	for _, scope := range configured {
		// An empty prefix explicitly means that portal scopes have no namespace.
		// In that mode, exact route-policy constants remain authoritative, but
		// arbitrary OIDC scopes cannot be classified as portal-owned by prefix.
		if prefix != "" && strings.HasPrefix(scope, prefix) {
			if _, ok := canonical[scope]; !ok {
				return errors.New("auth scopes contain an unknown portal scope")
			}
		}
	}
	return nil
}

// RegisterRoutes registers browser-facing authentication endpoints.
func (m *Manager) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /auth/login", m.Login)
	mux.HandleFunc("GET /auth/callback", m.Callback)
	mux.HandleFunc("POST /auth/refresh", m.Refresh)
	mux.Handle("POST /auth/logout", m.Require(http.HandlerFunc(m.Logout)))
	mux.Handle("GET /auth/userinfo", m.Require(http.HandlerFunc(m.UserInfo)))
}

// Require authenticates a request and optionally enforces scopes.
func (m *Manager) Require(next http.Handler, requiredScopes ...string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		principal, token, err := m.authenticate(r)
		if err != nil {
			w.Header().Set("WWW-Authenticate", "Bearer")
			writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
			return
		}
		for _, required := range requiredScopes {
			if _, ok := principal.Scopes[required]; !ok {
				writeError(w, http.StatusForbidden, "FORBIDDEN", "insufficient permissions")
				return
			}
		}
		ctx := systemcontext.WithPrincipal(r.Context(), principal)
		if token != "" {
			ctx = systemcontext.WithAccessToken(ctx, token)
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireAPI authenticates an /api request and enforces its canonical route scope.
func (m *Manager) RequireAPI(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		required, ok := ScopeForAPIRequest(r.Method, r.URL.Path)
		if !ok {
			if isKnownAPIPath(r.URL.Path) {
				writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed")
				return
			}
			writeError(w, http.StatusNotFound, "NOT_FOUND", "route not found")
			return
		}
		m.Require(next, required).ServeHTTP(w, r)
	})
}

// authenticate validates the incoming request and returns the derived principal along with
// the caller's own reconstructed access token (empty in placeholder mode, since there is no
// real IS-issued token to forward to downstream services like SCIM2).
func (m *Manager) authenticate(r *http.Request) (systemcontext.Principal, string, error) {
	if !m.cfg.Enabled {
		if !m.proxyCfg.PlaceholderModeEnabled {
			return systemcontext.Principal{}, "", errInvalidCredentials
		}
		userID := strings.TrimSpace(m.proxyCfg.PlaceholderUserID)
		orgID := strings.TrimSpace(m.proxyCfg.PlaceholderOrgID)
		if userID == "" || orgID == "" {
			return systemcontext.Principal{}, "", errInvalidCredentials
		}
		scopes := make(map[string]struct{}, len(AllPortalScopes))
		for _, scope := range AllPortalScopes {
			scopes[scope] = struct{}{}
		}
		return systemcontext.Principal{UserID: userID, OrgID: orgID, Scopes: scopes}, "", nil
	}
	part1, err := bearerPart(r)
	if err != nil {
		return systemcontext.Principal{}, "", err
	}
	part2, err := exactlyOneCookie(r, m.cfg.AccessTokenPart2Cookie)
	if err != nil {
		return systemcontext.Principal{}, "", err
	}
	token, err := reconstructToken(part1, part2, m.cfg)
	if err != nil {
		return systemcontext.Principal{}, "", err
	}
	principal, _, err := m.validateAccessToken(r.Context(), token)
	if err != nil {
		return systemcontext.Principal{}, "", err
	}
	return principal, token, nil
}

func (m *Manager) validateAccessToken(ctx context.Context, raw string) (systemcontext.Principal, time.Time, error) {
	ctx = oidc.ClientContext(ctx, m.httpClient)
	verified, err := m.accessVerifier.Verify(ctx, raw)
	if err != nil {
		return systemcontext.Principal{}, time.Time{}, errInvalidCredentials
	}
	claims := map[string]json.RawMessage{}
	if err := verified.Claims(&claims); err != nil {
		return systemcontext.Principal{}, time.Time{}, errInvalidCredentials
	}
	subject := strings.TrimSpace(verified.Subject)
	orgID, err := stringClaim(claims, m.cfg.OrgIDClaim)
	if err != nil || subject == "" || orgID == "" {
		return systemcontext.Principal{}, time.Time{}, errInvalidCredentials
	}
	if err := validateNotBefore(claims, m.cfg.ClockSkew); err != nil {
		return systemcontext.Principal{}, time.Time{}, errInvalidCredentials
	}
	if m.cfg.RequireAccessTokenType {
		value, claimErr := stringClaim(claims, m.cfg.AccessTokenTypeClaim)
		if claimErr != nil || value != m.cfg.AccessTokenTypeValue {
			return systemcontext.Principal{}, time.Time{}, errInvalidCredentials
		}
	}
	scopes, err := scopeClaim(claims, m.cfg.ScopeClaim)
	if err != nil {
		return systemcontext.Principal{}, time.Time{}, errInvalidCredentials
	}
	return systemcontext.Principal{UserID: subject, OrgID: orgID, Scopes: scopes}, verified.Expiry, nil
}

func (m *Manager) validateIDToken(ctx context.Context, raw string) (time.Time, error) {
	ctx = oidc.ClientContext(ctx, m.httpClient)
	verified, err := m.idVerifier.Verify(ctx, raw)
	if err != nil || strings.TrimSpace(verified.Subject) == "" {
		return time.Time{}, errInvalidCredentials
	}
	claims := map[string]json.RawMessage{}
	if err := verified.Claims(&claims); err != nil || validateNotBefore(claims, m.cfg.ClockSkew) != nil {
		return time.Time{}, errInvalidCredentials
	}
	return verified.Expiry, nil
}

func stringClaim(claims map[string]json.RawMessage, name string) (string, error) {
	raw, ok := claims[name]
	if !ok {
		return "", errInvalidCredentials
	}
	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		return "", errInvalidCredentials
	}
	return strings.TrimSpace(value), nil
}

func scopeClaim(claims map[string]json.RawMessage, name string) (map[string]struct{}, error) {
	raw, ok := claims[name]
	if !ok {
		return map[string]struct{}{}, nil
	}
	var text string
	values := []string{}
	if err := json.Unmarshal(raw, &text); err == nil {
		values = strings.Fields(text)
	} else if err := json.Unmarshal(raw, &values); err != nil {
		return nil, errInvalidCredentials
	}
	result := make(map[string]struct{}, len(values))
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			result[value] = struct{}{}
		}
	}
	return result, nil
}

func validateNotBefore(claims map[string]json.RawMessage, skew time.Duration) error {
	raw, ok := claims["nbf"]
	if !ok {
		return nil
	}
	var seconds json.Number
	if err := json.Unmarshal(raw, &seconds); err != nil {
		return errInvalidCredentials
	}
	value, err := seconds.Int64()
	if err != nil || time.Unix(value, 0).After(time.Now().Add(skew)) {
		return errInvalidCredentials
	}
	return nil
}
