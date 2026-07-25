/*
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
 * Licensed under the Apache License, Version 2.0.
 */

package grievance

import (
	"context"
	"encoding/base64"
	"fmt"
	"strconv"
	"time"

	"github.com/wso2/openfgc/internal/grievance/model"
	dbmodel "github.com/wso2/openfgc/internal/system/database/model"
	"github.com/wso2/openfgc/internal/system/error/serviceerror"
	"github.com/wso2/openfgc/internal/system/log"
	"github.com/wso2/openfgc/internal/system/stores"
	"github.com/wso2/openfgc/internal/system/utils"
)

// GrievanceService defines the exported service interface.
type GrievanceService interface {
	CreateGrievance(ctx context.Context, input model.CreateGrievanceInput, orgID string) (*model.GrievanceOutput, *serviceerror.ServiceError)
	GetGrievance(ctx context.Context, grievanceID, orgID string, includeInternal bool) (*model.GrievanceOutput, *serviceerror.ServiceError)
	ListGrievances(ctx context.Context, filter model.SearchFilter) (*model.ListOutput, *serviceerror.ServiceError)
	PostMessage(ctx context.Context, grievanceID, orgID string, input model.PostMessageInput) (*model.GrievanceOutput, *serviceerror.ServiceError)
	TransitionStatus(ctx context.Context, grievanceID, orgID string, input model.TransitionStatusInput) (*model.GrievanceOutput, *serviceerror.ServiceError)
	GetStats(ctx context.Context, orgID string) (*model.StatsOutput, *serviceerror.ServiceError)
	GetAttachmentContent(ctx context.Context, grievanceID, attachmentID, orgID string) (*model.AttachmentContent, *serviceerror.ServiceError)
}

type grievanceService struct {
	stores *stores.StoreRegistry
}

func newGrievanceService(registry *stores.StoreRegistry) GrievanceService {
	return &grievanceService{stores: registry}
}

// =============================================================================
// CreateGrievance
// =============================================================================

// CreateGrievance submits a new grievance, deriving priority from category and
// generating a human-facing sequential reference ID scoped to org+year.
func (s *grievanceService) CreateGrievance(ctx context.Context, input model.CreateGrievanceInput, orgID string) (*model.GrievanceOutput, *serviceerror.ServiceError) {
	logger := log.GetLogger().WithContext(ctx)

	priorityDB, ok := categoryToPriorityDB[input.Category]
	if !ok {
		return nil, serviceerror.CustomServiceError(ErrorValidationFailed,
			fmt.Sprintf("category %q is not a recognized grievance category", input.Category))
	}

	grievanceID := utils.GenerateUUID()
	currentTime := utils.GetCurrentTimeMillis()
	yearValue := strconv.Itoa(time.UnixMilli(currentTime).UTC().Year())
	dueTime := currentTime + int64(statutoryDueDays)*24*3600*1000

	g := &model.Grievance{
		GrievanceID:      grievanceID,
		OrgID:            orgID,
		UserID:           input.UserID,
		Category:         input.Category,
		Priority:         priorityDB,
		Status:           statusDBOpen,
		Description:      input.Description,
		SubmittedTime:    currentTime,
		UpdatedTime:      currentTime,
		StatutoryDueTime: dueTime,
	}

	attachmentRows, err := buildAttachmentRows(input.Attachments, grievanceID, orgID, nil, currentTime)
	if err != nil {
		return nil, serviceerror.CustomServiceError(ErrorValidationFailed, err.Error())
	}

	ackEntry := &model.TimelineEntry{
		EntryID:     utils.GenerateUUID(),
		OrgID:       orgID,
		GrievanceID: grievanceID,
		EntryType:   entryTypeDBSystemAck,
		Visibility:  visibilityDBShared,
		ActorRole:   actorRoleDBSystem,
		Message:     "Grievance received and acknowledged.",
		CreatedTime: currentTime,
	}

	queries := []func(tx dbmodel.TxInterface) error{
		func(tx dbmodel.TxInterface) error {
			seq, seqErr := s.stores.Grievance.NextReferenceSequence(tx, orgID, yearValue)
			if seqErr != nil {
				return seqErr
			}
			g.ReferenceID = fmt.Sprintf("GRV-%s-%05d", yearValue, seq)
			return s.stores.Grievance.Create(tx, g)
		},
		func(tx dbmodel.TxInterface) error {
			return s.stores.Grievance.CreateTimelineEntry(tx, ackEntry)
		},
	}
	for i := range attachmentRows {
		a := attachmentRows[i]
		queries = append(queries, func(tx dbmodel.TxInterface) error {
			return s.stores.Grievance.CreateAttachment(tx, &a)
		})
	}

	if err := s.stores.ExecuteTransaction(queries); err != nil {
		logger.Error("Create grievance transaction failed", log.Error(err))
		return nil, serviceerror.CustomServiceError(ErrorInternalServerError,
			fmt.Sprintf("failed to create grievance: %v", err))
	}

	logger.Info("Grievance created", log.String("grievance_id", grievanceID), log.String("reference_id", g.ReferenceID))
	return s.loadGrievanceOutput(ctx, g, orgID, true)
}

// =============================================================================
// GetGrievance / ListGrievances
// =============================================================================

// GetGrievance retrieves a grievance with its timeline and attachments.
// When includeInternal is false, INTERNAL-visibility timeline entries (and any
// attachments carried only on them) are excluded server-side.
func (s *grievanceService) GetGrievance(ctx context.Context, grievanceID, orgID string, includeInternal bool) (*model.GrievanceOutput, *serviceerror.ServiceError) {
	g, err := s.stores.Grievance.GetByID(ctx, grievanceID, orgID)
	if err != nil {
		return nil, serviceerror.CustomServiceError(ErrorInternalServerError, err.Error())
	}
	if g == nil {
		return nil, serviceerror.CustomServiceError(ErrorGrievanceNotFound,
			fmt.Sprintf("grievance with ID '%s' not found", grievanceID))
	}
	return s.loadGrievanceOutput(ctx, g, orgID, includeInternal)
}

// ListGrievances returns grievances matching the filter (Status/Priority given in API vocabulary).
func (s *grievanceService) ListGrievances(ctx context.Context, filter model.SearchFilter) (*model.ListOutput, *serviceerror.ServiceError) {
	dbFilter := filter
	if filter.Status != "" {
		v, err := mapValue(statusAPIToDB, filter.Status, "status")
		if err != nil {
			return nil, serviceerror.CustomServiceError(ErrorValidationFailed, err.Error())
		}
		dbFilter.Status = v
	}
	if filter.Priority != "" {
		v, err := mapValue(priorityAPIToDB, filter.Priority, "priority")
		if err != nil {
			return nil, serviceerror.CustomServiceError(ErrorValidationFailed, err.Error())
		}
		dbFilter.Priority = v
	}
	if dbFilter.PageSize <= 0 {
		dbFilter.PageSize = 10
	}
	if dbFilter.Page < 0 {
		dbFilter.Page = 0
	}

	rows, total, err := s.stores.Grievance.Search(ctx, dbFilter)
	if err != nil {
		return nil, serviceerror.CustomServiceError(ErrorInternalServerError, err.Error())
	}

	data := make([]model.GrievanceOutput, 0, len(rows))
	for _, row := range rows {
		priorityAPI, err1 := mapValue(priorityDBToAPI, row.Priority, "priority")
		statusAPI, err2 := mapValue(statusDBToAPI, row.Status, "status")
		if err1 != nil || err2 != nil {
			log.GetLogger().WithContext(ctx).Warn("Skipping grievance with unrecognized stored value",
				log.String("grievance_id", row.GrievanceID))
			continue
		}
		data = append(data, model.GrievanceOutput{
			GrievanceID:      row.GrievanceID,
			ReferenceID:      row.ReferenceID,
			Category:         row.Category,
			Priority:         priorityAPI,
			Status:           statusAPI,
			Description:      row.Description,
			UserID:           row.UserID,
			SubmittedTime:    row.SubmittedTime,
			UpdatedTime:      row.UpdatedTime,
			StatutoryDueTime: row.StatutoryDueTime,
		})
	}

	return &model.ListOutput{Data: data, Total: total, Page: dbFilter.Page, PageSize: dbFilter.PageSize}, nil
}

// =============================================================================
// PostMessage
// =============================================================================

// PostMessage appends a reply (visibility=shared) or internal note (visibility=internal,
// GrievanceOfficer only) to a grievance's timeline. A Data Principal's shared reply
// auto-transitions the grievance to "Waiting on DPO" (including reopening a Resolved
// grievance), a deliberate exception to the normal transition table.
func (s *grievanceService) PostMessage(ctx context.Context, grievanceID, orgID string, input model.PostMessageInput) (*model.GrievanceOutput, *serviceerror.ServiceError) {
	logger := log.GetLogger().WithContext(ctx)

	g, err := s.stores.Grievance.GetByID(ctx, grievanceID, orgID)
	if err != nil {
		return nil, serviceerror.CustomServiceError(ErrorInternalServerError, err.Error())
	}
	if g == nil {
		return nil, serviceerror.CustomServiceError(ErrorGrievanceNotFound,
			fmt.Sprintf("grievance with ID '%s' not found", grievanceID))
	}

	actorRoleDB, mapErr := mapValue(actorRoleAPIToDB, input.ActorRole, "actorRole")
	if mapErr != nil {
		return nil, serviceerror.CustomServiceError(ErrorValidationFailed, mapErr.Error())
	}
	visibilityDB, mapErr := mapValue(visibilityAPIToDB, input.Visibility, "visibility")
	if mapErr != nil {
		return nil, serviceerror.CustomServiceError(ErrorValidationFailed, mapErr.Error())
	}
	if visibilityDB == visibilityDBInternal && actorRoleDB != actorRoleDBGrievanceOfficer {
		return nil, serviceerror.CustomServiceError(ErrorValidationFailed,
			"only a Grievance Officer may post an internal-visibility entry")
	}

	entryType := entryTypeDBCommunication
	if visibilityDB == visibilityDBInternal {
		entryType = entryTypeDBNote
	}

	entryID := utils.GenerateUUID()
	currentTime := utils.GetCurrentTimeMillis()
	actorUserID := input.ActorUserID

	messageEntry := &model.TimelineEntry{
		EntryID:     entryID,
		OrgID:       orgID,
		GrievanceID: grievanceID,
		EntryType:   entryType,
		Visibility:  visibilityDB,
		ActorUserID: &actorUserID,
		ActorRole:   actorRoleDB,
		Message:     input.Message,
		CreatedTime: currentTime,
	}

	attachmentRows, buildErr := buildAttachmentRows(input.Attachments, grievanceID, orgID, &entryID, currentTime)
	if buildErr != nil {
		return nil, serviceerror.CustomServiceError(ErrorValidationFailed, buildErr.Error())
	}

	queries := []func(tx dbmodel.TxInterface) error{
		func(tx dbmodel.TxInterface) error {
			return s.stores.Grievance.CreateTimelineEntry(tx, messageEntry)
		},
	}
	for i := range attachmentRows {
		a := attachmentRows[i]
		queries = append(queries, func(tx dbmodel.TxInterface) error {
			return s.stores.Grievance.CreateAttachment(tx, &a)
		})
	}

	autoTransition := actorRoleDB == actorRoleDBDataPrincipal && visibilityDB == visibilityDBShared && g.Status != statusDBWaitingOnDPO
	if autoTransition {
		fromStatus := g.Status
		toStatus := statusDBWaitingOnDPO
		transitionEntry := &model.TimelineEntry{
			EntryID:     utils.GenerateUUID(),
			OrgID:       orgID,
			GrievanceID: grievanceID,
			EntryType:   entryTypeDBStatusChange,
			Visibility:  visibilityDBShared,
			ActorRole:   actorRoleDBSystem,
			Message:     autoTransitionMessage(fromStatus),
			FromStatus:  &fromStatus,
			ToStatus:    &toStatus,
			CreatedTime: currentTime,
		}
		queries = append(queries,
			func(tx dbmodel.TxInterface) error {
				return s.stores.Grievance.CreateTimelineEntry(tx, transitionEntry)
			},
			func(tx dbmodel.TxInterface) error {
				return s.stores.Grievance.UpdateStatus(tx, grievanceID, orgID, statusDBWaitingOnDPO, currentTime)
			},
		)
	} else {
		queries = append(queries, func(tx dbmodel.TxInterface) error {
			return s.stores.Grievance.TouchUpdatedTime(tx, grievanceID, orgID, currentTime)
		})
	}

	if err := s.stores.ExecuteTransaction(queries); err != nil {
		logger.Error("Post message transaction failed", log.Error(err), log.String("grievance_id", grievanceID))
		return nil, serviceerror.CustomServiceError(ErrorInternalServerError,
			fmt.Sprintf("failed to post message: %v", err))
	}

	refreshed, err := s.stores.Grievance.GetByID(ctx, grievanceID, orgID)
	if err != nil || refreshed == nil {
		return nil, serviceerror.CustomServiceError(ErrorInternalServerError, "failed to reload grievance after posting message")
	}
	return s.loadGrievanceOutput(ctx, refreshed, orgID, actorRoleDB == actorRoleDBGrievanceOfficer)
}

func autoTransitionMessage(fromStatus string) string {
	if fromStatus == statusDBResolved {
		return "Grievance reopened following a Data Principal reply; status set to Waiting on DPO."
	}
	return "Status updated to Waiting on DPO following a Data Principal reply."
}

// =============================================================================
// TransitionStatus
// =============================================================================

// TransitionStatus applies an explicit status change, enforcing the transition
// table server-side (the frontend's dropdown is a convenience, not the authority).
func (s *grievanceService) TransitionStatus(ctx context.Context, grievanceID, orgID string, input model.TransitionStatusInput) (*model.GrievanceOutput, *serviceerror.ServiceError) {
	logger := log.GetLogger().WithContext(ctx)

	g, err := s.stores.Grievance.GetByID(ctx, grievanceID, orgID)
	if err != nil {
		return nil, serviceerror.CustomServiceError(ErrorInternalServerError, err.Error())
	}
	if g == nil {
		return nil, serviceerror.CustomServiceError(ErrorGrievanceNotFound,
			fmt.Sprintf("grievance with ID '%s' not found", grievanceID))
	}

	actorRoleDB, mapErr := mapValue(actorRoleAPIToDB, input.ActorRole, "actorRole")
	if mapErr != nil {
		return nil, serviceerror.CustomServiceError(ErrorValidationFailed, mapErr.Error())
	}
	nextStatusDB, mapErr := mapValue(statusAPIToDB, input.NextStatus, "nextStatus")
	if mapErr != nil {
		return nil, serviceerror.CustomServiceError(ErrorValidationFailed, mapErr.Error())
	}

	if !containsString(nextStatusesDB[g.Status], nextStatusDB) {
		return nil, serviceerror.CustomServiceError(ErrorInvalidStatusTransition,
			fmt.Sprintf("cannot transition from %q to %q", g.Status, nextStatusDB))
	}

	entryType := entryTypeDBStatusChange
	if nextStatusDB == statusDBResolved {
		entryType = entryTypeDBResolution
	}

	currentTime := utils.GetCurrentTimeMillis()
	fromStatus := g.Status
	actorUserID := input.ActorUserID
	entry := &model.TimelineEntry{
		EntryID:     utils.GenerateUUID(),
		OrgID:       orgID,
		GrievanceID: grievanceID,
		EntryType:   entryType,
		Visibility:  visibilityDBShared,
		ActorUserID: &actorUserID,
		ActorRole:   actorRoleDB,
		Message:     "",
		FromStatus:  &fromStatus,
		ToStatus:    &nextStatusDB,
		CreatedTime: currentTime,
	}

	err = s.stores.ExecuteTransaction([]func(tx dbmodel.TxInterface) error{
		func(tx dbmodel.TxInterface) error {
			return s.stores.Grievance.CreateTimelineEntry(tx, entry)
		},
		func(tx dbmodel.TxInterface) error {
			return s.stores.Grievance.UpdateStatus(tx, grievanceID, orgID, nextStatusDB, currentTime)
		},
	})
	if err != nil {
		logger.Error("Transition status transaction failed", log.Error(err), log.String("grievance_id", grievanceID))
		return nil, serviceerror.CustomServiceError(ErrorInternalServerError,
			fmt.Sprintf("failed to transition status: %v", err))
	}

	refreshed, err := s.stores.Grievance.GetByID(ctx, grievanceID, orgID)
	if err != nil || refreshed == nil {
		return nil, serviceerror.CustomServiceError(ErrorInternalServerError, "failed to reload grievance after status transition")
	}
	return s.loadGrievanceOutput(ctx, refreshed, orgID, actorRoleDB == actorRoleDBGrievanceOfficer)
}

// =============================================================================
// GetStats / GetAttachmentContent
// =============================================================================

// GetStats returns the officer queue's aggregate summary counts.
func (s *grievanceService) GetStats(ctx context.Context, orgID string) (*model.StatsOutput, *serviceerror.ServiceError) {
	counts, err := s.stores.Grievance.CountByStatus(ctx, orgID)
	if err != nil {
		return nil, serviceerror.CustomServiceError(ErrorInternalServerError, err.Error())
	}
	slaBreached, err := s.stores.Grievance.CountSLABreached(ctx, orgID, utils.GetCurrentTimeMillis())
	if err != nil {
		return nil, serviceerror.CustomServiceError(ErrorInternalServerError, err.Error())
	}
	return &model.StatsOutput{
		Open:         counts[statusDBOpen] + counts[statusDBInProgress] + counts[statusDBWaitingOnDPO],
		AwaitingInfo: counts[statusDBWaitingInfo],
		Resolved:     counts[statusDBResolved],
		SLABreached:  slaBreached,
	}, nil
}

// GetAttachmentContent retrieves an attachment's full content, verifying it belongs to grievanceID.
func (s *grievanceService) GetAttachmentContent(ctx context.Context, grievanceID, attachmentID, orgID string) (*model.AttachmentContent, *serviceerror.ServiceError) {
	content, err := s.stores.Grievance.GetAttachmentContent(ctx, attachmentID, orgID)
	if err != nil {
		return nil, serviceerror.CustomServiceError(ErrorInternalServerError, err.Error())
	}
	if content == nil || content.GrievanceID != grievanceID {
		return nil, serviceerror.CustomServiceError(ErrorAttachmentNotFound,
			fmt.Sprintf("attachment with ID '%s' not found", attachmentID))
	}
	return content, nil
}

// =============================================================================
// Private helpers
// =============================================================================

// loadGrievanceOutput fetches the timeline and attachments for a grievance and assembles
// the API-vocabulary GrievanceOutput. When includeInternal is false, INTERNAL-visibility
// entries and attachments carried only on them are excluded.
func (s *grievanceService) loadGrievanceOutput(ctx context.Context, g *model.Grievance, orgID string, includeInternal bool) (*model.GrievanceOutput, *serviceerror.ServiceError) {
	entries, err := s.stores.Grievance.GetTimelineByGrievanceID(ctx, g.GrievanceID, orgID, includeInternal)
	if err != nil {
		return nil, serviceerror.CustomServiceError(ErrorInternalServerError, err.Error())
	}
	attachments, err := s.stores.Grievance.GetAttachmentsByGrievanceID(ctx, g.GrievanceID, orgID)
	if err != nil {
		return nil, serviceerror.CustomServiceError(ErrorInternalServerError, err.Error())
	}

	visibleEntryIDs := make(map[string]struct{}, len(entries))
	for _, e := range entries {
		visibleEntryIDs[e.EntryID] = struct{}{}
	}

	topLevelAttachments := make([]model.AttachmentOutput, 0)
	attachmentsByEntry := make(map[string][]model.AttachmentOutput)
	for _, a := range attachments {
		out := model.AttachmentOutput{
			AttachmentID:  a.AttachmentID,
			FileName:      a.FileName,
			FileSizeBytes: a.FileSizeBytes,
			ContentType:   a.ContentType,
		}
		if a.TimelineEntryID == nil {
			topLevelAttachments = append(topLevelAttachments, out)
			continue
		}
		if _, ok := visibleEntryIDs[*a.TimelineEntryID]; ok {
			attachmentsByEntry[*a.TimelineEntryID] = append(attachmentsByEntry[*a.TimelineEntryID], out)
		}
	}

	timeline := make([]model.TimelineEntryOutput, 0, len(entries))
	for _, e := range entries {
		out, serviceErr := timelineEntryToOutput(e, attachmentsByEntry[e.EntryID])
		if serviceErr != nil {
			return nil, serviceErr
		}
		timeline = append(timeline, *out)
	}

	priorityAPI, err1 := mapValue(priorityDBToAPI, g.Priority, "priority")
	statusAPI, err2 := mapValue(statusDBToAPI, g.Status, "status")
	if err1 != nil || err2 != nil {
		return nil, serviceerror.CustomServiceError(ErrorInternalServerError,
			"stored grievance has an unrecognized priority/status value")
	}

	return &model.GrievanceOutput{
		GrievanceID:      g.GrievanceID,
		ReferenceID:      g.ReferenceID,
		Category:         g.Category,
		Priority:         priorityAPI,
		Status:           statusAPI,
		Description:      g.Description,
		UserID:           g.UserID,
		SubmittedTime:    g.SubmittedTime,
		UpdatedTime:      g.UpdatedTime,
		StatutoryDueTime: g.StatutoryDueTime,
		Timeline:         timeline,
		Attachments:      topLevelAttachments,
	}, nil
}

func timelineEntryToOutput(e model.TimelineEntry, attachments []model.AttachmentOutput) (*model.TimelineEntryOutput, *serviceerror.ServiceError) {
	entryTypeAPI, ok := entryTypeDBToAPI[e.EntryType]
	if !ok {
		return nil, serviceerror.CustomServiceError(ErrorInternalServerError, "stored timeline entry has an unrecognized entry type")
	}
	visibilityAPI, ok := visibilityDBToAPI[e.Visibility]
	if !ok {
		return nil, serviceerror.CustomServiceError(ErrorInternalServerError, "stored timeline entry has an unrecognized visibility")
	}
	actorRoleAPI, ok := actorRoleDBToAPI[e.ActorRole]
	if !ok {
		return nil, serviceerror.CustomServiceError(ErrorInternalServerError, "stored timeline entry has an unrecognized actor role")
	}
	var fromStatusAPI, toStatusAPI *string
	if e.FromStatus != nil {
		if v, ok := statusDBToAPI[*e.FromStatus]; ok {
			fromStatusAPI = &v
		}
	}
	if e.ToStatus != nil {
		if v, ok := statusDBToAPI[*e.ToStatus]; ok {
			toStatusAPI = &v
		}
	}
	return &model.TimelineEntryOutput{
		EntryID:     e.EntryID,
		EntryType:   entryTypeAPI,
		Visibility:  visibilityAPI,
		ActorUserID: e.ActorUserID,
		ActorRole:   actorRoleAPI,
		Message:     e.Message,
		FromStatus:  fromStatusAPI,
		ToStatus:    toStatusAPI,
		CreatedTime: e.CreatedTime,
		Attachments: attachments,
	}, nil
}

// buildAttachmentRows decodes base64 attachment content and assembles DB rows ready for insertion.
func buildAttachmentRows(inputs []model.AttachmentInput, grievanceID, orgID string, timelineEntryID *string, createdTime int64) ([]model.AttachmentContent, error) {
	rows := make([]model.AttachmentContent, 0, len(inputs))
	for _, in := range inputs {
		data, err := base64.StdEncoding.DecodeString(in.ContentBase64)
		if err != nil {
			return nil, fmt.Errorf("attachment %q: invalid base64 content: %w", in.FileName, err)
		}
		if int64(len(data)) > maxAttachmentSizeBytes {
			return nil, fmt.Errorf("attachment %q exceeds the maximum size of %d bytes", in.FileName, maxAttachmentSizeBytes)
		}
		rows = append(rows, model.AttachmentContent{
			Attachment: model.Attachment{
				AttachmentID:    utils.GenerateUUID(),
				OrgID:           orgID,
				GrievanceID:     grievanceID,
				TimelineEntryID: timelineEntryID,
				FileName:        in.FileName,
				FileSizeBytes:   int64(len(data)),
				ContentType:     in.ContentType,
				CreatedTime:     createdTime,
			},
			FileData: data,
		})
	}
	return rows, nil
}

func containsString(list []string, target string) bool {
	for _, v := range list {
		if v == target {
			return true
		}
	}
	return false
}
