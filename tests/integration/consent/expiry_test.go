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
	"net/url"
	"time"
)

// TestConsentExpiry covers the auto-expiry behaviour triggered on read operations.
//
// The service checks whether a consent's expirationTime has passed whenever the
// consent is loaded (GET, validate, search). If so, it atomically transitions
// the consent and all its auth resources to their respective expired statuses
// before returning the response.
//
// Rules under test:
//   - Creating a consent with a past expirationTime immediately marks it EXPIRED.
//   - GET on an EXPIRED consent returns status EXPIRED.
//   - Auth resources of an expired consent are marked SYS_EXPIRED.
//   - Validate on an EXPIRED consent returns isValid=false with errorCode=401.
//   - Creating a consent with a future expirationTime keeps it CREATED/ACTIVE.
//   - List (/consents) with consentStatuses=EXPIRED returns expired consents.
func (ts *ConsentAPITestSuite) TestConsentExpiry() {
	// pastMs returns a Unix millisecond timestamp in the past.
	pastMs := func(d time.Duration) int64 {
		return time.Now().Add(-d).UnixMilli()
	}
	// futureMs returns a Unix millisecond timestamp in the future.
	futureMs := func(d time.Duration) int64 {
		return time.Now().Add(d).UnixMilli()
	}

	// -----------------------------------------------------------------------
	// Create with past expirationTime → immediately EXPIRED
	// -----------------------------------------------------------------------
	ts.Run("create with past expirationTime → create response has status EXPIRED", func() {
		orgID := freshOrgID()
		past := pastMs(1 * time.Minute)
		c := ts.mustCreateConsent(orgID, "grp-exp-create", ConsentCreateRequest{
			Type:           "accounts",
			ExpirationTime: &past,
			Authorizations: []AuthorizationRequest{{Status: "APPROVED"}},
		})
		ts.Equal("EXPIRED", c.Status,
			"consent with past expirationTime must be EXPIRED immediately")
	})

	// -----------------------------------------------------------------------
	// GET on an expired consent
	// -----------------------------------------------------------------------
	ts.Run("GET expired consent returns status EXPIRED", func() {
		orgID := freshOrgID()
		past := pastMs(1 * time.Minute)
		created := ts.mustCreateConsent(orgID, "grp-exp-get", ConsentCreateRequest{
			Type:           "accounts",
			ExpirationTime: &past,
			Authorizations: []AuthorizationRequest{{Status: "APPROVED"}},
		})

		status, got := ts.doGetConsent(orgID, created.ID)
		ts.Require().Equal(http.StatusOK, status)
		ts.Require().NotNil(got)
		ts.Equal("EXPIRED", got.Status, "GET on expired consent must return EXPIRED status")
	})

	ts.Run("GET expired consent — expirationTime is present in response", func() {
		orgID := freshOrgID()
		past := pastMs(30 * time.Second)
		created := ts.mustCreateConsent(orgID, "grp-exp-etime", ConsentCreateRequest{
			Type:           "accounts",
			ExpirationTime: &past,
		})

		_, got := ts.doGetConsent(orgID, created.ID)
		ts.Require().NotNil(got)
		ts.Require().NotNil(got.ExpirationTime, "expirationTime must be returned in response")
		ts.Equal(past, *got.ExpirationTime)
	})

	// -----------------------------------------------------------------------
	// Auth resources get SYS_EXPIRED when consent expires
	// -----------------------------------------------------------------------
	ts.Run("auth resources of expired consent are marked SYS_EXPIRED", func() {
		orgID := freshOrgID()
		past := pastMs(1 * time.Minute)
		created := ts.mustCreateConsent(orgID, "grp-exp-auth", ConsentCreateRequest{
			Type:           "accounts",
			ExpirationTime: &past,
			Authorizations: []AuthorizationRequest{
				{Type: "accounts", Status: "APPROVED"},
				{Type: "accounts", Status: "APPROVED"},
			},
		})

		// GET triggers expiry check; auth resources should be SYS_EXPIRED
		_, got := ts.doGetConsent(orgID, created.ID)
		ts.Require().NotNil(got)
		ts.Equal("EXPIRED", got.Status)
		ts.Require().Len(got.Authorizations, 2,
			"both auth resources must still be present after expiry")
		for _, auth := range got.Authorizations {
			ts.Equal("SYS_EXPIRED", auth.Status,
				"auth resource must be SYS_EXPIRED after consent expires")
		}
	})

	// -----------------------------------------------------------------------
	// Validate on an expired consent
	// -----------------------------------------------------------------------
	ts.Run("validate expired consent → isValid=false, errorCode=401, status EXPIRED in consentInfo", func() {
		orgID := freshOrgID()
		past := pastMs(1 * time.Minute)
		created := ts.mustCreateConsent(orgID, "grp-exp-val", ConsentCreateRequest{
			Type:           "accounts",
			ExpirationTime: &past,
			Authorizations: []AuthorizationRequest{{Status: "APPROVED"}},
		})

		_, body := ts.doValidateConsent(orgID, ConsentValidateRequest{ConsentID: created.ID})
		var resp ConsentValidateResponse
		ts.Require().NoError(json.Unmarshal(body, &resp))
		ts.False(resp.IsValid)
		ts.Equal(401, resp.ErrorCode)
		ts.Equal("invalid_consent_status", resp.ErrorMessage)
		ts.Require().NotNil(resp.ConsentInfo)
		ts.Equal("EXPIRED", resp.ConsentInfo.Status)
	})

	// -----------------------------------------------------------------------
	// Future expirationTime does not trigger expiry
	// -----------------------------------------------------------------------
	ts.Run("future expirationTime — consent status is not EXPIRED", func() {
		orgID := freshOrgID()
		future := futureMs(24 * time.Hour)
		c := ts.mustCreateConsent(orgID, "grp-exp-future", ConsentCreateRequest{
			Type:           "accounts",
			ExpirationTime: &future,
			Authorizations: []AuthorizationRequest{{Status: "APPROVED"}},
		})
		ts.NotEqual("EXPIRED", c.Status,
			"consent with future expirationTime must not be EXPIRED")
		ts.Equal("ACTIVE", c.Status)
	})

	ts.Run("no expirationTime — consent never expires", func() {
		orgID := freshOrgID()
		c := ts.mustCreateConsent(orgID, "grp-no-exp", ConsentCreateRequest{
			Type:           "accounts",
			Authorizations: []AuthorizationRequest{{Status: "APPROVED"}},
		})
		ts.Equal("ACTIVE", c.Status)
		ts.Nil(c.ExpirationTime, "expirationTime must be absent when not set")
	})

	// -----------------------------------------------------------------------
	// Expiry status visible in list / search
	// -----------------------------------------------------------------------
	ts.Run("list with consentStatuses=EXPIRED returns expired consents", func() {
		orgID := freshOrgID()
		past := pastMs(1 * time.Minute)
		ts.mustCreateConsent(orgID, "grp-list-exp", ConsentCreateRequest{
			Type:           "accounts",
			ExpirationTime: &past,
			Authorizations: []AuthorizationRequest{{Status: "APPROVED"}},
		})
		// Non-expired consent — must NOT appear in EXPIRED filter
		ts.mustCreateConsent(orgID, "grp-list-active", ConsentCreateRequest{
			Type:           "accounts",
			Authorizations: []AuthorizationRequest{{Status: "APPROVED"}},
		})

		status, resp := ts.doListConsents(orgID, url.Values{"consentStatuses": {"EXPIRED"}})
		ts.Require().Equal(http.StatusOK, status)
		ts.Require().NotNil(resp)
		ts.Equal(1, resp.Metadata.Total)
		ts.Require().Len(resp.Data, 1)
		ts.Equal("EXPIRED", resp.Data[0].Status)
	})

	ts.Run("list without status filter includes expired consents alongside active ones", func() {
		orgID := freshOrgID()
		past := pastMs(1 * time.Minute)
		ts.mustCreateConsent(orgID, "grp-all-exp", ConsentCreateRequest{
			Type:           "accounts",
			ExpirationTime: &past,
		})
		ts.mustCreateConsent(orgID, "grp-all-act", ConsentCreateRequest{
			Type: "accounts",
			Authorizations: []AuthorizationRequest{{Status: "APPROVED"}},
		})

		status, resp := ts.doListConsents(orgID, nil)
		ts.Require().Equal(http.StatusOK, status)
		ts.Require().NotNil(resp)
		ts.Equal(2, resp.Metadata.Total)
	})
}
