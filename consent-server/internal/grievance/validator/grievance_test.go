/*
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
 * Licensed under the Apache License, Version 2.0.
 */

package validator

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/require"
)

var validCategories = map[string]struct{}{"dataBreachComplaint": {}}

func validAttachment() AttachmentRequest {
	return AttachmentRequest{
		FileName:      "evidence.txt",
		ContentType:   "text/plain",
		ContentBase64: base64.StdEncoding.EncodeToString([]byte("hello")),
	}
}

// =============================================================================
// ValidateCreateRequest
// =============================================================================

func TestValidateCreateRequest_Success(t *testing.T) {
	err := ValidateCreateRequest("asha@fgc.com", "dataBreachComplaint", "My data was shared.",
		[]AttachmentRequest{validAttachment()}, validCategories, 1024)
	require.NoError(t, err)
}

func TestValidateCreateRequest_MissingUserID(t *testing.T) {
	err := ValidateCreateRequest("", "dataBreachComplaint", "desc", nil, validCategories, 1024)
	require.Error(t, err)
}

func TestValidateCreateRequest_UnknownCategory(t *testing.T) {
	err := ValidateCreateRequest("asha@fgc.com", "notACategory", "desc", nil, validCategories, 1024)
	require.Error(t, err)
}

func TestValidateCreateRequest_MissingDescription(t *testing.T) {
	err := ValidateCreateRequest("asha@fgc.com", "dataBreachComplaint", "", nil, validCategories, 1024)
	require.Error(t, err)
}

func TestValidateCreateRequest_AttachmentTooLarge(t *testing.T) {
	big := AttachmentRequest{
		FileName:      "big.bin",
		ContentType:   "application/octet-stream",
		ContentBase64: base64.StdEncoding.EncodeToString(make([]byte, 100)),
	}
	err := ValidateCreateRequest("asha@fgc.com", "dataBreachComplaint", "desc",
		[]AttachmentRequest{big}, validCategories, 10)
	require.Error(t, err)
}

func TestValidateCreateRequest_AttachmentInvalidBase64(t *testing.T) {
	bad := AttachmentRequest{FileName: "f.txt", ContentType: "text/plain", ContentBase64: "not-base64!!"}
	err := ValidateCreateRequest("asha@fgc.com", "dataBreachComplaint", "desc",
		[]AttachmentRequest{bad}, validCategories, 1024)
	require.Error(t, err)
}

// =============================================================================
// ValidateMessageRequest
// =============================================================================

func TestValidateMessageRequest_Success(t *testing.T) {
	err := ValidateMessageRequest("asha@fgc.com", "DataPrincipal", "hello", "shared", nil, 1024)
	require.NoError(t, err)
}

func TestValidateMessageRequest_InternalRequiresOfficer(t *testing.T) {
	err := ValidateMessageRequest("asha@fgc.com", "DataPrincipal", "hello", "internal", nil, 1024)
	require.Error(t, err)
}

func TestValidateMessageRequest_InternalAllowedForOfficer(t *testing.T) {
	err := ValidateMessageRequest("dpo@fgc.com", "GrievanceOfficer", "note", "internal", nil, 1024)
	require.NoError(t, err)
}

func TestValidateMessageRequest_InvalidActorRole(t *testing.T) {
	err := ValidateMessageRequest("dpo@fgc.com", "SomeoneElse", "note", "shared", nil, 1024)
	require.Error(t, err)
}

func TestValidateMessageRequest_InvalidVisibility(t *testing.T) {
	err := ValidateMessageRequest("dpo@fgc.com", "GrievanceOfficer", "note", "public", nil, 1024)
	require.Error(t, err)
}

func TestValidateMessageRequest_MissingMessage(t *testing.T) {
	err := ValidateMessageRequest("asha@fgc.com", "DataPrincipal", "", "shared", nil, 1024)
	require.Error(t, err)
}

// =============================================================================
// ValidateStatusTransitionRequest
// =============================================================================

func TestValidateStatusTransitionRequest_Success(t *testing.T) {
	err := ValidateStatusTransitionRequest("dpo@fgc.com", "GrievanceOfficer", "Investigation")
	require.NoError(t, err)
}

func TestValidateStatusTransitionRequest_MissingNextStatus(t *testing.T) {
	err := ValidateStatusTransitionRequest("dpo@fgc.com", "GrievanceOfficer", "")
	require.Error(t, err)
}
