/*
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
 * Licensed under the Apache License, Version 2.0.
 */

package anonymization

import (
	"context"
	"fmt"

	"github.com/wso2/openfgc/internal/anonymization/model"
	"github.com/wso2/openfgc/internal/consent"
	consentModel "github.com/wso2/openfgc/internal/consent/model"
	dbmodel "github.com/wso2/openfgc/internal/system/database/model"
	"github.com/wso2/openfgc/internal/system/error/serviceerror"
	"github.com/wso2/openfgc/internal/system/log"
	"github.com/wso2/openfgc/internal/system/stores"
	"github.com/wso2/openfgc/internal/system/utils"
)

// userIDAttributeKey is the CONSENT_ATTRIBUTE key that mirrors a consent's owning user —
// the same key used throughout the consent module's attribute-based search (e.g.
// SearchConsentsByAttribute). It must be kept in sync with the original identity here too,
// or the anonymized user could still be found by attribute lookup after "deletion".
const userIDAttributeKey = "userId"

// anonymizeActionBy/Reason document why the consent was revoked, for the CONSENT_STATUS_AUDIT
// trail this reuses via the standard RevokeConsent flow.
const anonymizeRevokeReason = "Account deleted by data principal"

// AnonymizationService anonymizes all OpenFGC data belonging to a user.
type AnonymizationService interface {
	// AnonymizeUser revokes the user's currently-active consents and then reassigns every
	// trace of their identity (auth resources, the "userId" consent attribute, grievances,
	// and grievance timeline entries) to a freshly generated UUID, scoped to orgID.
	AnonymizeUser(ctx context.Context, orgID, userID string) (*model.AnonymizeOutput, *serviceerror.ServiceError)
}

type service struct {
	stores     *stores.StoreRegistry
	consentSvc consent.ConsentService
}

// NewAnonymizationService creates a new anonymization service.
func NewAnonymizationService(registry *stores.StoreRegistry, consentSvc consent.ConsentService) AnonymizationService {
	return &service{stores: registry, consentSvc: consentSvc}
}

func (s *service) AnonymizeUser(ctx context.Context, orgID, userID string) (*model.AnonymizeOutput, *serviceerror.ServiceError) {
	logger := log.GetLogger().WithContext(ctx)
	logger.Info("Anonymizing user", log.String("org_id", orgID))

	// Phase 1: revoke every currently-active consent this user holds an authorization on,
	// via the existing single-consent revoke flow — reused as-is (rather than reimplementing
	// its status-audit/history logic) and run while the original userID is still attached, so
	// a partial failure here is always safely retriable on a subsequent anonymize attempt.
	authResources, err := s.stores.AuthResource.GetByUserID(ctx, orgID, userID)
	if err != nil {
		logger.Error("Failed to look up auth resources for user", log.Error(err))
		return nil, serviceerror.CustomServiceError(ErrorInternalServerError, err.Error())
	}

	seenConsents := make(map[string]struct{}, len(authResources))
	revokedCount := 0
	for _, ar := range authResources {
		if _, seen := seenConsents[ar.ConsentID]; seen {
			continue
		}
		seenConsents[ar.ConsentID] = struct{}{}

		_, revokeErr := s.consentSvc.RevokeConsent(ctx, ar.ConsentID, orgID, consentModel.ConsentRevokeInput{
			ActionBy: userID,
			Reason:   anonymizeRevokeReason,
		})
		if revokeErr != nil {
			if revokeErr.Code == consent.ErrorConsentAlreadyRevoked.Code {
				continue
			}
			logger.Error("Failed to revoke consent during anonymization",
				log.String("consent_id", ar.ConsentID), log.Error(revokeErr))
			return nil, serviceerror.CustomServiceError(ErrorInternalServerError,
				fmt.Sprintf("failed to revoke consent '%s': %v", ar.ConsentID, revokeErr))
		}
		revokedCount++
	}

	// Phase 2: reassign every remaining trace of the original identity to a single new
	// UUID, in one transaction — either all of it lands, or none of it does.
	newUserID := utils.GenerateUUID()
	currentTime := utils.GetCurrentTimeMillis()

	err = s.stores.ExecuteTransaction([]func(tx dbmodel.TxInterface) error{
		func(tx dbmodel.TxInterface) error {
			return s.stores.AuthResource.UpdateUserIDByUserID(tx, orgID, userID, newUserID, currentTime)
		},
		func(tx dbmodel.TxInterface) error {
			return s.stores.Consent.UpdateAttributeValue(tx, userIDAttributeKey, userID, newUserID, orgID)
		},
		func(tx dbmodel.TxInterface) error {
			return s.stores.Consent.UpdateStatusAuditActionByValue(tx, userID, newUserID, orgID)
		},
		func(tx dbmodel.TxInterface) error {
			return s.stores.Consent.UpdateHistoryActionByValue(tx, userID, newUserID, orgID)
		},
		func(tx dbmodel.TxInterface) error {
			return s.stores.Grievance.UpdateUserIDByUserID(tx, orgID, userID, newUserID)
		},
		func(tx dbmodel.TxInterface) error {
			return s.stores.Grievance.UpdateTimelineActorByUserID(tx, orgID, userID, newUserID)
		},
	})
	if err != nil {
		logger.Error("Anonymize transaction failed", log.Error(err))
		return nil, serviceerror.CustomServiceError(ErrorInternalServerError,
			fmt.Sprintf("failed to anonymize user data: %v", err))
	}

	logger.Info("User anonymized",
		log.String("org_id", orgID),
		log.Int("consents_revoked", revokedCount))

	return &model.AnonymizeOutput{
		AnonymizedUserID: newUserID,
		ConsentsRevoked:  revokedCount,
	}, nil
}
