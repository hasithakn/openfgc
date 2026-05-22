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
	"net/http"
	"net/url"
)

// TestVersionLifecycle exercises the full version management flow for a single element
// in one sequential test. This is the kind of cross-endpoint contract that individual
// endpoint tests cannot cover.
//
// Flow: create (v1) → add version (v2) → list (both) → delete v1 → v1 gone, v2 intact
//
//	→ delete v2 (last version) → element gone
func (ts *ElementAPITestSuite) TestVersionLifecycle() {
	orgID := freshOrgID()

	// Step 1: Create element — v1 is created automatically
	elemID := ts.mustCreateElement(orgID, "lifecycle-elem", "basic")
	ts.T().Logf("Created element: %s", elemID)

	// Step 2: Verify v1 exists and is the latest
	statusGet, v1 := ts.doGetElement(orgID, elemID)
	ts.Require().Equal(http.StatusOK, statusGet)
	ts.Equal("v1", v1.Version)

	// Step 3: Create v2
	v2 := ts.mustCreateVersion(orgID, elemID, CreateElementVersionRequest{
		DisplayName: ptr("Version Two"),
	})
	ts.Equal("v2", v2.Version)

	// Step 4: GET now returns v2 (latest)
	_, latest := ts.doGetElement(orgID, elemID)
	ts.Equal("v2", latest.Version, "GET must return the latest version after v2 is created")

	// Step 5: List versions — both v1 and v2 present in ascending order
	_, versions := ts.doListVersions(orgID, elemID)
	ts.Require().Len(versions.Versions, 2)
	ts.Equal("v1", versions.Versions[0].Version)
	ts.Equal("v2", versions.Versions[1].Version)

	// Step 6: Delete v1
	ts.mustDeleteVersion(orgID, elemID, "v1")

	// Step 7: v1 is gone
	statusV1, _ := ts.doGetVersion(orgID, elemID, "v1")
	ts.Equal(http.StatusNotFound, statusV1, "v1 must not be accessible after deletion")

	// Step 8: v2 still intact and is still the latest
	statusV2, v2After := ts.doGetVersion(orgID, elemID, "v2")
	ts.Equal(http.StatusOK, statusV2)
	ts.Equal("v2", v2After.Version)
	ts.Equal("Version Two", *v2After.DisplayName)

	// Step 9: List versions — only v2 remains
	_, versionsAfter := ts.doListVersions(orgID, elemID)
	ts.Require().Len(versionsAfter.Versions, 1)
	ts.Equal("v2", versionsAfter.Versions[0].Version)

	// Step 10: Delete v2 — this is the last version, so the element itself is also removed
	ts.mustDeleteVersion(orgID, elemID, "v2")

	// Step 11: Element is completely gone
	statusElem, _ := ts.doGetElement(orgID, elemID)
	ts.Equal(http.StatusNotFound, statusElem,
		"element must be gone after its last version is deleted")
}

// TestNamespaceScopedUniqueness verifies that element name uniqueness is scoped to
// (name + namespace + orgId). The same name in different namespaces creates two
// independent elements; the same name in the same namespace is rejected.
func (ts *ElementAPITestSuite) TestNamespaceScopedUniqueness() {
	orgID := freshOrgID()

	// Same name, different namespaces — both must succeed
	idHR := ts.mustCreateElementWith(orgID, CreateElementRequest{
		Name:      "employee-id",
		Type:      "basic",
		Namespace: "hr",
	})
	idFinance := ts.mustCreateElementWith(orgID, CreateElementRequest{
		Name:      "employee-id",
		Type:      "basic",
		Namespace: "finance",
	})
	ts.NotEqual(idHR, idFinance, "same name in different namespaces must create distinct elements")

	// Verify both exist independently
	_, hrElem := ts.doGetElement(orgID, idHR)
	_, finElem := ts.doGetElement(orgID, idFinance)
	ts.Equal("hr", hrElem.Namespace)
	ts.Equal("finance", finElem.Namespace)

	// Same name, same namespace (default) — second must fail
	status, resp := ts.doBatchCreate(orgID, []CreateElementRequest{
		{Name: "duplicate-name", Type: "basic"},
		{Name: "duplicate-name", Type: "basic"}, // same name, same default namespace
	})
	ts.Require().Equal(http.StatusOK, status)
	ts.assertBatchSuccess(resp.Results[0], "duplicate-name", "basic")
	ts.assertBatchFailed(resp.Results[1], "already exists")
}

// TestListReflectsLatestVersions verifies that the list endpoint always shows the
// latest version of each element, and that metadata (total, count) is accurate.
func (ts *ElementAPITestSuite) TestListReflectsLatestVersions() {
	orgID := freshOrgID()

	// Create two elements
	id1 := ts.mustCreateElement(orgID, "list-latest-a", "basic")
	ts.mustCreateElement(orgID, "list-latest-b", "basic")

	// Create v2 on the first element
	ts.mustCreateVersion(orgID, id1, CreateElementVersionRequest{
		DisplayName: ptr("A Version Two"),
	})

	// List — must return 2 items (one per element, each at their latest version)
	_, listResp := ts.doListElements(orgID, url.Values{})
	ts.Require().Equal(2, listResp.Metadata.Total,
		"list must return one item per element, not per version")

	// Find element A in the results and verify it shows v2
	var elemA *ElementResponse
	for i := range listResp.Data {
		if listResp.Data[i].Name == "list-latest-a" {
			elemA = &listResp.Data[i]
			break
		}
	}
	ts.Require().NotNil(elemA, "list-latest-a must be in the results")
	ts.Equal("v2", elemA.Version, "list must reflect the latest version for each element")
	ts.Require().NotNil(elemA.DisplayName)
	ts.Equal("A Version Two", *elemA.DisplayName)
}
