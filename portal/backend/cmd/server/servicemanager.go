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

package main

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/wso2/openfgc/portal/backend/internal/grievance"
	"github.com/wso2/openfgc/portal/backend/internal/me"
	"github.com/wso2/openfgc/portal/backend/internal/proxy"
	"github.com/wso2/openfgc/portal/backend/internal/system/auth"
	"github.com/wso2/openfgc/portal/backend/internal/system/config"
	"github.com/wso2/openfgc/portal/backend/internal/system/healthcheck"
)

func registerServices(mux *http.ServeMux, log *slog.Logger, cfg config.Config) error {
	healthHandler := healthcheck.NewHandler()
	mux.HandleFunc("GET /health/liveness", healthHandler.Liveness)
	mux.HandleFunc("GET /health/readiness", healthHandler.Readiness)
	mux.HandleFunc("GET /health", healthHandler.Liveness)
	log.Debug("registered health endpoints")

	authManager, err := auth.NewManager(context.Background(), cfg.Auth, cfg.Proxy, log)
	if err != nil {
		return err
	}
	authManager.RegisterRoutes(mux)
	log.Debug("registered auth module")

	if err := proxy.Initialize(mux, cfg.Proxy, authManager); err != nil {
		return err
	}
	log.Debug("registered proxy module")

	if err := me.Initialize(mux, cfg, authManager); err != nil {
		return err
	}
	log.Debug("registered me module")

	if err := grievance.Initialize(mux, cfg, authManager); err != nil {
		return err
	}
	log.Debug("registered grievance module")

	return nil
}
