/*
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
 * Licensed under the Apache License, Version 2.0.
 */

package profile

import (
	"log/slog"
	"net/http"

	"github.com/wso2/openfgc/portal/backend/internal/proxy"
	"github.com/wso2/openfgc/portal/backend/internal/system/auth"
	"github.com/wso2/openfgc/portal/backend/internal/system/config"
)

// Initialize sets up the profile module and registers self-service PII routes. No portal
// scope is required beyond authentication — every user manages their own profile.
func Initialize(mux *http.ServeMux, cfg config.Config, authManager *auth.Manager, log *slog.Logger) error {
	// A dedicated proxy instance (mirroring the grievance module's own pattern) rather than
	// the generic allowlisted /api/* passthrough: the anonymize call's userId must always be
	// the caller's own trusted principal, never a client-controlled value, so it can't be
	// exposed as an arbitrary passthrough route.
	anonymizeProxy, err := proxy.NewService(cfg.Proxy)
	if err != nil {
		return err
	}

	handler, err := NewHandler(cfg, log, authManager, anonymizeProxy)
	if err != nil {
		return err
	}

	mux.Handle("GET /profile", authManager.Require(http.HandlerFunc(handler.GetProfile)))
	mux.Handle("PATCH /profile", authManager.Require(http.HandlerFunc(handler.UpdateProfile)))
	mux.Handle("DELETE /profile", authManager.Require(http.HandlerFunc(handler.DeleteAccount)))

	return nil
}
