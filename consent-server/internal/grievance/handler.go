/*
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
 * Licensed under the Apache License, Version 2.0.
 */

package grievance

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/wso2/openfgc/internal/grievance/model"
	"github.com/wso2/openfgc/internal/grievance/validator"
	"github.com/wso2/openfgc/internal/system/constants"
	"github.com/wso2/openfgc/internal/system/error/serviceerror"
	"github.com/wso2/openfgc/internal/system/utils"
)

type grievanceHandler struct {
	service GrievanceService
}

func newGrievanceHandler(service GrievanceService) *grievanceHandler {
	return &grievanceHandler{service: service}
}

// =============================================================================
// HTTP handlers
// =============================================================================

// createGrievance handles POST /grievances
func (h *grievanceHandler) createGrievance(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	orgID := r.Header.Get(constants.HeaderOrgID)
	if err := utils.ValidateOrgID(orgID); err != nil {
		utils.SendError(w, r, serviceerror.CustomServiceError(ErrorValidationFailed, err.Error()))
		return
	}

	var req model.GrievanceCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(w, r, serviceerror.CustomServiceError(ErrorInvalidRequestBody, "invalid request body"))
		return
	}

	if err := validator.ValidateCreateRequest(
		req.UserID, req.Category, req.Description,
		attachmentRequestsToValidator(req.Attachments), validCategorySet(), maxAttachmentSizeBytes,
	); err != nil {
		utils.SendError(w, r, serviceerror.CustomServiceError(ErrorValidationFailed, err.Error()))
		return
	}

	input := model.CreateGrievanceInput{
		UserID:      req.UserID,
		Category:    req.Category,
		Description: req.Description,
		Attachments: attachmentRequestsToInput(req.Attachments),
	}

	out, serviceErr := h.service.CreateGrievance(ctx, input, orgID)
	if serviceErr != nil {
		utils.SendError(w, r, serviceErr)
		return
	}

	w.Header().Set(constants.HeaderContentType, constants.ContentTypeJSON)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(grievanceOutputToResponse(out))
}

// getGrievance handles GET /grievances/{grievanceId}
func (h *grievanceHandler) getGrievance(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	grievanceID := r.PathValue("grievanceId")
	orgID := r.Header.Get(constants.HeaderOrgID)
	if err := utils.ValidateOrgID(orgID); err != nil {
		utils.SendError(w, r, serviceerror.CustomServiceError(ErrorValidationFailed, err.Error()))
		return
	}

	includeInternal := r.URL.Query().Get("includeInternal") == "true"
	out, serviceErr := h.service.GetGrievance(ctx, grievanceID, orgID, includeInternal)
	if serviceErr != nil {
		utils.SendError(w, r, serviceErr)
		return
	}

	w.Header().Set(constants.HeaderContentType, constants.ContentTypeJSON)
	json.NewEncoder(w).Encode(grievanceOutputToResponse(out))
}

// listGrievances handles GET /grievances
func (h *grievanceHandler) listGrievances(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	orgID := r.Header.Get(constants.HeaderOrgID)
	if err := utils.ValidateOrgID(orgID); err != nil {
		utils.SendError(w, r, serviceerror.CustomServiceError(ErrorValidationFailed, err.Error()))
		return
	}

	q := r.URL.Query()
	filter := model.SearchFilter{
		OrgID:    orgID,
		Status:   q.Get("status"),
		Priority: q.Get("priority"),
		Query:    q.Get("q"),
		SortBy:   q.Get("sortBy"),
		Order:    q.Get("order"),
		Page:     0,
		PageSize: 10,
	}
	if s := q.Get("userIds"); s != "" {
		parts := strings.Split(s, ",")
		for i := range parts {
			parts[i] = strings.TrimSpace(parts[i])
		}
		filter.UserIDs = parts
	}
	if s := q.Get("page"); s != "" {
		if p, err := strconv.Atoi(s); err == nil && p >= 0 {
			filter.Page = p
		}
	}
	if s := q.Get("pageSize"); s != "" {
		if ps, err := strconv.Atoi(s); err == nil && ps > 0 {
			const maxPageSize = 100
			if ps > maxPageSize {
				ps = maxPageSize
			}
			filter.PageSize = ps
		}
	}

	out, serviceErr := h.service.ListGrievances(ctx, filter)
	if serviceErr != nil {
		utils.SendError(w, r, serviceErr)
		return
	}

	w.Header().Set(constants.HeaderContentType, constants.ContentTypeJSON)
	json.NewEncoder(w).Encode(listOutputToResponse(out))
}

// postMessage handles POST /grievances/{grievanceId}/messages
func (h *grievanceHandler) postMessage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	grievanceID := r.PathValue("grievanceId")
	orgID := r.Header.Get(constants.HeaderOrgID)
	if err := utils.ValidateOrgID(orgID); err != nil {
		utils.SendError(w, r, serviceerror.CustomServiceError(ErrorValidationFailed, err.Error()))
		return
	}

	var req model.GrievanceMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(w, r, serviceerror.CustomServiceError(ErrorInvalidRequestBody, "invalid request body"))
		return
	}

	if err := validator.ValidateMessageRequest(
		req.ActorUserID, req.ActorRole, req.Message, req.Visibility,
		attachmentRequestsToValidator(req.Attachments), maxAttachmentSizeBytes,
	); err != nil {
		utils.SendError(w, r, serviceerror.CustomServiceError(ErrorValidationFailed, err.Error()))
		return
	}

	input := model.PostMessageInput{
		ActorUserID: req.ActorUserID,
		ActorRole:   req.ActorRole,
		Message:     req.Message,
		Visibility:  req.Visibility,
		Attachments: attachmentRequestsToInput(req.Attachments),
	}

	out, serviceErr := h.service.PostMessage(ctx, grievanceID, orgID, input)
	if serviceErr != nil {
		utils.SendError(w, r, serviceErr)
		return
	}

	w.Header().Set(constants.HeaderContentType, constants.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(grievanceOutputToResponse(out))
}

// transitionStatus handles POST /grievances/{grievanceId}/status
func (h *grievanceHandler) transitionStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	grievanceID := r.PathValue("grievanceId")
	orgID := r.Header.Get(constants.HeaderOrgID)
	if err := utils.ValidateOrgID(orgID); err != nil {
		utils.SendError(w, r, serviceerror.CustomServiceError(ErrorValidationFailed, err.Error()))
		return
	}

	var req model.GrievanceStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(w, r, serviceerror.CustomServiceError(ErrorInvalidRequestBody, "invalid request body"))
		return
	}
	if err := validator.ValidateStatusTransitionRequest(req.ActorUserID, req.ActorRole, req.NextStatus); err != nil {
		utils.SendError(w, r, serviceerror.CustomServiceError(ErrorValidationFailed, err.Error()))
		return
	}

	input := model.TransitionStatusInput{
		ActorUserID: req.ActorUserID,
		ActorRole:   req.ActorRole,
		NextStatus:  req.NextStatus,
	}
	out, serviceErr := h.service.TransitionStatus(ctx, grievanceID, orgID, input)
	if serviceErr != nil {
		utils.SendError(w, r, serviceErr)
		return
	}

	w.Header().Set(constants.HeaderContentType, constants.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(grievanceOutputToResponse(out))
}

// getStats handles GET /grievances/stats
func (h *grievanceHandler) getStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	orgID := r.Header.Get(constants.HeaderOrgID)
	if err := utils.ValidateOrgID(orgID); err != nil {
		utils.SendError(w, r, serviceerror.CustomServiceError(ErrorValidationFailed, err.Error()))
		return
	}

	out, serviceErr := h.service.GetStats(ctx, orgID)
	if serviceErr != nil {
		utils.SendError(w, r, serviceErr)
		return
	}

	w.Header().Set(constants.HeaderContentType, constants.ContentTypeJSON)
	json.NewEncoder(w).Encode(model.GrievanceStatsResponse{
		Open:         out.Open,
		AwaitingInfo: out.AwaitingInfo,
		Resolved:     out.Resolved,
		SLABreached:  out.SLABreached,
	})
}

// getAttachment handles GET /grievances/{grievanceId}/attachments/{attachmentId}
func (h *grievanceHandler) getAttachment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	grievanceID := r.PathValue("grievanceId")
	attachmentID := r.PathValue("attachmentId")
	orgID := r.Header.Get(constants.HeaderOrgID)
	if err := utils.ValidateOrgID(orgID); err != nil {
		utils.SendError(w, r, serviceerror.CustomServiceError(ErrorValidationFailed, err.Error()))
		return
	}

	content, serviceErr := h.service.GetAttachmentContent(ctx, grievanceID, attachmentID, orgID)
	if serviceErr != nil {
		utils.SendError(w, r, serviceErr)
		return
	}

	w.Header().Set(constants.HeaderContentType, content.ContentType)
	w.Header().Set("Content-Disposition", `attachment; filename="`+content.FileName+`"`)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(content.FileData)
}

// =============================================================================
// Converters
// =============================================================================

func validCategorySet() map[string]struct{} {
	set := make(map[string]struct{}, len(categoryToPriorityDB))
	for k := range categoryToPriorityDB {
		set[k] = struct{}{}
	}
	return set
}

func attachmentRequestsToValidator(reqs []model.AttachmentAPI) []validator.AttachmentRequest {
	out := make([]validator.AttachmentRequest, 0, len(reqs))
	for _, a := range reqs {
		out = append(out, validator.AttachmentRequest{
			FileName:      a.FileName,
			ContentType:   a.ContentType,
			ContentBase64: a.ContentBase64,
		})
	}
	return out
}

func attachmentRequestsToInput(reqs []model.AttachmentAPI) []model.AttachmentInput {
	inputs := make([]model.AttachmentInput, 0, len(reqs))
	for _, a := range reqs {
		inputs = append(inputs, model.AttachmentInput{
			FileName:      a.FileName,
			ContentType:   a.ContentType,
			ContentBase64: a.ContentBase64,
		})
	}
	return inputs
}

func attachmentOutputsToResponse(outs []model.AttachmentOutput) []model.AttachmentResponse {
	resp := make([]model.AttachmentResponse, 0, len(outs))
	for _, a := range outs {
		resp = append(resp, model.AttachmentResponse{
			AttachmentID:  a.AttachmentID,
			FileName:      a.FileName,
			FileSizeBytes: a.FileSizeBytes,
			ContentType:   a.ContentType,
		})
	}
	return resp
}

func timelineOutputsToResponse(entries []model.TimelineEntryOutput) []model.TimelineEntryResponse {
	resp := make([]model.TimelineEntryResponse, 0, len(entries))
	for _, e := range entries {
		resp = append(resp, model.TimelineEntryResponse{
			EntryID:     e.EntryID,
			EntryType:   e.EntryType,
			Visibility:  e.Visibility,
			ActorUserID: e.ActorUserID,
			ActorRole:   e.ActorRole,
			Message:     e.Message,
			FromStatus:  e.FromStatus,
			ToStatus:    e.ToStatus,
			CreatedTime: e.CreatedTime,
			Attachments: attachmentOutputsToResponse(e.Attachments),
		})
	}
	return resp
}

func grievanceOutputToResponse(out *model.GrievanceOutput) *model.GrievanceResponse {
	if out == nil {
		return nil
	}
	return &model.GrievanceResponse{
		GrievanceID:      out.GrievanceID,
		ReferenceID:      out.ReferenceID,
		Category:         out.Category,
		Priority:         out.Priority,
		Status:           out.Status,
		Description:      out.Description,
		UserID:           out.UserID,
		SubmittedTime:    out.SubmittedTime,
		UpdatedTime:      out.UpdatedTime,
		StatutoryDueTime: out.StatutoryDueTime,
		Timeline:         timelineOutputsToResponse(out.Timeline),
		Attachments:      attachmentOutputsToResponse(out.Attachments),
	}
}

func listOutputToResponse(out *model.ListOutput) *model.GrievanceListResponse {
	data := make([]model.GrievanceResponse, 0, len(out.Data))
	for i := range out.Data {
		data = append(data, *grievanceOutputToResponse(&out.Data[i]))
	}
	return &model.GrievanceListResponse{
		Data: data,
		Metadata: model.GrievanceListMetadata{
			Total:    out.Total,
			Page:     out.Page,
			PageSize: out.PageSize,
		},
	}
}
