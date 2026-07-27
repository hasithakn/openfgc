/*
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
 * Licensed under the Apache License, Version 2.0.
 */

package profile

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/wso2/openfgc/portal/backend/internal/eventpublish"
	"github.com/wso2/openfgc/portal/backend/internal/proxy"
	"github.com/wso2/openfgc/portal/backend/internal/system/auth"
	"github.com/wso2/openfgc/portal/backend/internal/system/config"
	systemcontext "github.com/wso2/openfgc/portal/backend/internal/system/context"
)

// Handler serves the self-service GET/PATCH/DELETE /profile routes.
type Handler struct {
	svc         *Service
	authManager *auth.Manager
	anonymize   *proxy.Service
	events      *eventpublish.Client
	log         *slog.Logger
}

type errorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

var errRequestBodyTooLarge = errors.New("request body too large")

// maxProfileRequestBytes is deliberately small: profile update bodies are a handful of short
// text fields, nowhere near the proxy's general-purpose 50MB passthrough cap.
const maxProfileRequestBytes = 65536

// NewHandler creates a profile handler with an initialized service. authManager provides
// admin-scoped access tokens for account deletion; anonymize forwards the account-deletion
// data-anonymization call to consent-server; events publishes USER_DATA_CHANGE.
func NewHandler(cfg config.Config, log *slog.Logger, authManager *auth.Manager, anonymize *proxy.Service, events *eventpublish.Client) (*Handler, error) {
	svc, err := NewService(cfg, log)
	if err != nil {
		return nil, err
	}
	return &Handler{svc: svc, authManager: authManager, anonymize: anonymize, events: events, log: log}, nil
}

// GetProfile handles GET /profile, returning the caller's own PII from WSO2 IS.
func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request) {
	token, ok := systemcontext.AccessTokenFromContext(r.Context())
	if !ok {
		writeJSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}
	user, err := h.svc.GetMe(r.Context(), token)
	if err != nil {
		writeProfileError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toProfileResponse(user))
}

// UpdateProfile handles PATCH /profile, applying only the sections present in the request body.
func (h *Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	token, ok := systemcontext.AccessTokenFromContext(r.Context())
	if !ok {
		writeJSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}
	body, err := readBoundedBody(r)
	if err != nil {
		writeBodyReadError(w, err)
		return
	}
	var update ProfileUpdateRequest
	if err := json.Unmarshal(body, &update); err != nil {
		writeJSONError(w, http.StatusBadRequest, "INVALID_REQUEST_BODY", "invalid request body")
		return
	}
	ops := buildPatchOperations(update, h.svc.AgeAttributePath(), h.svc.BirthdayAttributePath())
	if len(ops) == 0 {
		writeJSONError(w, http.StatusBadRequest, "INVALID_REQUEST_BODY", "no profile fields to update")
		return
	}
	if _, err := h.svc.UpdateMe(r.Context(), token, ops); err != nil {
		writeProfileError(w, err)
		return
	}
	// Re-fetch rather than trust the PATCH response: SCIM implementations vary on
	// whether they return the full updated resource, a partial one, or no body at all.
	user, err := h.svc.GetMe(r.Context(), token)
	if err != nil {
		writeProfileError(w, err)
		return
	}

	if principal, ok := systemcontext.PrincipalFromContext(r.Context()); ok {
		h.events.Publish(principal.OrgID, "", eventpublish.Event{
			Topic:    eventpublish.TopicUserDataChange,
			Purposes: []string{},
			Payload: map[string]any{
				"userId":     principal.UserID,
				"changed":    changedProfileFields(update),
				"actionTime": time.Now().UnixMilli(),
			},
		})
	}

	writeJSON(w, http.StatusOK, toProfileResponse(user))
}

// changedProfileFields lists which sections of a ProfileUpdateRequest were present in the
// request (using the same field names as the public API), for the USER_DATA_CHANGE event.
func changedProfileFields(update ProfileUpdateRequest) []string {
	changed := make([]string, 0, 7)
	if update.Name != nil {
		changed = append(changed, "name")
	}
	if update.NickName != nil {
		changed = append(changed, "nickName")
	}
	if update.Emails != nil {
		changed = append(changed, "emails")
	}
	if update.PhoneNumbers != nil {
		changed = append(changed, "phoneNumbers")
	}
	if update.Addresses != nil {
		changed = append(changed, "addresses")
	}
	if update.Age != nil {
		changed = append(changed, "age")
	}
	if update.Birthday != nil {
		changed = append(changed, "birthday")
	}
	return changed
}

// DeleteAccount handles DELETE /profile — the end-user "delete my account" action.
// Irreversible: anonymizes the caller's consent/grievance data in consent-server, then
// permanently deletes their WSO2 IS identity. Anonymization runs first — if the identity
// delete step fails afterward, the caller still has a live account and can safely retry,
// whereas the reverse order would leave orphaned real-identity data unreachable forever.
func (h *Handler) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	token, ok := systemcontext.AccessTokenFromContext(r.Context())
	if !ok {
		writeJSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}
	principal, ok := systemcontext.PrincipalFromContext(r.Context())
	if !ok {
		writeJSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	resp, err := h.anonymize.ForwardRaw(r, http.MethodPost, "/api/v1/anonymize", func(q url.Values) {
		q.Set("userId", principal.UserID)
	}, nil)
	if err != nil {
		h.log.Error("account deletion: anonymize call failed", "error", err)
		writeJSONError(w, http.StatusBadGateway, "UPSTREAM_UNAVAILABLE", "unable to reach consent service")
		return
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		h.log.Error("account deletion: anonymize call returned an error status",
			"status", resp.StatusCode, "body", string(resp.Body))
		writeJSONError(w, http.StatusBadGateway, "UPSTREAM_ERROR", "unable to anonymize account data")
		return
	}

	user, err := h.svc.GetMe(r.Context(), token)
	if err != nil {
		h.log.Error("account deletion: failed to look up SCIM user id after anonymizing", "error", err)
		writeJSONError(w, http.StatusBadGateway, "UPSTREAM_ERROR",
			"account data was anonymized but the identity could not be looked up; please try again")
		return
	}
	if user.ID == "" {
		h.log.Error("account deletion: SCIM user record has no id")
		writeJSONError(w, http.StatusBadGateway, "UPSTREAM_ERROR",
			"account data was anonymized but the identity could not be resolved; please try again")
		return
	}

	adminToken, err := h.authManager.AdminAccessToken(r.Context())
	if err != nil {
		h.log.Error("account deletion: failed to obtain admin access token", "error", err)
		writeJSONError(w, http.StatusBadGateway, "UPSTREAM_ERROR",
			"account data was anonymized but the identity could not be deleted; please try again")
		return
	}

	if err := h.svc.DeleteUser(r.Context(), adminToken, user.ID); err != nil {
		writeProfileError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func readBoundedBody(r *http.Request) ([]byte, error) {
	if r.Body == nil {
		return nil, nil
	}
	defer func() {
		_ = r.Body.Close()
	}()
	limited := io.LimitReader(r.Body, maxProfileRequestBytes+1)
	body, err := io.ReadAll(limited)
	if err != nil {
		return nil, err
	}
	if int64(len(body)) > maxProfileRequestBytes {
		return nil, errRequestBodyTooLarge
	}
	return body, nil
}

func writeProfileError(w http.ResponseWriter, err error) {
	var statusErr *UpstreamStatusError
	if errors.As(err, &statusErr) {
		if statusErr.StatusCode == http.StatusUnauthorized || statusErr.StatusCode == http.StatusForbidden {
			writeJSONError(w, http.StatusUnauthorized, "UPSTREAM_UNAUTHORIZED", "identity server rejected the request")
			return
		}
		writeJSONError(w, http.StatusBadGateway, "UPSTREAM_ERROR", "identity server request failed")
		return
	}
	if errors.Is(err, ErrUpstreamTimeout) {
		writeJSONError(w, http.StatusGatewayTimeout, "UPSTREAM_TIMEOUT", "upstream timeout")
		return
	}
	writeJSONError(w, http.StatusBadGateway, "UPSTREAM_UNAVAILABLE", "upstream unavailable")
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
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
