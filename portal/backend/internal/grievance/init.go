/*
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
 * Licensed under the Apache License, Version 2.0.
 */

package grievance

import (
	"net/http"

	"github.com/wso2/openfgc/portal/backend/internal/system/auth"
	"github.com/wso2/openfgc/portal/backend/internal/system/config"
)

// Initialize sets up the grievance module and registers routes.
func Initialize(mux *http.ServeMux, cfg config.Config, authManager *auth.Manager) error {
	handler, err := NewHandler(cfg.Proxy)
	if err != nil {
		return err
	}

	// Data Principal self-serve routes.
	mux.Handle("GET /me/grievances", authManager.Require(http.HandlerFunc(handler.MyGrievances), auth.ScopeGrievancesReadSelf))
	mux.Handle("POST /me/grievances", authManager.Require(http.HandlerFunc(handler.SubmitGrievance), auth.ScopeGrievancesWriteSelf))
	mux.Handle("GET /me/grievances/{grievanceId}", authManager.Require(http.HandlerFunc(handler.MyGrievanceByID), auth.ScopeGrievancesReadSelf))
	mux.Handle("POST /me/grievances/{grievanceId}/messages", authManager.Require(http.HandlerFunc(handler.ReplyToMyGrievance), auth.ScopeGrievancesWriteSelf))
	mux.Handle("GET /me/grievances/{grievanceId}/attachments/{attachmentId}", authManager.Require(http.HandlerFunc(handler.MyGrievanceAttachment), auth.ScopeGrievancesReadSelf))

	// Grievance Officer routes.
	mux.Handle("GET /grievance-management/cases", authManager.Require(http.HandlerFunc(handler.CaseQueue), auth.ScopeGrievancesManage))
	mux.Handle("GET /grievance-management/cases/stats", authManager.Require(http.HandlerFunc(handler.CaseQueueStats), auth.ScopeGrievancesManage))
	mux.Handle("GET /grievance-management/cases/{grievanceId}", authManager.Require(http.HandlerFunc(handler.CaseByID), auth.ScopeGrievancesManage))
	mux.Handle("POST /grievance-management/cases/{grievanceId}/messages", authManager.Require(http.HandlerFunc(handler.ReplyToCase), auth.ScopeGrievancesManage))
	mux.Handle("POST /grievance-management/cases/{grievanceId}/status", authManager.Require(http.HandlerFunc(handler.TransitionCaseStatus), auth.ScopeGrievancesManage))
	mux.Handle("GET /grievance-management/cases/{grievanceId}/attachments/{attachmentId}", authManager.Require(http.HandlerFunc(handler.CaseAttachment), auth.ScopeGrievancesManage))

	return nil
}
