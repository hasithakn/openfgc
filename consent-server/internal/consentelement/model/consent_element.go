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

// Package model provides data models for consent elements.
package model

// Element type constants — matches the API type values.
const (
	ElementTypeBasic = "basic"
	ElementTypeJSON  = "json"
	ElementTypeXML   = "xml"
)

const DefaultNamespace = "default"

// ElementVersion represents one version of a consent element — one row in the ELEMENT table.
// All versions sharing the same ID belong to the same logical element.
// Version numbers start at 1 and increment monotonically; they are never reused.
type ElementVersion struct {
	VersionID   string            `json:"-" db:"VERSION_ID"`
	ID          string            `json:"elementId" db:"ID"`
	Name        string            `json:"name" db:"NAME"`
	Namespace   string            `json:"namespace" db:"NAMESPACE"`
	Type        string            `json:"type" db:"TYPE"`
	Version     int               `json:"version" db:"VERSION"`
	DisplayName *string           `json:"displayName,omitempty" db:"DISPLAY_NAME"`
	Description *string           `json:"description,omitempty" db:"DESCRIPTION"`
	Schema      *string           `json:"schema,omitempty" db:"ELEMENT_SCHEMA"`
	CreatedTime int64             `json:"createdTime" db:"CREATED_TIME"`
	OrgID       string            `json:"-" db:"ORG_ID"`
	Properties  map[string]string `json:"properties,omitempty" db:"-"`
}

// ElementVersionProperty is one row in the ELEMENT_PROPERTY table.
// It is used internally by the store; callers see properties on ElementVersion.Properties.
type ElementVersionProperty struct {
	ElementVersionID string `db:"ELEMENT_VERSION_ID"`
	Key              string `db:"ATT_KEY"`
	Value            string `db:"ATT_VALUE"`
	OrgID            string `db:"ORG_ID"`
}

// ElementListFilters holds query parameters for GET /consent-elements.
type ElementListFilters struct {
	Name      string
	Namespace string
	Type      string
	Version   *int
	Details   bool // when true, populate Schema and Properties
	Limit     int
	Offset    int
}

// ConsentElementCreateRequest is one item in the POST /consent-elements batch request body.
type ConsentElementCreateRequest struct {
	Name        string            `json:"name"`
	Namespace   string            `json:"namespace,omitempty"`
	DisplayName *string           `json:"displayName,omitempty"`
	Description *string           `json:"description,omitempty"`
	Type        string            `json:"type"`
	Schema      *string           `json:"schema,omitempty"`
	Properties  map[string]string `json:"properties,omitempty"`
}

// ElementVersionCreateRequest is the body for POST /consent-elements/{elementId}/versions.
// Name, Namespace, and Type are immutable identifiers inherited from the element — they cannot
// change across versions and are not accepted in this request.
type ElementVersionCreateRequest struct {
	DisplayName *string           `json:"displayName,omitempty"`
	Description *string           `json:"description,omitempty"`
	Schema      *string           `json:"schema,omitempty"`
	Properties  map[string]string `json:"properties,omitempty"`
}

// BulkCreateResultItem is one entry in the BulkCreateResponse.Results slice.
type BulkCreateResultItem struct {
	Status  string          `json:"status"` // "SUCCESS" or "FAILED"
	Element *ElementVersion `json:"element,omitempty"`
	Error   *string         `json:"error,omitempty"`
}

// BulkCreateResponse is the response body for POST /consent-elements (HTTP 200).
type BulkCreateResponse struct {
	Results []BulkCreateResultItem `json:"results"`
}

// ListResponse is the response body for GET /consent-elements.
type ListResponse struct {
	Elements []ElementVersion `json:"elements"`
	Total    int              `json:"total"`
}

// VersionListResponse is the response body for GET /consent-elements/{elementId}/versions.
type VersionListResponse struct {
	ElementID string           `json:"elementId"`
	Versions  []ElementVersion `json:"versions"`
}
