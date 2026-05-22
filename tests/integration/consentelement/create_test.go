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

import (
	"encoding/json"
	"net/http"
	"strings"
)

// TestBatchCreate covers POST /consent-elements scenarios that return HTTP 200
// with per-item results. Validation failures (missing fields, bad type, schema
// errors, name conflicts) appear as FAILED items in the results array — not as
// HTTP-level errors.
//
// Each sub-test uses its own freshOrgID() for full isolation — no cleanup needed.
func (ts *ElementAPITestSuite) TestBatchCreate() {
	type testCase struct {
		name        string
		elements    []CreateElementRequest
		checkResult func(results []BatchResultItem)
	}

	cases := []testCase{
		// -------------------------------------------------------------------------
		// Success paths
		// -------------------------------------------------------------------------
		{
			name:     "basic element — created as v1 with default namespace",
			elements: []CreateElementRequest{{Name: "user-email", Type: "basic"}},
			checkResult: func(results []BatchResultItem) {
				ts.Require().Len(results, 1)
				ts.assertBatchSuccess(results[0], "user-email", "basic")
				ts.Equal("v1", results[0].Element.Version)
				ts.Equal("default", results[0].Element.Namespace)
			},
		},
		{
			name: "json element with schema — schema preserved in response",
			elements: []CreateElementRequest{{
				Name:   "account-payload",
				Type:   "json",
				Schema: json.RawMessage(`{"type":"object","properties":{"id":{"type":"string"}}}`),
			}},
			checkResult: func(results []BatchResultItem) {
				ts.Require().Len(results, 1)
				ts.assertBatchSuccess(results[0], "account-payload", "json")
				ts.Require().NotNil(results[0].Element.Schema, "schema must be returned for json type")
			},
		},
		{
			name: "xml element with schema — success",
			elements: []CreateElementRequest{{
				Name:   "account-xml",
				Type:   "xml",
				Schema: json.RawMessage(`"<xs:schema xmlns:xs=\"http://www.w3.org/2001/XMLSchema\"/>"`),
			}},
			checkResult: func(results []BatchResultItem) {
				ts.Require().Len(results, 1)
				ts.assertBatchSuccess(results[0], "account-xml", "xml")
			},
		},
		{
			name: "explicit namespace — stored and returned",
			elements: []CreateElementRequest{{
				Name:      "salary-amount",
				Type:      "basic",
				Namespace: "finance",
			}},
			checkResult: func(results []BatchResultItem) {
				ts.Require().Len(results, 1)
				ts.assertBatchSuccess(results[0], "salary-amount", "basic")
				ts.Equal("finance", results[0].Element.Namespace)
			},
		},
		{
			name: "all optional fields — displayName, description, properties stored and returned",
			elements: []CreateElementRequest{{
				Name:        "annotated-elem",
				Type:        "basic",
				DisplayName: ptr("Annotated Element"),
				Description: ptr("Has all optional fields"),
				Properties:  map[string]string{"env": "prod", "owner": "team-a"},
			}},
			checkResult: func(results []BatchResultItem) {
				ts.Require().Len(results, 1)
				ts.assertBatchSuccess(results[0], "annotated-elem", "basic")
				e := results[0].Element
				ts.Require().NotNil(e.DisplayName)
				ts.Equal("Annotated Element", *e.DisplayName)
				ts.Require().NotNil(e.Description)
				ts.Equal("Has all optional fields", *e.Description)
				ts.Equal("prod", e.Properties["env"])
				ts.Equal("team-a", e.Properties["owner"])
			},
		},
		{
			name: "batch of three — each succeeds independently with a unique elementId",
			elements: []CreateElementRequest{
				{Name: "first-name", Type: "basic"},
				{Name: "last-name", Type: "basic"},
				{Name: "date-of-birth", Type: "basic"},
			},
			checkResult: func(results []BatchResultItem) {
				ts.Require().Len(results, 3)
				ts.assertBatchSuccess(results[0], "first-name", "basic")
				ts.assertBatchSuccess(results[1], "last-name", "basic")
				ts.assertBatchSuccess(results[2], "date-of-birth", "basic")
				// Swagger contract: every element gets a unique UUID
				ts.NotEqual(results[0].Element.ElementID, results[1].Element.ElementID)
				ts.NotEqual(results[1].Element.ElementID, results[2].Element.ElementID)
			},
		},

		// -------------------------------------------------------------------------
		// Per-item validation failures (HTTP 200, item.Status = "FAILED")
		// -------------------------------------------------------------------------
		{
			name:     "missing name — FAILED",
			elements: []CreateElementRequest{{Type: "basic"}},
			checkResult: func(results []BatchResultItem) {
				ts.Require().Len(results, 1)
				ts.assertBatchFailed(results[0], "name is required")
			},
		},
		{
			name:     "missing type — FAILED",
			elements: []CreateElementRequest{{Name: "no-type-elem"}},
			checkResult: func(results []BatchResultItem) {
				ts.Require().Len(results, 1)
				ts.assertBatchFailed(results[0], "type is required")
			},
		},
		{
			name:     "invalid type value — FAILED",
			elements: []CreateElementRequest{{Name: "bad-type-elem", Type: "INVALID"}},
			checkResult: func(results []BatchResultItem) {
				ts.Require().Len(results, 1)
				ts.assertBatchFailed(results[0], "invalid element type")
			},
		},
		{
			name:     "old type 'json-payload' is rejected — FAILED",
			elements: []CreateElementRequest{{Name: "legacy-json", Type: "json-payload"}},
			checkResult: func(results []BatchResultItem) {
				ts.Require().Len(results, 1)
				ts.assertBatchFailed(results[0], "invalid element type")
			},
		},
		{
			name:     "old type 'resource-field' is rejected — FAILED",
			elements: []CreateElementRequest{{Name: "legacy-rf", Type: "resource-field"}},
			checkResult: func(results []BatchResultItem) {
				ts.Require().Len(results, 1)
				ts.assertBatchFailed(results[0], "invalid element type")
			},
		},
		{
			name:     "name exceeds 255 chars — FAILED",
			elements: []CreateElementRequest{{Name: strings.Repeat("a", 256), Type: "basic"}},
			checkResult: func(results []BatchResultItem) {
				ts.Require().Len(results, 1)
				ts.assertBatchFailed(results[0], "255")
			},
		},
		{
			name: "description exceeds 1024 chars — FAILED",
			elements: []CreateElementRequest{{
				Name:        "long-desc-elem",
				Type:        "basic",
				Description: ptr(strings.Repeat("x", 1025)),
			}},
			checkResult: func(results []BatchResultItem) {
				ts.Require().Len(results, 1)
				ts.assertBatchFailed(results[0], "1024")
			},
		},
		{
			name:     "json type without schema — FAILED",
			elements: []CreateElementRequest{{Name: "json-no-schema", Type: "json"}},
			checkResult: func(results []BatchResultItem) {
				ts.Require().Len(results, 1)
				ts.Equal("FAILED", results[0].Status)
			},
		},
		{
			name:     "xml type without schema — FAILED",
			elements: []CreateElementRequest{{Name: "xml-no-schema", Type: "xml"}},
			checkResult: func(results []BatchResultItem) {
				ts.Require().Len(results, 1)
				ts.Equal("FAILED", results[0].Status)
			},
		},
		{
			name: "same name twice in one batch — first SUCCESS, second FAILED (name already exists)",
			elements: []CreateElementRequest{
				{Name: "dupe-elem", Type: "basic"},
				{Name: "dupe-elem", Type: "basic"},
			},
			checkResult: func(results []BatchResultItem) {
				ts.Require().Len(results, 2)
				ts.assertBatchSuccess(results[0], "dupe-elem", "basic")
				ts.assertBatchFailed(results[1], "already exists")
			},
		},
		{
			name: "valid and invalid interleaved — failure does not block sibling items",
			elements: []CreateElementRequest{
				{Name: "interleaved-a", Type: "basic"},
				{Type: "basic"}, // missing name
				{Name: "interleaved-b", Type: "basic"},
			},
			checkResult: func(results []BatchResultItem) {
				ts.Require().Len(results, 3)
				ts.assertBatchSuccess(results[0], "interleaved-a", "basic")
				ts.assertBatchFailed(results[1], "name is required")
				ts.assertBatchSuccess(results[2], "interleaved-b", "basic")
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		ts.Run(tc.name, func() {
			orgID := freshOrgID() // isolated per sub-test — no cleanup needed
			status, resp := ts.doBatchCreate(orgID, tc.elements)
			ts.Require().Equal(http.StatusOK, status,
				"batch create must always return HTTP 200 (per-item failures are in results[])")
			ts.Require().NotNil(resp)
			tc.checkResult(resp.Results)
		})
	}
}

// TestBatchCreateHTTPErrors covers requests rejected before batch processing begins.
// These return a non-200 status with a structured error body — the entire request
// fails, not individual items.
func (ts *ElementAPITestSuite) TestBatchCreateHTTPErrors() {
	type testCase struct {
		name          string
		omitOrgID     bool
		body          any // string → sent as-is; []CreateElementRequest → JSON-marshalled
		wantStatus    int
		wantErrorCode string
	}

	orgID := freshOrgID()

	cases := []testCase{
		{
			name:          "missing org-id header — 400 CE-1003",
			omitOrgID:     true,
			body:          []CreateElementRequest{{Name: "x", Type: "basic"}},
			wantStatus:    http.StatusBadRequest,
			wantErrorCode: "CE-1003",
		},
		{
			name:          "malformed JSON body — 400 CE-1001",
			body:          `{not valid json`,
			wantStatus:    http.StatusBadRequest,
			wantErrorCode: "CE-1001",
		},
		{
			name:          "empty array — 400 CE-1002",
			body:          []CreateElementRequest{},
			wantStatus:    http.StatusBadRequest,
			wantErrorCode: "CE-1002",
		},
	}

	for _, tc := range cases {
		tc := tc
		ts.Run(tc.name, func() {
			requestOrgID := orgID
			if tc.omitOrgID {
				requestOrgID = ""
			}
			status, body := ts.doRequest(http.MethodPost, "/api/v1/consent-elements", requestOrgID, tc.body)
			ts.Require().Equal(tc.wantStatus, status)
			ts.assertAPIError(body, tc.wantErrorCode)
		})
	}
}
