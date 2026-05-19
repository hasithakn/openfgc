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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/wso2/openfgc/internal/consentelement/model"
	"github.com/wso2/openfgc/internal/system/constants"
)

const (
	testOrgID     = "test-org-123"
	testElementID = "elem-123"
)

func stringPtr(s string) *string { return &s }

// --- createElements ---

func TestCreateElement_Success(t *testing.T) {
	mockService := NewMockConsentElementService(t)

	requests := []model.ConsentElementCreateRequest{
		{Name: "email", Type: "basic"},
	}
	result := &model.BulkCreateResponse{
		Results: []model.BulkCreateResultItem{
			{Status: "SUCCESS", Element: &model.ElementVersion{ID: testElementID, Name: "email", Type: "basic", Version: 1}},
		},
	}
	mockService.On("CreateElementsInBatch", mock.Anything, requests, testOrgID).Return(result, nil)

	body, _ := json.Marshal(requests)
	req := httptest.NewRequest(http.MethodPost, "/consent-elements", bytes.NewBuffer(body))
	req.Header.Set(constants.HeaderOrgID, testOrgID)
	rr := httptest.NewRecorder()

	newConsentElementHandler(mockService).createElements(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	var resp model.BulkCreateResponse
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	require.Len(t, resp.Results, 1)
	require.Equal(t, "SUCCESS", resp.Results[0].Status)
}

func TestCreateElement_MissingOrgID(t *testing.T) {
	mockService := NewMockConsentElementService(t)
	body, _ := json.Marshal([]model.ConsentElementCreateRequest{{Name: "x", Type: "basic"}})
	req := httptest.NewRequest(http.MethodPost, "/consent-elements", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	newConsentElementHandler(mockService).createElements(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)
	mockService.AssertNotCalled(t, "CreateElementsInBatch")
}

func TestCreateElement_InvalidJSON(t *testing.T) {
	mockService := NewMockConsentElementService(t)
	req := httptest.NewRequest(http.MethodPost, "/consent-elements", bytes.NewBufferString("{bad"))
	req.Header.Set(constants.HeaderOrgID, testOrgID)
	rr := httptest.NewRecorder()

	newConsentElementHandler(mockService).createElements(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestCreateElement_EmptyArray(t *testing.T) {
	mockService := NewMockConsentElementService(t)
	body, _ := json.Marshal([]model.ConsentElementCreateRequest{})
	req := httptest.NewRequest(http.MethodPost, "/consent-elements", bytes.NewBuffer(body))
	req.Header.Set(constants.HeaderOrgID, testOrgID)
	rr := httptest.NewRecorder()

	newConsentElementHandler(mockService).createElements(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestCreateElement_ServiceError(t *testing.T) {
	mockService := NewMockConsentElementService(t)
	requests := []model.ConsentElementCreateRequest{{Name: "email", Type: "basic"}}
	mockService.On("CreateElementsInBatch", mock.Anything, requests, testOrgID).Return(nil, &ErrorCreateElement)

	body, _ := json.Marshal(requests)
	req := httptest.NewRequest(http.MethodPost, "/consent-elements", bytes.NewBuffer(body))
	req.Header.Set(constants.HeaderOrgID, testOrgID)
	rr := httptest.NewRecorder()

	newConsentElementHandler(mockService).createElements(rr, req)

	require.Equal(t, http.StatusInternalServerError, rr.Code)
}

// --- getElement ---

func TestGetElement_Success(t *testing.T) {
	mockService := NewMockConsentElementService(t)

	expected := &model.ElementVersion{ID: testElementID, Name: "email", Type: "basic", Version: 1}
	mockService.On("GetElement", mock.Anything, testElementID, testOrgID).Return(expected, nil)

	req := httptest.NewRequest(http.MethodGet, "/consent-elements/"+testElementID, nil)
	req.SetPathValue("elementId", testElementID)
	req.Header.Set(constants.HeaderOrgID, testOrgID)
	rr := httptest.NewRecorder()

	newConsentElementHandler(mockService).getElement(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	var resp model.ElementVersion
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	require.Equal(t, testElementID, resp.ID)
}

func TestGetElement_MissingOrgID(t *testing.T) {
	mockService := NewMockConsentElementService(t)
	req := httptest.NewRequest(http.MethodGet, "/consent-elements/"+testElementID, nil)
	req.SetPathValue("elementId", testElementID)
	rr := httptest.NewRecorder()

	newConsentElementHandler(mockService).getElement(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestGetElement_NotFound(t *testing.T) {
	mockService := NewMockConsentElementService(t)
	mockService.On("GetElement", mock.Anything, testElementID, testOrgID).Return(nil, &ErrorElementNotFound)

	req := httptest.NewRequest(http.MethodGet, "/consent-elements/"+testElementID, nil)
	req.SetPathValue("elementId", testElementID)
	req.Header.Set(constants.HeaderOrgID, testOrgID)
	rr := httptest.NewRecorder()

	newConsentElementHandler(mockService).getElement(rr, req)

	require.Equal(t, http.StatusNotFound, rr.Code)
}

// --- listElements ---

func TestListElements_Success(t *testing.T) {
	mockService := NewMockConsentElementService(t)

	result := &model.ListResponse{
		Elements: []model.ElementVersion{
			{ID: "e1", Name: "email", Type: "basic", Version: 1},
			{ID: "e2", Name: "age", Type: "json", Version: 1},
		},
		Total: 2,
	}
	mockService.On("ListElements", mock.Anything, testOrgID, model.ElementListFilters{Limit: 100}).Return(result, nil)

	req := httptest.NewRequest(http.MethodGet, "/consent-elements", nil)
	req.Header.Set(constants.HeaderOrgID, testOrgID)
	rr := httptest.NewRecorder()

	newConsentElementHandler(mockService).listElements(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	var resp model.ListResponse
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	require.Len(t, resp.Elements, 2)
	require.Equal(t, 2, resp.Total)
}

func TestListElements_MissingOrgID(t *testing.T) {
	mockService := NewMockConsentElementService(t)
	req := httptest.NewRequest(http.MethodGet, "/consent-elements", nil)
	rr := httptest.NewRecorder()

	newConsentElementHandler(mockService).listElements(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestListElements_ServiceError(t *testing.T) {
	mockService := NewMockConsentElementService(t)
	mockService.On("ListElements", mock.Anything, testOrgID, model.ElementListFilters{Limit: 100}).Return(nil, &ErrorReadElement)

	req := httptest.NewRequest(http.MethodGet, "/consent-elements", nil)
	req.Header.Set(constants.HeaderOrgID, testOrgID)
	rr := httptest.NewRecorder()

	newConsentElementHandler(mockService).listElements(rr, req)

	require.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestListElements_WithFilters(t *testing.T) {
	mockService := NewMockConsentElementService(t)

	v := 2
	expected := model.ElementListFilters{Name: "em", Namespace: "ns1", Type: "basic", Version: &v, Details: true, Limit: 10, Offset: 5}
	mockService.On("ListElements", mock.Anything, testOrgID, expected).Return(&model.ListResponse{}, nil)

	req := httptest.NewRequest(http.MethodGet, "/consent-elements?name=em&namespace=ns1&type=basic&version=2&details=true&limit=10&offset=5", nil)
	req.Header.Set(constants.HeaderOrgID, testOrgID)
	rr := httptest.NewRecorder()

	newConsentElementHandler(mockService).listElements(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	mockService.AssertExpectations(t)
}

// --- listElementVersions ---

func TestListElementVersions_Success(t *testing.T) {
	mockService := NewMockConsentElementService(t)

	result := &model.VersionListResponse{
		ElementID: testElementID,
		Versions: []model.ElementVersion{
			{ID: testElementID, Version: 1},
			{ID: testElementID, Version: 2},
		},
	}
	mockService.On("ListElementVersions", mock.Anything, testElementID, testOrgID).Return(result, nil)

	req := httptest.NewRequest(http.MethodGet, "/consent-elements/"+testElementID+"/versions", nil)
	req.SetPathValue("elementId", testElementID)
	req.Header.Set(constants.HeaderOrgID, testOrgID)
	rr := httptest.NewRecorder()

	newConsentElementHandler(mockService).listElementVersions(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	var resp model.VersionListResponse
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	require.Equal(t, testElementID, resp.ElementID)
	require.Len(t, resp.Versions, 2)
}

func TestListElementVersions_NotFound(t *testing.T) {
	mockService := NewMockConsentElementService(t)
	mockService.On("ListElementVersions", mock.Anything, testElementID, testOrgID).Return(nil, &ErrorElementNotFound)

	req := httptest.NewRequest(http.MethodGet, "/consent-elements/"+testElementID+"/versions", nil)
	req.SetPathValue("elementId", testElementID)
	req.Header.Set(constants.HeaderOrgID, testOrgID)
	rr := httptest.NewRecorder()

	newConsentElementHandler(mockService).listElementVersions(rr, req)

	require.Equal(t, http.StatusNotFound, rr.Code)
}

// --- createElementVersion ---

func TestCreateElementVersion_Success(t *testing.T) {
	mockService := NewMockConsentElementService(t)

	vreq := model.ElementVersionCreateRequest{DisplayName: stringPtr("v2")}
	created := &model.ElementVersion{ID: testElementID, Version: 2, DisplayName: stringPtr("v2")}
	mockService.On("CreateElementVersion", mock.Anything, testElementID, vreq, testOrgID).Return(created, nil)

	body, _ := json.Marshal(vreq)
	req := httptest.NewRequest(http.MethodPost, "/consent-elements/"+testElementID+"/versions", bytes.NewBuffer(body))
	req.SetPathValue("elementId", testElementID)
	req.Header.Set(constants.HeaderOrgID, testOrgID)
	rr := httptest.NewRecorder()

	newConsentElementHandler(mockService).createElementVersion(rr, req)

	require.Equal(t, http.StatusCreated, rr.Code)
	var resp model.ElementVersion
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	require.Equal(t, 2, resp.Version)
}

func TestCreateElementVersion_InvalidJSON(t *testing.T) {
	mockService := NewMockConsentElementService(t)
	req := httptest.NewRequest(http.MethodPost, "/consent-elements/"+testElementID+"/versions", bytes.NewBufferString("{bad"))
	req.SetPathValue("elementId", testElementID)
	req.Header.Set(constants.HeaderOrgID, testOrgID)
	rr := httptest.NewRecorder()

	newConsentElementHandler(mockService).createElementVersion(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestCreateElementVersion_ServiceError(t *testing.T) {
	mockService := NewMockConsentElementService(t)
	vreq := model.ElementVersionCreateRequest{}
	mockService.On("CreateElementVersion", mock.Anything, testElementID, vreq, testOrgID).Return(nil, &ErrorCreateElement)

	body, _ := json.Marshal(vreq)
	req := httptest.NewRequest(http.MethodPost, "/consent-elements/"+testElementID+"/versions", bytes.NewBuffer(body))
	req.SetPathValue("elementId", testElementID)
	req.Header.Set(constants.HeaderOrgID, testOrgID)
	rr := httptest.NewRecorder()

	newConsentElementHandler(mockService).createElementVersion(rr, req)

	require.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestCreateElementVersion_MissingOrgID(t *testing.T) {
	mockService := NewMockConsentElementService(t)
	body, _ := json.Marshal(model.ElementVersionCreateRequest{})
	req := httptest.NewRequest(http.MethodPost, "/consent-elements/"+testElementID+"/versions", bytes.NewBuffer(body))
	req.SetPathValue("elementId", testElementID)
	rr := httptest.NewRecorder()

	newConsentElementHandler(mockService).createElementVersion(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)
}

// --- getElementVersion ---

func TestGetElementVersion_Success(t *testing.T) {
	mockService := NewMockConsentElementService(t)

	expected := &model.ElementVersion{ID: testElementID, Version: 2}
	mockService.On("GetElementVersion", mock.Anything, testElementID, 2, testOrgID).Return(expected, nil)

	req := httptest.NewRequest(http.MethodGet, "/consent-elements/"+testElementID+"/versions/2", nil)
	req.SetPathValue("elementId", testElementID)
	req.SetPathValue("version", "2")
	req.Header.Set(constants.HeaderOrgID, testOrgID)
	rr := httptest.NewRecorder()

	newConsentElementHandler(mockService).getElementVersion(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	var resp model.ElementVersion
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	require.Equal(t, 2, resp.Version)
}

func TestGetElementVersion_InvalidVersion(t *testing.T) {
	mockService := NewMockConsentElementService(t)
	req := httptest.NewRequest(http.MethodGet, "/consent-elements/"+testElementID+"/versions/abc", nil)
	req.SetPathValue("elementId", testElementID)
	req.SetPathValue("version", "abc")
	req.Header.Set(constants.HeaderOrgID, testOrgID)
	rr := httptest.NewRecorder()

	newConsentElementHandler(mockService).getElementVersion(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestGetElementVersion_NotFound(t *testing.T) {
	mockService := NewMockConsentElementService(t)
	mockService.On("GetElementVersion", mock.Anything, testElementID, 99, testOrgID).Return(nil, &ErrorElementNotFound)

	req := httptest.NewRequest(http.MethodGet, "/consent-elements/"+testElementID+"/versions/99", nil)
	req.SetPathValue("elementId", testElementID)
	req.SetPathValue("version", "99")
	req.Header.Set(constants.HeaderOrgID, testOrgID)
	rr := httptest.NewRecorder()

	newConsentElementHandler(mockService).getElementVersion(rr, req)

	require.Equal(t, http.StatusNotFound, rr.Code)
}

// --- deleteElementVersion ---

func TestDeleteElementVersion_Success(t *testing.T) {
	mockService := NewMockConsentElementService(t)
	mockService.On("DeleteElementVersion", mock.Anything, testElementID, 1, testOrgID).Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/consent-elements/"+testElementID+"/versions/1", nil)
	req.SetPathValue("elementId", testElementID)
	req.SetPathValue("version", "1")
	req.Header.Set(constants.HeaderOrgID, testOrgID)
	rr := httptest.NewRecorder()

	newConsentElementHandler(mockService).deleteElementVersion(rr, req)

	require.Equal(t, http.StatusNoContent, rr.Code)
}

func TestDeleteElementVersion_MissingOrgID(t *testing.T) {
	mockService := NewMockConsentElementService(t)
	req := httptest.NewRequest(http.MethodDelete, "/consent-elements/"+testElementID+"/versions/1", nil)
	req.SetPathValue("elementId", testElementID)
	req.SetPathValue("version", "1")
	rr := httptest.NewRecorder()

	newConsentElementHandler(mockService).deleteElementVersion(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestDeleteElementVersion_InvalidVersion(t *testing.T) {
	mockService := NewMockConsentElementService(t)
	req := httptest.NewRequest(http.MethodDelete, "/consent-elements/"+testElementID+"/versions/0", nil)
	req.SetPathValue("elementId", testElementID)
	req.SetPathValue("version", "0")
	req.Header.Set(constants.HeaderOrgID, testOrgID)
	rr := httptest.NewRecorder()

	newConsentElementHandler(mockService).deleteElementVersion(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestDeleteElementVersion_ReferencedByPurpose(t *testing.T) {
	mockService := NewMockConsentElementService(t)
	mockService.On("DeleteElementVersion", mock.Anything, testElementID, 1, testOrgID).Return(&ErrorVersionReferencedByPurpose)

	req := httptest.NewRequest(http.MethodDelete, "/consent-elements/"+testElementID+"/versions/1", nil)
	req.SetPathValue("elementId", testElementID)
	req.SetPathValue("version", "1")
	req.Header.Set(constants.HeaderOrgID, testOrgID)
	rr := httptest.NewRecorder()

	newConsentElementHandler(mockService).deleteElementVersion(rr, req)

	require.Equal(t, http.StatusConflict, rr.Code)
}
