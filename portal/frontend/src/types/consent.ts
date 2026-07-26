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

export type ConsentStatus = 'Active' | 'Pending' | 'Rejected' | 'Revoked' | 'Expired'

export const CONSENT_API_STATUSES = ['CREATED', 'ACTIVE', 'REJECTED', 'REVOKED', 'EXPIRED'] as const

export type ConsentAPIStatus = (typeof CONSENT_API_STATUSES)[number]

export function isConsentAPIStatus(status: string): status is ConsentAPIStatus {
  return CONSENT_API_STATUSES.includes(status as ConsentAPIStatus)
}

export interface ConsentRecord {
  id: string
  groupId: string
  status: ConsentAPIStatus
  purposes: string[]
  updatedAt: string
  expirationTime?: number
  canRevoke: boolean
  canApprove: boolean
  userId?: string
}

export interface ConsentRegistryFilters {
  status: 'All' | ConsentStatus
  startDate: string
  endDate: string
  purpose: string
  userId?: string
}

export interface ConsentListQueryParams {
  consentStatuses?: string
  purposeName?: string
  fromTime?: number
  toTime?: number
  limit: number
  offset: number
  userId?: string
}

export interface ConsentElementApprovalItem {
  elementId: string
  name: string
  namespace: string
  version: string
  displayName?: string
  approved: boolean
  mandatory: boolean
  value?: unknown
  description?: string
}

export interface ConsentApprovalSelection {
  purposeId: string
  purposeVersion: string
  elementId: string
  elementVersion: string
}

export interface ConsentPurposeItem {
  purposeId: string
  name: string
  version: string
  displayName?: string
  description?: string
  elements: ConsentElementApprovalItem[]
}

export interface ConsentAuthorizationResource {
  id: string
  userId?: string
  type: string
  status: string
  updatedTime: number
  resources?: unknown
}

export interface ConsentStatusAudit {
  statusAuditId: string
  previousStatus?: string
  currentStatus: string
  actionTime: number
  actionBy?: string
  reason?: string
}

export interface ConsentDetailAPI {
  id: string
  groupId: string
  type: string
  status: ConsentAPIStatus | string
  createdTime: number
  updatedTime: number
  expirationTime?: number
  recurringIndicator?: boolean
  frequency?: number
  dataAccessValidityDuration?: number
  purposes: ConsentPurposeItem[]
  attributes?: Record<string, string>
  authorizations?: ConsentAuthorizationResource[]
  statusHistory?: ConsentStatusAudit[]
}

export interface ConsentHistoryEntry {
  historyId: string
  actionTime: number
  actionBy?: string
  reason?: string
  snapshot?: ConsentDetailAPI
}

export interface ConsentHistoryListResponse {
  id: string
  history: ConsentHistoryEntry[]
}

export interface ConsentSearchMetadata {
  total: number
  offset: number
  count: number
  limit: number
}

export interface ConsentSearchResponse {
  data: ConsentDetailAPI[]
  metadata: ConsentSearchMetadata
}
