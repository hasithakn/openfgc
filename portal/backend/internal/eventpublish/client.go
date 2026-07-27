/*
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
 * Licensed under the Apache License, Version 2.0.
 */

// Package eventpublish publishes business events (profile/PII changes) to the Event
// Notification Framework (ENF) so subscribers can react to them. The consent-lifecycle
// events (revoke/update/expire) and account deletion are published from consent-server
// instead, since the BFF has no visibility into consent data — only profile/PII does.
package eventpublish

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// TopicUserDataChange is the ENF topic name for profile/PII change events — must match the
// DefaultTopics seeded by the eventsubscription module (internal/eventsubscription/service.go).
const TopicUserDataChange = "USER_DATA_CHANGE"

// Event is one business event to publish to ENF.
type Event struct {
	Topic    string
	Purposes []string
	Payload  map[string]any
}

type eventRequest struct {
	Topic    string         `json:"topic"`
	Purposes []string       `json:"purposes"`
	Payload  map[string]any `json:"payload"`
}

// Client publishes events to ENF. Publishing is always best-effort: a failure is logged and
// never propagated to the caller — a notification going out is never a reason to fail the
// underlying profile update that already succeeded.
type Client struct {
	baseURL *url.URL
	http    *http.Client
	timeout time.Duration
	log     *slog.Logger
}

// NewClient builds an ENF publish client. Returns (nil, nil) when baseURL is blank —
// publishing is then a no-op; Publish is safe to call on a nil *Client.
func NewClient(baseURL string, timeout time.Duration, log *slog.Logger) (*Client, error) {
	trimmed := strings.TrimSpace(baseURL)
	if trimmed == "" {
		return nil, nil
	}
	parsed, err := url.Parse(trimmed)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return nil, fmt.Errorf("invalid event_framework.api_url: %q", baseURL)
	}
	return &Client{
		baseURL: parsed,
		http:    &http.Client{Timeout: timeout},
		timeout: timeout,
		log:     log,
	}, nil
}

// Publish fires evt to ENF in the background (fire-and-forget) using a detached context, so
// the caller's request isn't held open waiting on it and isn't affected by the caller's own
// context being cancelled once its response has been written. Safe to call on a nil Client.
func (c *Client) Publish(orgID, groupID string, evt Event) {
	if c == nil {
		return
	}
	go c.publish(orgID, groupID, evt)
}

func (c *Client) publish(orgID, groupID string, evt Event) {
	purposes := evt.Purposes
	if purposes == nil {
		purposes = []string{}
	}
	// Stamp "topic" into the payload itself too — ENF's delivered webhook body flattens
	// payload fields alongside its own deliveryId, with no other indication of event type,
	// so this makes each delivered body self-describing without extra correlation.
	payload := evt.Payload
	if payload == nil {
		payload = map[string]any{}
	}
	payload["topic"] = evt.Topic
	body, err := json.Marshal(eventRequest{Topic: evt.Topic, Purposes: purposes, Payload: payload})
	if err != nil {
		c.log.Error("failed to marshal event payload", "topic", evt.Topic, "error", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	target := *c.baseURL
	target.Path = strings.TrimRight(c.baseURL.Path, "/") + "/events"

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, target.String(), bytes.NewReader(body))
	if err != nil {
		c.log.Error("failed to build event publish request", "topic", evt.Topic, "error", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("org-id", orgID)
	// ENF's EVENT table has GROUP_ID as NOT NULL — org-level events (e.g. profile changes)
	// have no natural consent group, so fall back to orgID rather than omitting the header,
	// which would insert NULL and fail the whole publish.
	if groupID == "" {
		groupID = orgID
	}
	req.Header.Set("group-id", groupID)

	resp, err := c.http.Do(req)
	if err != nil {
		c.log.Error("event publish request failed", "topic", evt.Topic, "error", err)
		return
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		c.log.Error("event publish returned an error status",
			"topic", evt.Topic, "status", resp.StatusCode, "body", string(respBody))
		return
	}

	c.log.Info("event published", "topic", evt.Topic, "status", resp.StatusCode)
}
