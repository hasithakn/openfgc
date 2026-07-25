/*
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
 * Licensed under the Apache License, Version 2.0.
 */

import type {
  GrievanceDetailAPI,
  GrievanceListQueryParams,
  GrievanceListResponseAPI,
  GrievanceStatsAPI,
  GrievanceStatus,
  GrievanceTimelineVisibility,
} from '../../../types/grievance'
import { apiRequest } from '../../../utils/apiClient'

interface AttachmentPayload {
  fileName: string
  contentType: string
  contentBase64: string
}

/** Reads a File's bytes and base64-encodes them, matching the BFF's JSON attachment contract. */
async function fileToBase64(file: File): Promise<string> {
  const buffer = await file.arrayBuffer()
  const bytes = new Uint8Array(buffer)
  let binary = ''
  const chunkSize = 0x8000

  for (let offset = 0; offset < bytes.length; offset += chunkSize) {
    binary += String.fromCharCode(...bytes.subarray(offset, offset + chunkSize))
  }

  return btoa(binary)
}

async function toAttachmentPayloads(files: File[]): Promise<AttachmentPayload[]> {
  return Promise.all(
    files.map(async (file) => ({
      fileName: file.name,
      contentType: file.type || 'application/octet-stream',
      contentBase64: await fileToBase64(file),
    })),
  )
}

// =============================================================================
// Data Principal self-serve endpoints
// =============================================================================

export async function fetchMyGrievances(
  params: GrievanceListQueryParams,
): Promise<GrievanceListResponseAPI> {
  return apiRequest<GrievanceListResponseAPI>('/me/grievances', {
    method: 'GET',
    query: {
      status: params.status,
      priority: params.priority,
      q: params.q,
      page: params.page,
      pageSize: params.pageSize,
    },
  })
}

export async function fetchMyGrievanceByID(grievanceId: string): Promise<GrievanceDetailAPI> {
  return apiRequest<GrievanceDetailAPI>(`/me/grievances/${encodeURIComponent(grievanceId)}`, {
    method: 'GET',
  })
}

export async function submitMyGrievance(
  category: string,
  description: string,
  attachments: File[],
): Promise<GrievanceDetailAPI> {
  return apiRequest<GrievanceDetailAPI>('/me/grievances', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      category,
      description,
      attachments: await toAttachmentPayloads(attachments),
    }),
  })
}

export async function replyToMyGrievance(
  grievanceId: string,
  message: string,
  attachments: File[],
): Promise<GrievanceDetailAPI> {
  return apiRequest<GrievanceDetailAPI>(
    `/me/grievances/${encodeURIComponent(grievanceId)}/messages`,
    {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        message,
        attachments: await toAttachmentPayloads(attachments),
      }),
    },
  )
}

// =============================================================================
// Grievance Officer endpoints
// =============================================================================

export async function fetchGrievanceQueue(
  params: GrievanceListQueryParams,
): Promise<GrievanceListResponseAPI> {
  return apiRequest<GrievanceListResponseAPI>('/grievance-management/cases', {
    method: 'GET',
    query: {
      status: params.status,
      priority: params.priority,
      q: params.q,
      page: params.page,
      pageSize: params.pageSize,
    },
  })
}

export async function fetchGrievanceQueueStats(): Promise<GrievanceStatsAPI> {
  return apiRequest<GrievanceStatsAPI>('/grievance-management/cases/stats', { method: 'GET' })
}

export async function fetchCaseByID(grievanceId: string): Promise<GrievanceDetailAPI> {
  return apiRequest<GrievanceDetailAPI>(
    `/grievance-management/cases/${encodeURIComponent(grievanceId)}`,
    { method: 'GET' },
  )
}

export async function replyToCase(
  grievanceId: string,
  message: string,
  visibility: GrievanceTimelineVisibility,
  attachments: File[],
): Promise<GrievanceDetailAPI> {
  return apiRequest<GrievanceDetailAPI>(
    `/grievance-management/cases/${encodeURIComponent(grievanceId)}/messages`,
    {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        message,
        visibility,
        attachments: await toAttachmentPayloads(attachments),
      }),
    },
  )
}

export async function transitionCaseStatus(
  grievanceId: string,
  nextStatus: GrievanceStatus,
): Promise<GrievanceDetailAPI> {
  return apiRequest<GrievanceDetailAPI>(
    `/grievance-management/cases/${encodeURIComponent(grievanceId)}/status`,
    {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ nextStatus }),
    },
  )
}
