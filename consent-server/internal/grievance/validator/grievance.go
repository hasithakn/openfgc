/*
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
 * Licensed under the Apache License, Version 2.0.
 */

// Package validator validates grievance request payloads.
package validator

import (
	"encoding/base64"
	"fmt"
	"strings"
)

// MaxDescriptionLength bounds the free-text description field.
const MaxDescriptionLength = 10_000

// MaxMessageLength bounds a timeline entry message.
const MaxMessageLength = 10_000

// AttachmentRequest is the shape validated for every submitted attachment,
// independent of the model package to avoid an import cycle with the service layer.
type AttachmentRequest struct {
	FileName      string
	ContentType   string
	ContentBase64 string
}

// ValidateCreateRequest validates a grievance submission's shape.
// validCategories is the set of accepted category tokens (provided by the caller so
// this package stays independent of the category→priority mapping table).
func ValidateCreateRequest(userID, category, description string, attachments []AttachmentRequest, validCategories map[string]struct{}, maxAttachmentSizeBytes int64) error {
	if strings.TrimSpace(userID) == "" {
		return fmt.Errorf("userId is required")
	}
	if _, ok := validCategories[category]; !ok {
		return fmt.Errorf("category %q is not a recognized grievance category", category)
	}
	if strings.TrimSpace(description) == "" {
		return fmt.Errorf("description is required")
	}
	if len(description) > MaxDescriptionLength {
		return fmt.Errorf("description exceeds maximum length of %d characters", MaxDescriptionLength)
	}
	return validateAttachments(attachments, maxAttachmentSizeBytes)
}

// ValidateMessageRequest validates a reply/internal-note submission's shape.
func ValidateMessageRequest(actorUserID, actorRole, message, visibility string, attachments []AttachmentRequest, maxAttachmentSizeBytes int64) error {
	if strings.TrimSpace(actorUserID) == "" {
		return fmt.Errorf("actorUserId is required")
	}
	if actorRole != "DataPrincipal" && actorRole != "GrievanceOfficer" {
		return fmt.Errorf("actorRole must be 'DataPrincipal' or 'GrievanceOfficer'")
	}
	if visibility != "shared" && visibility != "internal" {
		return fmt.Errorf("visibility must be 'shared' or 'internal'")
	}
	if visibility == "internal" && actorRole != "GrievanceOfficer" {
		return fmt.Errorf("only a Grievance Officer may post an internal-visibility entry")
	}
	if strings.TrimSpace(message) == "" {
		return fmt.Errorf("message is required")
	}
	if len(message) > MaxMessageLength {
		return fmt.Errorf("message exceeds maximum length of %d characters", MaxMessageLength)
	}
	return validateAttachments(attachments, maxAttachmentSizeBytes)
}

// ValidateStatusTransitionRequest validates a status-transition submission's shape.
// The actual transition-table check happens in the service layer, which owns the table.
func ValidateStatusTransitionRequest(actorUserID, actorRole, nextStatus string) error {
	if strings.TrimSpace(actorUserID) == "" {
		return fmt.Errorf("actorUserId is required")
	}
	if actorRole != "DataPrincipal" && actorRole != "GrievanceOfficer" {
		return fmt.Errorf("actorRole must be 'DataPrincipal' or 'GrievanceOfficer'")
	}
	if strings.TrimSpace(nextStatus) == "" {
		return fmt.Errorf("nextStatus is required")
	}
	return nil
}

func validateAttachments(attachments []AttachmentRequest, maxSizeBytes int64) error {
	for _, a := range attachments {
		if strings.TrimSpace(a.FileName) == "" {
			return fmt.Errorf("attachment fileName is required")
		}
		if strings.TrimSpace(a.ContentType) == "" {
			return fmt.Errorf("attachment %q: contentType is required", a.FileName)
		}
		decoded, err := base64.StdEncoding.DecodeString(a.ContentBase64)
		if err != nil {
			return fmt.Errorf("attachment %q: contentBase64 is not valid base64: %w", a.FileName, err)
		}
		if int64(len(decoded)) > maxSizeBytes {
			return fmt.Errorf("attachment %q exceeds the maximum size of %d bytes", a.FileName, maxSizeBytes)
		}
	}
	return nil
}
