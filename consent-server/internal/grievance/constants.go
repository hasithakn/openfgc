/*
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
 * Licensed under the Apache License, Version 2.0.
 */

package grievance

import "fmt"

// DB-vocabulary priority values (stored in GRIEVANCE.PRIORITY).
const (
	priorityDBCritical = "Critical"
	priorityDBHigh     = "High"
	priorityDBMedium   = "Medium"
	priorityDBLow      = "Low"
)

// API-vocabulary priority values (matches the frontend's existing P0-P3 chip/i18n keys).
const (
	PriorityAPIP0 = "P0"
	PriorityAPIP1 = "P1"
	PriorityAPIP2 = "P2"
	PriorityAPIP3 = "P3"
)

var priorityDBToAPI = map[string]string{
	priorityDBCritical: PriorityAPIP0,
	priorityDBHigh:     PriorityAPIP1,
	priorityDBMedium:   PriorityAPIP2,
	priorityDBLow:      PriorityAPIP3,
}

var priorityAPIToDB = map[string]string{
	PriorityAPIP0: priorityDBCritical,
	PriorityAPIP1: priorityDBHigh,
	PriorityAPIP2: priorityDBMedium,
	PriorityAPIP3: priorityDBLow,
}

// DB-vocabulary status values (stored in GRIEVANCE.STATUS).
const (
	statusDBOpen         = "Open"
	statusDBInProgress   = "In Progress"
	statusDBWaitingInfo  = "Waiting Info"
	statusDBWaitingOnDPO = "Waiting on DPO"
	statusDBResolved     = "Resolved"
)

// API-vocabulary status values (matches the frontend's existing GrievanceStatus enum).
const (
	StatusAPIOpen          = "Open"
	StatusAPIInvestigation = "Investigation"
	StatusAPIAwaitingInfo  = "AwaitingInfo"
	StatusAPIWaitingOnDpo  = "WaitingOnDpo"
	StatusAPIResolved      = "Resolved"
)

var statusDBToAPI = map[string]string{
	statusDBOpen:         StatusAPIOpen,
	statusDBInProgress:   StatusAPIInvestigation,
	statusDBWaitingInfo:  StatusAPIAwaitingInfo,
	statusDBWaitingOnDPO: StatusAPIWaitingOnDpo,
	statusDBResolved:     StatusAPIResolved,
}

var statusAPIToDB = map[string]string{
	StatusAPIOpen:          statusDBOpen,
	StatusAPIInvestigation: statusDBInProgress,
	StatusAPIAwaitingInfo:  statusDBWaitingInfo,
	StatusAPIWaitingOnDpo:  statusDBWaitingOnDPO,
	StatusAPIResolved:      statusDBResolved,
}

// nextStatusesDB is the server-enforced status transition table (DB vocabulary).
// Reopening from statusDBResolved via a Data Principal reply is a deliberate exception
// handled in service.go, not part of this table.
var nextStatusesDB = map[string][]string{
	statusDBOpen:         {statusDBInProgress, statusDBWaitingInfo},
	statusDBInProgress:   {statusDBWaitingInfo, statusDBResolved},
	statusDBWaitingInfo:  {statusDBInProgress},
	statusDBWaitingOnDPO: {statusDBInProgress, statusDBWaitingInfo, statusDBResolved},
	statusDBResolved:     {},
}

// DB-vocabulary timeline entry types (stored in GRIEVANCE_TIMELINE_ENTRY.ENTRY_TYPE).
const (
	entryTypeDBSystemAck     = "SYSTEM_ACK"
	entryTypeDBStatusChange  = "STATUS_CHANGE"
	entryTypeDBResolution    = "RESOLUTION"
	entryTypeDBCommunication = "COMMUNICATION"
	entryTypeDBNote          = "NOTE"
)

// API-vocabulary timeline entry types (matches the frontend's GrievanceTimelineEntryType).
const (
	entryTypeAPISystemAck     = "systemAcknowledgement"
	entryTypeAPIStatusChange  = "statusChange"
	entryTypeAPIResolution    = "resolution"
	entryTypeAPICommunication = "communication"
	entryTypeAPINote          = "note"
)

var entryTypeDBToAPI = map[string]string{
	entryTypeDBSystemAck:     entryTypeAPISystemAck,
	entryTypeDBStatusChange:  entryTypeAPIStatusChange,
	entryTypeDBResolution:    entryTypeAPIResolution,
	entryTypeDBCommunication: entryTypeAPICommunication,
	entryTypeDBNote:          entryTypeAPINote,
}

// DB-vocabulary visibility values (stored in GRIEVANCE_TIMELINE_ENTRY.VISIBILITY).
const (
	visibilityDBShared   = "SHARED"
	visibilityDBInternal = "INTERNAL"
)

// API-vocabulary visibility values.
const (
	VisibilityAPIShared   = "shared"
	VisibilityAPIInternal = "internal"
)

var visibilityDBToAPI = map[string]string{
	visibilityDBShared:   VisibilityAPIShared,
	visibilityDBInternal: VisibilityAPIInternal,
}

var visibilityAPIToDB = map[string]string{
	VisibilityAPIShared:   visibilityDBShared,
	VisibilityAPIInternal: visibilityDBInternal,
}

// DB-vocabulary actor role values (stored in GRIEVANCE_TIMELINE_ENTRY.ACTOR_ROLE).
const (
	actorRoleDBDataPrincipal    = "DATA_PRINCIPAL"
	actorRoleDBGrievanceOfficer = "GRIEVANCE_OFFICER"
	actorRoleDBSystem           = "SYSTEM"
)

// API-vocabulary actor role values (matches the frontend's GrievanceActorRole).
const (
	ActorRoleAPIDataPrincipal    = "DataPrincipal"
	ActorRoleAPIGrievanceOfficer = "GrievanceOfficer"
	actorRoleAPISystem           = "System"
)

var actorRoleDBToAPI = map[string]string{
	actorRoleDBDataPrincipal:    ActorRoleAPIDataPrincipal,
	actorRoleDBGrievanceOfficer: ActorRoleAPIGrievanceOfficer,
	actorRoleDBSystem:           actorRoleAPISystem,
}

var actorRoleAPIToDB = map[string]string{
	ActorRoleAPIDataPrincipal:    actorRoleDBDataPrincipal,
	ActorRoleAPIGrievanceOfficer: actorRoleDBGrievanceOfficer,
	actorRoleAPISystem:           actorRoleDBSystem,
}

// categoryToPriorityDB derives the DB priority value from the API category token at
// submission time. The 10 map keys are the complete set of valid categories.
var categoryToPriorityDB = map[string]string{
	"dataBreachComplaint":               priorityDBCritical,
	"unauthorizedDataSharing":           priorityDBCritical,
	"consentWithdrawalIssue":            priorityDBHigh,
	"purposeViolation":                  priorityDBHigh,
	"dataErasureRequestNotFulfilled":    priorityDBHigh,
	"dataCorrectionRequestNotFulfilled": priorityDBMedium,
	"consentManagerIssue":               priorityDBMedium,
	"dataAccessRequestDenial":           priorityDBMedium,
	"excessiveDataCollection":           priorityDBLow,
	"other":                             priorityDBLow,
}

// statutoryDueDays is the number of days after submission a grievance's statutory due date falls.
const statutoryDueDays = 90

// maxAttachmentSizeBytes is the maximum accepted size (in bytes) for a single attachment.
const maxAttachmentSizeBytes = 10 * 1024 * 1024

func mapValue(m map[string]string, key, kind string) (string, error) {
	v, ok := m[key]
	if !ok {
		return "", fmt.Errorf("unknown %s %q", kind, key)
	}
	return v, nil
}
