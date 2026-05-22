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
	"net/url"
)

// TestListElements covers GET /consent-elements with all filter and pagination combinations.
func (ts *ElementAPITestSuite) TestListElements() {
	type testCase struct {
		name          string
		setup         func(orgID string)
		params        url.Values
		omitOrgID     bool
		wantStatus    int
		wantErrorCode string
		checkResult   func(resp *ElementListResponse)
	}

	cases := []testCase{
		// -------------------------------------------------------------------------
		// Baseline responses
		// -------------------------------------------------------------------------
		{
			name:       "empty org — returns empty data with correct metadata",
			setup:      func(orgID string) {}, // no elements created
			params:     nil,
			wantStatus: http.StatusOK,
			checkResult: func(resp *ElementListResponse) {
				ts.Empty(resp.Data, "no elements should exist for a fresh org")
				ts.Equal(0, resp.Metadata.Total)
				ts.Equal(0, resp.Metadata.Count)
			},
		},
		{
			name: "multiple elements — metadata reflects correct counts",
			setup: func(orgID string) {
				ts.mustCreateElement(orgID, "list-elem-a", "basic")
				ts.mustCreateElement(orgID, "list-elem-b", "basic")
				ts.mustCreateElement(orgID, "list-elem-c", "basic")
			},
			params:     nil,
			wantStatus: http.StatusOK,
			checkResult: func(resp *ElementListResponse) {
				ts.Equal(3, resp.Metadata.Total)
				ts.Equal(3, resp.Metadata.Count)
				ts.Len(resp.Data, 3)
			},
		},

		// -------------------------------------------------------------------------
		// Filter: name (substring / LIKE match)
		// -------------------------------------------------------------------------
		{
			name: "name filter — returns only matching elements",
			setup: func(orgID string) {
				ts.mustCreateElement(orgID, "filter-salary", "basic")
				ts.mustCreateElement(orgID, "filter-bonus", "basic")
				ts.mustCreateElement(orgID, "filter-salary-cap", "basic")
			},
			params:     url.Values{"name": {"salary"}},
			wantStatus: http.StatusOK,
			checkResult: func(resp *ElementListResponse) {
				ts.Equal(2, resp.Metadata.Total, "only 'filter-salary' and 'filter-salary-cap' match")
				for _, e := range resp.Data {
					ts.Contains(e.Name, "salary")
				}
			},
		},

		// -------------------------------------------------------------------------
		// Filter: namespace (exact match)
		// -------------------------------------------------------------------------
		{
			name: "namespace filter — returns only elements in that namespace",
			setup: func(orgID string) {
				ts.mustCreateElementWith(orgID, CreateElementRequest{Name: "ns-elem-a", Type: "basic", Namespace: "hr"})
				ts.mustCreateElementWith(orgID, CreateElementRequest{Name: "ns-elem-b", Type: "basic", Namespace: "hr"})
				ts.mustCreateElementWith(orgID, CreateElementRequest{Name: "ns-elem-c", Type: "basic", Namespace: "finance"})
			},
			params:     url.Values{"namespace": {"hr"}},
			wantStatus: http.StatusOK,
			checkResult: func(resp *ElementListResponse) {
				ts.Equal(2, resp.Metadata.Total)
				for _, e := range resp.Data {
					ts.Equal("hr", e.Namespace)
				}
			},
		},

		// -------------------------------------------------------------------------
		// Filter: type (exact match)
		// -------------------------------------------------------------------------
		{
			name: "type filter — returns only elements of that type",
			setup: func(orgID string) {
				ts.mustCreateElement(orgID, "type-basic-elem", "basic")
				ts.mustCreateElementWith(orgID, CreateElementRequest{
					Name:   "type-json-elem",
					Type:   "json",
					Schema: json.RawMessage(`{"type":"object"}`),
				})
			},
			params:     url.Values{"type": {"json"}},
			wantStatus: http.StatusOK,
			checkResult: func(resp *ElementListResponse) {
				ts.Equal(1, resp.Metadata.Total)
				ts.Equal("json", resp.Data[0].Type)
			},
		},

		// -------------------------------------------------------------------------
		// Pagination: limit and offset
		// -------------------------------------------------------------------------
		{
			name: "limit — returns only the requested number of elements",
			setup: func(orgID string) {
				for i := 0; i < 5; i++ {
					ts.mustCreateElement(orgID, "paged-elem-"+string(rune('a'+i)), "basic")
				}
			},
			params:     url.Values{"limit": {"3"}},
			wantStatus: http.StatusOK,
			checkResult: func(resp *ElementListResponse) {
				ts.Equal(3, resp.Metadata.Limit)
				ts.Len(resp.Data, 3)
				ts.Equal(3, resp.Metadata.Count)
				ts.Equal(5, resp.Metadata.Total, "total reflects all matching, not just this page")
			},
		},
		{
			name: "offset — skips the first N elements",
			setup: func(orgID string) {
				for i := 0; i < 4; i++ {
					ts.mustCreateElement(orgID, "offset-elem-"+string(rune('a'+i)), "basic")
				}
			},
			params:     url.Values{"limit": {"10"}, "offset": {"2"}},
			wantStatus: http.StatusOK,
			checkResult: func(resp *ElementListResponse) {
				ts.Equal(2, resp.Metadata.Count, "2 elements after skipping the first 2")
				ts.Equal(4, resp.Metadata.Total)
				ts.Equal(2, resp.Metadata.Offset)
			},
		},

		// -------------------------------------------------------------------------
		// details flag
		// -------------------------------------------------------------------------
		{
			name: "details=false (default) — schema and properties omitted",
			setup: func(orgID string) {
				ts.mustCreateElementWith(orgID, CreateElementRequest{
					Name:       "detail-elem",
					Type:       "basic",
					Properties: map[string]string{"key": "val"},
				})
			},
			params:     url.Values{"details": {"false"}},
			wantStatus: http.StatusOK,
			checkResult: func(resp *ElementListResponse) {
				ts.Require().Len(resp.Data, 1)
				ts.Nil(resp.Data[0].Schema, "schema must be omitted when details=false")
				ts.Empty(resp.Data[0].Properties, "properties must be omitted when details=false")
			},
		},
		{
			name: "details=true — schema and properties included",
			setup: func(orgID string) {
				ts.mustCreateElementWith(orgID, CreateElementRequest{
					Name:       "detail-full-elem",
					Type:       "basic",
					Properties: map[string]string{"owner": "team-b"},
				})
			},
			params:     url.Values{"details": {"true"}},
			wantStatus: http.StatusOK,
			checkResult: func(resp *ElementListResponse) {
				ts.Require().Len(resp.Data, 1)
				ts.Equal("team-b", resp.Data[0].Properties["owner"],
					"properties must be included when details=true")
			},
		},

		// -------------------------------------------------------------------------
		// Error cases
		// -------------------------------------------------------------------------
		{
			name:          "version filter without name or namespace — 400 CE-1017",
			setup:         func(orgID string) {},
			params:        url.Values{"version": {"v1"}},
			wantStatus:    http.StatusBadRequest,
			wantErrorCode: "CE-1017",
		},
		{
			name:          "missing org-id header — 400 CE-1003",
			setup:         func(orgID string) {},
			omitOrgID:     true,
			wantStatus:    http.StatusBadRequest,
			wantErrorCode: "CE-1003",
		},
	}

	for _, tc := range cases {
		tc := tc
		ts.Run(tc.name, func() {
			orgID := freshOrgID()
			if tc.setup != nil {
				tc.setup(orgID)
			}

			requestOrgID := orgID
			if tc.omitOrgID {
				requestOrgID = ""
			}

			status, body := ts.doRequest(http.MethodGet, buildListPath(tc.params), requestOrgID, nil)
			ts.Require().Equal(tc.wantStatus, status)

			if tc.wantErrorCode != "" {
				ts.assertAPIError(body, tc.wantErrorCode)
				return
			}

			var resp ElementListResponse
			ts.Require().NoError(json.Unmarshal(body, &resp))
			if tc.checkResult != nil {
				tc.checkResult(&resp)
			}
		})
	}
}

// buildListPath constructs the path with query params for GET /consent-elements.
func buildListPath(params url.Values) string {
	if len(params) == 0 {
		return "/api/v1/consent-elements"
	}
	return "/api/v1/consent-elements?" + params.Encode()
}
