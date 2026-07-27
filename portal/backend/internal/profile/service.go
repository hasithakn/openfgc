/*
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
 * Licensed under the Apache License, Version 2.0.
 */

// Package profile lets an authenticated user view and update their own PII, sourced live from
// WSO2 IS's SCIM2 self-service API (GET/PATCH /scim2/Me) rather than duplicated in our own DB.
// Unlike internal/proxy, this forwards the caller's own bearer token — not a trusted org-id/
// group-id header pair — since /scim2/Me resolves the user and org from the token itself.
package profile

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/wso2/openfgc/portal/backend/internal/system/config"
)

var (
	// ErrUpstreamTimeout is returned when the identity server request times out.
	ErrUpstreamTimeout = errors.New("identity server request timed out")
	// ErrUpstreamUnavailable is returned when the identity server cannot be reached.
	ErrUpstreamUnavailable = errors.New("identity server unavailable")
)

const scimContentType = "application/scim+json"

// UpstreamStatusError represents a non-2xx response from WSO2 IS's SCIM2 API.
type UpstreamStatusError struct {
	StatusCode int
}

func (e *UpstreamStatusError) Error() string {
	return fmt.Sprintf("identity server responded with status %d", e.StatusCode)
}

// ScimName maps the SCIM2 Core User "name" complex attribute.
type ScimName struct {
	Formatted  string `json:"formatted,omitempty"`
	GivenName  string `json:"givenName,omitempty"`
	FamilyName string `json:"familyName,omitempty"`
}

// ScimEmail maps one entry of the SCIM2 Core User "emails" multi-valued attribute.
type ScimEmail struct {
	Value   string `json:"value,omitempty"`
	Type    string `json:"type,omitempty"`
	Primary bool   `json:"primary,omitempty"`
}

// ScimPhoneNumber maps one entry of the SCIM2 Core User "phoneNumbers" multi-valued attribute.
type ScimPhoneNumber struct {
	Value   string `json:"value,omitempty"`
	Type    string `json:"type,omitempty"`
	Primary bool   `json:"primary,omitempty"`
}

// ScimAddress maps one entry of the SCIM2 Core User "addresses" multi-valued attribute.
type ScimAddress struct {
	Formatted string `json:"formatted,omitempty"`
	Type      string `json:"type,omitempty"`
	Primary   bool   `json:"primary,omitempty"`
}

// ScimUser is the subset of the SCIM2 Core User schema this module treats as PII, plus
// Age and Birthday, neither of which is part of any static SCIM schema — they live under
// whatever custom extension schema/attribute the deployment configured (identity_server.
// scim_age_attribute_path / scim_birthday_attribute_path), so they're populated separately
// by Service, not by json.Unmarshal.
type ScimUser struct {
	UserName     string            `json:"userName,omitempty"`
	Name         ScimName          `json:"name,omitempty"`
	NickName     string            `json:"nickName,omitempty"`
	Emails       []ScimEmail       `json:"emails,omitempty"`
	PhoneNumbers []ScimPhoneNumber `json:"phoneNumbers,omitempty"`
	Addresses    []ScimAddress     `json:"addresses,omitempty"`
	Age          *int              `json:"-"`
	Birthday     string            `json:"-"`
}

// PatchOperation is a single SCIM PatchOp operation (RFC 7644 §3.5.2).
type PatchOperation struct {
	Op    string `json:"op"`
	Path  string `json:"path,omitempty"`
	Value any    `json:"value,omitempty"`
}

type patchRequest struct {
	Schemas    []string         `json:"schemas"`
	Operations []PatchOperation `json:"Operations"`
}

// Service calls WSO2 IS's SCIM2 self-service (/scim2/Me) API on behalf of the caller.
type Service struct {
	baseURL        *url.URL
	http           *http.Client
	ageSchema      string // e.g. "urn:scim:schemas:extension:custom:User"; empty disables age entirely
	ageAttr        string // e.g. "age"
	birthdaySchema string // empty disables birthday entirely
	birthdayAttr   string // e.g. "birthday"
}

// NewService builds a profile service targeting the configured Identity Server SCIM2 API.
func NewService(cfg config.Config) (*Service, error) {
	parsed, err := url.Parse(strings.TrimSpace(cfg.IdentityServer.SCIMBaseURL))
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return nil, fmt.Errorf("invalid identity_server.scim_base_url: %q", cfg.IdentityServer.SCIMBaseURL)
	}
	ageSchema, ageAttr := splitAttributePath(strings.TrimSpace(cfg.IdentityServer.SCIMAgeAttributePath))
	birthdaySchema, birthdayAttr := splitAttributePath(strings.TrimSpace(cfg.IdentityServer.SCIMBirthdayAttributePath))
	return &Service{
		baseURL:        parsed,
		http:           &http.Client{Timeout: cfg.IdentityServer.SCIMTimeout},
		ageSchema:      ageSchema,
		ageAttr:        ageAttr,
		birthdaySchema: birthdaySchema,
		birthdayAttr:   birthdayAttr,
	}, nil
}

// AgeAttributePath returns the full SCIM PATCH path for the age attribute (e.g.
// "urn:scim:schemas:extension:custom:User:age"), or "" if not configured.
func (s *Service) AgeAttributePath() string {
	if s.ageSchema == "" || s.ageAttr == "" {
		return ""
	}
	return s.ageSchema + ":" + s.ageAttr
}

// BirthdayAttributePath returns the full SCIM PATCH path for the birthday attribute,
// or "" if not configured.
func (s *Service) BirthdayAttributePath() string {
	if s.birthdaySchema == "" || s.birthdayAttr == "" {
		return ""
	}
	return s.birthdaySchema + ":" + s.birthdayAttr
}

// splitAttributePath splits a full SCIM attribute path at its last ":" — schema URNs
// always contain colons themselves, so only the final segment is the attribute name.
func splitAttributePath(path string) (schema, attr string) {
	idx := strings.LastIndex(path, ":")
	if idx < 0 {
		return "", ""
	}
	return path[:idx], path[idx+1:]
}

// GetMe fetches the caller's own SCIM2 user record.
func (s *Service) GetMe(ctx context.Context, accessToken string) (*ScimUser, error) {
	var user ScimUser
	if err := s.do(ctx, http.MethodGet, accessToken, nil, &user); err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateMe applies the given SCIM PatchOp operations to the caller's own user record.
func (s *Service) UpdateMe(ctx context.Context, accessToken string, ops []PatchOperation) (*ScimUser, error) {
	body, err := json.Marshal(patchRequest{
		Schemas:    []string{"urn:ietf:params:scim:api:messages:2.0:PatchOp"},
		Operations: ops,
	})
	if err != nil {
		return nil, err
	}
	var user ScimUser
	if err := s.do(ctx, http.MethodPatch, accessToken, body, &user); err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *Service) do(ctx context.Context, method, accessToken string, body []byte, out any) error {
	target := *s.baseURL
	target.Path = strings.TrimRight(s.baseURL.Path, "/") + "/scim2/Me"

	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, target.String(), reader)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", scimContentType)
	if body != nil {
		req.Header.Set("Content-Type", scimContentType)
	}

	resp, err := s.http.Do(req)
	if err != nil {
		var netErr net.Error
		if errors.Is(err, context.DeadlineExceeded) || (errors.As(err, &netErr) && netErr.Timeout()) {
			return ErrUpstreamTimeout
		}
		return ErrUpstreamUnavailable
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read identity server response: %w", ErrUpstreamUnavailable)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &UpstreamStatusError{StatusCode: resp.StatusCode}
	}
	if out != nil {
		if err := json.Unmarshal(respBody, out); err != nil {
			return fmt.Errorf("parse identity server response: %w", err)
		}
		if user, ok := out.(*ScimUser); ok {
			user.Age = s.extractAge(respBody)
			user.Birthday = s.extractBirthday(respBody)
		}
	}
	return nil
}

// extractCustomAttribute pulls a configured custom-schema attribute out of a raw SCIM2
// response. Done via a generic map (not a static json tag) since the schema/attribute
// names are deployment-configured, not fixed by any spec.
func extractCustomAttribute(body []byte, schema, attr string) (json.RawMessage, bool) {
	if schema == "" || attr == "" {
		return nil, false
	}
	var top map[string]json.RawMessage
	if err := json.Unmarshal(body, &top); err != nil {
		return nil, false
	}
	schemaRaw, ok := top[schema]
	if !ok {
		return nil, false
	}
	var attrs map[string]json.RawMessage
	if err := json.Unmarshal(schemaRaw, &attrs); err != nil {
		return nil, false
	}
	attrRaw, ok := attrs[attr]
	if !ok {
		return nil, false
	}
	return attrRaw, true
}

func (s *Service) extractAge(body []byte) *int {
	attrRaw, ok := extractCustomAttribute(body, s.ageSchema, s.ageAttr)
	if !ok {
		return nil
	}
	var age int
	if err := json.Unmarshal(attrRaw, &age); err != nil {
		return nil
	}
	return &age
}

func (s *Service) extractBirthday(body []byte) string {
	attrRaw, ok := extractCustomAttribute(body, s.birthdaySchema, s.birthdayAttr)
	if !ok {
		return ""
	}
	var birthday string
	if err := json.Unmarshal(attrRaw, &birthday); err != nil {
		return ""
	}
	return birthday
}
