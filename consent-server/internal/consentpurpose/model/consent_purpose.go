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

// Package model provides data models for consent purposes.
package model

// =============================================================================
// DB types — store layer only, db tags, no json tags
// =============================================================================

// PurposeVersion is one row from the PURPOSE table.
// Properties is populated separately from the PURPOSE_PROPERTY table.
// Elements is populated separately from the PURPOSE_ELEMENT_MAPPING table.
type PurposeVersion struct {
	VersionID   string                `db:"VERSION_ID"`
	ID          string                `db:"ID"`
	Name        string                `db:"NAME"`
	GroupID     string                `db:"GROUP_ID"`
	VersionNum  int                   `db:"VERSION"`
	DisplayName *string               `db:"DISPLAY_NAME"`
	Description *string               `db:"DESCRIPTION"`
	CreatedTime int64                 `db:"CREATED_TIME"`
	OrgID       string                `db:"ORG_ID"`
	Properties  map[string]string     `db:"-"`
	Elements    []PurposeMappedElement `db:"-"`
}

// PurposeVersionProperty is one row from the PURPOSE_PROPERTY table.
type PurposeVersionProperty struct {
	PurposeVersionID string `db:"PURPOSE_VERSION_ID"`
	Key              string `db:"ATT_KEY"`
	Value            string `db:"ATT_VALUE"`
	OrgID            string `db:"ORG_ID"`
}

// PurposeMappedElement is the result of joining PURPOSE_ELEMENT_MAPPING with the ELEMENT table.
// Used by store.GetPurposeVersionElements to load element details alongside the mandatory flag.
type PurposeMappedElement struct {
	ElementVersionID string `db:"ELEMENT_VERSION_ID"`
	ElementID        string `db:"ELEMENT_ID"`
	Name             string `db:"NAME"`
	Namespace        string `db:"NAMESPACE"`
	VersionNum       int    `db:"VERSION"`
	Mandatory        bool   `db:"MANDATORY"`
}

// =============================================================================
// Service input types — handler → service, no tags
// =============================================================================

// ElementRef identifies an element version by name and namespace.
// When Version is nil the service resolves to the latest available version.
type ElementRef struct {
	Name      string
	Namespace string // defaults to "default" if empty
	Version   *int   // nil = use latest version
	Mandatory bool
}

// CreatePurposeInput is the input to the CreatePurpose service method.
// GroupID is read from the group-id request header; when absent the service
// sets it to orgID (org-level purpose).
type CreatePurposeInput struct {
	Name        string
	GroupID     string
	DisplayName *string
	Description *string
	Properties  map[string]string
	Elements    []ElementRef
}

// CreatePurposeVersionInput is the input to the CreatePurposeVersion service method.
type CreatePurposeVersionInput struct {
	DisplayName *string
	Description *string
	Properties  map[string]string
	Elements    []ElementRef
}

// PurposeListFilter holds query parameters for the ListPurposes service method.
type PurposeListFilter struct {
	GroupIDs         []string
	PurposeName      string
	PurposeVersion   *int
	ElementName      string
	ElementNamespace string
	ElementVersion   *int
	Details          bool
	Limit            int
	Offset           int
}

// =============================================================================
// Service return types — service → handler, no tags
// =============================================================================

// PurposeElementOutput is the service-layer representation of one element mapped to a purpose version.
type PurposeElementOutput struct {
	ElementVersionID string
	ElementID        string
	Name             string
	Namespace        string
	VersionNum       int
	Mandatory        bool
}

// PurposeOutput is the service-layer output for a purpose at a specific version (no db tags).
type PurposeOutput struct {
	VersionID   string
	ID          string
	Name        string
	GroupID     string
	VersionNum  int
	DisplayName *string
	Description *string
	CreatedTime int64
	OrgID       string
	Properties  map[string]string
	Elements    []PurposeElementOutput
}

// PurposeListOutput is the return type from ListPurposes.
type PurposeListOutput struct {
	Data   []PurposeOutput
	Total  int
	Offset int
	Count  int
	Limit  int
}

// PurposeVersionListOutput is the return type from GetPurposeVersions.
// Common purpose fields are hoisted to the top level; Versions contains the per-version items.
type PurposeVersionListOutput struct {
	PurposeID string
	Name      string
	GroupID   string
	Versions  []PurposeOutput
}

// =============================================================================
// API request types — HTTP boundary, handler only, json tags, no db tags
// =============================================================================

// ElementRefRequest identifies an element within a purpose create or version request body.
type ElementRefRequest struct {
	Name      string  `json:"name"`
	Namespace string  `json:"namespace,omitempty"` // defaults to "default" when absent
	Version   *string `json:"version,omitempty"`   // nil = use latest; accepted formats: "v1", "v2", …
	Mandatory bool    `json:"mandatory"`
}

// CreatePurposeRequest is the body for POST /consent-purposes.
// The group-id is read from the request header, not this body.
type CreatePurposeRequest struct {
	Name        string             `json:"name"`
	DisplayName *string            `json:"displayName,omitempty"`
	Description *string            `json:"description,omitempty"`
	Properties  map[string]string  `json:"properties,omitempty"`
	Elements    []ElementRefRequest `json:"elements"`
}

// CreatePurposeVersionRequest is the body for POST /consent-purposes/{purposeId}/versions.
type CreatePurposeVersionRequest struct {
	DisplayName *string            `json:"displayName,omitempty"`
	Description *string            `json:"description,omitempty"`
	Properties  map[string]string  `json:"properties,omitempty"`
	Elements    []ElementRefRequest `json:"elements"`
}

// =============================================================================
// API response types — HTTP boundary, handler only, json tags, no db tags
// =============================================================================

// PurposeElementResponse is one element entry within a PurposeResponse or PurposeVersionItem.
type PurposeElementResponse struct {
	ElementID string `json:"elementId"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Version   string `json:"version"` // "v1", "v2", …
	Mandatory bool   `json:"mandatory"`
}

// PurposeResponse is the response body for:
//   - POST   /consent-purposes
//   - GET    /consent-purposes/{purposeId}
//   - POST   /consent-purposes/{purposeId}/versions
//   - GET    /consent-purposes/{purposeId}/versions/{version}
type PurposeResponse struct {
	PurposeID   string                   `json:"purposeId"`
	Name        string                   `json:"name"`
	GroupID     string                   `json:"groupId"`
	Version     string                   `json:"version"` // "v1", "v2", …
	DisplayName *string                  `json:"displayName,omitempty"`
	Description *string                  `json:"description,omitempty"`
	Properties  map[string]string        `json:"properties,omitempty"`
	Elements    []PurposeElementResponse `json:"elements,omitempty"`
	CreatedTime int64                    `json:"createdTime"`
}

// PageMetadata holds pagination metadata for list responses.
type PageMetadata struct {
	Total  int `json:"total"`
	Offset int `json:"offset"`
	Count  int `json:"count"`
	Limit  int `json:"limit"`
}

// PurposeListResponse is the response body for GET /consent-purposes.
type PurposeListResponse struct {
	Data     []PurposeResponse `json:"data"`
	Metadata PageMetadata      `json:"metadata"`
}

// PurposeVersionItem is one entry in PurposeVersionListResponse.
// Purpose-level fields (Name, GroupID) are hoisted to the parent object.
type PurposeVersionItem struct {
	Version     string                   `json:"version"`
	DisplayName *string                  `json:"displayName,omitempty"`
	Description *string                  `json:"description,omitempty"`
	Properties  map[string]string        `json:"properties,omitempty"`
	Elements    []PurposeElementResponse `json:"elements,omitempty"`
	CreatedTime int64                    `json:"createdTime"`
}

// PurposeVersionListResponse is the response body for GET /consent-purposes/{purposeId}/versions.
type PurposeVersionListResponse struct {
	PurposeID string               `json:"purposeId"`
	Name      string               `json:"name"`
	GroupID   string               `json:"groupId"`
	Versions  []PurposeVersionItem `json:"versions"`
}
