/*
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
 * Licensed under the Apache License, Version 2.0.
 */

package eventsubscription

import (
	"net/http"

	"github.com/wso2/openfgc/portal/backend/internal/system/auth"
	"github.com/wso2/openfgc/portal/backend/internal/system/config"
)

// Initialize sets up the event-subscription module and registers admin-only routes.
func Initialize(mux *http.ServeMux, cfg config.Config, authManager *auth.Manager) error {
	handler, err := NewHandler(cfg)
	if err != nil {
		return err
	}

	mux.Handle("POST /admin/event-subscriptions", authManager.Require(http.HandlerFunc(handler.CreateSubscription), auth.ScopeEventSubscriptionsManage))
	mux.Handle("GET /admin/event-subscriptions", authManager.Require(http.HandlerFunc(handler.ListSubscriptions), auth.ScopeEventSubscriptionsManage))
	mux.Handle("GET /admin/event-subscriptions/{subscriptionId}", authManager.Require(http.HandlerFunc(handler.GetSubscription), auth.ScopeEventSubscriptionsManage))
	mux.Handle("DELETE /admin/event-subscriptions/{subscriptionId}", authManager.Require(http.HandlerFunc(handler.DeleteSubscription), auth.ScopeEventSubscriptionsManage))
	mux.Handle("GET /admin/event-subscriptions/{subscriptionId}/events", authManager.Require(http.HandlerFunc(handler.ListSubscriptionEvents), auth.ScopeEventSubscriptionsManage))
	mux.Handle("GET /admin/event-subscriptions/{subscriptionId}/events/{deliveryId}/history", authManager.Require(http.HandlerFunc(handler.GetSubscriptionEventHistory), auth.ScopeEventSubscriptionsManage))
	mux.Handle("GET /admin/events", authManager.Require(http.HandlerFunc(handler.ListOrgEvents), auth.ScopeEventSubscriptionsManage))
	mux.Handle("GET /admin/events/{deliveryId}/history", authManager.Require(http.HandlerFunc(handler.GetOrgEventHistory), auth.ScopeEventSubscriptionsManage))
	mux.Handle("GET /admin/topics", authManager.Require(http.HandlerFunc(handler.ListTopics), auth.ScopeEventSubscriptionsManage))

	return nil
}
