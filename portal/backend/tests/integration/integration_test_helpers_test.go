package integration

import (
	"context"
	"net/http"

	"github.com/wso2/openfgc/portal/backend/internal/me"
	"github.com/wso2/openfgc/portal/backend/internal/proxy"
	"github.com/wso2/openfgc/portal/backend/internal/system/auth"
	"github.com/wso2/openfgc/portal/backend/internal/system/config"
	"github.com/wso2/openfgc/portal/backend/internal/system/healthcheck"
	systemlog "github.com/wso2/openfgc/portal/backend/internal/system/log"
	"github.com/wso2/openfgc/portal/backend/internal/system/middleware"
)

func newIntegrationHandler(cfg config.Config) (http.Handler, error) {
	mux := http.NewServeMux()

	healthHandler := healthcheck.NewHandler()
	mux.HandleFunc("GET /health/liveness", healthHandler.Liveness)
	mux.HandleFunc("GET /health/readiness", healthHandler.Readiness)
	mux.HandleFunc("GET /health", healthHandler.Liveness)

	log := systemlog.New(cfg.Log.Level)
	authManager, err := auth.NewManager(context.Background(), cfg.Auth, cfg.Proxy, false, log)
	if err != nil {
		return nil, err
	}
	authManager.RegisterRoutes(mux)
	if err := proxy.Initialize(mux, cfg.Proxy, authManager); err != nil {
		return nil, err
	}
	if err := me.Initialize(mux, cfg, authManager); err != nil {
		return nil, err
	}

	withCORS := middleware.CORS(mux, middleware.CORSOptions{
		AllowedOrigins:   cfg.CORS.AllowedOrigins,
		AllowedMethods:   cfg.CORS.AllowedMethods,
		AllowedHeaders:   cfg.CORS.AllowedHeaders,
		AllowCredentials: cfg.CORS.AllowCredentials,
	})

	return middleware.CorrelationID(log, withCORS), nil
}
