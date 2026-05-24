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
	"strings"
	"time"
)

// TestCreateConsent covers POST /consents.
//
// Isolation: every sub-test gets a fresh org via freshOrgID() and no DB cleanup is needed.
//
// Layout:
//   - buildBody: creates any required elements/purposes and returns the request body.
//     Use when the sub-test needs real DB state before sending the request.
//   - rawBody: used for static bodies (JSON parse errors, validation errors that fire before
//     any DB look-up) — avoids unnecessary element/purpose setup.
//   - setup: runs before buildBody; use for pre-conditions beyond the request body.
//   - groupID: value for the group-id header; "" = omit (triggers CS-4002).
func (ts *ConsentAPITestSuite) TestCreateConsent() {
	type testCase struct {
		name string

		// buildBody creates any required elements / purposes and returns the body.
		// Receives the fresh orgID so it can scope its DB setup.
		buildBody func(orgID string) any

		// rawBody is a static body string (used for parse/header errors).
		rawBody string

		// setup runs before buildBody; use for pre-conditions (e.g. existing purpose).
		setup func(orgID string)

		groupID       string // group-id header value; "" = omit
		omitOrgID     bool
		wantStatus    int
		wantErrorCode string
		checkResult   func(orgID, groupID string, resp *ConsentResponse)
	}

	cases := []testCase{

		// -----------------------------------------------------------------------
		// Minimal happy-path
		// -----------------------------------------------------------------------
		{
			name:    "minimal: type only, no authorizations, no purposes",
			groupID: "grp-minimal",
			buildBody: func(_ string) any {
				return ConsentCreateRequest{Type: "accounts"}
			},
			wantStatus: http.StatusCreated,
			checkResult: func(_, groupID string, resp *ConsentResponse) {
				ts.assertConsentResponse(resp, "accounts", groupID)
				ts.Empty(resp.Authorizations, "no authorizations expected")
				ts.Empty(resp.Purposes, "no purposes expected")
				ts.NotNil(resp.Attributes, "attributes must be a non-nil map (empty)")
			},
		},

		// -----------------------------------------------------------------------
		// group-id header behaviour
		// -----------------------------------------------------------------------
		{
			name:    "group-id header is stored as groupId in the response",
			groupID: "my-tenant",
			buildBody: func(_ string) any {
				return ConsentCreateRequest{Type: "payments"}
			},
			wantStatus: http.StatusCreated,
			checkResult: func(_, groupID string, resp *ConsentResponse) {
				ts.Equal("my-tenant", resp.GroupID, "groupId must equal the group-id header value")
			},
		},

		// -----------------------------------------------------------------------
		// Authorizations
		// -----------------------------------------------------------------------
		{
			name:    "authorization type and status default when absent",
			groupID: "grp-auth-defaults",
			buildBody: func(_ string) any {
				// Neither type nor status — server defaults type to "default" and status to "APPROVED"
				return ConsentCreateRequest{
					Type:           "accounts",
					Authorizations: []AuthorizationRequest{{}},
				}
			},
			wantStatus: http.StatusCreated,
			checkResult: func(_, _ string, resp *ConsentResponse) {
				ts.Require().Len(resp.Authorizations, 1)
				ts.Equal("default", resp.Authorizations[0].Type)
				ts.Equal("APPROVED", resp.Authorizations[0].Status)
				ts.NotEmpty(resp.Authorizations[0].ID, "authorization id must not be empty")
			},
		},
		{
			name:    "single authorization with type, status, and userId",
			groupID: "grp-single-auth",
			buildBody: func(_ string) any {
				return ConsentCreateRequest{
					Type: "accounts",
					Authorizations: []AuthorizationRequest{
						{UserID: "user-1", Type: "accounts", Status: "APPROVED"},
					},
				}
			},
			wantStatus: http.StatusCreated,
			checkResult: func(_, _ string, resp *ConsentResponse) {
				ts.Require().Len(resp.Authorizations, 1)
				a := resp.Authorizations[0]
				ts.NotEmpty(a.ID)
				ts.Equal("accounts", a.Type)
				ts.Equal("APPROVED", a.Status)
				ts.Require().NotNil(a.UserID)
				ts.Equal("user-1", *a.UserID)
			},
		},
		{
			name:    "multiple authorizations — all returned with correct ids",
			groupID: "grp-multi-auth",
			buildBody: func(_ string) any {
				return ConsentCreateRequest{
					Type: "accounts",
					Authorizations: []AuthorizationRequest{
						{Type: "accounts", Status: "APPROVED"},
						{Type: "savings", Status: "CREATED"},
					},
				}
			},
			wantStatus: http.StatusCreated,
			checkResult: func(_, _ string, resp *ConsentResponse) {
				ts.Require().Len(resp.Authorizations, 2)
				for _, a := range resp.Authorizations {
					ts.NotEmpty(a.ID, "every authorization must have a non-empty id")
				}
				// Verify IDs are distinct
				ts.NotEqual(resp.Authorizations[0].ID, resp.Authorizations[1].ID)
			},
		},
		{
			name:    "authorization with resources object round-trips correctly",
			groupID: "grp-auth-resources",
			buildBody: func(_ string) any {
				return ConsentCreateRequest{
					Type: "accounts",
					Authorizations: []AuthorizationRequest{
						{
							Type:      "accounts",
							Status:    "APPROVED",
							Resources: map[string]interface{}{"accountIds": []string{"acc-1", "acc-2"}},
						},
					},
				}
			},
			wantStatus: http.StatusCreated,
			checkResult: func(_, _ string, resp *ConsentResponse) {
				ts.Require().Len(resp.Authorizations, 1)
				ts.NotNil(resp.Authorizations[0].Resources, "resources must be returned")
			},
		},

		// -----------------------------------------------------------------------
		// Status derivation from authorizations
		// -----------------------------------------------------------------------
		{
			name:    "no authorizations → status is CREATED",
			groupID: "grp-status-no-auth",
			buildBody: func(_ string) any {
				return ConsentCreateRequest{Type: "accounts"}
			},
			wantStatus: http.StatusCreated,
			checkResult: func(_, _ string, resp *ConsentResponse) {
				ts.Equal("CREATED", resp.Status)
			},
		},
		{
			name:    "all authorizations APPROVED → status is ACTIVE",
			groupID: "grp-status-all-approved",
			buildBody: func(_ string) any {
				return ConsentCreateRequest{
					Type: "accounts",
					Authorizations: []AuthorizationRequest{
						{Status: "APPROVED"},
						{Status: "APPROVED"},
					},
				}
			},
			wantStatus: http.StatusCreated,
			checkResult: func(_, _ string, resp *ConsentResponse) {
				ts.Equal("ACTIVE", resp.Status)
			},
		},
		{
			name:    "any authorization REJECTED → status is REJECTED",
			groupID: "grp-status-rejected",
			buildBody: func(_ string) any {
				return ConsentCreateRequest{
					Type: "accounts",
					Authorizations: []AuthorizationRequest{
						{Status: "APPROVED"},
						{Status: "REJECTED"},
					},
				}
			},
			wantStatus: http.StatusCreated,
			checkResult: func(_, _ string, resp *ConsentResponse) {
				ts.Equal("REJECTED", resp.Status)
			},
		},
		{
			name:    "all authorizations CREATED → status is CREATED",
			groupID: "grp-status-all-created",
			buildBody: func(_ string) any {
				return ConsentCreateRequest{
					Type: "accounts",
					Authorizations: []AuthorizationRequest{
						{Status: "CREATED"},
						{Status: "CREATED"},
					},
				}
			},
			wantStatus: http.StatusCreated,
			checkResult: func(_, _ string, resp *ConsentResponse) {
				ts.Equal("CREATED", resp.Status)
			},
		},
		{
			name:    "mix of APPROVED and CREATED → status is CREATED",
			groupID: "grp-status-mixed",
			buildBody: func(_ string) any {
				return ConsentCreateRequest{
					Type: "accounts",
					Authorizations: []AuthorizationRequest{
						{Status: "APPROVED"},
						{Status: "CREATED"},
					},
				}
			},
			wantStatus: http.StatusCreated,
			checkResult: func(_, _ string, resp *ConsentResponse) {
				ts.Equal("CREATED", resp.Status)
			},
		},

		// -----------------------------------------------------------------------
		// Purposes
		// -----------------------------------------------------------------------
		{
			name:    "single purpose with element approval — purpose and element returned",
			groupID: "grp-purpose",
			buildBody: func(orgID string) any {
				ts.mustCreateElement(orgID, "cs-email", "basic")
				ts.mustCreatePurpose(orgID, "cs-marketing", "cs-email")
				return ConsentCreateRequest{
					Type: "accounts",
					Purposes: []PurposeRefRequest{
						{
							Name: "cs-marketing",
							Elements: []ElementApprovalRequest{
								{Name: "cs-email", Approved: true, Value: "user@example.com"},
							},
						},
					},
				}
			},
			wantStatus: http.StatusCreated,
			checkResult: func(_, _ string, resp *ConsentResponse) {
				ts.Require().Len(resp.Purposes, 1)
				p := resp.Purposes[0]
				ts.Equal("cs-marketing", p.Name)
				ts.NotEmpty(p.PurposeID)
				ts.NotEmpty(p.Version)
				ts.Require().Len(p.Elements, 1)
				e := p.Elements[0]
				ts.Equal("cs-email", e.Name)
				ts.True(e.Approved)
				ts.NotEmpty(e.ElementID)
			},
		},
		{
			// The same element may appear in more than one purpose within a single
			// consent. Each purpose stores its own independent approval state for
			// that element — they do not interfere with each other.
			name:    "same element in two purposes — each purpose stores its own approval state independently",
			groupID: "grp-shared-elem",
			buildBody: func(orgID string) any {
				ts.mustCreateElement(orgID, "cs-se-shared", "basic")
				ts.mustCreateElement(orgID, "cs-se-only-a", "basic")
				ts.mustCreateElement(orgID, "cs-se-only-b", "basic")
				ts.mustCreatePurpose(orgID, "cs-se-purpose-a", "cs-se-shared")
				ts.doRequest(http.MethodPost, "/api/v1/consent-purposes", orgID, "", map[string]any{
					"name": "cs-se-purpose-b",
					"elements": []map[string]any{
						{"name": "cs-se-shared"},
						{"name": "cs-se-only-b"},
					},
				})
				return ConsentCreateRequest{
					Type: "accounts",
					Purposes: []PurposeRefRequest{
						{
							Name: "cs-se-purpose-a",
							Elements: []ElementApprovalRequest{
								{Name: "cs-se-shared", Approved: true}, // approved in purpose-a
							},
						},
						{
							Name: "cs-se-purpose-b",
							Elements: []ElementApprovalRequest{
								{Name: "cs-se-shared", Approved: false}, // not approved in purpose-b
								{Name: "cs-se-only-b", Approved: true},
							},
						},
					},
				}
			},
			wantStatus: http.StatusCreated,
			checkResult: func(_, _ string, resp *ConsentResponse) {
				ts.Require().Len(resp.Purposes, 2, "both purposes must be returned")

				byPurpose := make(map[string]PurposeResponse)
				for _, p := range resp.Purposes {
					byPurpose[p.Name] = p
				}

				// purpose-a: shared element approved
				purposeA := byPurpose["cs-se-purpose-a"]
				ts.Require().Len(purposeA.Elements, 1)
				ts.Equal("cs-se-shared", purposeA.Elements[0].Name)
				ts.True(purposeA.Elements[0].Approved,
					"shared element must be approved in purpose-a")

				// purpose-b: shared element NOT approved; only-b element approved
				purposeB := byPurpose["cs-se-purpose-b"]
				ts.Require().Len(purposeB.Elements, 2)
				elemByName := make(map[string]ElementApprovalResponse)
				for _, e := range purposeB.Elements {
					elemByName[e.Name] = e
				}
				ts.False(elemByName["cs-se-shared"].Approved,
					"shared element must not be approved in purpose-b")
				ts.True(elemByName["cs-se-only-b"].Approved,
					"only-b element must be approved in purpose-b")
			},
		},
		{
			name:    "multiple purposes — all returned",
			groupID: "grp-multi-purpose",
			buildBody: func(orgID string) any {
				ts.mustCreateElement(orgID, "cs-mp-elem1", "basic")
				ts.mustCreateElement(orgID, "cs-mp-elem2", "basic")
				ts.mustCreatePurpose(orgID, "cs-mp-purpose1", "cs-mp-elem1")
				ts.mustCreatePurpose(orgID, "cs-mp-purpose2", "cs-mp-elem2")
				return ConsentCreateRequest{
					Type: "accounts",
					Purposes: []PurposeRefRequest{
						{
							Name:     "cs-mp-purpose1",
							Elements: []ElementApprovalRequest{{Name: "cs-mp-elem1", Approved: true}},
						},
						{
							Name:     "cs-mp-purpose2",
							Elements: []ElementApprovalRequest{{Name: "cs-mp-elem2", Approved: false}},
						},
					},
				}
			},
			wantStatus: http.StatusCreated,
			checkResult: func(_, _ string, resp *ConsentResponse) {
				ts.Require().Len(resp.Purposes, 2)
			},
		},
		{
			name:    "purpose version pinned to v1 — resolved version in response is v1",
			groupID: "grp-purpose-version",
			buildBody: func(orgID string) any {
				elemID := ts.mustCreateElement(orgID, "cs-ver-elem", "basic")
				ts.mustCreatePurpose(orgID, "cs-ver-purpose", "cs-ver-elem")
				// Create a v2 of the purpose so "latest" would be v2
				body := map[string]any{
					"elements": []map[string]any{{"name": "cs-ver-elem"}},
				}
				ts.doRequest(http.MethodPost, "/api/v1/consent-purposes/"+
					func() string {
						_, resp := ts.doRequest(http.MethodGet, "/api/v1/consent-purposes", orgID, "", nil)
						var list struct {
							Data []struct {
								PurposeID string `json:"purposeId"`
								Name      string `json:"name"`
							} `json:"data"`
						}
						_ = json.Unmarshal(resp, &list)
						for _, p := range list.Data {
							if p.Name == "cs-ver-purpose" {
								return p.PurposeID
							}
						}
						return "unknown"
					}()+"/versions", orgID, "", body)
				_ = elemID
				return ConsentCreateRequest{
					Type: "accounts",
					Purposes: []PurposeRefRequest{
						{
							Name:    "cs-ver-purpose",
							Version: ptr("v1"), // pin to v1
							Elements: []ElementApprovalRequest{
								{Name: "cs-ver-elem", Approved: true},
							},
						},
					},
				}
			},
			wantStatus: http.StatusCreated,
			checkResult: func(_, _ string, resp *ConsentResponse) {
				ts.Require().Len(resp.Purposes, 1)
				ts.Equal("v1", resp.Purposes[0].Version)
			},
		},

		// -----------------------------------------------------------------------
		// Optional scalar fields
		// -----------------------------------------------------------------------
		{
			name:    "expirationTime round-trips (value in Unix seconds — server converts to millis)",
			groupID: "grp-expiry",
			buildBody: func(_ string) any {
				// Server: values < 10^11 are seconds and get multiplied by 1000.
				expirySecs := int64(1_800_000_000) // year ~2027 in seconds
				return ConsentCreateRequest{
					Type:           "accounts",
					ExpirationTime: &expirySecs,
				}
			},
			wantStatus: http.StatusCreated,
			checkResult: func(_, _ string, resp *ConsentResponse) {
				ts.Require().NotNil(resp.ExpirationTime)
				// Server stored it as millis
				ts.Greater(*resp.ExpirationTime, int64(946684800000),
					"expirationTime in response must be a Unix millisecond value")
			},
		},
		{
			name:    "frequency round-trips",
			groupID: "grp-freq",
			buildBody: func(_ string) any {
				return ConsentCreateRequest{Type: "accounts", Frequency: intPtr(5)}
			},
			wantStatus: http.StatusCreated,
			checkResult: func(_, _ string, resp *ConsentResponse) {
				ts.Require().NotNil(resp.Frequency)
				ts.Equal(5, *resp.Frequency)
			},
		},
		{
			name:    "recurringIndicator=true round-trips",
			groupID: "grp-recurring",
			buildBody: func(_ string) any {
				return ConsentCreateRequest{Type: "accounts", RecurringIndicator: boolPtr(true)}
			},
			wantStatus: http.StatusCreated,
			checkResult: func(_, _ string, resp *ConsentResponse) {
				ts.Require().NotNil(resp.RecurringIndicator)
				ts.True(*resp.RecurringIndicator)
			},
		},
		{
			name:    "dataAccessValidityDuration round-trips",
			groupID: "grp-davd",
			buildBody: func(_ string) any {
				return ConsentCreateRequest{Type: "accounts", DataAccessValidityDuration: int64Ptr(86400000)}
			},
			wantStatus: http.StatusCreated,
			checkResult: func(_, _ string, resp *ConsentResponse) {
				ts.Require().NotNil(resp.DataAccessValidityDuration)
				ts.Equal(int64(86400000), *resp.DataAccessValidityDuration)
			},
		},
		{
			name:    "attributes round-trip as key-value pairs",
			groupID: "grp-attrs",
			buildBody: func(_ string) any {
				return ConsentCreateRequest{
					Type:       "accounts",
					Attributes: map[string]string{"merchantId": "M-123", "channel": "mobile"},
				}
			},
			wantStatus: http.StatusCreated,
			checkResult: func(_, _ string, resp *ConsentResponse) {
				ts.Require().Len(resp.Attributes, 2)
				ts.Equal("M-123", resp.Attributes["merchantId"])
				ts.Equal("mobile", resp.Attributes["channel"])
			},
		},
		{
			name:    "all optional fields populated — all round-trip",
			groupID: "grp-full",
			buildBody: func(orgID string) any {
				ts.mustCreateElement(orgID, "cs-full-elem", "basic")
				ts.mustCreatePurpose(orgID, "cs-full-purpose", "cs-full-elem")
				return ConsentCreateRequest{
					Type:                       "accounts",
					ExpirationTime:             int64Ptr(1_800_000_000),
					Frequency:                  intPtr(10),
					RecurringIndicator:         boolPtr(true),
					DataAccessValidityDuration: int64Ptr(3_600_000),
					Attributes:                 map[string]string{"k": "v"},
					Authorizations:             []AuthorizationRequest{{Type: "accounts", Status: "APPROVED"}},
					Purposes: []PurposeRefRequest{
						{
							Name:     "cs-full-purpose",
							Elements: []ElementApprovalRequest{{Name: "cs-full-elem", Approved: true}},
						},
					},
				}
			},
			wantStatus: http.StatusCreated,
			checkResult: func(_, groupID string, resp *ConsentResponse) {
				ts.assertConsentResponse(resp, "accounts", groupID)
				ts.Require().NotNil(resp.ExpirationTime)
				ts.Require().NotNil(resp.Frequency)
				ts.Equal(10, *resp.Frequency)
				ts.Require().NotNil(resp.RecurringIndicator)
				ts.True(*resp.RecurringIndicator)
				ts.Require().NotNil(resp.DataAccessValidityDuration)
				ts.Equal(int64(3_600_000), *resp.DataAccessValidityDuration)
				ts.Len(resp.Attributes, 1)
				ts.Len(resp.Authorizations, 1)
				ts.Len(resp.Purposes, 1)
			},
		},

		// -----------------------------------------------------------------------
		// Response contract
		// -----------------------------------------------------------------------
		{
			name:    "response timestamps are Unix milliseconds",
			groupID: "grp-ts-check",
			buildBody: func(_ string) any {
				return ConsentCreateRequest{Type: "accounts"}
			},
			wantStatus: http.StatusCreated,
			checkResult: func(_, _ string, resp *ConsentResponse) {
				// Unix millis for 2000-01-01 = 946684800000. Values less than this are seconds.
				ts.Greater(resp.CreatedTime, int64(946684800000),
					"createdTime must be a Unix millisecond timestamp, not seconds")
				ts.Greater(resp.UpdatedTime, int64(946684800000),
					"updatedTime must be a Unix millisecond timestamp, not seconds")
			},
		},
		{
			name:    "id in response is a non-empty UUID",
			groupID: "grp-id-check",
			buildBody: func(_ string) any {
				return ConsentCreateRequest{Type: "accounts"}
			},
			wantStatus: http.StatusCreated,
			checkResult: func(_, _ string, resp *ConsentResponse) {
				ts.NotEmpty(resp.ID)
				// UUIDs are 36 chars: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
				ts.Len(resp.ID, 36, "consent id must be a UUID (36 chars)")
			},
		},

		// -----------------------------------------------------------------------
		// Header validation errors
		// -----------------------------------------------------------------------
		{
			name:          "missing org-id header → 400 CS-4002",
			omitOrgID:     true,
			groupID:       "grp-test",
			rawBody:       `{"type":"accounts"}`,
			wantStatus:    http.StatusBadRequest,
			wantErrorCode: "CS-4002",
		},
		{
			name:          "missing group-id header → 400 CS-4002",
			groupID:       "", // omit group-id
			rawBody:       `{"type":"accounts"}`,
			wantStatus:    http.StatusBadRequest,
			wantErrorCode: "CS-4002",
		},

		// -----------------------------------------------------------------------
		// Body validation errors
		// -----------------------------------------------------------------------
		{
			name:          "malformed JSON body → 400 CS-4001",
			groupID:       "grp-test",
			rawBody:       `{bad json`,
			wantStatus:    http.StatusBadRequest,
			wantErrorCode: "CS-4001",
		},
		{
			name:          "missing type → 400 CS-4002",
			groupID:       "grp-test",
			rawBody:       `{"authorizations":[{"type":"accounts"}]}`,
			wantStatus:    http.StatusBadRequest,
			wantErrorCode: "CS-4002",
		},
		{
			name:          "empty type → 400 CS-4002",
			groupID:       "grp-test",
			rawBody:       `{"type":"","authorizations":[{"type":"accounts"}]}`,
			wantStatus:    http.StatusBadRequest,
			wantErrorCode: "CS-4002",
		},
		{
			name:    "type exceeds 64 characters → 400 CS-4002",
			groupID: "grp-test",
			rawBody: func() string {
				longType := strings.Repeat("x", 65)
				return `{"type":"` + longType + `"}`
			}(),
			wantStatus:    http.StatusBadRequest,
			wantErrorCode: "CS-4002",
		},
		{
			name:          "expirationTime negative → 400 CS-4002",
			groupID:       "grp-test",
			rawBody:       `{"type":"accounts","expirationTime":-1}`,
			wantStatus:    http.StatusBadRequest,
			wantErrorCode: "CS-4002",
		},
		{
			// Values like 123 have too few digits to be a valid Unix timestamp in
			// either seconds (10 digits) or milliseconds (13 digits) format.
			name:          "expirationTime too small (not a valid timestamp) → 400 CS-4002",
			groupID:       "grp-test",
			rawBody:       `{"type":"accounts","expirationTime":123}`,
			wantStatus:    http.StatusBadRequest,
			wantErrorCode: "CS-4002",
		},
		{
			name:          "frequency negative → 400 CS-4002",
			groupID:       "grp-test",
			rawBody:       `{"type":"accounts","frequency":-1}`,
			wantStatus:    http.StatusBadRequest,
			wantErrorCode: "CS-4002",
		},

		// -----------------------------------------------------------------------
		// Authorization status validation
		// -----------------------------------------------------------------------
		{
			name:    "authorization with system-reserved status SYS_EXPIRED → 400 CS-4002",
			groupID: "grp-sys-status",
			buildBody: func(_ string) any {
				return ConsentCreateRequest{
					Type: "accounts",
					Authorizations: []AuthorizationRequest{
						{Type: "accounts", Status: "SYS_EXPIRED"},
					},
				}
			},
			wantStatus:    http.StatusBadRequest,
			wantErrorCode: "CS-4002",
		},
		{
			name:    "authorization with system-reserved status SYS_REVOKED → 400 CS-4002",
			groupID: "grp-sys-revoked",
			buildBody: func(_ string) any {
				return ConsentCreateRequest{
					Type: "accounts",
					Authorizations: []AuthorizationRequest{
						{Type: "accounts", Status: "SYS_REVOKED"},
					},
				}
			},
			wantStatus:    http.StatusBadRequest,
			wantErrorCode: "CS-4002",
		},

		// -----------------------------------------------------------------------
		// Purpose validation errors
		// -----------------------------------------------------------------------
		{
			name:    "purpose name not found → 400 CS-4002",
			groupID: "grp-purpose-missing",
			buildBody: func(_ string) any {
				return ConsentCreateRequest{
					Type: "accounts",
					Purposes: []PurposeRefRequest{
						{
							Name:     "no-such-purpose-xyz",
							Elements: []ElementApprovalRequest{{Name: "some-elem", Approved: true}},
						},
					},
				}
			},
			wantStatus:    http.StatusBadRequest,
			wantErrorCode: "CS-4002",
		},
		{
			name:    "element not in purpose → 400 CS-4002",
			groupID: "grp-elem-not-in-purpose",
			buildBody: func(orgID string) any {
				ts.mustCreateElement(orgID, "cs-enp-elem", "basic")
				ts.mustCreatePurpose(orgID, "cs-enp-purpose", "cs-enp-elem")
				return ConsentCreateRequest{
					Type: "accounts",
					Purposes: []PurposeRefRequest{
						{
							Name: "cs-enp-purpose",
							Elements: []ElementApprovalRequest{
								{Name: "cs-enp-nonexistent-elem", Approved: true},
							},
						},
					},
				}
			},
			wantStatus:    http.StatusBadRequest,
			wantErrorCode: "CS-4002",
		},
		{
			name:    "purpose version in invalid format → 400 CS-4002",
			groupID: "grp-bad-ver",
			buildBody: func(orgID string) any {
				ts.mustCreateElement(orgID, "cs-bv-elem", "basic")
				ts.mustCreatePurpose(orgID, "cs-bv-purpose", "cs-bv-elem")
				return ConsentCreateRequest{
					Type: "accounts",
					Purposes: []PurposeRefRequest{
						{
							Name:     "cs-bv-purpose",
							Version:  ptr("abc"), // invalid — must be "v1", "v2", …
							Elements: []ElementApprovalRequest{{Name: "cs-bv-elem", Approved: true}},
						},
					},
				}
			},
			wantStatus:    http.StatusBadRequest,
			wantErrorCode: "CS-4002",
		},
		{
			name:    "purpose version v0 (must be ≥ v1) → 400 CS-4002",
			groupID: "grp-ver-zero",
			buildBody: func(orgID string) any {
				ts.mustCreateElement(orgID, "cs-vz-elem", "basic")
				ts.mustCreatePurpose(orgID, "cs-vz-purpose", "cs-vz-elem")
				return ConsentCreateRequest{
					Type: "accounts",
					Purposes: []PurposeRefRequest{
						{
							Name:     "cs-vz-purpose",
							Version:  ptr("v0"),
							Elements: []ElementApprovalRequest{{Name: "cs-vz-elem", Approved: true}},
						},
					},
				}
			},
			wantStatus:    http.StatusBadRequest,
			wantErrorCode: "CS-4002",
		},

		// -----------------------------------------------------------------------
		// Element namespace disambiguation
		//
		// An element is uniquely identified by (name, namespace). Two elements can
		// share the same name if they live in different namespaces. The approval
		// request must supply namespace to identify the right element; when absent
		// the server defaults it to "default".
		// -----------------------------------------------------------------------
		{
			name:    "same element name in two namespaces — namespace disambiguates correctly",
			groupID: "grp-ns-two",
			buildBody: func(orgID string) any {
				ts.mustCreateElementFull(orgID, map[string]any{"name": "ns-shared", "type": "basic", "namespace": "default"})
				ts.mustCreateElementFull(orgID, map[string]any{"name": "ns-shared", "type": "basic", "namespace": "finance"})
				ts.doRequest(http.MethodPost, "/api/v1/consent-purposes", orgID, "", map[string]any{
					"name": "ns-two-ns-purp",
					"elements": []map[string]any{
						{"name": "ns-shared", "namespace": "default"},
						{"name": "ns-shared", "namespace": "finance"},
					},
				})
				return ConsentCreateRequest{
					Type: "accounts",
					Purposes: []PurposeRefRequest{{
						Name: "ns-two-ns-purp",
						Elements: []ElementApprovalRequest{
							{Name: "ns-shared", Namespace: "default", Approved: true},
							{Name: "ns-shared", Namespace: "finance", Approved: false},
						},
					}},
				}
			},
			wantStatus: http.StatusCreated,
			checkResult: func(_, _ string, resp *ConsentResponse) {
				ts.Require().Len(resp.Purposes, 1)
				ts.Require().Len(resp.Purposes[0].Elements, 2)
				byNS := make(map[string]ElementApprovalResponse)
				for _, e := range resp.Purposes[0].Elements {
					byNS[e.Namespace] = e
				}
				ts.True(byNS["default"].Approved, "default namespace element must be approved")
				ts.False(byNS["finance"].Approved, "finance namespace element must not be approved")
			},
		},
		{
			// Element lives in "finance"; the purpose references it with that namespace.
			// The approval omits namespace → server defaults to "default" → no element
			// named (ns-finance-only, "default") exists in the purpose → 400 CS-4002.
			name:    "element only in non-default namespace — omitting namespace defaults to 'default', not found → 400 CS-4002",
			groupID: "grp-ns-miss",
			buildBody: func(orgID string) any {
				ts.mustCreateElementFull(orgID, map[string]any{"name": "ns-finance-only", "type": "basic", "namespace": "finance"})
				ts.doRequest(http.MethodPost, "/api/v1/consent-purposes", orgID, "", map[string]any{
					"name":     "ns-finance-only-purp",
					"elements": []map[string]any{{"name": "ns-finance-only", "namespace": "finance"}},
				})
				return ConsentCreateRequest{
					Type: "accounts",
					Purposes: []PurposeRefRequest{{
						Name: "ns-finance-only-purp",
						Elements: []ElementApprovalRequest{
							{Name: "ns-finance-only", Approved: true}, // namespace absent → "default"
						},
					}},
				}
			},
			wantStatus:    http.StatusBadRequest,
			wantErrorCode: "CS-4002",
		},
		{
			// Element is in "default" namespace. Omitting namespace in the approval
			// defaults to "default" and matches correctly.
			name:    "element in 'default' namespace — omitting namespace in approval matches correctly",
			groupID: "grp-ns-def",
			buildBody: func(orgID string) any {
				ts.mustCreateElementFull(orgID, map[string]any{"name": "ns-def-elem", "type": "basic", "namespace": "default"})
				ts.doRequest(http.MethodPost, "/api/v1/consent-purposes", orgID, "", map[string]any{
					"name":     "ns-def-purp",
					"elements": []map[string]any{{"name": "ns-def-elem"}}, // purpose also omits → "default"
				})
				return ConsentCreateRequest{
					Type: "accounts",
					Purposes: []PurposeRefRequest{{
						Name:     "ns-def-purp",
						Elements: []ElementApprovalRequest{{Name: "ns-def-elem", Approved: true}},
					}},
				}
			},
			wantStatus: http.StatusCreated,
			checkResult: func(_, _ string, resp *ConsentResponse) {
				ts.Require().Len(resp.Purposes[0].Elements, 1)
				elem := resp.Purposes[0].Elements[0]
				ts.Equal("default", elem.Namespace)
				ts.True(elem.Approved)
			},
		},
		{
			// Element is in "default" namespace. Sending namespace: "finance" in the
			// approval references a (name, "finance") pair that does not belong to
			// the purpose → 400 CS-4002.
			name:    "explicit wrong namespace — element not in purpose → 400 CS-4002",
			groupID: "grp-ns-wrong",
			buildBody: func(orgID string) any {
				ts.mustCreateElementFull(orgID, map[string]any{"name": "ns-wrong-elem", "type": "basic", "namespace": "default"})
				ts.doRequest(http.MethodPost, "/api/v1/consent-purposes", orgID, "", map[string]any{
					"name":     "ns-wrong-purp",
					"elements": []map[string]any{{"name": "ns-wrong-elem"}}, // in "default"
				})
				return ConsentCreateRequest{
					Type: "accounts",
					Purposes: []PurposeRefRequest{{
						Name: "ns-wrong-purp",
						Elements: []ElementApprovalRequest{
							{Name: "ns-wrong-elem", Namespace: "finance", Approved: true},
						},
					}},
				}
			},
			wantStatus:    http.StatusBadRequest,
			wantErrorCode: "CS-4002",
		},

		// -----------------------------------------------------------------------
		// expirationTime normalization
		//
		// Values < 10^11 are treated as Unix seconds and multiplied by 1000.
		// Values ≥ 10^11 are treated as Unix milliseconds and stored as-is.
		// Each case uses an IIFE so buildBody and checkResult share a captured variable.
		// -----------------------------------------------------------------------
		func() testCase {
			var sentSecs int64
			return testCase{
				name:    "expirationTime in Unix seconds (< 10^11) — server converts to milliseconds",
				groupID: "grp-exp-secs-norm",
				buildBody: func(_ string) any {
					sentSecs = time.Now().Add(24 * time.Hour).Unix()
					return ConsentCreateRequest{Type: "accounts", ExpirationTime: &sentSecs}
				},
				wantStatus: http.StatusCreated,
				checkResult: func(_, _ string, resp *ConsentResponse) {
					ts.Require().NotNil(resp.ExpirationTime)
					ts.Equal(sentSecs*1000, *resp.ExpirationTime,
						"seconds input must be stored as milliseconds (input × 1000)")
				},
			}
		}(),
		func() testCase {
			var sentMs int64
			return testCase{
				name:    "expirationTime in Unix milliseconds (≥ 10^11) — stored unchanged",
				groupID: "grp-exp-ms-norm",
				buildBody: func(_ string) any {
					sentMs = time.Now().Add(24 * time.Hour).UnixMilli()
					return ConsentCreateRequest{Type: "accounts", ExpirationTime: &sentMs}
				},
				wantStatus: http.StatusCreated,
				checkResult: func(_, _ string, resp *ConsentResponse) {
					ts.Require().NotNil(resp.ExpirationTime)
					ts.Equal(sentMs, *resp.ExpirationTime,
						"milliseconds input must be stored unchanged")
				},
			}
		}(),
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

			// buildBody takes priority over rawBody.
			var reqBody any
			if tc.buildBody != nil {
				reqBody = tc.buildBody(orgID)
			} else {
				reqBody = tc.rawBody
			}

			status, body := ts.doCreateConsentRaw(requestOrgID, tc.groupID, reqBody)
			ts.Require().Equal(tc.wantStatus, status, "unexpected status; body: %s", body)

			if tc.wantErrorCode != "" {
				ts.assertAPIError(body, tc.wantErrorCode)
				return
			}

			var resp ConsentResponse
			ts.Require().NoError(json.Unmarshal(body, &resp), "unmarshal ConsentResponse: %s", body)
			if tc.checkResult != nil {
				tc.checkResult(orgID, tc.groupID, &resp)
			}
		})
	}
}
