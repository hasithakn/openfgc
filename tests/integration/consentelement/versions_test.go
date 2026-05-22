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

// TestListElementVersions covers GET /consent-elements/{elementId}/versions.
func (ts *ElementAPITestSuite) TestListElementVersions() {
	type testCase struct {
		name          string
		setup         func(orgID string) string // returns elementId
		elementID     string                    // used when setup is nil
		omitOrgID     bool
		wantStatus    int
		wantErrorCode string
		checkResult   func(resp *ElementVersionListResponse)
	}

	cases := []testCase{
		{
			name: "single version — returns list with one entry",
			setup: func(orgID string) string {
				return ts.mustCreateElement(orgID, "lv-single", "basic")
			},
			wantStatus: http.StatusOK,
			checkResult: func(resp *ElementVersionListResponse) {
				ts.Equal("lv-single", resp.Name)
				ts.Equal("basic", resp.Type)
				ts.Require().Len(resp.Versions, 1)
				ts.Equal("v1", resp.Versions[0].Version)
				ts.Greater(resp.Versions[0].CreatedTime, int64(0))
			},
		},
		{
			name: "multiple versions — all returned in ascending version order",
			setup: func(orgID string) string {
				id := ts.mustCreateElement(orgID, "lv-multi", "basic")
				ts.mustCreateVersion(orgID, id, CreateElementVersionRequest{DisplayName: ptr("V2")})
				ts.mustCreateVersion(orgID, id, CreateElementVersionRequest{DisplayName: ptr("V3")})
				return id
			},
			wantStatus: http.StatusOK,
			checkResult: func(resp *ElementVersionListResponse) {
				ts.Require().Len(resp.Versions, 3)
				ts.Equal("v1", resp.Versions[0].Version)
				ts.Equal("v2", resp.Versions[1].Version)
				ts.Equal("v3", resp.Versions[2].Version)
			},
		},
		{
			name: "element-level fields hoisted to response root",
			setup: func(orgID string) string {
				return ts.mustCreateElementWith(orgID, CreateElementRequest{
					Name:      "lv-namespaced",
					Type:      "basic",
					Namespace: "payroll",
				})
			},
			wantStatus: http.StatusOK,
			checkResult: func(resp *ElementVersionListResponse) {
				ts.NotEmpty(resp.ElementID)
				ts.Equal("lv-namespaced", resp.Name)
				ts.Equal("payroll", resp.Namespace)
				ts.Equal("basic", resp.Type)
			},
		},
		{
			name:          "non-existent element — 404 CE-1016",
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

			status, body := ts.doListVersionsRaw(requestOrgID, elemID)
			ts.Require().Equal(tc.wantStatus, status)

			if tc.wantErrorCode != "" {
				ts.assertAPIError(body, tc.wantErrorCode)
				return
			}

			var resp ElementVersionListResponse
			ts.Require().NoError(json.Unmarshal(body, &resp))
			if tc.checkResult != nil {
				tc.checkResult(&resp)
			}
		})
	}
}

// TestCreateElementVersion covers POST /consent-elements/{elementId}/versions.
func (ts *ElementAPITestSuite) TestCreateElementVersion() {
	type testCase struct {
		name          string
		setup         func(orgID string) string // returns elementId
		elementID     string
		req           CreateElementVersionRequest
		rawBody       string // for malformed JSON tests
		omitOrgID     bool
		wantStatus    int
		wantErrorCode string
		checkResult   func(elem *ElementResponse)
	}

	cases := []testCase{
		{
			name: "creates v2 — version number auto-increments",
			setup: func(orgID string) string {
				return ts.mustCreateElement(orgID, "cv-auto-inc", "basic")
			},
			req:        CreateElementVersionRequest{DisplayName: ptr("Second Version")},
			wantStatus: http.StatusCreated,
			checkResult: func(elem *ElementResponse) {
				ts.assertElementResponse(elem, "cv-auto-inc", "basic")
				ts.Equal("v2", elem.Version)
				ts.Require().NotNil(elem.DisplayName)
				ts.Equal("Second Version", *elem.DisplayName)
			},
		},
		{
			name: "creates v3 after v2 — continues incrementing",
			setup: func(orgID string) string {
				id := ts.mustCreateElement(orgID, "cv-triple", "basic")
				ts.mustCreateVersion(orgID, id, CreateElementVersionRequest{})
				return id
			},
			req:        CreateElementVersionRequest{Description: ptr("Third")},
			wantStatus: http.StatusCreated,
			checkResult: func(elem *ElementResponse) {
				ts.Equal("v3", elem.Version)
			},
		},
		{
			name: "immutable fields inherited — name, namespace, type unchanged",
			setup: func(orgID string) string {
				return ts.mustCreateElementWith(orgID, CreateElementRequest{
					Name:      "cv-immutable",
					Type:      "basic",
					Namespace: "security",
				})
			},
			req:        CreateElementVersionRequest{DisplayName: ptr("New display")},
			wantStatus: http.StatusCreated,
			checkResult: func(elem *ElementResponse) {
				ts.Equal("cv-immutable", elem.Name, "name must be inherited from element")
				ts.Equal("security", elem.Namespace, "namespace must be inherited")
				ts.Equal("basic", elem.Type, "type must be inherited")
			},
		},
		{
			name: "new version with properties — stored and returned",
			setup: func(orgID string) string {
				return ts.mustCreateElement(orgID, "cv-props", "basic")
			},
			req: CreateElementVersionRequest{
				Properties: map[string]string{"reviewed": "true"},
			},
			wantStatus: http.StatusCreated,
			checkResult: func(elem *ElementResponse) {
				ts.Equal("true", elem.Properties["reviewed"])
			},
		},
		{
			name:          "non-existent element — 404 CE-1016",
			elementID:     "00000000-0000-0000-0000-000000000000",
			req:           CreateElementVersionRequest{},
			wantStatus:    http.StatusNotFound,
			wantErrorCode: "CE-1016",
		},
		{
			name:          "malformed JSON body — 400 CE-1001",
			elementID:     "00000000-0000-0000-0000-000000000000",
			rawBody:       `{bad json`,
			wantStatus:    http.StatusBadRequest,
			wantErrorCode: "CE-1001",
		},
		{
			name:          "missing org-id header — 400 CE-1003",
			elementID:     "00000000-0000-0000-0000-000000000000",
			req:           CreateElementVersionRequest{},
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

			var body any = tc.req
			if tc.rawBody != "" {
				body = tc.rawBody
			}

			path := "/api/v1/consent-elements/" + elemID + "/versions"
			status, respBody := ts.doRequest(http.MethodPost, path, requestOrgID, body)
			ts.Require().Equal(tc.wantStatus, status)

			if tc.wantErrorCode != "" {
				ts.assertAPIError(respBody, tc.wantErrorCode)
				return
			}

			var elem ElementResponse
			ts.Require().NoError(json.Unmarshal(respBody, &elem))
			if tc.checkResult != nil {
				tc.checkResult(&elem)
			}
		})
	}
}

// TestGetElementVersion covers GET /consent-elements/{elementId}/versions/{version}.
func (ts *ElementAPITestSuite) TestGetElementVersion() {
	type testCase struct {
		name          string
		setup         func(orgID string) (elemID, version string)
		omitOrgID     bool
		wantStatus    int
		wantErrorCode string
		checkResult   func(elem *ElementResponse)
	}

	cases := []testCase{
		{
			name: "get v1 — returns correct version data",
			setup: func(orgID string) (string, string) {
				id := ts.mustCreateElement(orgID, "gv-v1", "basic")
				ts.mustCreateVersion(orgID, id, CreateElementVersionRequest{DisplayName: ptr("V2")})
				return id, "v1"
			},
			wantStatus: http.StatusOK,
			checkResult: func(elem *ElementResponse) {
				ts.Equal("v1", elem.Version)
				ts.Nil(elem.DisplayName, "v1 has no displayName; v2 does")
			},
		},
		{
			name: "get v2 — returns v2 not v1",
			setup: func(orgID string) (string, string) {
				id := ts.mustCreateElement(orgID, "gv-v2", "basic")
				ts.mustCreateVersion(orgID, id, CreateElementVersionRequest{DisplayName: ptr("V2 display")})
				return id, "v2"
			},
			wantStatus: http.StatusOK,
			checkResult: func(elem *ElementResponse) {
				ts.Equal("v2", elem.Version)
				ts.Require().NotNil(elem.DisplayName)
				ts.Equal("V2 display", *elem.DisplayName)
			},
		},
		{
			name: "invalid version format — 400",
			setup: func(orgID string) (string, string) {
				id := ts.mustCreateElement(orgID, "gv-bad-fmt", "basic")
				return id, "2" // must be "v2", not "2"
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "version not found — 404 CE-1016",
			setup: func(orgID string) (string, string) {
				id := ts.mustCreateElement(orgID, "gv-no-v99", "basic")
				return id, "v99"
			},
			wantStatus:    http.StatusNotFound,
			wantErrorCode: "CE-1016",
		},
		{
			name: "missing org-id header — 400 CE-1003",
			setup: func(orgID string) (string, string) {
				return "00000000-0000-0000-0000-000000000000", "v1"
			},
			omitOrgID:     true,
			wantStatus:    http.StatusBadRequest,
			wantErrorCode: "CE-1003",
		},
	}

	for _, tc := range cases {
		tc := tc
		ts.Run(tc.name, func() {
			orgID := freshOrgID()
			elemID, version := tc.setup(orgID)

			requestOrgID := orgID
			if tc.omitOrgID {
				requestOrgID = ""
			}

			statusCode, respBody := ts.doRequest(
				http.MethodGet,
				"/api/v1/consent-elements/"+elemID+"/versions/"+version,
				requestOrgID,
				nil,
			)
			ts.Require().Equal(tc.wantStatus, statusCode)

			if tc.wantErrorCode != "" {
				ts.assertAPIError(respBody, tc.wantErrorCode)
				return
			}

			if tc.checkResult != nil {
				var elem ElementResponse
				ts.Require().NoError(json.Unmarshal(respBody, &elem))
				tc.checkResult(&elem)
			}
		})
	}
}

// TestDeleteElementVersion covers DELETE /consent-elements/{elementId}/versions/{version}.
func (ts *ElementAPITestSuite) TestDeleteElementVersion() {
	type testCase struct {
		name          string
		setup         func(orgID string) (elemID, version string)
		omitOrgID     bool
		wantStatus    int
		wantErrorCode string
		afterDelete   func(orgID, elemID string) // optional assertions after successful delete
	}

	cases := []testCase{
		{
			name: "delete v1 when v2 exists — 204, v2 still accessible",
			setup: func(orgID string) (string, string) {
				id := ts.mustCreateElement(orgID, "dv-two-vers", "basic")
				ts.mustCreateVersion(orgID, id, CreateElementVersionRequest{})
				return id, "v1"
			},
			wantStatus: http.StatusNoContent,
			afterDelete: func(orgID, elemID string) {
				// v1 gone
				statusV1, _ := ts.doGetVersion(orgID, elemID, "v1")
				ts.Equal(http.StatusNotFound, statusV1, "v1 must be gone after deletion")
				// v2 still present
				statusV2, v2 := ts.doGetVersion(orgID, elemID, "v2")
				ts.Equal(http.StatusOK, statusV2, "v2 must still be accessible")
				ts.Equal("v2", v2.Version)
			},
		},
		{
			name: "delete last version — 204, element itself is also removed",
			setup: func(orgID string) (string, string) {
				id := ts.mustCreateElement(orgID, "dv-last-ver", "basic")
				return id, "v1"
			},
			wantStatus: http.StatusNoContent,
			afterDelete: func(orgID, elemID string) {
				status, _ := ts.doGetElement(orgID, elemID)
				ts.Equal(http.StatusNotFound, status,
					"element must be gone after its last version is deleted")
			},
		},
		{
			name: "non-existent version — 404 CE-1016",
			setup: func(orgID string) (string, string) {
				id := ts.mustCreateElement(orgID, "dv-no-v99", "basic")
				return id, "v99"
			},
			wantStatus:    http.StatusNotFound,
			wantErrorCode: "CE-1016",
		},
		{
			name: "invalid version format — 400",
			setup: func(orgID string) (string, string) {
				id := ts.mustCreateElement(orgID, "dv-bad-fmt", "basic")
				return id, "1" // must be "v1"
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "missing org-id header — 400 CE-1003",
			setup: func(orgID string) (string, string) {
				return "00000000-0000-0000-0000-000000000000", "v1"
			},
			omitOrgID:     true,
			wantStatus:    http.StatusBadRequest,
			wantErrorCode: "CE-1003",
		},
	}

	for _, tc := range cases {
		tc := tc
		ts.Run(tc.name, func() {
			orgID := freshOrgID()
			elemID, version := tc.setup(orgID)

			requestOrgID := orgID
			if tc.omitOrgID {
				requestOrgID = ""
			}

			path := "/api/v1/consent-elements/" + elemID + "/versions/" + version
			status, body := ts.doRequest(http.MethodDelete, path, requestOrgID, nil)
			ts.Require().Equal(tc.wantStatus, status)

			if tc.wantErrorCode != "" {
				ts.assertAPIError(body, tc.wantErrorCode)
				return
			}

			if tc.afterDelete != nil {
				tc.afterDelete(orgID, elemID)
			}
		})
	}
}

// doListVersionsRaw executes GET /consent-elements/{elementId}/versions and returns
// the raw (status, body) without parsing — used by TestListElementVersions to handle
// both success and error cases uniformly.
func (ts *ElementAPITestSuite) doListVersionsRaw(orgID, elementID string) (int, []byte) {
	return ts.doRequest(
		http.MethodGet,
		"/api/v1/consent-elements/"+elementID+"/versions",
		orgID,
		nil,
	)
}
