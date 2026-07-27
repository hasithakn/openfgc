/*
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
 * Licensed under the Apache License, Version 2.0.
 */

package anonymization

import (
	"net/http"

	"github.com/wso2/openfgc/internal/consent"
	"github.com/wso2/openfgc/internal/eventpublish"
	"github.com/wso2/openfgc/internal/system/constants"
	"github.com/wso2/openfgc/internal/system/stores"
)

// Initialize sets up the anonymization module and registers its route. consentSvc is the
// already-constructed consent module service, reused here to revoke consents rather than
// duplicating that logic. eventClient publishes ACCOUNT_DELETE; nil disables publishing.
func Initialize(mux *http.ServeMux, registry *stores.StoreRegistry, consentSvc consent.ConsentService, eventClient *eventpublish.Client) {
	service := NewAnonymizationService(registry, consentSvc, eventClient)
	h := newAnonymizationHandler(service)

	mux.HandleFunc("POST "+constants.APIBasePath+"/anonymize", h.anonymize)
}
