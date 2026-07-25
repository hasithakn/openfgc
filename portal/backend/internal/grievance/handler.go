/*
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
 * Licensed under the Apache License, Version 2.0.
 */

package grievance

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"

	"github.com/wso2/openfgc/portal/backend/internal/proxy"
	"github.com/wso2/openfgc/portal/backend/internal/system/config"
	systemcontext "github.com/wso2/openfgc/portal/backend/internal/system/context"
)

// Handler serves the /me/grievances and /grievance-management route groups.
type Handler struct {
	svc *Service
	cfg config.ProxyConfig
}

type errorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

var errRequestBodyTooLarge = errors.New("request body too large")

// NewHandler creates a grievance handler with initialized service.
func NewHandler(cfg config.ProxyConfig) (*Handler, error) {
	svc, err := NewService(cfg)
	if err != nil {
		return nil, err
	}
	return &Handler{svc: svc, cfg: cfg}, nil
}

// =============================================================================
// Self-serve (/me/grievances) handlers — Data Principal only.
// =============================================================================

// MyGrievances handles GET /me/grievances, forcing userIds to the caller's own subject.
func (h *Handler) MyGrievances(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.resolveUserID(w, r)
	if !ok {
		return
	}
	if err := h.svc.proxy.Forward(w, r, http.MethodGet, "/api/v1/grievances", func(q url.Values) {
		q.Set("userIds", userID)
	}, nil); err != nil {
		writeProxyError(w, err)
	}
}

// SubmitGrievance handles POST /me/grievances, forcing userId to the caller's own subject.
func (h *Handler) SubmitGrievance(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.resolveUserID(w, r)
	if !ok {
		return
	}
	body, err := h.readBoundedBody(r)
	if err != nil {
		writeBodyReadError(w, err)
		return
	}
	rewritten, err := withJSONFields(body, map[string]any{"userId": userID})
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "INVALID_PAYLOAD", "invalid request payload")
		return
	}
	if err := h.svc.proxy.Forward(w, r, http.MethodPost, "/api/v1/grievances", nil, rewritten); err != nil {
		writeProxyError(w, err)
	}
}

// MyGrievanceByID handles GET /me/grievances/{grievanceId}, verifying ownership before returning it.
func (h *Handler) MyGrievanceByID(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.resolveUserID(w, r)
	if !ok {
		return
	}
	grievanceID := r.PathValue("grievanceId")
	if grievanceID == "" {
		writeJSONError(w, http.StatusNotFound, "NOT_FOUND", "grievance id not found")
		return
	}
	baseResp, err := h.svc.proxy.ForwardRaw(r, http.MethodGet, "/api/v1/grievances/"+url.PathEscape(grievanceID), func(q url.Values) {
		q.Set("includeInternal", "false")
	}, nil)
	if err != nil {
		writeProxyError(w, err)
		return
	}
	if baseResp.StatusCode != http.StatusOK {
		_ = h.svc.proxy.WriteUpstreamResponse(w, baseResp)
		return
	}
	owned, err := isOwnedByUser(baseResp.Body, userID)
	if err != nil {
		writeJSONError(w, http.StatusBadGateway, "UPSTREAM_UNAVAILABLE", "upstream unavailable")
		return
	}
	if !owned {
		writeJSONError(w, http.StatusNotFound, "NOT_FOUND", "grievance not found")
		return
	}
	_ = h.svc.proxy.WriteUpstreamResponse(w, baseResp)
}

// ReplyToMyGrievance handles POST /me/grievances/{grievanceId}/messages.
// The actor identity and visibility are always server-derived — a Data Principal
// can never post as anyone else, nor post an internal-visibility entry.
func (h *Handler) ReplyToMyGrievance(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.resolveUserID(w, r)
	if !ok {
		return
	}
	grievanceID := r.PathValue("grievanceId")
	if grievanceID == "" {
		writeJSONError(w, http.StatusNotFound, "NOT_FOUND", "grievance id not found")
		return
	}
	if !h.verifyOwnership(w, r, grievanceID, userID) {
		return
	}
	body, err := h.readBoundedBody(r)
	if err != nil {
		writeBodyReadError(w, err)
		return
	}
	rewritten, err := withJSONFields(body, map[string]any{
		"actorUserId": userID,
		"actorRole":   "DataPrincipal",
		"visibility":  "shared",
	})
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "INVALID_PAYLOAD", "invalid request payload")
		return
	}
	if err := h.svc.proxy.Forward(w, r, http.MethodPost, "/api/v1/grievances/"+url.PathEscape(grievanceID)+"/messages", nil, rewritten); err != nil {
		writeProxyError(w, err)
	}
}

// MyGrievanceAttachment handles GET /me/grievances/{grievanceId}/attachments/{attachmentId}.
func (h *Handler) MyGrievanceAttachment(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.resolveUserID(w, r)
	if !ok {
		return
	}
	grievanceID := r.PathValue("grievanceId")
	attachmentID := r.PathValue("attachmentId")
	if grievanceID == "" || attachmentID == "" {
		writeJSONError(w, http.StatusNotFound, "NOT_FOUND", "attachment not found")
		return
	}
	if !h.verifyOwnership(w, r, grievanceID, userID) {
		return
	}
	if err := h.svc.proxy.Forward(w, r, http.MethodGet,
		"/api/v1/grievances/"+url.PathEscape(grievanceID)+"/attachments/"+url.PathEscape(attachmentID), nil, nil); err != nil {
		writeProxyError(w, err)
	}
}

// =============================================================================
// Officer (/grievance-management) handlers — Grievance Officer only.
// =============================================================================

// CaseQueue handles GET /grievance-management/cases (no self-filtering — any grievance in the org).
func (h *Handler) CaseQueue(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.proxy.Forward(w, r, http.MethodGet, "/api/v1/grievances", nil, nil); err != nil {
		writeProxyError(w, err)
	}
}

// CaseQueueStats handles GET /grievance-management/cases/stats.
func (h *Handler) CaseQueueStats(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.proxy.Forward(w, r, http.MethodGet, "/api/v1/grievances/stats", nil, nil); err != nil {
		writeProxyError(w, err)
	}
}

// CaseByID handles GET /grievance-management/cases/{grievanceId}, forcing the full (internal-inclusive) timeline.
func (h *Handler) CaseByID(w http.ResponseWriter, r *http.Request) {
	grievanceID := r.PathValue("grievanceId")
	if grievanceID == "" {
		writeJSONError(w, http.StatusNotFound, "NOT_FOUND", "grievance id not found")
		return
	}
	if err := h.svc.proxy.Forward(w, r, http.MethodGet, "/api/v1/grievances/"+url.PathEscape(grievanceID), func(q url.Values) {
		q.Set("includeInternal", "true")
	}, nil); err != nil {
		writeProxyError(w, err)
	}
}

// ReplyToCase handles POST /grievance-management/cases/{grievanceId}/messages.
// The actor identity is always server-derived; the client still chooses visibility
// (shared reply vs. internal note) — consent-server enforces that only a
// GrievanceOfficer actor may use "internal".
func (h *Handler) ReplyToCase(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.resolveUserID(w, r)
	if !ok {
		return
	}
	grievanceID := r.PathValue("grievanceId")
	if grievanceID == "" {
		writeJSONError(w, http.StatusNotFound, "NOT_FOUND", "grievance id not found")
		return
	}
	body, err := h.readBoundedBody(r)
	if err != nil {
		writeBodyReadError(w, err)
		return
	}
	rewritten, err := withJSONFields(body, map[string]any{
		"actorUserId": userID,
		"actorRole":   "GrievanceOfficer",
	})
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "INVALID_PAYLOAD", "invalid request payload")
		return
	}
	if err := h.svc.proxy.Forward(w, r, http.MethodPost, "/api/v1/grievances/"+url.PathEscape(grievanceID)+"/messages", nil, rewritten); err != nil {
		writeProxyError(w, err)
	}
}

// TransitionCaseStatus handles POST /grievance-management/cases/{grievanceId}/status.
func (h *Handler) TransitionCaseStatus(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.resolveUserID(w, r)
	if !ok {
		return
	}
	grievanceID := r.PathValue("grievanceId")
	if grievanceID == "" {
		writeJSONError(w, http.StatusNotFound, "NOT_FOUND", "grievance id not found")
		return
	}
	body, err := h.readBoundedBody(r)
	if err != nil {
		writeBodyReadError(w, err)
		return
	}
	rewritten, err := withJSONFields(body, map[string]any{
		"actorUserId": userID,
		"actorRole":   "GrievanceOfficer",
	})
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "INVALID_PAYLOAD", "invalid request payload")
		return
	}
	if err := h.svc.proxy.Forward(w, r, http.MethodPost, "/api/v1/grievances/"+url.PathEscape(grievanceID)+"/status", nil, rewritten); err != nil {
		writeProxyError(w, err)
	}
}

// CaseAttachment handles GET /grievance-management/cases/{grievanceId}/attachments/{attachmentId}.
func (h *Handler) CaseAttachment(w http.ResponseWriter, r *http.Request) {
	grievanceID := r.PathValue("grievanceId")
	attachmentID := r.PathValue("attachmentId")
	if grievanceID == "" || attachmentID == "" {
		writeJSONError(w, http.StatusNotFound, "NOT_FOUND", "attachment not found")
		return
	}
	if err := h.svc.proxy.Forward(w, r, http.MethodGet,
		"/api/v1/grievances/"+url.PathEscape(grievanceID)+"/attachments/"+url.PathEscape(attachmentID), nil, nil); err != nil {
		writeProxyError(w, err)
	}
}

// =============================================================================
// Shared helpers
// =============================================================================

// verifyOwnership fetches the grievance and writes a 404 (existence-hiding) response if it
// doesn't belong to userID. Returns true only when the caller should proceed.
func (h *Handler) verifyOwnership(w http.ResponseWriter, r *http.Request, grievanceID, userID string) bool {
	baseResp, err := h.svc.proxy.ForwardRaw(r, http.MethodGet, "/api/v1/grievances/"+url.PathEscape(grievanceID), nil, nil)
	if err != nil {
		writeProxyError(w, err)
		return false
	}
	if baseResp.StatusCode != http.StatusOK {
		_ = h.svc.proxy.WriteUpstreamResponse(w, baseResp)
		return false
	}
	owned, err := isOwnedByUser(baseResp.Body, userID)
	if err != nil {
		writeJSONError(w, http.StatusBadGateway, "UPSTREAM_UNAVAILABLE", "upstream unavailable")
		return false
	}
	if !owned {
		writeJSONError(w, http.StatusNotFound, "NOT_FOUND", "grievance not found")
		return false
	}
	return true
}

func (h *Handler) resolveUserID(w http.ResponseWriter, r *http.Request) (string, bool) {
	principal, ok := systemcontext.PrincipalFromContext(r.Context())
	if !ok {
		writeJSONError(w, http.StatusServiceUnavailable, "PLACEHOLDER_UNAVAILABLE", "placeholder identity unavailable")
		return "", false
	}
	return principal.UserID, true
}

func (h *Handler) readBoundedBody(r *http.Request) ([]byte, error) {
	if r.Body == nil {
		return nil, nil
	}
	defer func() {
		_ = r.Body.Close()
	}()
	limited := io.LimitReader(r.Body, h.cfg.MaxRequestBytes+1)
	body, err := io.ReadAll(limited)
	if err != nil {
		return nil, err
	}
	if int64(len(body)) > h.cfg.MaxRequestBytes {
		return nil, errRequestBodyTooLarge
	}
	return body, nil
}

func writeProxyError(w http.ResponseWriter, err error) {
	if errors.Is(err, proxy.ErrUpstreamTimeout) {
		writeJSONError(w, http.StatusGatewayTimeout, "UPSTREAM_TIMEOUT", "upstream timeout")
		return
	}
	if errors.Is(err, proxy.ErrUpstreamResponseTooLarge) {
		writeJSONError(w, http.StatusBadGateway, "UPSTREAM_RESPONSE_TOO_LARGE", "upstream response too large")
		return
	}
	writeJSONError(w, http.StatusBadGateway, "UPSTREAM_UNAVAILABLE", "upstream unavailable")
}

func writeJSONError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(errorResponse{Code: code, Message: message})
}

func writeBodyReadError(w http.ResponseWriter, err error) {
	if errors.Is(err, errRequestBodyTooLarge) {
		writeJSONError(w, http.StatusRequestEntityTooLarge, "REQUEST_TOO_LARGE", "request entity too large")
		return
	}
	writeJSONError(w, http.StatusBadRequest, "INVALID_REQUEST_BODY", "invalid request body")
}
