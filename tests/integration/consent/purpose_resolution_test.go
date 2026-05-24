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

package consent

import (
	"encoding/json"
	"net/http"
)

// TestPurposeResolution covers the two-step lookup the consent service uses when
// resolving a purpose name at consent-create/update time:
//
//  1. First, look for a purpose owned by the consent's group (group-scoped purpose).
//  2. If not found, fall back to the org-level purpose (groupId stored as orgId).
//
// Rules under test:
//   - Org-level purpose  → accessible to a consent from any group.
//   - Group-scoped purpose → accessible only to a consent from the same group.
//   - Same-name clash → group-scoped purpose shadows the org-level purpose for that group.
//   - Wrong group → 400 CS-4002 (purpose not accessible).
func (ts *ConsentAPITestSuite) TestPurposeResolution() {
	type testCase struct {
		name        string
		setup       func(orgID string)
		consentReq  func(orgID string) ConsentCreateRequest
		groupID     string // group-id header for the consent
		wantStatus  int
		wantError   string
		checkResult func(resp *ConsentResponse)
	}

	cases := []testCase{
		// -----------------------------------------------------------------------
		// Org-level purpose accessible to any group
		// -----------------------------------------------------------------------
		{
			name: "org-level purpose is accessible to a consent from any group",
			setup: func(orgID string) {
				ts.mustCreateElement(orgID, "pr-elem-org", "basic")
				// No group-id → org-level purpose (groupId stored as orgId)
				ts.mustCreatePurpose(orgID, "pr-purpose-org", "pr-elem-org")
			},
			consentReq: func(_ string) ConsentCreateRequest {
				return ConsentCreateRequest{
					Type: "accounts",
					Purposes: []PurposeRefRequest{
						{
							Name:     "pr-purpose-org",
							Elements: []ElementApprovalRequest{{Name: "pr-elem-org", Approved: true}},
						},
					},
				}
			},
			groupID:    "any-group-123", // different from the purpose's "owner"
			wantStatus: http.StatusCreated,
			checkResult: func(resp *ConsentResponse) {
				ts.Require().Len(resp.Purposes, 1)
				ts.Equal("pr-purpose-org", resp.Purposes[0].Name)
			},
		},
		{
			name: "org-level purpose is accessible to a second, distinct group",
			setup: func(orgID string) {
				ts.mustCreateElement(orgID, "pr-elem-org2", "basic")
				ts.mustCreatePurpose(orgID, "pr-purpose-org2", "pr-elem-org2")
			},
			consentReq: func(_ string) ConsentCreateRequest {
				return ConsentCreateRequest{
					Type: "accounts",
					Purposes: []PurposeRefRequest{
						{
							Name:     "pr-purpose-org2",
							Elements: []ElementApprovalRequest{{Name: "pr-elem-org2", Approved: false}},
						},
					},
				}
			},
			groupID:    "another-group-456",
			wantStatus: http.StatusCreated,
			checkResult: func(resp *ConsentResponse) {
				ts.Require().Len(resp.Purposes, 1)
				ts.Equal("pr-purpose-org2", resp.Purposes[0].Name)
			},
		},

		// -----------------------------------------------------------------------
		// Group-scoped purpose accessible only to matching group
		// -----------------------------------------------------------------------
		{
			name: "group-scoped purpose is accessible to a consent from the same group",
			setup: func(orgID string) {
				ts.mustCreateElement(orgID, "pr-elem-grp", "basic")
				// group-id = "grp-owner" → group-scoped purpose
				ts.mustCreatePurposeWithGroup(orgID, "grp-owner", "pr-purpose-grp", "pr-elem-grp")
			},
			consentReq: func(_ string) ConsentCreateRequest {
				return ConsentCreateRequest{
					Type: "accounts",
					Purposes: []PurposeRefRequest{
						{
							Name:     "pr-purpose-grp",
							Elements: []ElementApprovalRequest{{Name: "pr-elem-grp", Approved: true}},
						},
					},
				}
			},
			groupID:    "grp-owner", // same group that owns the purpose
			wantStatus: http.StatusCreated,
			checkResult: func(resp *ConsentResponse) {
				ts.Require().Len(resp.Purposes, 1)
				ts.Equal("pr-purpose-grp", resp.Purposes[0].Name)
			},
		},
		{
			name: "group-scoped purpose is NOT accessible to a consent from a different group",
			setup: func(orgID string) {
				ts.mustCreateElement(orgID, "pr-elem-grp-x", "basic")
				ts.mustCreatePurposeWithGroup(orgID, "grp-owner-x", "pr-purpose-grp-x", "pr-elem-grp-x")
			},
			consentReq: func(_ string) ConsentCreateRequest {
				return ConsentCreateRequest{
					Type: "accounts",
					Purposes: []PurposeRefRequest{
						{
							Name:     "pr-purpose-grp-x",
							Elements: []ElementApprovalRequest{{Name: "pr-elem-grp-x", Approved: true}},
						},
					},
				}
			},
			groupID:    "grp-other", // different group
			wantStatus: http.StatusBadRequest,
			wantError:  "CS-4002",
		},

		// -----------------------------------------------------------------------
		// Name uniqueness is org-wide: same name not allowed across any scope
		// -----------------------------------------------------------------------
		{
			// Creating a group-scoped purpose when an org-level purpose with the same
			// name already exists must fail — tested via mustCreatePurposeWithGroup
			// inside the consent test setup. We verify the blocking at the consent
			// level by confirming the purpose API rejects the duplicate.
			name: "cannot create group-scoped purpose when org-level with same name exists",
			setup: func(orgID string) {
				ts.mustCreateElement(orgID, "pr-block-elem", "basic")
				// Org-level purpose created first
				ts.mustCreatePurpose(orgID, "pr-block-purpose", "pr-block-elem")

				// Attempting to create a group-scoped purpose with the same name should fail.
				body := map[string]any{
					"name":     "pr-block-purpose",
					"elements": []map[string]any{{"name": "pr-block-elem"}},
				}
				status, respBody := ts.doRequest(http.MethodPost, "/api/v1/consent-purposes", orgID, "some-group", body)
				ts.Require().Equal(http.StatusConflict, status,
					"expected 409 when creating group-scoped purpose whose name exists at org level; body: %s", respBody)
			},
			// Consent references the org-level purpose (which still exists and is valid)
			consentReq: func(_ string) ConsentCreateRequest {
				return ConsentCreateRequest{
					Type: "accounts",
					Purposes: []PurposeRefRequest{
						{
							Name:     "pr-block-purpose",
							Elements: []ElementApprovalRequest{{Name: "pr-block-elem", Approved: true}},
						},
					},
				}
			},
			groupID:    "any-group",
			wantStatus: http.StatusCreated,
			checkResult: func(resp *ConsentResponse) {
				ts.Require().Len(resp.Purposes, 1)
				ts.Equal("pr-block-purpose", resp.Purposes[0].Name)
			},
		},
		{
			name: "cannot create org-level purpose when a same-name purpose exists in any group",
			setup: func(orgID string) {
				ts.mustCreateElement(orgID, "pr-block2-elem", "basic")
				// Group-scoped purpose created first
				ts.mustCreatePurposeWithGroup(orgID, "grp-first", "pr-block2-purpose", "pr-block2-elem")

				// Attempting to create an org-level purpose with the same name should also fail.
				body := map[string]any{
					"name":     "pr-block2-purpose",
					"elements": []map[string]any{{"name": "pr-block2-elem"}},
				}
				// No group-id header → org-level creation attempt
				status, respBody := ts.doRequest(http.MethodPost, "/api/v1/consent-purposes", orgID, "", body)
				ts.Require().Equal(http.StatusConflict, status,
					"expected 409 when creating org-level purpose whose name exists in another group; body: %s", respBody)
			},
			// Consent references the group-scoped purpose (still exists and is valid)
			consentReq: func(_ string) ConsentCreateRequest {
				return ConsentCreateRequest{
					Type: "accounts",
					Purposes: []PurposeRefRequest{
						{
							Name:     "pr-block2-purpose",
							Elements: []ElementApprovalRequest{{Name: "pr-block2-elem", Approved: true}},
						},
					},
				}
			},
			groupID:    "grp-first", // same group that owns the purpose
			wantStatus: http.StatusCreated,
			checkResult: func(resp *ConsentResponse) {
				ts.Require().Len(resp.Purposes, 1)
				ts.Equal("pr-block2-purpose", resp.Purposes[0].Name)
			},
		},

		// -----------------------------------------------------------------------
		// No matching purpose at all
		// -----------------------------------------------------------------------
		{
			name: "purpose not found in group or org-level → 400 CS-4002",
			setup: func(_ string) {
				// No purpose created
			},
			consentReq: func(_ string) ConsentCreateRequest {
				return ConsentCreateRequest{
					Type: "accounts",
					Purposes: []PurposeRefRequest{
						{
							Name:     "does-not-exist-anywhere",
							Elements: []ElementApprovalRequest{{Name: "any-elem", Approved: true}},
						},
					},
				}
			},
			groupID:    "some-group",
			wantStatus: http.StatusBadRequest,
			wantError:  "CS-4002",
		},
	}

	for _, tc := range cases {
		tc := tc
		ts.Run(tc.name, func() {
			orgID := freshOrgID()

			if tc.setup != nil {
				tc.setup(orgID)
			}

			req := tc.consentReq(orgID)
			status, body := ts.doCreateConsentRaw(orgID, tc.groupID, req)
			ts.Require().Equal(tc.wantStatus, status, "unexpected status; body: %s", body)

			if tc.wantError != "" {
				ts.assertAPIError(body, tc.wantError)
				return
			}

			var resp ConsentResponse
			ts.Require().NoError(json.Unmarshal(body, &resp), "unmarshal ConsentResponse: %s", body)
			if tc.checkResult != nil {
				tc.checkResult(&resp)
			}
		})
	}
}
