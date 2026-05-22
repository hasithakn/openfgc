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
)

// TestGetElement covers GET /consent-elements/{elementId}.
// This endpoint always returns the latest version of the element.
func (ts *ElementAPITestSuite) TestGetElement() {
	type testCase struct {
		name          string
		setup         func(orgID string) string // creates data, returns elementId to fetch
		elementID     string                    // used when setup is nil (error-case shortcuts)
		omitOrgID     bool
		wantStatus    int
		wantErrorCode string
		checkResult   func(elem *ElementResponse)
	}

	cases := []testCase{
		{
			name: "basic element — all required swagger fields present",
			setup: func(orgID string) string {
				return ts.mustCreateElement(orgID, "get-basic", "basic")
			},
			wantStatus: http.StatusOK,
			checkResult: func(elem *ElementResponse) {
				ts.assertElementResponse(elem, "get-basic", "basic")
				ts.Equal("v1", elem.Version)
				ts.Equal("default", elem.Namespace)
			},
		},
		{
			name: "json element — schema included in response",
			setup: func(orgID string) string {
				return ts.mustCreateElementWith(orgID, CreateElementRequest{
					Name:   "get-json",
					Type:   "json",
					Schema: json.RawMessage(`{"type":"object"}`),
				})
			},
			wantStatus: http.StatusOK,
			checkResult: func(elem *ElementResponse) {
				ts.assertElementResponse(elem, "get-json", "json")
				ts.Require().NotNil(elem.Schema, "schema must be returned for json type")
			},
		},
		{
			name: "element with properties — properties returned in response",
			setup: func(orgID string) string {
				return ts.mustCreateElementWith(orgID, CreateElementRequest{
					Name:       "get-with-props",
					Type:       "basic",
					Properties: map[string]string{"source": "hr-system"},
				})
			},
			wantStatus: http.StatusOK,
			checkResult: func(elem *ElementResponse) {
				ts.Equal("hr-system", elem.Properties["source"])
			},
		},
		{
			name: "after creating v2 — returns v2 (latest), not v1",
			setup: func(orgID string) string {
				id := ts.mustCreateElement(orgID, "get-versioned", "basic")
				ts.mustCreateVersion(orgID, id, CreateElementVersionRequest{
					DisplayName: ptr("Version Two"),
				})
				return id
			},
			wantStatus: http.StatusOK,
			checkResult: func(elem *ElementResponse) {
				ts.Equal("v2", elem.Version, "GET must return the latest version")
				ts.Require().NotNil(elem.DisplayName)
				ts.Equal("Version Two", *elem.DisplayName)
			},
		},
		{
			name:          "non-existent elementId — 404 CE-1016",
			elementID:     "00000000-0000-0000-0000-000000000000",
			wantStatus:    http.StatusNotFound,
			wantErrorCode: "CE-1016",
		},
		{
			name:          "missing org-id header — 400 CE-1003",
			elementID:     "00000000-0000-0000-0000-000000000000",
			omitOrgID:     true,
			wantStatus:    http.StatusBadRequest,
			wantErrorCode: "CE-1003",
		},
	}

	for _, tc := range cases {
		tc := tc
		ts.Run(tc.name, func() {
			orgID := freshOrgID()

			var elemID string
			if tc.setup != nil {
				elemID = tc.setup(orgID)
			} else {
				elemID = tc.elementID
			}

			requestOrgID := orgID
			if tc.omitOrgID {
				requestOrgID = ""
			}

			status, body := ts.doRequest(http.MethodGet, "/api/v1/consent-elements/"+elemID, requestOrgID, nil)
			ts.Require().Equal(tc.wantStatus, status)

			if tc.wantErrorCode != "" {
				ts.assertAPIError(body, tc.wantErrorCode)
				return
			}

			var elem ElementResponse
			ts.Require().NoError(json.Unmarshal(body, &elem))
			if tc.checkResult != nil {
				tc.checkResult(&elem)
			}
		})
	}
}
