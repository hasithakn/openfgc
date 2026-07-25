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

import type {
  GrievanceActorRole,
  GrievancePriority,
  GrievanceSlaState,
  GrievanceStatus,
} from '../../../types/grievance'

type ChipColor = 'success' | 'warning' | 'error' | 'info' | 'default'

export function getGrievancePriorityChipColor(priority: GrievancePriority): ChipColor {
  switch (priority) {
    case 'P0':
      return 'error'
    case 'P1':
      return 'warning'
    case 'P2':
      return 'info'
    case 'P3':
    default:
      return 'default'
  }
}

export function getGrievanceStatusChipColor(
  status: GrievanceStatus,
  viewerRole: Extract<GrievanceActorRole, 'DataPrincipal' | 'GrievanceOfficer'>,
): ChipColor {
  switch (status) {
    case 'Open':
      return 'info'
    case 'Investigation':
      return 'warning'
    case 'WaitingOnDpo':
      return viewerRole === 'DataPrincipal' ? 'default' : 'error'
    case 'AwaitingInfo':
      return viewerRole === 'DataPrincipal' ? 'error' : 'default'
    case 'Resolved':
      return 'success'
    default:
      return 'default'
  }
}

export function getGrievanceStatusLabelKey(status: GrievanceStatus): string {
  return status.charAt(0).toLowerCase() + status.slice(1)
}

const SLA_AT_RISK_THRESHOLD_HOURS = 24 * 14
const DAY_IN_MS = 1000 * 60 * 60 * 24

export function getGrievanceSlaState(
  statutoryDueDate: string,
  status: GrievanceStatus,
): GrievanceSlaState {
  if (status === 'Resolved') {
    return 'met'
  }

  const hoursRemaining = (new Date(statutoryDueDate).getTime() - Date.now()) / (1000 * 60 * 60)

  if (hoursRemaining < 0) {
    return 'breached'
  }

  if (hoursRemaining <= SLA_AT_RISK_THRESHOLD_HOURS) {
    return 'atRisk'
  }

  return 'onTrack'
}

export function getGrievanceSlaDaysRemaining(statutoryDueDate: string): number {
  return Math.ceil((new Date(statutoryDueDate).getTime() - Date.now()) / DAY_IN_MS)
}

const FILE_SIZE_UNITS = ['B', 'KB', 'MB', 'GB'] as const

export function formatAttachmentSize(bytes: number): string {
  if (!Number.isFinite(bytes) || bytes <= 0) {
    return '0 B'
  }

  let value = bytes
  let unitIndex = 0

  while (value >= 1024 && unitIndex < FILE_SIZE_UNITS.length - 1) {
    value /= 1024
    unitIndex += 1
  }

  return `${unitIndex === 0 ? value : value.toFixed(1)} ${FILE_SIZE_UNITS[unitIndex]}`
}
