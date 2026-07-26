/*
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
 * Licensed under the Apache License, Version 2.0.
 */

package profile

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/wso2/openfgc/portal/backend/internal/system/config"
	systemcontext "github.com/wso2/openfgc/portal/backend/internal/system/context"
)

// Handler serves the self-service GET/PATCH /profile routes.
type Handler struct {
	svc *Service
}

type errorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

var errRequestBodyTooLarge = errors.New("request body too large")

// maxProfileRequestBytes is deliberately small: profile update bodies are a handful of short
// text fields, nowhere near the proxy's general-purpose 50MB passthrough cap.
const maxProfileRequestBytes = 65536

// NewHandler creates a profile handler with an initialized service.
func NewHandler(cfg config.Config) (*Handler, error) {
	svc, err := NewService(cfg)
	if err != nil {
		return nil, err
	}
	return &Handler{svc: svc}, nil
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
	ops := buildPatchOperations(update)
	if len(ops) == 0 {
		writeJSONError(w, http.StatusBadRequest, "INVALID_REQUEST_BODY", "no profile fields to update")
		return
	}
	user, err := h.svc.UpdateMe(r.Context(), token, ops)
	if err != nil {
		writeProfileError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toProfileResponse(user))
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
