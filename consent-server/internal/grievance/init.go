/*
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
 * Licensed under the Apache License, Version 2.0.
 */

package grievance

import (
	"net/http"

	"github.com/wso2/openfgc/internal/system/constants"
	"github.com/wso2/openfgc/internal/system/stores"
)

// Initialize sets up the grievance module and registers routes.
func Initialize(mux *http.ServeMux, registry *stores.StoreRegistry) GrievanceService {
	service := newGrievanceService(registry)
	handler := newGrievanceHandler(service)

	registerRoutes(mux, handler)

	return service
}

// registerRoutes registers all grievance routes.
func registerRoutes(mux *http.ServeMux, handler *grievanceHandler) {
	// POST /api/v1/grievances - Submit a new grievance
	mux.HandleFunc("POST "+constants.APIBasePath+"/grievances", handler.createGrievance)

	// GET /api/v1/grievances - List/search grievances
	mux.HandleFunc("GET "+constants.APIBasePath+"/grievances", handler.listGrievances)

	// GET /api/v1/grievances/stats - Officer queue aggregate summary
	mux.HandleFunc("GET "+constants.APIBasePath+"/grievances/stats", handler.getStats)

	// GET /api/v1/grievances/{grievanceId} - Get grievance detail
	mux.HandleFunc("GET "+constants.APIBasePath+"/grievances/{grievanceId}", handler.getGrievance)

	// POST /api/v1/grievances/{grievanceId}/messages - Post a reply or internal note
	mux.HandleFunc("POST "+constants.APIBasePath+"/grievances/{grievanceId}/messages", handler.postMessage)

	// POST /api/v1/grievances/{grievanceId}/status - Transition status
	mux.HandleFunc("POST "+constants.APIBasePath+"/grievances/{grievanceId}/status", handler.transitionStatus)

	// GET /api/v1/grievances/{grievanceId}/attachments/{attachmentId} - Download an attachment
	mux.HandleFunc("GET "+constants.APIBasePath+"/grievances/{grievanceId}/attachments/{attachmentId}", handler.getAttachment)
}
