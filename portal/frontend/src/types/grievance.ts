/*
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
 *
 * WSO2 LLC. licenses this file to you under the Apache License,
 * Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

export const GRIEVANCE_CATEGORIES = [
  'dataBreachComplaint',
  'unauthorizedDataSharing',
  'consentWithdrawalIssue',
  'purposeViolation',
  'dataErasureRequestNotFulfilled',
  'dataCorrectionRequestNotFulfilled',
  'consentManagerIssue',
  'dataAccessRequestDenial',
  'excessiveDataCollection',
  'other',
] as const

export type GrievanceCategory = (typeof GRIEVANCE_CATEGORIES)[number]

export const GRIEVANCE_PRIORITIES = ['P0', 'P1', 'P2', 'P3'] as const

export type GrievancePriority = (typeof GRIEVANCE_PRIORITIES)[number]

export const GRIEVANCE_STATUSES = [
  'Open',
  'Investigation',
  'AwaitingInfo',
  'WaitingOnDpo',
  'Resolved',
] as const

export type GrievanceStatus = (typeof GRIEVANCE_STATUSES)[number]

export type GrievanceSlaState = 'onTrack' | 'atRisk' | 'breached' | 'met'

export type GrievanceTimelineEntryType =
  | 'systemAcknowledgement'
  | 'statusChange'
  | 'communication'
  | 'note'
  | 'resolution'

export type GrievanceActorRole = 'DataPrincipal' | 'GrievanceOfficer' | 'System'

export type GrievanceTimelineVisibility = 'shared' | 'internal'

export interface GrievanceTimelineEntry {
  id: string
  type: GrievanceTimelineEntryType
  actorName: string
  actorRole: GrievanceActorRole
  message: string
  timestamp: string
  visibility: GrievanceTimelineVisibility
  fromStatus?: GrievanceStatus
  toStatus?: GrievanceStatus
  attachments?: GrievanceAttachment[]
}

export interface GrievanceAttachment {
  id: string
  fileName: string
  fileSizeLabel: string
}

export interface GrievanceRecord {
  id: string
  referenceId: string
  category: GrievanceCategory
  priority: GrievancePriority
  status: GrievanceStatus
  dataPrincipalName: string
  submittedAt: string
  updatedAt: string
  statutoryDueDate: string
  relatedConsentId?: string
  relatedConsentPurposeName?: string
}

export interface GrievanceDetail extends GrievanceRecord {
  description: string
  attachments: GrievanceAttachment[]
  timeline: GrievanceTimelineEntry[]
}

export interface GrievanceSubmissionInput {
  category: GrievanceCategory
  description: string
  attachments: File[]
}

// =============================================================================
// API-shape types — raw BFF response/request shapes, consumed only by the
// api/ and hooks/ layers, which map them into the domain types above.
// =============================================================================

export interface GrievanceAttachmentAPI {
  attachmentId: string
  fileName: string
  fileSizeBytes: number
  contentType: string
}

export interface GrievanceTimelineEntryAPI {
  entryId: string
  entryType: GrievanceTimelineEntryType
  visibility: GrievanceTimelineVisibility
  actorUserId?: string
  actorRole: GrievanceActorRole
  message: string
  fromStatus?: GrievanceStatus
  toStatus?: GrievanceStatus
  createdTime: number
  attachments?: GrievanceAttachmentAPI[]
}

export interface GrievanceDetailAPI {
  grievanceId: string
  referenceId: string
  category: GrievanceCategory
  priority: GrievancePriority
  status: GrievanceStatus
  description?: string
  userId: string
  submittedTime: number
  updatedTime: number
  statutoryDueTime: number
  timeline?: GrievanceTimelineEntryAPI[]
  attachments?: GrievanceAttachmentAPI[]
}

export interface GrievanceListMetadataAPI {
  total: number
  page: number
  pageSize: number
}

export interface GrievanceListResponseAPI {
  data: GrievanceDetailAPI[]
  metadata: GrievanceListMetadataAPI
}

export interface GrievanceStatsAPI {
  open: number
  awaitingInfo: number
  resolved: number
  slaBreached: number
}

export interface GrievanceListQueryParams {
  status?: GrievanceStatus
  priority?: GrievancePriority
  q?: string
  page: number
  pageSize: number
}
