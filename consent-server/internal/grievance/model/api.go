/*
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
 * Licensed under the Apache License, Version 2.0.
 */

package model

// AttachmentAPI is a client-submitted attachment as it appears in a JSON request body.
type AttachmentAPI struct {
	FileName      string `json:"fileName"`
	ContentType   string `json:"contentType"`
	ContentBase64 string `json:"contentBase64"`
}

// GrievanceCreateRequest is the POST /grievances request body.
type GrievanceCreateRequest struct {
	UserID      string          `json:"userId"`
	Category    string          `json:"category"`
	Description string          `json:"description"`
	Attachments []AttachmentAPI `json:"attachments,omitempty"`
}

// GrievanceMessageRequest is the POST /grievances/{id}/messages request body.
type GrievanceMessageRequest struct {
	ActorUserID string          `json:"actorUserId"`
	ActorRole   string          `json:"actorRole"`
	Message     string          `json:"message"`
	Visibility  string          `json:"visibility"`
	Attachments []AttachmentAPI `json:"attachments,omitempty"`
}

// GrievanceStatusRequest is the POST /grievances/{id}/status request body.
type GrievanceStatusRequest struct {
	ActorUserID string `json:"actorUserId"`
	ActorRole   string `json:"actorRole"`
	NextStatus  string `json:"nextStatus"`
}

// AttachmentResponse is the metadata-only JSON view of an attachment.
type AttachmentResponse struct {
	AttachmentID  string `json:"attachmentId"`
	FileName      string `json:"fileName"`
	FileSizeBytes int64  `json:"fileSizeBytes"`
	ContentType   string `json:"contentType"`
}

// TimelineEntryResponse is the JSON view of one timeline entry.
type TimelineEntryResponse struct {
	EntryID     string               `json:"entryId"`
	EntryType   string               `json:"entryType"`
	Visibility  string               `json:"visibility"`
	ActorUserID *string              `json:"actorUserId,omitempty"`
	ActorRole   string               `json:"actorRole"`
	Message     string               `json:"message"`
	FromStatus  *string              `json:"fromStatus,omitempty"`
	ToStatus    *string              `json:"toStatus,omitempty"`
	CreatedTime int64                `json:"createdTime"`
	Attachments []AttachmentResponse `json:"attachments,omitempty"`
}

// GrievanceResponse is the JSON view of a grievance. Timeline/Attachments are populated
// for detail fetches only; list results omit them.
type GrievanceResponse struct {
	GrievanceID      string                  `json:"grievanceId"`
	ReferenceID      string                  `json:"referenceId"`
	Category         string                  `json:"category"`
	Priority         string                  `json:"priority"`
	Status           string                  `json:"status"`
	Description      string                  `json:"description,omitempty"`
	UserID           string                  `json:"userId"`
	SubmittedTime    int64                   `json:"submittedTime"`
	UpdatedTime      int64                   `json:"updatedTime"`
	StatutoryDueTime int64                   `json:"statutoryDueTime"`
	Timeline         []TimelineEntryResponse `json:"timeline,omitempty"`
	Attachments      []AttachmentResponse    `json:"attachments,omitempty"`
}

// GrievanceListMetadata is the pagination envelope for a list response.
type GrievanceListMetadata struct {
	Total    int `json:"total"`
	Page     int `json:"page"`
	PageSize int `json:"pageSize"`
}

// GrievanceListResponse is the GET /grievances response body.
type GrievanceListResponse struct {
	Data     []GrievanceResponse   `json:"data"`
	Metadata GrievanceListMetadata `json:"metadata"`
}

// GrievanceStatsResponse is the GET /grievances/stats response body.
type GrievanceStatsResponse struct {
	Open         int `json:"open"`
	AwaitingInfo int `json:"awaitingInfo"`
	Resolved     int `json:"resolved"`
	SLABreached  int `json:"slaBreached"`
}
