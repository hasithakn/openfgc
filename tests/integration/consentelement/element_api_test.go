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
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/wso2/openfgc/tests/integration/testutils"
)

var serverURL = testutils.GetTestServerURL()

// orgCounter drives freshOrgID — a monotonically increasing counter.
// No UUID library needed; the value is unique within a test run.
var orgCounter atomic.Int64

// freshOrgID returns a unique org ID for each call.
// Tests use this instead of a shared constant so they never share DB state
// and never need per-test cleanup.
func freshOrgID() string {
	return fmt.Sprintf("test-ce-%d", orgCounter.Add(1))
}

// ptr converts a string literal to *string, used when building request bodies.
func ptr(s string) *string { return &s }

// =============================================================================
// Suite
// =============================================================================

// ElementAPITestSuite is the testify suite for all consent element integration tests.
type ElementAPITestSuite struct {
	suite.Suite
}

func TestElementAPITestSuite(t *testing.T) {
	suite.Run(t, new(ElementAPITestSuite))
}

func (ts *ElementAPITestSuite) SetupSuite() {
	ts.T().Log("=== ConsentElement Integration Test Suite Starting ===")
}

// =============================================================================
// Core HTTP helper
// =============================================================================

// doRequest executes an HTTP request and returns (statusCode, responseBody).
//
//   - orgID: written as the org-id header; pass "" to omit it entirely
//     (use this for missing-header error-case tests).
//   - body: nil for GET/DELETE; a struct (JSON-marshalled) or a raw string for POST/PUT.
func (ts *ElementAPITestSuite) doRequest(method, path, orgID string, body any) (int, []byte) {
	var rawBody []byte
	if body != nil { //nolint:nestif
		if s, ok := body.(string); ok {
			rawBody = []byte(s)
		} else {
			var err error
			rawBody, err = json.Marshal(body)
			ts.Require().NoError(err, "marshal request body")
		}
	}

	req, err := http.NewRequest(method, serverURL+path, bytes.NewReader(rawBody))
	ts.Require().NoError(err)

	if orgID != "" {
		req.Header.Set(testutils.HeaderOrgID, orgID)
	}
	if len(rawBody) > 0 {
		req.Header.Set(testutils.HeaderContentType, "application/json")
	}

	resp, err := testutils.GetHTTPClient().Do(req)
	ts.Require().NoError(err)
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	ts.Require().NoError(err)
	return resp.StatusCode, respBody
}

// =============================================================================
// Typed endpoint helpers
//
// Each helper returns (httpStatus, parsedResponse).
// The parsed response is nil when the status code does not match the expected
// success code — use doRequest directly to access the raw body in those cases.
// =============================================================================

func (ts *ElementAPITestSuite) doBatchCreate(orgID string, elements []CreateElementRequest) (int, *BatchCreateResponse) {
	status, body := ts.doRequest(http.MethodPost, "/api/v1/consent-elements", orgID, elements)
	if status != http.StatusOK {
		return status, nil
	}
	var resp BatchCreateResponse
	ts.Require().NoError(json.Unmarshal(body, &resp), "unmarshal BatchCreateResponse: %s", body)
	return status, &resp
}

func (ts *ElementAPITestSuite) doGetElement(orgID, elementID string) (int, *ElementResponse) {
	status, body := ts.doRequest(http.MethodGet, "/api/v1/consent-elements/"+elementID, orgID, nil)
	if status != http.StatusOK {
		return status, nil
	}
	var resp ElementResponse
	ts.Require().NoError(json.Unmarshal(body, &resp), "unmarshal ElementResponse")
	return status, &resp
}

func (ts *ElementAPITestSuite) doListElements(orgID string, params url.Values) (int, *ElementListResponse) {
	path := "/api/v1/consent-elements"
	if len(params) > 0 {
		path += "?" + params.Encode()
	}
	status, body := ts.doRequest(http.MethodGet, path, orgID, nil)
	if status != http.StatusOK {
		return status, nil
	}
	var resp ElementListResponse
	ts.Require().NoError(json.Unmarshal(body, &resp), "unmarshal ElementListResponse")
	return status, &resp
}

func (ts *ElementAPITestSuite) doListVersions(orgID, elementID string) (int, *ElementVersionListResponse) {
	path := fmt.Sprintf("/api/v1/consent-elements/%s/versions", elementID)
	status, body := ts.doRequest(http.MethodGet, path, orgID, nil)
	if status != http.StatusOK {
		return status, nil
	}
	var resp ElementVersionListResponse
	ts.Require().NoError(json.Unmarshal(body, &resp), "unmarshal ElementVersionListResponse")
	return status, &resp
}

func (ts *ElementAPITestSuite) doCreateVersion(orgID, elementID string, req CreateElementVersionRequest) (int, *ElementResponse) {
	path := fmt.Sprintf("/api/v1/consent-elements/%s/versions", elementID)
	status, body := ts.doRequest(http.MethodPost, path, orgID, req)
	if status != http.StatusCreated {
		return status, nil
	}
	var resp ElementResponse
	ts.Require().NoError(json.Unmarshal(body, &resp), "unmarshal ElementResponse (createVersion)")
	return status, &resp
}

func (ts *ElementAPITestSuite) doGetVersion(orgID, elementID, version string) (int, *ElementResponse) {
	path := fmt.Sprintf("/api/v1/consent-elements/%s/versions/%s", elementID, version)
	status, body := ts.doRequest(http.MethodGet, path, orgID, nil)
	if status != http.StatusOK {
		return status, nil
	}
	var resp ElementResponse
	ts.Require().NoError(json.Unmarshal(body, &resp), "unmarshal ElementResponse (getVersion)")
	return status, &resp
}

func (ts *ElementAPITestSuite) doDeleteVersion(orgID, elementID, version string) int {
	path := fmt.Sprintf("/api/v1/consent-elements/%s/versions/%s", elementID, version)
	status, _ := ts.doRequest(http.MethodDelete, path, orgID, nil)
	return status
}

// =============================================================================
// Must-helpers
//
// These are for test setup steps, not the operation under test.
// They call Require internally so the test stops immediately if setup fails,
// keeping failure messages focused on the actual assertion being tested.
// =============================================================================

// mustCreateElement creates a single element (name + type) and returns its elementId.
func (ts *ElementAPITestSuite) mustCreateElement(orgID, name, elemType string) string {
	return ts.mustCreateElementWith(orgID, CreateElementRequest{Name: name, Type: elemType})
}

// mustCreateElementWith creates a single element from a full request and returns its elementId.
func (ts *ElementAPITestSuite) mustCreateElementWith(orgID string, req CreateElementRequest) string {
	status, resp := ts.doBatchCreate(orgID, []CreateElementRequest{req})
	ts.Require().Equal(http.StatusOK, status, "mustCreateElement: unexpected HTTP status")
	ts.Require().NotNil(resp)
	ts.Require().Len(resp.Results, 1)
	ts.Require().Equal("SUCCESS", resp.Results[0].Status,
		"mustCreateElement: element creation FAILED — error: %v", resp.Results[0].Error)
	ts.Require().NotNil(resp.Results[0].Element)
	return resp.Results[0].Element.ElementID
}

// mustCreateVersion creates a new version on an existing element and returns the response.
func (ts *ElementAPITestSuite) mustCreateVersion(orgID, elementID string, req CreateElementVersionRequest) *ElementResponse {
	status, resp := ts.doCreateVersion(orgID, elementID, req)
	ts.Require().Equal(http.StatusCreated, status, "mustCreateVersion: unexpected HTTP status")
	ts.Require().NotNil(resp)
	return resp
}

// mustDeleteVersion deletes a specific version and fails if it doesn't return 204.
func (ts *ElementAPITestSuite) mustDeleteVersion(orgID, elementID, version string) {
	status := ts.doDeleteVersion(orgID, elementID, version)
	ts.Require().Equal(http.StatusNoContent, status,
		"mustDeleteVersion: unexpected HTTP status for element=%s version=%s", elementID, version)
}

// =============================================================================
// Assertion helpers
// =============================================================================

// assertAPIError parses body as an ErrorResponse, asserts the error code, and
// returns the parsed struct so callers can make additional assertions.
func (ts *ElementAPITestSuite) assertAPIError(body []byte, wantCode string) ErrorResponse {
	var errResp ErrorResponse
	ts.Require().NoError(json.Unmarshal(body, &errResp),
		"body is not a valid ErrorResponse: %s", string(body))
	ts.Require().Equal(wantCode, errResp.Code, "unexpected error code")
	ts.Require().NotEmpty(errResp.Message, "error response must have a non-empty message")
	return errResp
}

// assertElementResponse validates the fields that the swagger spec mandates are
// always present on an ElementResponse, regardless of which endpoint returned it.
// Call this from checkResult closures after asserting operation-specific fields.
func (ts *ElementAPITestSuite) assertElementResponse(e *ElementResponse, wantName, wantType string) {
	ts.Require().NotNil(e)
	ts.Require().NotEmpty(e.ElementID, "elementId must not be empty")
	ts.Require().NotEmpty(e.Namespace, "namespace must not be empty (defaults to 'default')")
	ts.Require().NotEmpty(e.Version, "version must not be empty (expected 'v1', 'v2', …)")
	ts.Require().Greater(e.CreatedTime, int64(0), "createdTime must be a positive Unix-ms timestamp")
	ts.Equal(wantName, e.Name, "name mismatch")
	ts.Equal(wantType, e.Type, "type mismatch")
}

// assertBatchSuccess asserts that a batch result item is SUCCESS and validates the
// swagger contract (all required fields present).
func (ts *ElementAPITestSuite) assertBatchSuccess(item BatchResultItem, wantName, wantType string) {
	ts.Require().Equal("SUCCESS", item.Status,
		"expected SUCCESS but got FAILED — error: %v", item.Error)
	ts.Require().NotNil(item.Element, "SUCCESS result must have an element")
	ts.Require().Nil(item.Error, "SUCCESS result must not have an error field")
	ts.assertElementResponse(item.Element, wantName, wantType)
}

// assertBatchFailed asserts that a batch result item is FAILED and that the error
// description contains wantDescContains (case-insensitive substring match).
// Pass "" to skip the description content check.
func (ts *ElementAPITestSuite) assertBatchFailed(item BatchResultItem, wantDescContains string) {
	ts.Require().Equal("FAILED", item.Status, "expected FAILED but got SUCCESS")
	ts.Require().Nil(item.Element, "FAILED result must not have an element")
	ts.Require().NotNil(item.Error, "FAILED result must have an error description")
	if wantDescContains != "" {
		ts.Require().Contains(
			strings.ToLower(*item.Error),
			strings.ToLower(wantDescContains),
			"error description mismatch",
		)
	}
}
