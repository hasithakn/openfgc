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

// jsonSchema is a JSON Schema that requires an object with a string field "id".
const jsonSchema = `{"type":"object","properties":{"id":{"type":"string"}},"required":["id"]}`

// xmlSchema is an XSD that declares a single <patient> element of type xs:string.
const xmlSchema = `<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema"><xs:element name="patient" type="xs:string"/></xs:schema>`

// TestElementValues covers the element value field on consent element approvals.
//
// When a consent is created, each element reference may include a value that is
// stored alongside the approval state. The server validates the value against the
// element's type and schema:
//   - basic:  any string is accepted; no schema validation.
//   - json:   value must be valid JSON; if element has a schema, must also match it.
//   - xml:    value must be well-formed XML; if element has a schema (XSD), validated against it.
//
// Values are returned in GET, list, and validate responses.
func (ts *ConsentAPITestSuite) TestElementValues() {
	type testCase struct {
		name string

		// setup creates elements + purpose, returns (orgID, purposeName already set up).
		// The test body then creates a consent with the given elementApprovals.
		setup func(orgID string) string // returns purpose name

		elementApprovals []ElementApprovalRequest // element approvals for the consent
		purposeName      string                   // set by setup return value

		rawBody    string // used for static validation-error cases
		omitOrgID  bool
		wantStatus int
		wantError  string
		checkResult func(resp *ConsentResponse)
	}

	// -----------------------------------------------------------------------
	// Helpers shared across cases
	// -----------------------------------------------------------------------

	// mustSetupBasicPurpose creates a basic element + purpose and returns the purpose name.
	mustSetupBasicPurpose := func(orgID, elemName, purposeName string) string {
		ts.mustCreateElement(orgID, elemName, "basic")
		ts.doRequest(http.MethodPost, "/api/v1/consent-purposes", orgID, "", map[string]any{
			"name":     purposeName,
			"elements": []map[string]any{{"name": elemName}},
		})
		return purposeName
	}

	// mustSetupJSONPurpose creates a JSON element with schema + purpose and returns the purpose name.
	mustSetupJSONPurpose := func(orgID, elemName, purposeName string) string {
		ts.mustCreateElementFull(orgID, map[string]any{
			"name":   elemName,
			"type":   "json",
			"schema": json.RawMessage(jsonSchema),
		})
		ts.doRequest(http.MethodPost, "/api/v1/consent-purposes", orgID, "", map[string]any{
			"name":     purposeName,
			"elements": []map[string]any{{"name": elemName}},
		})
		return purposeName
	}

	// mustSetupXMLPurpose creates an XML element with XSD + purpose and returns the purpose name.
	mustSetupXMLPurpose := func(orgID, elemName, purposeName string) string {
		ts.mustCreateElementFull(orgID, map[string]any{
			"name":   elemName,
			"type":   "xml",
			"schema": xmlSchema,
		})
		ts.doRequest(http.MethodPost, "/api/v1/consent-purposes", orgID, "", map[string]any{
			"name":     purposeName,
			"elements": []map[string]any{{"name": elemName}},
		})
		return purposeName
	}

	cases := []testCase{
		// -----------------------------------------------------------------------
		// basic element
		// -----------------------------------------------------------------------
		{
			name: "basic element with string value — stored and returned in create response",
			setup: func(orgID string) string {
				return mustSetupBasicPurpose(orgID, "ev-basic-store", "ev-purp-basic-store")
			},
			elementApprovals: []ElementApprovalRequest{
				{Name: "ev-basic-store", Approved: true, Value: "hello-world"},
			},
			wantStatus: http.StatusCreated,
			checkResult: func(resp *ConsentResponse) {
				ts.Require().Len(resp.Purposes, 1)
				ts.Require().Len(resp.Purposes[0].Elements, 1)
				elem := resp.Purposes[0].Elements[0]
				ts.Equal("ev-basic-store", elem.Name)
				ts.Require().NotNil(elem.Value, "value must be returned")
				ts.Equal("hello-world", elem.Value)
			},
		},
		{
			name: "basic element without value — value absent in response",
			setup: func(orgID string) string {
				return mustSetupBasicPurpose(orgID, "ev-basic-nil", "ev-purp-basic-nil")
			},
			elementApprovals: []ElementApprovalRequest{
				{Name: "ev-basic-nil", Approved: true}, // no Value field
			},
			wantStatus: http.StatusCreated,
			checkResult: func(resp *ConsentResponse) {
				ts.Require().Len(resp.Purposes, 1)
				ts.Require().Len(resp.Purposes[0].Elements, 1)
				elem := resp.Purposes[0].Elements[0]
				ts.Nil(elem.Value, "value must be absent when not provided")
			},
		},
		{
			name: "basic element value round-trips through GET",
			setup: func(orgID string) string {
				return mustSetupBasicPurpose(orgID, "ev-basic-rt", "ev-purp-basic-rt")
			},
			elementApprovals: []ElementApprovalRequest{
				{Name: "ev-basic-rt", Approved: true, Value: "round-trip-value"},
			},
			wantStatus: http.StatusCreated,
			checkResult: func(resp *ConsentResponse) {
				// The resp here is the create response; we also GET to verify round-trip.
				// (The GET assertion is done in the test body below using extra state.)
				ts.Require().Len(resp.Purposes[0].Elements, 1)
				ts.Equal("round-trip-value", resp.Purposes[0].Elements[0].Value)
			},
		},

		// -----------------------------------------------------------------------
		// json element
		// -----------------------------------------------------------------------
		{
			name: "json element with valid JSON object value and schema — accepted",
			setup: func(orgID string) string {
				return mustSetupJSONPurpose(orgID, "ev-json-valid", "ev-purp-json-valid")
			},
			elementApprovals: []ElementApprovalRequest{
				// value is a JSON object satisfying the schema {"type":"object","required":["id"]}
				{Name: "ev-json-valid", Approved: true, Value: map[string]string{"id": "abc-123"}},
			},
			wantStatus: http.StatusCreated,
			checkResult: func(resp *ConsentResponse) {
				ts.Require().Len(resp.Purposes, 1)
				ts.Require().Len(resp.Purposes[0].Elements, 1)
				elem := resp.Purposes[0].Elements[0]
				ts.NotNil(elem.Value, "json element value must be returned")
				// The server unmarshals the stored JSON string back to an interface{} in responses.
				asMap, ok := elem.Value.(map[string]interface{})
				ts.Require().True(ok, "json element value must be returned as an object")
				ts.Equal("abc-123", asMap["id"])
			},
		},
		{
			name: "json element value not matching schema (missing required field) → 400 CS-4002",
			setup: func(orgID string) string {
				return mustSetupJSONPurpose(orgID, "ev-json-schema-fail", "ev-purp-json-schema-fail")
			},
			elementApprovals: []ElementApprovalRequest{
				// value is valid JSON but missing the required "id" field
				{Name: "ev-json-schema-fail", Approved: true, Value: map[string]string{"name": "missing-id"}},
			},
			wantStatus: http.StatusBadRequest,
			wantError:  "CS-4002",
		},
		{
			name: "json element with invalid (non-JSON) string value → 400 CS-4002",
			setup: func(orgID string) string {
				return mustSetupJSONPurpose(orgID, "ev-json-invalid", "ev-purp-json-invalid")
			},
			elementApprovals: []ElementApprovalRequest{
				// plain string → stored as "not-valid-json" (no quotes) → fails JSON parse
				{Name: "ev-json-invalid", Approved: true, Value: "not-valid-json"},
			},
			wantStatus: http.StatusBadRequest,
			wantError:  "CS-4002",
		},
		{
			name: "json element without value — skips schema validation",
			setup: func(orgID string) string {
				return mustSetupJSONPurpose(orgID, "ev-json-noval", "ev-purp-json-noval")
			},
			elementApprovals: []ElementApprovalRequest{
				// no Value → validation skipped entirely
				{Name: "ev-json-noval", Approved: false},
			},
			wantStatus: http.StatusCreated,
			checkResult: func(resp *ConsentResponse) {
				ts.Require().Len(resp.Purposes[0].Elements, 1)
				ts.Nil(resp.Purposes[0].Elements[0].Value, "value must be nil when not provided")
			},
		},

		// -----------------------------------------------------------------------
		// xml element
		// -----------------------------------------------------------------------
		{
			name: "xml element with valid XML matching XSD — accepted",
			setup: func(orgID string) string {
				return mustSetupXMLPurpose(orgID, "ev-xml-valid", "ev-purp-xml-valid")
			},
			elementApprovals: []ElementApprovalRequest{
				// valid XML conforming to the XSD (root element is <patient>)
				{Name: "ev-xml-valid", Approved: true, Value: "<patient>hello</patient>"},
			},
			wantStatus: http.StatusCreated,
			checkResult: func(resp *ConsentResponse) {
				ts.Require().Len(resp.Purposes[0].Elements, 1)
				elem := resp.Purposes[0].Elements[0]
				ts.NotNil(elem.Value, "xml element value must be returned")
				// XML values are returned as strings (not parsed further)
				ts.Equal("<patient>hello</patient>", elem.Value)
			},
		},
		{
			name: "xml element with malformed XML → 400 CS-4002",
			setup: func(orgID string) string {
				return mustSetupXMLPurpose(orgID, "ev-xml-bad", "ev-purp-xml-bad")
			},
			elementApprovals: []ElementApprovalRequest{
				{Name: "ev-xml-bad", Approved: true, Value: "<patient>not closed"},
			},
			wantStatus: http.StatusBadRequest,
			wantError:  "CS-4002",
		},
		{
			name: "xml element without value — skips schema validation",
			setup: func(orgID string) string {
				return mustSetupXMLPurpose(orgID, "ev-xml-noval", "ev-purp-xml-noval")
			},
			elementApprovals: []ElementApprovalRequest{
				{Name: "ev-xml-noval", Approved: false}, // no Value
			},
			wantStatus: http.StatusCreated,
			checkResult: func(resp *ConsentResponse) {
				ts.Require().Len(resp.Purposes[0].Elements, 1)
				ts.Nil(resp.Purposes[0].Elements[0].Value)
			},
		},

		// -----------------------------------------------------------------------
		// Multiple elements with mixed types and values
		// -----------------------------------------------------------------------
		{
			name: "multiple elements — values stored independently per element",
			setup: func(orgID string) string {
				ts.mustCreateElement(orgID, "ev-multi-basic", "basic")
				ts.mustCreateElementFull(orgID, map[string]any{
					"name":   "ev-multi-json",
					"type":   "json",
					"schema": json.RawMessage(jsonSchema),
				})
				ts.doRequest(http.MethodPost, "/api/v1/consent-purposes", orgID, "", map[string]any{
					"name": "ev-purp-multi",
					"elements": []map[string]any{
						{"name": "ev-multi-basic"},
						{"name": "ev-multi-json"},
					},
				})
				return "ev-purp-multi"
			},
			elementApprovals: []ElementApprovalRequest{
				{Name: "ev-multi-basic", Approved: true, Value: "basic-value"},
				{Name: "ev-multi-json", Approved: true, Value: map[string]string{"id": "json-val"}},
			},
			wantStatus: http.StatusCreated,
			checkResult: func(resp *ConsentResponse) {
				ts.Require().Len(resp.Purposes, 1)
				ts.Require().Len(resp.Purposes[0].Elements, 2)
				byName := make(map[string]ElementApprovalResponse)
				for _, e := range resp.Purposes[0].Elements {
					byName[e.Name] = e
				}
				ts.Equal("basic-value", byName["ev-multi-basic"].Value)
				jsonElem := byName["ev-multi-json"]
				asMap, ok := jsonElem.Value.(map[string]interface{})
				ts.Require().True(ok, "json element value must be an object")
				ts.Equal("json-val", asMap["id"])
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		ts.Run(tc.name, func() {
			orgID := freshOrgID()

			var purposeName string
			if tc.setup != nil {
				purposeName = tc.setup(orgID)
			}

			var reqBody any
			if purposeName != "" {
				reqBody = ConsentCreateRequest{
					Type: "accounts",
					Purposes: []PurposeRefRequest{{
						Name:     purposeName,
						Elements: tc.elementApprovals,
					}},
				}
			}

			status, body := ts.doCreateConsentRaw(orgID, "grp-ev", reqBody)
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

	// -----------------------------------------------------------------------
	// Extra: element value in GET response
	// -----------------------------------------------------------------------
	ts.Run("element value persists — GET returns same value as create", func() {
		orgID := freshOrgID()
		ts.mustCreateElement(orgID, "ev-persist-elem", "basic")
		ts.doRequest(http.MethodPost, "/api/v1/consent-purposes", orgID, "", map[string]any{
			"name":     "ev-persist-purp",
			"elements": []map[string]any{{"name": "ev-persist-elem"}},
		})
		c := ts.mustCreateConsent(orgID, "grp-ev-persist", ConsentCreateRequest{
			Type: "accounts",
			Purposes: []PurposeRefRequest{{
				Name:     "ev-persist-purp",
				Elements: []ElementApprovalRequest{{Name: "ev-persist-elem", Approved: true, Value: "persist-me"}},
			}},
		})

		_, got := ts.doGetConsent(orgID, c.ID)
		ts.Require().NotNil(got)
		ts.Require().Len(got.Purposes, 1)
		ts.Require().Len(got.Purposes[0].Elements, 1)
		ts.Equal("persist-me", got.Purposes[0].Elements[0].Value,
			"element value must be the same in GET as in create response")
	})

	// -----------------------------------------------------------------------
	// Extra: element value in validate consentInformation
	// -----------------------------------------------------------------------
	ts.Run("element value appears in validate consentInformation", func() {
		orgID := freshOrgID()
		ts.mustCreateElement(orgID, "ev-val-elem", "basic")
		ts.doRequest(http.MethodPost, "/api/v1/consent-purposes", orgID, "", map[string]any{
			"name":     "ev-val-purp",
			"elements": []map[string]any{{"name": "ev-val-elem"}},
		})
		c := ts.mustCreateConsent(orgID, "grp-ev-val", ConsentCreateRequest{
			Type: "accounts",
			Authorizations: []AuthorizationRequest{{Status: "APPROVED"}},
			Purposes: []PurposeRefRequest{{
				Name:     "ev-val-purp",
				Elements: []ElementApprovalRequest{{Name: "ev-val-elem", Approved: true, Value: "in-validate"}},
			}},
		})

		_, body := ts.doValidateConsent(orgID, ConsentValidateRequest{ConsentID: c.ID})
		var valResp ConsentValidateResponse
		ts.Require().NoError(json.Unmarshal(body, &valResp))
		ts.True(valResp.IsValid)
		ts.Require().NotNil(valResp.ConsentInfo)
		ts.Require().Len(valResp.ConsentInfo.Purposes, 1)
		ts.Require().Len(valResp.ConsentInfo.Purposes[0].Elements, 1)
		ts.Equal("in-validate", valResp.ConsentInfo.Purposes[0].Elements[0].Value,
			"element value must appear in the validate consentInformation response")
	})
}
