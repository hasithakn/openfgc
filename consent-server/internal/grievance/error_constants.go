/*
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
 * Licensed under the Apache License, Version 2.0.
 */

// Package grievance provides grievance (complaint) management functionality.
package grievance

import "github.com/wso2/openfgc/internal/system/error/serviceerror"

// Client errors for grievance operations.
var (
	// ErrorInvalidRequestBody is the error returned when the request body is invalid or malformed.
	ErrorInvalidRequestBody = serviceerror.ServiceError{
		Type:        serviceerror.ClientErrorType,
		Code:        "GRV-4001",
		Message:     "Invalid request body",
		Description: "The request body is malformed or contains invalid data",
	}
	// ErrorValidationFailed is the error returned when request validation fails.
	ErrorValidationFailed = serviceerror.ServiceError{
		Type:        serviceerror.ClientErrorType,
		Code:        "GRV-4002",
		Message:     "Validation failed",
		Description: "Request validation failed",
	}
	// ErrorGrievanceNotFound is the error returned when a grievance is not found.
	ErrorGrievanceNotFound = serviceerror.ServiceError{
		Type:        serviceerror.ClientErrorType,
		Code:        "GRV-4040",
		Message:     "Grievance not found",
		Description: "The requested grievance could not be found",
	}
	// ErrorAttachmentNotFound is the error returned when an attachment is not found.
	ErrorAttachmentNotFound = serviceerror.ServiceError{
		Type:        serviceerror.ClientErrorType,
		Code:        "GRV-4042",
		Message:     "Attachment not found",
		Description: "The requested attachment could not be found",
	}
	// ErrorInvalidStatusTransition is the error returned when a status transition is not allowed.
	ErrorInvalidStatusTransition = serviceerror.ServiceError{
		Type:        serviceerror.ClientErrorType,
		Code:        "GRV-4090",
		Message:     "Invalid status transition",
		Description: "The requested status transition is not allowed from the grievance's current status",
	}
)

// Server errors for grievance operations.
var (
	// ErrorInternalServerError is the error returned when an internal operation fails.
	ErrorInternalServerError = serviceerror.ServiceError{
		Type:        serviceerror.ServerErrorType,
		Code:        "GRV-5000",
		Message:     "Internal server error",
		Description: "An unexpected internal error occurred",
	}
)
