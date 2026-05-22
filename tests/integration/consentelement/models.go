/*
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
 *
 * WSO2 LLC. licenses this file to you under the Apache License,
 * Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package consentelement

import "encoding/json"

// =============================================================================
// Request types — what we send to the server.
// These mirror the server's model/consent_element.go API request types exactly.
// =============================================================================

// CreateElementRequest is one item in the POST /consent-elements batch body.
// Schema accepts a JSON object ({"type":"object"}) or a plain JSON string value.
type CreateElementRequest struct {
	Name        string            `json:"name,omitempty"`
	Namespace   string            `json:"namespace,omitempty"`
	DisplayName *string           `json:"displayName,omitempty"`
	Description *string           `json:"description,omitempty"`
	Type        string            `json:"type,omitempty"`
	Schema      json.RawMessage   `json:"schema,omitempty"`
	Properties  map[string]string `json:"properties,omitempty"`
}

// CreateElementVersionRequest is the body for POST /consent-elements/{elementId}/versions.
type CreateElementVersionRequest struct {
	DisplayName *string           `json:"displayName,omitempty"`
	Description *string           `json:"description,omitempty"`
	Schema      json.RawMessage   `json:"schema,omitempty"`
	Properties  map[string]string `json:"properties,omitempty"`
}

// =============================================================================
// Response types — what we receive from the server.
// Field names mirror the server's model/consent_element.go response types exactly.
// If the server renames a field, unmarshalling will silently zero it — the swagger
// contract assertions in suite_test.go will catch the drift.
// =============================================================================

// ElementResponse is returned by:
//   - GET  /consent-elements/{elementId}              (latest version)
//   - GET  /consent-elements/{elementId}/versions/{v} (specific version)
//   - POST /consent-elements/{elementId}/versions     (new version, HTTP 201)
//
// It is also the item type inside BatchResultItem.Element and ElementListResponse.Data.
type ElementResponse struct {
	ElementID   string            `json:"elementId"`
	Name        string            `json:"name"`
	Namespace   string            `json:"namespace"`
	Type        string            `json:"type"`
	Version     string            `json:"version"` // "v1", "v2", …
	DisplayName *string           `json:"displayName"`
	Description *string           `json:"description"`
	Schema      *string           `json:"schema"`
	Properties  map[string]string `json:"properties"`
	CreatedTime int64             `json:"createdTime"` // Unix milliseconds
}

// BatchResultItem is one entry in BatchCreateResponse.Results.
// Status is always present; Element is populated on SUCCESS, Error on FAILED.
type BatchResultItem struct {
	Status  string           `json:"status"`  // "SUCCESS" | "FAILED"
	Element *ElementResponse `json:"element"` // nil when FAILED
	Error   *string          `json:"error"`   // nil when SUCCESS
}

// BatchCreateResponse is the body returned by POST /consent-elements (HTTP 200).
// The response is always 200; per-item success/failure is in Results.
type BatchCreateResponse struct {
	Results []BatchResultItem `json:"results"`
}

// ElementListResponse is the body returned by GET /consent-elements.
type ElementListResponse struct {
	Data     []ElementResponse `json:"data"`
	Metadata PageMetadata      `json:"metadata"`
}

// ElementVersionItem is one entry inside ElementVersionListResponse.Versions.
// Element-level fields (name, namespace, type) are hoisted to the parent object.
type ElementVersionItem struct {
	Version     string            `json:"version"`
	DisplayName *string           `json:"displayName"`
	Description *string           `json:"description"`
	Schema      *string           `json:"schema"`
	Properties  map[string]string `json:"properties"`
	CreatedTime int64             `json:"createdTime"`
}

// ElementVersionListResponse is the body returned by
// GET /consent-elements/{elementId}/versions.
type ElementVersionListResponse struct {
	ElementID string               `json:"elementId"`
	Name      string               `json:"name"`
	Namespace string               `json:"namespace"`
	Type      string               `json:"type"`
	Versions  []ElementVersionItem `json:"versions"`
}

// PageMetadata carries pagination state in all list responses.
type PageMetadata struct {
	Total  int `json:"total"`
	Offset int `json:"offset"`
	Count  int `json:"count"`
	Limit  int `json:"limit"`
}

// ErrorResponse is the structured error body the server returns on HTTP 4xx/5xx.
type ErrorResponse struct {
	Code        string `json:"code"`
	Message     string `json:"message"`
	Description string `json:"description"`
	TraceID     string `json:"traceId"`
}
