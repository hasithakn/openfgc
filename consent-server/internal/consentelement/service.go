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

// Package consentelement provides consent element management functionality.
package consentelement

import (
	"context"
	"fmt"
	"time"

	"github.com/wso2/openfgc/internal/consentelement/model"
	"github.com/wso2/openfgc/internal/consentelement/validators"
	dbmodel "github.com/wso2/openfgc/internal/system/database/model"
	"github.com/wso2/openfgc/internal/system/error/serviceerror"
	"github.com/wso2/openfgc/internal/system/stores"
	"github.com/wso2/openfgc/internal/system/utils"
)

// ConsentElementService defines the exported service interface.
type ConsentElementService interface {
	// CreateElementsInBatch creates elements with partial success — failures do not block other items.
	CreateElementsInBatch(ctx context.Context, requests []model.ConsentElementCreateRequest, orgID string) (*model.BulkCreateResponse, *serviceerror.ServiceError)

	// GetElement returns the latest version of an element.
	GetElement(ctx context.Context, elementID, orgID string) (*model.ElementVersion, *serviceerror.ServiceError)

	// GetElementVersion returns a specific version of an element.
	GetElementVersion(ctx context.Context, elementID string, version int, orgID string) (*model.ElementVersion, *serviceerror.ServiceError)

	// ListElementVersions returns all versions of one element ordered ascending.
	ListElementVersions(ctx context.Context, elementID, orgID string) (*model.VersionListResponse, *serviceerror.ServiceError)

	// ListElements returns paginated latest versions matching the given filters.
	ListElements(ctx context.Context, orgID string, filters model.ElementListFilters) (*model.ListResponse, *serviceerror.ServiceError)

	// CreateElementVersion appends a new version to an existing element.
	CreateElementVersion(ctx context.Context, elementID string, req model.ElementVersionCreateRequest, orgID string) (*model.ElementVersion, *serviceerror.ServiceError)

	// DeleteElementVersion deletes a specific version. Returns 409 if referenced by a purpose.
	// Deleting the last version also deletes the element.
	DeleteElementVersion(ctx context.Context, elementID string, version int, orgID string) *serviceerror.ServiceError
}

// consentElementService implements the ConsentElementService interface.
type consentElementService struct {
	stores *stores.StoreRegistry
}

// newConsentElementService creates a new consent element service.
func newConsentElementService(registry *stores.StoreRegistry) ConsentElementService {
	return &consentElementService{stores: registry}
}

// CreateElementsInBatch creates multiple elements. Each item is processed independently;
// per-item failures are collected and returned as FAILED results, not as a top-level error.
func (s *consentElementService) CreateElementsInBatch(ctx context.Context, requests []model.ConsentElementCreateRequest, orgID string) (*model.BulkCreateResponse, *serviceerror.ServiceError) {
	if len(requests) == 0 {
		return nil, &ErrorAtLeastOneElement
	}

	results := make([]model.BulkCreateResultItem, 0, len(requests))
	for _, req := range requests {
		elementVersion, svcErr := s.createSingleElement(ctx, req, orgID)
		if svcErr != nil {
			msg := svcErr.Description
			results = append(results, model.BulkCreateResultItem{Status: "FAILED", Error: &msg})
		} else {
			results = append(results, model.BulkCreateResultItem{Status: "SUCCESS", Element: elementVersion})
		}
	}

	return &model.BulkCreateResponse{Results: results}, nil
}

func (s *consentElementService) createSingleElement(ctx context.Context, req model.ConsentElementCreateRequest, orgID string) (*model.ElementVersion, *serviceerror.ServiceError) {
	if svcErr := validateCreateRequest(req); svcErr != nil {
		return nil, svcErr
	}

	if req.Namespace == "" {
		req.Namespace = model.DefaultNamespace
	}

	store := s.stores.ConsentElement
	existing, err := store.GetByNameAndNamespace(ctx, req.Name, req.Namespace, orgID)
	if err != nil {
		return nil, serviceerror.CustomServiceError(ErrorCreateElement, fmt.Sprintf("failed to check name existence: %v", err))
	}
	if existing != nil {
		return nil, serviceerror.CustomServiceError(ErrorElementNameExists,
			fmt.Sprintf("element with name '%s' and namespace '%s' already exists", req.Name, req.Namespace))
	}

	elementVersion := &model.ElementVersion{
		VersionID:   utils.GenerateUUID(),
		ID:          utils.GenerateUUID(),
		Name:        req.Name,
		Namespace:   req.Namespace,
		Type:        req.Type,
		Version:     1,
		DisplayName: req.DisplayName,
		Description: req.Description,
		Schema:      req.Schema,
		CreatedTime: time.Now().UnixMilli(),
		OrgID:       orgID,
		Properties:  req.Properties,
	}

	if err := s.stores.ExecuteTransaction([]func(tx dbmodel.TxInterface) error{
		func(tx dbmodel.TxInterface) error { return store.CreateVersion(tx, elementVersion) },
	}); err != nil {
		return nil, serviceerror.CustomServiceError(ErrorCreateElement, fmt.Sprintf("failed to create element: %v", err))
	}
	return elementVersion, nil
}

// GetElement returns the latest version of an element.
func (s *consentElementService) GetElement(ctx context.Context, elementID, orgID string) (*model.ElementVersion, *serviceerror.ServiceError) {
	elementVersion, err := s.stores.ConsentElement.GetLatestVersion(ctx, elementID, orgID)
	if err != nil {
		return nil, serviceerror.CustomServiceError(ErrorReadElement, fmt.Sprintf("failed to retrieve element: %v", err))
	}
	if elementVersion == nil {
		return nil, serviceerror.CustomServiceError(ErrorElementNotFound, fmt.Sprintf("element '%s' not found", elementID))
	}
	return elementVersion, nil
}

// GetElementVersion returns a specific version of an element.
func (s *consentElementService) GetElementVersion(ctx context.Context, elementID string, version int, orgID string) (*model.ElementVersion, *serviceerror.ServiceError) {
	elementVersion, err := s.stores.ConsentElement.GetVersion(ctx, elementID, version, orgID)
	if err != nil {
		return nil, serviceerror.CustomServiceError(ErrorReadElement, fmt.Sprintf("failed to retrieve element version: %v", err))
	}
	if elementVersion == nil {
		return nil, serviceerror.CustomServiceError(ErrorElementNotFound,
			fmt.Sprintf("element '%s' version %d not found", elementID, version))
	}
	return elementVersion, nil
}

// ListElementVersions returns all versions of one element ordered ascending.
func (s *consentElementService) ListElementVersions(ctx context.Context, elementID, orgID string) (*model.VersionListResponse, *serviceerror.ServiceError) {
	store := s.stores.ConsentElement
	exists, err := store.ElementExists(ctx, elementID, orgID)
	if err != nil {
		return nil, serviceerror.CustomServiceError(ErrorReadElement, fmt.Sprintf("failed to check element: %v", err))
	}
	if !exists {
		return nil, serviceerror.CustomServiceError(ErrorElementNotFound, fmt.Sprintf("element '%s' not found", elementID))
	}

	versions, err := store.ListVersions(ctx, elementID, orgID)
	if err != nil {
		return nil, serviceerror.CustomServiceError(ErrorReadElement, fmt.Sprintf("failed to list element versions: %v", err))
	}
	return &model.VersionListResponse{ElementID: elementID, Versions: versions}, nil
}

// ListElements returns paginated latest versions matching the given filters.
func (s *consentElementService) ListElements(ctx context.Context, orgID string, filters model.ElementListFilters) (*model.ListResponse, *serviceerror.ServiceError) {
	if filters.Limit <= 0 {
		filters.Limit = 100
	}
	if filters.Offset < 0 {
		filters.Offset = 0
	}

	versions, total, err := s.stores.ConsentElement.List(ctx, orgID, filters)
	if err != nil {
		return nil, serviceerror.CustomServiceError(ErrorReadElement, fmt.Sprintf("failed to list elements: %v", err))
	}
	return &model.ListResponse{Elements: versions, Total: total}, nil
}

// CreateElementVersion appends a new immutable version to an existing element.
// Name, Namespace, and Type are inherited from the element and cannot change.
func (s *consentElementService) CreateElementVersion(ctx context.Context, elementID string, req model.ElementVersionCreateRequest, orgID string) (*model.ElementVersion, *serviceerror.ServiceError) {
	store := s.stores.ConsentElement

	latest, err := store.GetLatestVersion(ctx, elementID, orgID)
	if err != nil {
		return nil, serviceerror.CustomServiceError(ErrorReadElement, fmt.Sprintf("failed to retrieve element: %v", err))
	}
	if latest == nil {
		return nil, serviceerror.CustomServiceError(ErrorElementNotFound, fmt.Sprintf("element '%s' not found", elementID))
	}

	elementVersion := &model.ElementVersion{
		VersionID:   utils.GenerateUUID(),
		ID:          elementID,
		Name:        latest.Name,
		Namespace:   latest.Namespace,
		Type:        latest.Type,
		Version:     latest.Version + 1,
		DisplayName: req.DisplayName,
		Description: req.Description,
		Schema:      req.Schema,
		CreatedTime: time.Now().UnixMilli(),
		OrgID:       orgID,
		Properties:  req.Properties,
	}

	if err := s.stores.ExecuteTransaction([]func(tx dbmodel.TxInterface) error{
		func(tx dbmodel.TxInterface) error { return store.CreateVersion(tx, elementVersion) },
	}); err != nil {
		return nil, serviceerror.CustomServiceError(ErrorCreateElement, fmt.Sprintf("failed to create element version: %v", err))
	}
	return elementVersion, nil
}

// DeleteElementVersion deletes a specific version. Returns 409 if referenced by a purpose.
// Deleting the last version also removes the element entity.
func (s *consentElementService) DeleteElementVersion(ctx context.Context, elementID string, version int, orgID string) *serviceerror.ServiceError {
	store := s.stores.ConsentElement

	elementVersion, err := store.GetVersion(ctx, elementID, version, orgID)
	if err != nil {
		return serviceerror.CustomServiceError(ErrorReadElement, fmt.Sprintf("failed to retrieve element version: %v", err))
	}
	if elementVersion == nil {
		return serviceerror.CustomServiceError(ErrorElementNotFound,
			fmt.Sprintf("element '%s' version %d not found", elementID, version))
	}

	referenced, err := store.IsVersionReferencedByPurpose(ctx, elementVersion.VersionID, orgID)
	if err != nil {
		return serviceerror.CustomServiceError(ErrorReadElement, fmt.Sprintf("failed to check version references: %v", err))
	}
	if referenced {
		return serviceerror.CustomServiceError(ErrorVersionReferencedByPurpose,
			fmt.Sprintf("element '%s' version %d is referenced by one or more purposes and cannot be deleted", elementID, version))
	}

	allVersions, err := store.ListVersions(ctx, elementID, orgID)
	if err != nil {
		return serviceerror.CustomServiceError(ErrorReadElement, fmt.Sprintf("failed to check versions: %v", err))
	}
	isLastVersion := len(allVersions) == 1

	txOps := []func(tx dbmodel.TxInterface) error{
		func(tx dbmodel.TxInterface) error { return store.DeleteVersion(tx, elementVersion.VersionID, orgID) },
	}
	if isLastVersion {
		txOps = append(txOps, func(tx dbmodel.TxInterface) error { return store.DeleteElement(tx, elementID, orgID) })
	}

	if err := s.stores.ExecuteTransaction(txOps); err != nil {
		return serviceerror.CustomServiceError(ErrorDeleteElement, fmt.Sprintf("failed to delete element version: %v", err))
	}
	return nil
}

// validateCreateRequest validates a single element create request.
func validateCreateRequest(req model.ConsentElementCreateRequest) *serviceerror.ServiceError {
	if req.Name == "" {
		return &ErrorElementNameRequired
	}
	if len(req.Name) > 255 {
		return &ErrorElementNameTooLong
	}
	if req.Description != nil && len(*req.Description) > 1024 {
		return &ErrorElementDescriptionTooLong
	}
	if req.Type == "" {
		return &ErrorElementTypeRequired
	}
	switch req.Type {
	case model.ElementTypeBasic, model.ElementTypeJSON, model.ElementTypeXML:
	default:
		return serviceerror.CustomServiceError(ErrorInvalidElementType, fmt.Sprintf("invalid element type: %s", req.Type))
	}
	if elementTypeDef, err := validators.GetTypeRegistry().Get(req.Type); err == nil {
		if verr := elementTypeDef.ValidateSchema(req.Schema); verr != nil {
			return serviceerror.CustomServiceError(ErrorValidateElement, verr.Message)
		}
		if errs := elementTypeDef.ValidateProperties(req.Properties); len(errs) > 0 {
			return serviceerror.CustomServiceError(ErrorValidateElement, errs[0].Message)
		}
	}
	return nil
}
