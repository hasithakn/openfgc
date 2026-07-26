/*
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
 * Licensed under the Apache License, Version 2.0.
 */

package profile

import (
	"net/http"

	"github.com/wso2/openfgc/portal/backend/internal/system/auth"
	"github.com/wso2/openfgc/portal/backend/internal/system/config"
)

// Initialize sets up the profile module and registers self-service PII routes. No portal
// scope is required beyond authentication — every user manages their own profile.
func Initialize(mux *http.ServeMux, cfg config.Config, authManager *auth.Manager) error {
	handler, err := NewHandler(cfg)
	if err != nil {
		return err
	}

	mux.Handle("GET /profile", authManager.Require(http.HandlerFunc(handler.GetProfile)))
	mux.Handle("PATCH /profile", authManager.Require(http.HandlerFunc(handler.UpdateProfile)))

	return nil
}
