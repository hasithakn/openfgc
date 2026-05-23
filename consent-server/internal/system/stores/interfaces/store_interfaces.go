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

// Package interfaces defines the store interfaces for data operations.
package interfaces

import (
	"context"

	authResourceModel "github.com/wso2/openfgc/internal/authresource/model"
	consentModel "github.com/wso2/openfgc/internal/consent/model"
	consentElementModel "github.com/wso2/openfgc/internal/consentelement/model"
	purposeModel "github.com/wso2/openfgc/internal/consentpurpose/model"
	dbmodel "github.com/wso2/openfgc/internal/system/database/model"
)

// ConsentStore defines the interface for consent data operations
type ConsentStore interface {
	GetByID(ctx context.Context, consentID, orgID string) (*consentModel.Consent, error)
	Search(ctx context.Context, filters consentModel.ConsentSearchFilters) ([]consentModel.Consent, int, error)
	GetAttributesByConsentID(ctx context.Context, consentID, orgID string) ([]consentModel.ConsentAttribute, error)
	GetAttributesByConsentIDs(ctx context.Context, consentIDs []string, orgID string) (map[string]map[string]string, error)
	FindConsentIDsByAttributeKey(ctx context.Context, key, orgID string) ([]string, error)
	FindConsentIDsByAttribute(ctx context.Context, key, value, orgID string) ([]string, error)
	Create(tx dbmodel.TxInterface, consent *consentModel.Consent) error
	Update(tx dbmodel.TxInterface, consent *consentModel.Consent) error
	UpdateStatus(tx dbmodel.TxInterface, consentID, orgID, status string, updatedTime int64) error
	CreateAttributes(tx dbmodel.TxInterface, attributes []consentModel.ConsentAttribute) error
	DeleteAttributesByConsentID(tx dbmodel.TxInterface, consentID, orgID string) error
	CreateStatusAudit(tx dbmodel.TxInterface, audit *consentModel.ConsentStatusAudit) error

	CreateConsentPurposeMapping(tx dbmodel.TxInterface, consentID, purposeID, orgID string) error
	CreatePurposeElementApproval(tx dbmodel.TxInterface, approval *consentModel.ConsentElementApprovalRecord) error
	GetConsentPurposeMappingsByConsentID(ctx context.Context, consentID, orgID string) ([]consentModel.ConsentPurposeMapping, error)
	GetPurposeElementApprovalsByConsentID(ctx context.Context, consentID, orgID string) ([]consentModel.ConsentElementApprovalRecord, error)
	DeleteConsentPurposeMappingsByConsentID(tx dbmodel.TxInterface, consentID, orgID string) error
	DeletePurposeElementApprovalsByConsentID(tx dbmodel.TxInterface, consentID, orgID string) error
	CheckPurposeUsedInConsents(ctx context.Context, purposeID, orgID string) (bool, error)
}

// AuthResourceStore defines the interface for authorization resource data operations
type AuthResourceStore interface {
	GetByID(ctx context.Context, authID, orgID string) (*authResourceModel.AuthResource, error)
	GetByConsentID(ctx context.Context, consentID, orgID string) ([]authResourceModel.AuthResource, error)
	GetByConsentIDs(ctx context.Context, consentIDs []string, orgID string) ([]authResourceModel.AuthResource, error)
	Create(tx dbmodel.TxInterface, authResource *authResourceModel.AuthResource) error
	Update(tx dbmodel.TxInterface, authResource *authResourceModel.AuthResource) error
	DeleteByConsentID(tx dbmodel.TxInterface, consentID, orgID string) error
	UpdateAllStatusByConsentID(tx dbmodel.TxInterface, consentID, orgID, status string, updatedTime int64) error
}

// ConsentElementStore defines the interface for consent element data operations.
// Each logical element is identified by an ID and has one or more immutable versions.
// Version 1 is created when the element is first created; subsequent versions are added via CreateVersion.
type ConsentElementStore interface {
	// CreateVersion inserts a new element version (ELEMENT row + ELEMENT_PROPERTY rows).
	// Used for the initial create (version=1) and all subsequent versions.
	CreateVersion(tx dbmodel.TxInterface, version *consentElementModel.ElementVersion) error

	// GetLatestVersion returns the highest-numbered version of an element, with properties populated.
	GetLatestVersion(ctx context.Context, elementID, orgID string) (*consentElementModel.ElementVersion, error)

	// GetVersion returns a specific version by version number, with properties populated.
	GetVersion(ctx context.Context, elementID string, version int, orgID string) (*consentElementModel.ElementVersion, error)

	// ListVersions returns all versions of one element ordered by version number ascending, with properties.
	ListVersions(ctx context.Context, elementID, orgID string) ([]consentElementModel.ElementVersion, error)

	// List returns the latest version of each element matching the filters, with total count for pagination.
	// When filters.Details is false, Schema and Properties are not populated.
	List(ctx context.Context, orgID string, filters consentElementModel.ElementListFilter) ([]consentElementModel.ElementVersion, int, error)

	// GetByNameAndNamespace returns the latest version of an element matching name+namespace, or nil if not found.
	// Used for duplicate-name checks on element create.
	GetByNameAndNamespace(ctx context.Context, name, namespace, orgID string) (*consentElementModel.ElementVersion, error)

	// ElementExists reports whether any version of the element exists.
	ElementExists(ctx context.Context, elementID, orgID string) (bool, error)

	// DeleteVersion deletes a specific version row (ELEMENT_PROPERTY rows cascade).
	DeleteVersion(tx dbmodel.TxInterface, versionID, orgID string) error

	// DeleteElement deletes all versions of an element. Called when the last version is removed.
	DeleteElement(tx dbmodel.TxInterface, elementID, orgID string) error

	// IsVersionReferencedByPurpose reports whether any purpose version references this element version.
	// Returns true → caller must reject the delete with 409 Conflict.
	IsVersionReferencedByPurpose(ctx context.Context, versionID, orgID string) (bool, error)
}

// ConsentPurposeStore defines the interface for purpose data operations.
// Each logical purpose is identified by an ID and has one or more immutable versions.
// Version 1 is created when the purpose is first created; subsequent versions are added via CreateVersion.
type ConsentPurposeStore interface {
	// CreateVersion inserts a new purpose version (PURPOSE row + PURPOSE_PROPERTY rows).
	// Element mappings are linked separately via LinkElementVersion.
	CreateVersion(tx dbmodel.TxInterface, version *purposeModel.PurposeVersion) error

	// GetLatestVersion returns the highest-numbered version of a purpose, with properties and elements.
	// Returns nil if not found.
	GetLatestVersion(ctx context.Context, purposeID, orgID string) (*purposeModel.PurposeVersion, error)

	// GetVersion returns a specific version by version number, with properties and elements.
	// Returns nil if not found.
	GetVersion(ctx context.Context, purposeID string, version int, orgID string) (*purposeModel.PurposeVersion, error)

	// GetVersionByID returns a purpose version by its VERSION_ID, with properties and elements.
	// Returns nil if not found.
	GetVersionByID(ctx context.Context, purposeVersionID, orgID string) (*purposeModel.PurposeVersion, error)

	// ListVersions returns all versions of one purpose ordered by version number ascending, with properties and elements.
	ListVersions(ctx context.Context, purposeID, orgID string) ([]purposeModel.PurposeVersion, error)

	// PurposeExists reports whether any version of the purpose exists.
	PurposeExists(ctx context.Context, purposeID, orgID string) (bool, error)

	// GetByNameAndGroupID returns the latest version of a purpose with the given name and groupID,
	// or nil if no such purpose exists. Used to enforce name uniqueness within a group before insert.
	GetByNameAndGroupID(ctx context.Context, name, groupID, orgID string) (*purposeModel.PurposeVersion, error)

	// DeleteVersion deletes a specific version row.
	// PURPOSE_PROPERTY and PURPOSE_ELEMENT_MAPPING rows cascade automatically.
	DeleteVersion(tx dbmodel.TxInterface, purposeVersionID, orgID string) error

	// DeletePurpose deletes all versions of a purpose. Called when the last version is removed.
	DeletePurpose(tx dbmodel.TxInterface, purposeID, orgID string) error

	// LinkElementVersion creates a mapping from a purpose version to an element version.
	LinkElementVersion(tx dbmodel.TxInterface, purposeVersionID, elementVersionID, orgID string, mandatory bool) error

	// GetVersionElements returns all element refs for a specific purpose version.
	GetVersionElements(ctx context.Context, purposeVersionID, orgID string) ([]purposeModel.PurposeMappedElement, error)

	// IsElementVersionUsed reports whether any purpose version references this element version.
	// Returns true → caller must reject the element version delete with 409 Conflict.
	IsElementVersionUsed(ctx context.Context, elementVersionID, orgID string) (bool, error)

	// IsVersionUsedInConsents reports whether any consent references this purpose version.
	// Returns true → caller must reject the purpose version delete with 409 Conflict.
	IsVersionUsedInConsents(ctx context.Context, purposeVersionID, orgID string) (bool, error)

	// List returns the latest version of each purpose matching the filters, with total count for pagination.
	// When filters.Details is false, Properties and Elements are not populated.
	List(ctx context.Context, orgID string, filters purposeModel.PurposeListFilter) ([]purposeModel.PurposeVersion, int, error)
}
