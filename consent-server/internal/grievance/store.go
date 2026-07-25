/*
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
 * Licensed under the Apache License, Version 2.0.
 */

package grievance

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/wso2/openfgc/internal/grievance/model"
	dbconst "github.com/wso2/openfgc/internal/system/database/constants"
	dbmodel "github.com/wso2/openfgc/internal/system/database/model"
	"github.com/wso2/openfgc/internal/system/database/provider"
	dbutils "github.com/wso2/openfgc/internal/system/database/utils"
	"github.com/wso2/openfgc/internal/system/stores/interfaces"
)

const grievanceColumns = "GRIEVANCE_ID, ORG_ID, USER_ID, REFERENCE_ID, CATEGORY, PRIORITY, STATUS, DESCRIPTION, SUBMITTED_TIME, UPDATED_TIME, STATUTORY_DUE_TIME"

const timelineColumns = "ENTRY_ID, ORG_ID, GRIEVANCE_ID, ENTRY_TYPE, VISIBILITY, ACTOR_USER_ID, ACTOR_ROLE, MESSAGE, FROM_STATUS, TO_STATUS, CREATED_TIME"

const attachmentMetaColumns = "ATTACHMENT_ID, ORG_ID, GRIEVANCE_ID, TIMELINE_ENTRY_ID, FILE_NAME, FILE_SIZE_BYTES, CONTENT_TYPE, CREATED_TIME"

const attachmentFullColumns = attachmentMetaColumns + ", FILE_DATA"

var (
	queryCreateGrievance = dbmodel.DBQuery{
		ID:            "CREATE_GRIEVANCE",
		Query:         "INSERT INTO GRIEVANCE (" + grievanceColumns + ") VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		PostgresQuery: "INSERT INTO GRIEVANCE (" + grievanceColumns + ") VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)",
	}

	queryGetGrievanceByID = dbmodel.DBQuery{
		ID:            "GET_GRIEVANCE_BY_ID",
		Query:         "SELECT " + grievanceColumns + " FROM GRIEVANCE WHERE GRIEVANCE_ID = ? AND ORG_ID = ?",
		PostgresQuery: "SELECT " + grievanceColumns + " FROM GRIEVANCE WHERE GRIEVANCE_ID = $1 AND ORG_ID = $2",
	}

	queryGetGrievanceByIDForUpdate = dbmodel.DBQuery{
		ID:            "GET_GRIEVANCE_BY_ID_FOR_UPDATE",
		Query:         "SELECT " + grievanceColumns + " FROM GRIEVANCE WHERE GRIEVANCE_ID = ? AND ORG_ID = ? FOR UPDATE",
		PostgresQuery: "SELECT " + grievanceColumns + " FROM GRIEVANCE WHERE GRIEVANCE_ID = $1 AND ORG_ID = $2 FOR UPDATE",
		SQLiteQuery:   "SELECT " + grievanceColumns + " FROM GRIEVANCE WHERE GRIEVANCE_ID = ? AND ORG_ID = ?",
	}

	queryUpdateGrievanceStatus = dbmodel.DBQuery{
		ID:            "UPDATE_GRIEVANCE_STATUS",
		Query:         "UPDATE GRIEVANCE SET STATUS = ?, UPDATED_TIME = ? WHERE GRIEVANCE_ID = ? AND ORG_ID = ?",
		PostgresQuery: "UPDATE GRIEVANCE SET STATUS = $1, UPDATED_TIME = $2 WHERE GRIEVANCE_ID = $3 AND ORG_ID = $4",
	}

	queryTouchGrievanceUpdatedTime = dbmodel.DBQuery{
		ID:            "TOUCH_GRIEVANCE_UPDATED_TIME",
		Query:         "UPDATE GRIEVANCE SET UPDATED_TIME = ? WHERE GRIEVANCE_ID = ? AND ORG_ID = ?",
		PostgresQuery: "UPDATE GRIEVANCE SET UPDATED_TIME = $1 WHERE GRIEVANCE_ID = $2 AND ORG_ID = $3",
	}

	queryCreateTimelineEntry = dbmodel.DBQuery{
		ID:            "CREATE_GRIEVANCE_TIMELINE_ENTRY",
		Query:         "INSERT INTO GRIEVANCE_TIMELINE_ENTRY (" + timelineColumns + ") VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		PostgresQuery: "INSERT INTO GRIEVANCE_TIMELINE_ENTRY (" + timelineColumns + ") VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)",
	}

	queryGetTimelineByGrievanceIDAll = dbmodel.DBQuery{
		ID:            "GET_TIMELINE_BY_GRIEVANCE_ID_ALL",
		Query:         "SELECT " + timelineColumns + " FROM GRIEVANCE_TIMELINE_ENTRY WHERE GRIEVANCE_ID = ? AND ORG_ID = ? ORDER BY CREATED_TIME ASC",
		PostgresQuery: "SELECT " + timelineColumns + " FROM GRIEVANCE_TIMELINE_ENTRY WHERE GRIEVANCE_ID = $1 AND ORG_ID = $2 ORDER BY CREATED_TIME ASC",
	}

	queryGetTimelineByGrievanceIDShared = dbmodel.DBQuery{
		ID:            "GET_TIMELINE_BY_GRIEVANCE_ID_SHARED",
		Query:         "SELECT " + timelineColumns + " FROM GRIEVANCE_TIMELINE_ENTRY WHERE GRIEVANCE_ID = ? AND ORG_ID = ? AND VISIBILITY = 'SHARED' ORDER BY CREATED_TIME ASC",
		PostgresQuery: "SELECT " + timelineColumns + " FROM GRIEVANCE_TIMELINE_ENTRY WHERE GRIEVANCE_ID = $1 AND ORG_ID = $2 AND VISIBILITY = 'SHARED' ORDER BY CREATED_TIME ASC",
	}

	queryCreateAttachment = dbmodel.DBQuery{
		ID:            "CREATE_GRIEVANCE_ATTACHMENT",
		Query:         "INSERT INTO GRIEVANCE_ATTACHMENT (" + attachmentFullColumns + ") VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		PostgresQuery: "INSERT INTO GRIEVANCE_ATTACHMENT (" + attachmentFullColumns + ") VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)",
	}

	queryGetAttachmentsMetaByGrievanceID = dbmodel.DBQuery{
		ID:            "GET_ATTACHMENTS_META_BY_GRIEVANCE_ID",
		Query:         "SELECT " + attachmentMetaColumns + " FROM GRIEVANCE_ATTACHMENT WHERE GRIEVANCE_ID = ? AND ORG_ID = ? ORDER BY CREATED_TIME ASC",
		PostgresQuery: "SELECT " + attachmentMetaColumns + " FROM GRIEVANCE_ATTACHMENT WHERE GRIEVANCE_ID = $1 AND ORG_ID = $2 ORDER BY CREATED_TIME ASC",
	}

	queryGetAttachmentContentByID = dbmodel.DBQuery{
		ID:            "GET_ATTACHMENT_CONTENT_BY_ID",
		Query:         "SELECT " + attachmentFullColumns + " FROM GRIEVANCE_ATTACHMENT WHERE ATTACHMENT_ID = ? AND ORG_ID = ?",
		PostgresQuery: "SELECT " + attachmentFullColumns + " FROM GRIEVANCE_ATTACHMENT WHERE ATTACHMENT_ID = $1 AND ORG_ID = $2",
	}

	queryGetReferenceCounterForUpdate = dbmodel.DBQuery{
		ID:            "GET_GRIEVANCE_REFERENCE_COUNTER_FOR_UPDATE",
		Query:         "SELECT NEXT_SEQUENCE FROM GRIEVANCE_REFERENCE_COUNTER WHERE ORG_ID = ? AND YEAR_VALUE = ? FOR UPDATE",
		PostgresQuery: "SELECT NEXT_SEQUENCE FROM GRIEVANCE_REFERENCE_COUNTER WHERE ORG_ID = $1 AND YEAR_VALUE = $2 FOR UPDATE",
		SQLiteQuery:   "SELECT NEXT_SEQUENCE FROM GRIEVANCE_REFERENCE_COUNTER WHERE ORG_ID = ? AND YEAR_VALUE = ?",
	}

	queryInsertReferenceCounter = dbmodel.DBQuery{
		ID:            "INSERT_GRIEVANCE_REFERENCE_COUNTER",
		Query:         "INSERT INTO GRIEVANCE_REFERENCE_COUNTER (ORG_ID, YEAR_VALUE, NEXT_SEQUENCE) VALUES (?, ?, ?)",
		PostgresQuery: "INSERT INTO GRIEVANCE_REFERENCE_COUNTER (ORG_ID, YEAR_VALUE, NEXT_SEQUENCE) VALUES ($1, $2, $3)",
	}

	queryUpdateReferenceCounter = dbmodel.DBQuery{
		ID:            "UPDATE_GRIEVANCE_REFERENCE_COUNTER",
		Query:         "UPDATE GRIEVANCE_REFERENCE_COUNTER SET NEXT_SEQUENCE = ? WHERE ORG_ID = ? AND YEAR_VALUE = ?",
		PostgresQuery: "UPDATE GRIEVANCE_REFERENCE_COUNTER SET NEXT_SEQUENCE = $1 WHERE ORG_ID = $2 AND YEAR_VALUE = $3",
	}

	queryCountGrievancesByStatus = dbmodel.DBQuery{
		ID:            "COUNT_GRIEVANCES_BY_STATUS",
		Query:         "SELECT STATUS, COUNT(*) AS count FROM GRIEVANCE WHERE ORG_ID = ? GROUP BY STATUS",
		PostgresQuery: "SELECT STATUS, COUNT(*) AS count FROM GRIEVANCE WHERE ORG_ID = $1 GROUP BY STATUS",
	}

	queryCountSLABreached = dbmodel.DBQuery{
		ID:            "COUNT_GRIEVANCES_SLA_BREACHED",
		Query:         "SELECT COUNT(*) AS count FROM GRIEVANCE WHERE ORG_ID = ? AND STATUTORY_DUE_TIME < ? AND STATUS != 'Resolved'",
		PostgresQuery: "SELECT COUNT(*) AS count FROM GRIEVANCE WHERE ORG_ID = $1 AND STATUTORY_DUE_TIME < $2 AND STATUS != 'Resolved'",
	}
)

// store implements interfaces.GrievanceStore.
type store struct{}

// NewGrievanceStore creates a new grievance store.
func NewGrievanceStore() interfaces.GrievanceStore {
	return &store{}
}

func (s *store) getDBClient() (provider.DBClientInterface, error) {
	return provider.GetDBProvider().GetConsentDBClient()
}

// =============================================================================
// Write operations (transactional)
// =============================================================================

func (s *store) Create(tx dbmodel.TxInterface, g *model.Grievance) error {
	_, err := tx.Exec(queryCreateGrievance,
		g.GrievanceID, g.OrgID, g.UserID, g.ReferenceID, g.Category, g.Priority, g.Status,
		g.Description, g.SubmittedTime, g.UpdatedTime, g.StatutoryDueTime)
	return err
}

func (s *store) UpdateStatus(tx dbmodel.TxInterface, grievanceID, orgID, status string, updatedTime int64) error {
	result, err := tx.Exec(queryUpdateGrievanceStatus, status, updatedTime, grievanceID, orgID)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("no grievance found with GRIEVANCE_ID=%s and ORG_ID=%s", grievanceID, orgID)
	}
	return nil
}

func (s *store) TouchUpdatedTime(tx dbmodel.TxInterface, grievanceID, orgID string, updatedTime int64) error {
	_, err := tx.Exec(queryTouchGrievanceUpdatedTime, updatedTime, grievanceID, orgID)
	return err
}

func (s *store) CreateTimelineEntry(tx dbmodel.TxInterface, e *model.TimelineEntry) error {
	_, err := tx.Exec(queryCreateTimelineEntry,
		e.EntryID, e.OrgID, e.GrievanceID, e.EntryType, e.Visibility, e.ActorUserID,
		e.ActorRole, e.Message, e.FromStatus, e.ToStatus, e.CreatedTime)
	return err
}

func (s *store) CreateAttachment(tx dbmodel.TxInterface, a *model.AttachmentContent) error {
	_, err := tx.Exec(queryCreateAttachment,
		a.AttachmentID, a.OrgID, a.GrievanceID, a.TimelineEntryID, a.FileName,
		a.FileSizeBytes, a.ContentType, a.CreatedTime, a.FileData)
	return err
}

// NextReferenceSequence returns the next sequence number for the given org+year,
// creating the counter row on first use, and increments it for the next caller.
// Must be called within the same transaction as the grievance insert it backs.
func (s *store) NextReferenceSequence(tx dbmodel.TxInterface, orgID, yearValue string) (int64, error) {
	rows, err := queryRowsInTx(tx, queryGetReferenceCounterForUpdate, orgID, yearValue)
	if err != nil {
		return 0, err
	}
	if len(rows) == 0 {
		if _, err := tx.Exec(queryInsertReferenceCounter, orgID, yearValue, int64(2)); err != nil {
			return 0, err
		}
		return 1, nil
	}
	current := getInt64(rows[0], "next_sequence")
	if _, err := tx.Exec(queryUpdateReferenceCounter, current+1, orgID, yearValue); err != nil {
		return 0, err
	}
	return current, nil
}

// =============================================================================
// Read operations
// =============================================================================

func (s *store) GetByID(ctx context.Context, grievanceID, orgID string) (*model.Grievance, error) {
	dbClient, err := s.getDBClient()
	if err != nil {
		return nil, fmt.Errorf("failed to get database client: %w", err)
	}
	rows, err := dbClient.Query(queryGetGrievanceByID, grievanceID, orgID)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	return mapToGrievance(rows[0]), nil
}

func (s *store) GetByIDForUpdate(tx dbmodel.TxInterface, grievanceID, orgID string) (*model.Grievance, error) {
	rows, err := queryRowsInTx(tx, queryGetGrievanceByIDForUpdate, grievanceID, orgID)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	return mapToGrievance(rows[0]), nil
}

func (s *store) GetTimelineByGrievanceID(ctx context.Context, grievanceID, orgID string, includeInternal bool) ([]model.TimelineEntry, error) {
	dbClient, err := s.getDBClient()
	if err != nil {
		return nil, fmt.Errorf("failed to get database client: %w", err)
	}
	query := queryGetTimelineByGrievanceIDShared
	if includeInternal {
		query = queryGetTimelineByGrievanceIDAll
	}
	rows, err := dbClient.Query(query, grievanceID, orgID)
	if err != nil {
		return nil, err
	}
	entries := make([]model.TimelineEntry, 0, len(rows))
	for _, row := range rows {
		entries = append(entries, mapToTimelineEntry(row))
	}
	return entries, nil
}

func (s *store) GetAttachmentsByGrievanceID(ctx context.Context, grievanceID, orgID string) ([]model.Attachment, error) {
	dbClient, err := s.getDBClient()
	if err != nil {
		return nil, fmt.Errorf("failed to get database client: %w", err)
	}
	rows, err := dbClient.Query(queryGetAttachmentsMetaByGrievanceID, grievanceID, orgID)
	if err != nil {
		return nil, err
	}
	attachments := make([]model.Attachment, 0, len(rows))
	for _, row := range rows {
		attachments = append(attachments, mapToAttachment(row))
	}
	return attachments, nil
}

func (s *store) GetAttachmentContent(ctx context.Context, attachmentID, orgID string) (*model.AttachmentContent, error) {
	dbClient, err := s.getDBClient()
	if err != nil {
		return nil, fmt.Errorf("failed to get database client: %w", err)
	}
	rows, err := dbClient.Query(queryGetAttachmentContentByID, attachmentID, orgID)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	return mapToAttachmentContent(rows[0]), nil
}

func (s *store) CountByStatus(ctx context.Context, orgID string) (map[string]int, error) {
	dbClient, err := s.getDBClient()
	if err != nil {
		return nil, fmt.Errorf("failed to get database client: %w", err)
	}
	rows, err := dbClient.Query(queryCountGrievancesByStatus, orgID)
	if err != nil {
		return nil, err
	}
	result := make(map[string]int, len(rows))
	for _, row := range rows {
		result[getString(row, "status")] = int(extractCount(row))
	}
	return result, nil
}

func (s *store) CountSLABreached(ctx context.Context, orgID string, now int64) (int, error) {
	dbClient, err := s.getDBClient()
	if err != nil {
		return 0, fmt.Errorf("failed to get database client: %w", err)
	}
	rows, err := dbClient.Query(queryCountSLABreached, orgID, now)
	if err != nil {
		return 0, err
	}
	if len(rows) == 0 {
		return 0, nil
	}
	return int(extractCount(rows[0])), nil
}

// Search returns grievances matching the filter along with the total match count for pagination.
func (s *store) Search(ctx context.Context, filter model.SearchFilter) ([]model.Grievance, int, error) {
	dbClient, err := s.getDBClient()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get database client: %w", err)
	}

	where := []string{"ORG_ID = ?"}
	args := []interface{}{filter.OrgID}

	if len(filter.UserIDs) > 0 {
		ph := strings.Repeat("?,", len(filter.UserIDs))
		where = append(where, fmt.Sprintf("USER_ID IN (%s)", ph[:len(ph)-1]))
		for _, id := range filter.UserIDs {
			args = append(args, id)
		}
	}
	if filter.Status != "" {
		where = append(where, "STATUS = ?")
		args = append(args, filter.Status)
	}
	if filter.Priority != "" {
		where = append(where, "PRIORITY = ?")
		args = append(args, filter.Priority)
	}
	if filter.Query != "" {
		pattern, escapeClause := likePattern(dbClient, filter.Query)
		where = append(where, fmt.Sprintf("(REFERENCE_ID LIKE ?%s OR USER_ID LIKE ?%s)", escapeClause, escapeClause))
		args = append(args, pattern, pattern)
	}

	whereClause := strings.Join(where, " AND ")
	orderClause := grievanceOrderClause(filter.SortBy, filter.Order)

	countSQL := "SELECT COUNT(*) AS count FROM GRIEVANCE WHERE " + whereClause
	countQuery := dbmodel.DBQuery{ID: "SEARCH_GRIEVANCES_COUNT", Query: countSQL, PostgresQuery: dbutils.ConvertToPostgresParams(countSQL)}
	countRows, err := dbClient.Query(countQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	total := 0
	if len(countRows) > 0 {
		total = int(extractCount(countRows[0]))
	}

	dataArgs := append(append([]interface{}{}, args...), filter.PageSize, filter.Page*filter.PageSize)
	dataSQL := "SELECT " + grievanceColumns + " FROM GRIEVANCE WHERE " + whereClause + orderClause + " LIMIT ? OFFSET ?"
	dataQuery := dbmodel.DBQuery{ID: "SEARCH_GRIEVANCES_DATA", Query: dataSQL, PostgresQuery: dbutils.ConvertToPostgresParams(dataSQL)}
	rows, err := dbClient.Query(dataQuery, dataArgs...)
	if err != nil {
		return nil, 0, err
	}

	result := make([]model.Grievance, 0, len(rows))
	for _, row := range rows {
		result = append(result, *mapToGrievance(row))
	}
	return result, total, nil
}

func grievanceOrderClause(sortBy, order string) string {
	column := "UPDATED_TIME"
	switch sortBy {
	case "submittedTime":
		column = "SUBMITTED_TIME"
	case "statutoryDueTime":
		column = "STATUTORY_DUE_TIME"
	case "priority":
		column = "PRIORITY"
	}
	direction := "DESC"
	if strings.EqualFold(order, "asc") {
		direction = "ASC"
	}
	return fmt.Sprintf(" ORDER BY %s %s", column, direction)
}

// likePattern escapes a string for a LIKE search, mirroring the escaping convention
// already used by consent/store.go's consentLikePattern.
func likePattern(dbClient provider.DBClientInterface, value string) (pattern, escapeClause string) {
	var escaped string
	switch dbClient.GetDBType() {
	case dbconst.DatabaseTypeSQLite, dbconst.DatabaseTypePostgres:
		r := strings.NewReplacer("|", "||", "%", "|%", "_", "|_")
		escaped = r.Replace(value)
		escapeClause = " ESCAPE '|'"
	default: // MySQL
		r := strings.NewReplacer("%", "\\%", "_", "\\_")
		escaped = r.Replace(value)
	}
	return "%" + escaped + "%", escapeClause
}

// =============================================================================
// Row mapping helpers
// =============================================================================

func queryRowsInTx(tx dbmodel.TxInterface, query dbmodel.DBQuery, args ...interface{}) ([]map[string]interface{}, error) {
	rows, err := tx.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, 0)
	for rows.Next() {
		values := make([]interface{}, len(columns))
		pointers := make([]interface{}, len(columns))
		for i := range values {
			pointers[i] = &values[i]
		}
		if err := rows.Scan(pointers...); err != nil {
			return nil, err
		}

		row := make(map[string]interface{}, len(columns))
		for i, column := range columns {
			row[strings.ToLower(column)] = values[i]
		}
		result = append(result, row)
	}
	return result, rows.Err()
}

func mapToGrievance(row map[string]interface{}) *model.Grievance {
	if row == nil {
		return nil
	}
	return &model.Grievance{
		GrievanceID:      getString(row, "grievance_id"),
		OrgID:            getString(row, "org_id"),
		UserID:           getString(row, "user_id"),
		ReferenceID:      getString(row, "reference_id"),
		Category:         getString(row, "category"),
		Priority:         getString(row, "priority"),
		Status:           getString(row, "status"),
		Description:      getString(row, "description"),
		SubmittedTime:    getInt64(row, "submitted_time"),
		UpdatedTime:      getInt64(row, "updated_time"),
		StatutoryDueTime: getInt64(row, "statutory_due_time"),
	}
}

func mapToTimelineEntry(row map[string]interface{}) model.TimelineEntry {
	return model.TimelineEntry{
		EntryID:     getString(row, "entry_id"),
		OrgID:       getString(row, "org_id"),
		GrievanceID: getString(row, "grievance_id"),
		EntryType:   getString(row, "entry_type"),
		Visibility:  getString(row, "visibility"),
		ActorUserID: getStringPtr(row, "actor_user_id"),
		ActorRole:   getString(row, "actor_role"),
		Message:     getString(row, "message"),
		FromStatus:  getStringPtr(row, "from_status"),
		ToStatus:    getStringPtr(row, "to_status"),
		CreatedTime: getInt64(row, "created_time"),
	}
}

func mapToAttachment(row map[string]interface{}) model.Attachment {
	return model.Attachment{
		AttachmentID:    getString(row, "attachment_id"),
		OrgID:           getString(row, "org_id"),
		GrievanceID:     getString(row, "grievance_id"),
		TimelineEntryID: getStringPtr(row, "timeline_entry_id"),
		FileName:        getString(row, "file_name"),
		FileSizeBytes:   getInt64(row, "file_size_bytes"),
		ContentType:     getString(row, "content_type"),
		CreatedTime:     getInt64(row, "created_time"),
	}
}

func mapToAttachmentContent(row map[string]interface{}) *model.AttachmentContent {
	if row == nil {
		return nil
	}
	var data []byte
	switch v := row["file_data"].(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	}
	return &model.AttachmentContent{
		Attachment: mapToAttachment(row),
		FileData:   data,
	}
}

func extractCount(row map[string]interface{}) int64 {
	if v, ok := row["count"].(int64); ok {
		return v
	}
	if v, ok := row["count"].([]uint8); ok {
		if parsed, err := strconv.ParseInt(string(v), 10, 64); err == nil {
			return parsed
		}
	}
	return 0
}

func getString(row map[string]interface{}, key string) string {
	switch v := row[key].(type) {
	case string:
		return v
	case []byte:
		return string(v)
	}
	return ""
}

func getStringPtr(row map[string]interface{}, key string) *string {
	switch v := row[key].(type) {
	case string:
		if v == "" {
			return nil
		}
		return &v
	case []byte:
		if len(v) == 0 {
			return nil
		}
		s := string(v)
		return &s
	}
	return nil
}

func getInt64(row map[string]interface{}, key string) int64 {
	switch v := row[key].(type) {
	case int64:
		return v
	case int32:
		return int64(v)
	case int:
		return int64(v)
	case float64:
		return int64(v)
	case []uint8:
		if parsed, err := strconv.ParseInt(string(v), 10, 64); err == nil {
			return parsed
		}
	case string:
		if parsed, err := strconv.ParseInt(v, 10, 64); err == nil {
			return parsed
		}
	}
	return 0
}
