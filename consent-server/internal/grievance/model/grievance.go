/*
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
 * Licensed under the Apache License, Version 2.0.
 */

// Package model provides data models for grievances.
package model

// =============================================================================
// DB types — store layer only, db tags, no json tags
// =============================================================================

// Grievance is one row from the GRIEVANCE table. Priority and Status hold DB values
// (e.g. "Critical", "Waiting on DPO") — the handler maps these to/from the API's
// P0-P3 / Open-Investigation-AwaitingInfo-WaitingOnDpo-Resolved vocabulary.
type Grievance struct {
	GrievanceID      string `db:"GRIEVANCE_ID"`
	OrgID            string `db:"ORG_ID"`
	UserID           string `db:"USER_ID"`
	ReferenceID      string `db:"REFERENCE_ID"`
	Category         string `db:"CATEGORY"`
	Priority         string `db:"PRIORITY"`
	Status           string `db:"STATUS"`
	Description      string `db:"DESCRIPTION"`
	SubmittedTime    int64  `db:"SUBMITTED_TIME"`
	UpdatedTime      int64  `db:"UPDATED_TIME"`
	StatutoryDueTime int64  `db:"STATUTORY_DUE_TIME"`
}

// TimelineEntry is one row from the GRIEVANCE_TIMELINE_ENTRY table.
type TimelineEntry struct {
	EntryID     string  `db:"ENTRY_ID"`
	OrgID       string  `db:"ORG_ID"`
	GrievanceID string  `db:"GRIEVANCE_ID"`
	EntryType   string  `db:"ENTRY_TYPE"`
	Visibility  string  `db:"VISIBILITY"`
	ActorUserID *string `db:"ACTOR_USER_ID"`
	ActorRole   string  `db:"ACTOR_ROLE"`
	Message     string  `db:"MESSAGE"`
	FromStatus  *string `db:"FROM_STATUS"`
	ToStatus    *string `db:"TO_STATUS"`
	CreatedTime int64   `db:"CREATED_TIME"`
}

// Attachment is one row from the GRIEVANCE_ATTACHMENT table, excluding FILE_DATA.
// Used for metadata listing (grievance/timeline-entry detail responses).
type Attachment struct {
	AttachmentID    string  `db:"ATTACHMENT_ID"`
	OrgID           string  `db:"ORG_ID"`
	GrievanceID     string  `db:"GRIEVANCE_ID"`
	TimelineEntryID *string `db:"TIMELINE_ENTRY_ID"`
	FileName        string  `db:"FILE_NAME"`
	FileSizeBytes   int64   `db:"FILE_SIZE_BYTES"`
	ContentType     string  `db:"CONTENT_TYPE"`
	CreatedTime     int64   `db:"CREATED_TIME"`
}

// AttachmentContent is a full GRIEVANCE_ATTACHMENT row including binary content,
// used only for the download endpoint.
type AttachmentContent struct {
	Attachment
	FileData []byte
}

// =============================================================================
// Service-layer input types
// =============================================================================

// AttachmentInput is a client-submitted attachment, transported as base64 in JSON.
type AttachmentInput struct {
	FileName      string
	ContentType   string
	ContentBase64 string
}

// CreateGrievanceInput is the service-layer input for submitting a new grievance.
// Category is the API-vocabulary token (e.g. "dataBreachComplaint"); priority is derived
// server-side from it and is not caller-supplied.
type CreateGrievanceInput struct {
	UserID      string
	Category    string
	Description string
	Attachments []AttachmentInput
}

// PostMessageInput is the service-layer input for posting a reply or internal note.
// ActorRole and Visibility are API-vocabulary values ("DataPrincipal"/"GrievanceOfficer", "shared"/"internal").
type PostMessageInput struct {
	ActorUserID string
	ActorRole   string
	Message     string
	Visibility  string
	Attachments []AttachmentInput
}

// TransitionStatusInput is the service-layer input for an explicit status transition.
// NextStatus is an API-vocabulary value.
type TransitionStatusInput struct {
	ActorUserID string
	ActorRole   string
	NextStatus  string
}

// SearchFilter carries list/query parameters. Status/Priority are API-vocabulary values when set.
type SearchFilter struct {
	OrgID    string
	UserIDs  []string
	Status   string
	Priority string
	Query    string
	Page     int
	PageSize int
	SortBy   string
	Order    string
}

// =============================================================================
// Service-layer output types
// =============================================================================

// GrievanceOutput is the fully assembled service-layer view of a grievance.
// All Priority/Status/EntryType/Visibility/ActorRole fields hold API-vocabulary values.
// Timeline/Attachments are populated for detail fetches only; list results omit them.
type GrievanceOutput struct {
	GrievanceID      string
	ReferenceID      string
	Category         string
	Priority         string
	Status           string
	Description      string
	UserID           string
	SubmittedTime    int64
	UpdatedTime      int64
	StatutoryDueTime int64
	Timeline         []TimelineEntryOutput
	Attachments      []AttachmentOutput
}

// TimelineEntryOutput is the API-vocabulary view of one timeline entry.
type TimelineEntryOutput struct {
	EntryID     string
	EntryType   string
	Visibility  string
	ActorUserID *string
	ActorRole   string
	Message     string
	FromStatus  *string
	ToStatus    *string
	CreatedTime int64
	Attachments []AttachmentOutput
}

// AttachmentOutput is the metadata-only view of an attachment (no file content).
type AttachmentOutput struct {
	AttachmentID  string
	FileName      string
	FileSizeBytes int64
	ContentType   string
}

// ListOutput is the paginated list result.
type ListOutput struct {
	Data     []GrievanceOutput
	Total    int
	Page     int
	PageSize int
}

// StatsOutput is the officer queue's aggregate summary.
type StatsOutput struct {
	Open         int
	AwaitingInfo int
	Resolved     int
	SLABreached  int
}
