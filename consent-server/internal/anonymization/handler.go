/*
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
 * Licensed under the Apache License, Version 2.0.
 */

package anonymization

import (
	"encoding/json"
	"net/http"

	"github.com/wso2/openfgc/internal/anonymization/model"
	"github.com/wso2/openfgc/internal/system/constants"
	"github.com/wso2/openfgc/internal/system/error/serviceerror"
	"github.com/wso2/openfgc/internal/system/utils"
)

type handler struct {
	service AnonymizationService
}

func newAnonymizationHandler(service AnonymizationService) *handler {
	return &handler{service: service}
}

// anonymize handles POST /anonymize?userId=... — reassigns every OpenFGC record belonging
// to userId (within orgID) to a freshly generated, unlinkable UUID.
func (h *handler) anonymize(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	orgID := r.Header.Get(constants.HeaderOrgID)
	if err := utils.ValidateOrgID(orgID); err != nil {
		utils.SendError(w, r, serviceerror.CustomServiceError(ErrorValidationFailed, err.Error()))
		return
	}

	userID := r.URL.Query().Get("userId")
	if err := utils.ValidateRequired("userId", userID); err != nil {
		utils.SendError(w, r, serviceerror.CustomServiceError(ErrorValidationFailed, err.Error()))
		return
	}

	out, svcErr := h.service.AnonymizeUser(ctx, orgID, userID)
	if svcErr != nil {
		utils.SendError(w, r, svcErr)
		return
	}

	w.Header().Set(constants.HeaderContentType, constants.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(model.AnonymizeResponse{
		AnonymizedUserID: out.AnonymizedUserID,
		ConsentsRevoked:  out.ConsentsRevoked,
	})
}
