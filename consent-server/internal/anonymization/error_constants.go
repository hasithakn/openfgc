/*
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
 * Licensed under the Apache License, Version 2.0.
 */

// Package anonymization provides account-anonymization functionality: disconnecting a
// user's consent and grievance records from their real identity by reassigning them to a
// freshly generated, unlinkable UUID.
package anonymization

import "github.com/wso2/openfgc/internal/system/error/serviceerror"

// Client errors for anonymization operations.
var (
	// ErrorValidationFailed is the error returned when request validation fails.
	ErrorValidationFailed = serviceerror.ServiceError{
		Type:        serviceerror.ClientErrorType,
		Code:        "AN-4002",
		Message:     "Validation failed",
		Description: "Request validation failed",
	}
)

// Server errors for anonymization operations.
var (
	// ErrorInternalServerError is the error returned when an internal operation fails.
	ErrorInternalServerError = serviceerror.ServiceError{
		Type:        serviceerror.ServerErrorType,
		Code:        "AN-5000",
		Message:     "Internal server error",
		Description: "An unexpected internal error occurred",
	}
)
