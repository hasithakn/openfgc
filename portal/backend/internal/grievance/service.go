/*
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
 * Licensed under the Apache License, Version 2.0.
 */

// Package grievance contains the BFF's self-serve (/me/grievances) and officer
// (/grievance-management) route groups. Every write forwarded upstream carries a
// server-derived actorUserId/actorRole pair, injected here and never taken from the
// client body, mirroring how the me package injects a trusted group-id on consent writes.
package grievance

import (
	"encoding/json"
	"fmt"

	"github.com/wso2/openfgc/portal/backend/internal/proxy"
	"github.com/wso2/openfgc/portal/backend/internal/system/config"
)

// Service wraps the generic proxy service with grievance-specific helpers.
type Service struct {
	proxy *proxy.Service
}

// NewService builds a grievance service from app config.
func NewService(cfg config.ProxyConfig) (*Service, error) {
	p, err := proxy.NewService(cfg)
	if err != nil {
		return nil, err
	}
	return &Service{proxy: p}, nil
}

type grievanceOwnerView struct {
	UserID string `json:"userId"`
}

// isOwnedByUser reports whether a fetched grievance detail body belongs to userID.
func isOwnedByUser(body []byte, userID string) (bool, error) {
	var view grievanceOwnerView
	if err := json.Unmarshal(body, &view); err != nil {
		return false, fmt.Errorf("parse grievance response: %w", err)
	}
	return view.UserID == userID, nil
}

// withJSONFields decodes body as a JSON object, overwrites the given keys, and re-encodes it.
// Used to inject trusted server-derived fields (actorUserId, actorRole, userId) into a
// client-submitted body before forwarding upstream.
func withJSONFields(body []byte, overrides map[string]any) ([]byte, error) {
	fields := map[string]any{}
	if len(body) > 0 {
		if err := json.Unmarshal(body, &fields); err != nil {
			return nil, fmt.Errorf("parse request body: %w", err)
		}
	}
	for k, v := range overrides {
		fields[k] = v
	}
	return json.Marshal(fields)
}
