/*
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
 *
 * WSO2 LLC. licenses this file to you under the Apache License,
 * Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

// Package context provides request-scoped values shared across modules.
package context

import (
	"context"
	"strings"
)

type principalKey struct{}

type accessTokenKey struct{}

// Principal contains identity and authorization data derived from a validated token.
type Principal struct {
	UserID string
	OrgID  string
	Scopes map[string]struct{}
}

// WithPrincipal stores an authenticated principal in request context.
func WithPrincipal(ctx context.Context, principal Principal) context.Context {
	return context.WithValue(ctx, principalKey{}, principal)
}

// PrincipalFromContext returns the authenticated principal.
func PrincipalFromContext(ctx context.Context) (Principal, bool) {
	if ctx == nil {
		return Principal{}, false
	}
	principal, ok := ctx.Value(principalKey{}).(Principal)
	if !ok || strings.TrimSpace(principal.UserID) == "" || strings.TrimSpace(principal.OrgID) == "" {
		return Principal{}, false
	}
	return principal, true
}

// WithAccessToken stores the caller's own validated, reconstructed access token in request
// context. Kept separate from Principal (which flows into places like /auth/userinfo's JSON
// response) so the raw credential is never casually serialized alongside identity claims.
func WithAccessToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, accessTokenKey{}, token)
}

// AccessTokenFromContext returns the caller's own access token, if one was stored.
func AccessTokenFromContext(ctx context.Context) (string, bool) {
	if ctx == nil {
		return "", false
	}
	token, ok := ctx.Value(accessTokenKey{}).(string)
	if !ok || strings.TrimSpace(token) == "" {
		return "", false
	}
	return token, true
}
