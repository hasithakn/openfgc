/*
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
 * Licensed under the Apache License, Version 2.0.
 */

import type {
  GrievanceAttachment,
  GrievanceAttachmentAPI,
  GrievanceDetail,
  GrievanceDetailAPI,
  GrievanceRecord,
  GrievanceTimelineEntry,
  GrievanceTimelineEntryAPI,
} from '../../../types/grievance'
import { formatAttachmentSize } from '../utils/grievanceDisplay'

function toAttachment(api: GrievanceAttachmentAPI): GrievanceAttachment {
  return {
    id: api.attachmentId,
    fileName: api.fileName,
    fileSizeLabel: formatAttachmentSize(api.fileSizeBytes),
  }
}

function toActorName(entry: GrievanceTimelineEntryAPI): string {
  return entry.actorRole === 'System' ? 'System' : (entry.actorUserId ?? 'Unknown')
}

function toTimelineEntry(api: GrievanceTimelineEntryAPI): GrievanceTimelineEntry {
  return {
    id: api.entryId,
    type: api.entryType,
    actorName: toActorName(api),
    actorRole: api.actorRole,
    message: api.message,
    timestamp: new Date(api.createdTime).toISOString(),
    visibility: api.visibility,
    fromStatus: api.fromStatus,
    toStatus: api.toStatus,
    attachments: api.attachments?.map(toAttachment),
  }
}

/** Maps a grievance API response to the frontend's list-row shape (no timeline/description). */
export function toGrievanceRecord(api: GrievanceDetailAPI): GrievanceRecord {
  return {
    id: api.grievanceId,
    referenceId: api.referenceId,
    category: api.category,
    priority: api.priority,
    status: api.status,
    dataPrincipalName: api.userId,
    submittedAt: new Date(api.submittedTime).toISOString(),
    updatedAt: new Date(api.updatedTime).toISOString(),
    statutoryDueDate: new Date(api.statutoryDueTime).toISOString(),
  }
}

/** Maps a grievance API response to the frontend's full detail shape (with timeline/attachments). */
export function toGrievanceDetail(api: GrievanceDetailAPI): GrievanceDetail {
  return {
    ...toGrievanceRecord(api),
    description: api.description ?? '',
    attachments: (api.attachments ?? []).map(toAttachment),
    timeline: (api.timeline ?? []).map(toTimelineEntry),
  }
}
