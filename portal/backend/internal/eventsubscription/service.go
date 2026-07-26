/*
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
 * Licensed under the Apache License, Version 2.0.
 */

// Package eventsubscription contains the BFF's admin-only passthrough routes for the
// Event Notification Framework (ENF): event subscriptions and event delivery history.
package eventsubscription

import (
	"github.com/wso2/openfgc/portal/backend/internal/proxy"
	"github.com/wso2/openfgc/portal/backend/internal/system/config"
)

// DefaultTopics are seeded for an org the first time it lists topics with none yet.
var DefaultTopics = []string{
	"CONSENT_REVOKE", "CONSENT_EXPIRE", "CONSENT_UPDATE", "USER_DATA_CHANGE", "ACCOUNT_DELETE",
}

// Service wraps a proxy.Service targeting the ENF upstream (a separate service from the
// consent-server), reused as-is since ENF now speaks the same org-id/group-id header
// convention as the consent-server.
type Service struct {
	proxy *proxy.Service
}

// NewService builds an event-subscription service targeting the configured ENF API.
func NewService(cfg config.Config) (*Service, error) {
	p, err := proxy.NewServiceForTarget(cfg.Proxy, cfg.EventFramework.APIURL, cfg.EventFramework.APITimeout)
	if err != nil {
		return nil, err
	}
	return &Service{proxy: p}, nil
}
