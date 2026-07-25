/*
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
 * Licensed under the Apache License, Version 2.0.
 */

import {
  keepPreviousData,
  queryOptions,
  type UseMutationResult,
  type UseQueryResult,
  useMutation,
  useQuery,
  useQueryClient,
} from '@tanstack/react-query'
import {
  fetchCaseByID,
  fetchGrievanceQueue,
  fetchGrievanceQueueStats,
  replyToCase,
  transitionCaseStatus,
} from '../../grievances/api/grievancesApi'
import { toGrievanceDetail, toGrievanceRecord } from '../../grievances/api/grievanceMappers'
import type {
  GrievanceDetail,
  GrievanceRecord,
  GrievanceStatsAPI,
  GrievanceStatus,
  GrievanceTimelineVisibility,
} from '../../../types/grievance'
import type { GrievanceQueueFiltersState } from '../types'

interface GrievanceQueueResult {
  rows: GrievanceRecord[]
  total: number
}

interface ReplyVariables {
  grievanceId: string
  message: string
  visibility: GrievanceTimelineVisibility
  attachments: File[]
}

interface TransitionStatusVariables {
  grievanceId: string
  nextStatus: GrievanceStatus
}

function grievanceQueueQueryOptions(
  filters: GrievanceQueueFiltersState,
  page: number,
  pageSize: number,
) {
  return queryOptions({
    queryKey: ['grievances', 'queue', filters, page, pageSize],
    queryFn: async (): Promise<GrievanceQueueResult> => {
      const response = await fetchGrievanceQueue({
        status: filters.status === 'All' ? undefined : filters.status,
        priority: filters.priority === 'All' ? undefined : filters.priority,
        q: filters.search.trim() || undefined,
        page,
        pageSize,
      })
      return {
        rows: response.data.map(toGrievanceRecord),
        total: response.metadata.total,
      }
    },
    placeholderData: keepPreviousData,
  })
}

export function useGrievanceQueueQuery(
  filters: GrievanceQueueFiltersState,
  page: number,
  pageSize: number,
): UseQueryResult<GrievanceQueueResult> {
  return useQuery(grievanceQueueQueryOptions(filters, page, pageSize))
}

export function useGrievanceQueueStatsQuery(): UseQueryResult<GrievanceStatsAPI> {
  return useQuery({
    queryKey: ['grievances', 'queue', 'stats'],
    queryFn: fetchGrievanceQueueStats,
  })
}

export function useCaseDetailQuery(
  grievanceId: string | undefined,
): UseQueryResult<GrievanceDetail> {
  return useQuery<GrievanceDetail>({
    queryKey: ['grievance', 'case', grievanceId],
    queryFn: async (): Promise<GrievanceDetail> =>
      toGrievanceDetail(await fetchCaseByID(String(grievanceId))),
    enabled: Boolean(grievanceId),
  })
}

export function useReplyToCaseMutation(): UseMutationResult<
  GrievanceDetail,
  Error,
  ReplyVariables
> {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({
      grievanceId,
      message,
      visibility,
      attachments,
    }: ReplyVariables): Promise<GrievanceDetail> =>
      toGrievanceDetail(await replyToCase(grievanceId, message, visibility, attachments)),
    onSuccess: async (_data, variables): Promise<void> => {
      await queryClient.invalidateQueries({ queryKey: ['grievances', 'queue'] })
      await queryClient.invalidateQueries({
        queryKey: ['grievance', 'case', variables.grievanceId],
      })
    },
  })
}

export function useTransitionStatusMutation(): UseMutationResult<
  GrievanceDetail,
  Error,
  TransitionStatusVariables
> {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({
      grievanceId,
      nextStatus,
    }: TransitionStatusVariables): Promise<GrievanceDetail> =>
      toGrievanceDetail(await transitionCaseStatus(grievanceId, nextStatus)),
    onSuccess: async (_data, variables): Promise<void> => {
      await queryClient.invalidateQueries({ queryKey: ['grievances', 'queue'] })
      await queryClient.invalidateQueries({
        queryKey: ['grievance', 'case', variables.grievanceId],
      })
    },
  })
}
