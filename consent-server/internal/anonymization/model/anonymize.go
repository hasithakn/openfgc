/*
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
 * Licensed under the Apache License, Version 2.0.
 */

// Package model provides data models for account anonymization.
package model

// AnonymizeOutput is the return type from AnonymizeUser.
type AnonymizeOutput struct {
	AnonymizedUserID string
	ConsentsRevoked  int
}

// AnonymizeResponse is the HTTP response body for POST /anonymize.
type AnonymizeResponse struct {
	AnonymizedUserID string `json:"anonymizedUserId"`
	ConsentsRevoked  int    `json:"consentsRevoked"`
}
