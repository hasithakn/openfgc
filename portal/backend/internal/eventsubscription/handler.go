/*
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
 * Licensed under the Apache License, Version 2.0.
 */

package eventsubscription

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

// Handler serves the admin-only /admin/event-subscriptions and /admin/events route groups.
type Handler struct {
	svc *Service
	cfg config.ProxyConfig
}

type errorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

var errRequestBodyTooLarge = errors.New("request body too large")

// NewHandler creates an event-subscription handler with an initialized service.
func NewHandler(cfg config.Config) (*Handler, error) {
	svc, err := NewService(cfg)
	if err != nil {
		return nil, err
	}
	return &Handler{svc: svc, cfg: cfg.Proxy}, nil
}

// CreateSubscription handles POST /admin/event-subscriptions.
//
// ENF's SUBSCRIPTION.GROUP_ID column is NOT NULL with no server-side default (unlike the
// consent-server, which defaults an absent group-id to the org-id). Since the portal admin
// UI has no independent notion of "group" for event subscriptions, the org-id is passed
// through as the group-id too.
func (h *Handler) CreateSubscription(w http.ResponseWriter, r *http.Request) {
	body, err := h.readBoundedBody(r)
	if err != nil {
		writeBodyReadError(w, err)
		return
	}
	principal, ok := systemcontext.PrincipalFromContext(r.Context())
	if !ok {
		writeJSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}
	if err := h.svc.proxy.ForwardWithGroupID(w, r, http.MethodPost, "/subscriptions", nil, body, principal.OrgID); err != nil {
		writeProxyError(w, err)
	}
}

// ListSubscriptions handles GET /admin/event-subscriptions.
func (h *Handler) ListSubscriptions(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.proxy.Forward(w, r, http.MethodGet, "/subscriptions", nil, nil); err != nil {
		writeProxyError(w, err)
	}
}

// GetSubscription handles GET /admin/event-subscriptions/{subscriptionId}.
func (h *Handler) GetSubscription(w http.ResponseWriter, r *http.Request) {
	subscriptionID := r.PathValue("subscriptionId")
	if subscriptionID == "" {
		writeJSONError(w, http.StatusNotFound, "NOT_FOUND", "subscription id not found")
		return
	}
	if err := h.svc.proxy.Forward(w, r, http.MethodGet, "/subscriptions/"+url.PathEscape(subscriptionID), nil, nil); err != nil {
		writeProxyError(w, err)
	}
}

// DeleteSubscription handles DELETE /admin/event-subscriptions/{subscriptionId}.
func (h *Handler) DeleteSubscription(w http.ResponseWriter, r *http.Request) {
	subscriptionID := r.PathValue("subscriptionId")
	if subscriptionID == "" {
		writeJSONError(w, http.StatusNotFound, "NOT_FOUND", "subscription id not found")
		return
	}
	if err := h.svc.proxy.Forward(w, r, http.MethodDelete, "/subscriptions/"+url.PathEscape(subscriptionID), nil, nil); err != nil {
		writeProxyError(w, err)
	}
}

// ListSubscriptionEvents handles GET /admin/event-subscriptions/{subscriptionId}/events.
func (h *Handler) ListSubscriptionEvents(w http.ResponseWriter, r *http.Request) {
	subscriptionID := r.PathValue("subscriptionId")
	if subscriptionID == "" {
		writeJSONError(w, http.StatusNotFound, "NOT_FOUND", "subscription id not found")
		return
	}
	if err := h.svc.proxy.Forward(w, r, http.MethodGet,
		"/subscriptions/"+url.PathEscape(subscriptionID)+"/events", nil, nil); err != nil {
		writeProxyError(w, err)
	}
}

// GetSubscriptionEventHistory handles GET /admin/event-subscriptions/{subscriptionId}/events/{deliveryId}/history.
func (h *Handler) GetSubscriptionEventHistory(w http.ResponseWriter, r *http.Request) {
	subscriptionID := r.PathValue("subscriptionId")
	deliveryID := r.PathValue("deliveryId")
	if subscriptionID == "" || deliveryID == "" {
		writeJSONError(w, http.StatusNotFound, "NOT_FOUND", "event delivery not found")
		return
	}
	if err := h.svc.proxy.Forward(w, r, http.MethodGet,
		"/subscriptions/"+url.PathEscape(subscriptionID)+"/events/"+url.PathEscape(deliveryID)+"/history", nil, nil); err != nil {
		writeProxyError(w, err)
	}
}

// ListTopics handles GET /admin/topics. Read-only — topic management stays a manual/Postman
// operation, this only backs the subscription-creation topic dropdown in the portal UI.
//
// The first time an org has zero topics, DefaultTopics are created for it before the list is
// returned. Org is resolved from the authenticated principal (the org_handle JWT claim), the
// same way every other route in this BFF resolves it — never client-supplied or hardcoded.
func (h *Handler) ListTopics(w http.ResponseWriter, r *http.Request) {
	if _, ok := systemcontext.PrincipalFromContext(r.Context()); !ok {
		writeJSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}
	if err := h.ensureDefaultTopics(r); err != nil {
		writeProxyError(w, err)
		return
	}
	if err := h.svc.proxy.Forward(w, r, http.MethodGet, "/topics", nil, nil); err != nil {
		writeProxyError(w, err)
	}
}

// ensureDefaultTopics creates DefaultTopics for the requesting org if it has none yet.
func (h *Handler) ensureDefaultTopics(r *http.Request) error {
	resp, err := h.svc.proxy.ForwardRaw(r, http.MethodGet, "/topics", func(q url.Values) {
		q.Set("limit", "1")
		q.Set("offset", "0")
	}, nil)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return nil
	}
	var listResp struct {
		Total int `json:"total"`
	}
	if err := json.Unmarshal(resp.Body, &listResp); err != nil || listResp.Total > 0 {
		return nil
	}
	// The outgoing POSTs are synthesized here, not forwarded client bodies, so unlike a normal
	// passthrough there's no client-supplied Content-Type to inherit — it must be set explicitly.
	originalContentType := r.Header.Get("Content-Type")
	r.Header.Set("Content-Type", "application/json")
	defer func() {
		if originalContentType == "" {
			r.Header.Del("Content-Type")
		} else {
			r.Header.Set("Content-Type", originalContentType)
		}
	}()

	for _, name := range DefaultTopics {
		payload, err := json.Marshal(map[string]string{"name": name})
		if err != nil {
			continue
		}
		// Best-effort: a concurrent request may have already created it (409), which is fine.
		_, _ = h.svc.proxy.ForwardRaw(r, http.MethodPost, "/topics", nil, payload)
	}
	return nil
}

// ListOrgEvents handles GET /admin/events.
func (h *Handler) ListOrgEvents(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.proxy.Forward(w, r, http.MethodGet, "/events", nil, nil); err != nil {
		writeProxyError(w, err)
	}
}

// GetOrgEventHistory handles GET /admin/events/{deliveryId}/history.
func (h *Handler) GetOrgEventHistory(w http.ResponseWriter, r *http.Request) {
	deliveryID := r.PathValue("deliveryId")
	if deliveryID == "" {
		writeJSONError(w, http.StatusNotFound, "NOT_FOUND", "event delivery not found")
		return
	}
	if err := h.svc.proxy.Forward(w, r, http.MethodGet, "/events/"+url.PathEscape(deliveryID)+"/history", nil, nil); err != nil {
		writeProxyError(w, err)
	}
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
