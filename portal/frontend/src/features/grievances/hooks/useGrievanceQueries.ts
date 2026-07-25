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
  fetchMyGrievanceByID,
  fetchMyGrievances,
  replyToMyGrievance,
  submitMyGrievance,
} from '../api/grievancesApi'
import { toGrievanceDetail, toGrievanceRecord } from '../api/grievanceMappers'
import type {
  GrievanceDetail,
  GrievanceRecord,
  GrievanceStatus,
  GrievanceSubmissionInput,
} from '../../../types/grievance'

interface GrievanceListResult {
  rows: GrievanceRecord[]
  total: number
}

interface ReplyVariables {
  grievanceId: string
  message: string
  attachments: File[]
}

function myGrievanceListQueryOptions(
  statusFilter: GrievanceStatus | 'All',
  page: number,
  pageSize: number,
) {
  return queryOptions({
    queryKey: ['grievances', 'mine', statusFilter, page, pageSize],
    queryFn: async (): Promise<GrievanceListResult> => {
      const response = await fetchMyGrievances({
        status: statusFilter === 'All' ? undefined : statusFilter,
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

export function useMyGrievanceListQuery(
  statusFilter: GrievanceStatus | 'All',
  page: number,
  pageSize: number,
): UseQueryResult<GrievanceListResult> {
  return useQuery(myGrievanceListQueryOptions(statusFilter, page, pageSize))
}

export function useMyGrievanceDetailQuery(
  grievanceId: string | undefined,
): UseQueryResult<GrievanceDetail> {
  return useQuery<GrievanceDetail>({
    queryKey: ['grievance', 'mine', grievanceId],
    queryFn: async (): Promise<GrievanceDetail> =>
      toGrievanceDetail(await fetchMyGrievanceByID(String(grievanceId))),
    enabled: Boolean(grievanceId),
  })
}

export function useSubmitGrievanceMutation(): UseMutationResult<
  GrievanceDetail,
  Error,
  GrievanceSubmissionInput
> {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (input: GrievanceSubmissionInput): Promise<GrievanceDetail> =>
      toGrievanceDetail(
        await submitMyGrievance(input.category, input.description, input.attachments),
      ),
    onSuccess: async (): Promise<void> => {
      await queryClient.invalidateQueries({ queryKey: ['grievances'] })
    },
  })
}

export function useReplyToGrievanceMutation(): UseMutationResult<
  GrievanceDetail,
  Error,
  ReplyVariables
> {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({
      grievanceId,
      message,
      attachments,
    }: ReplyVariables): Promise<GrievanceDetail> =>
      toGrievanceDetail(await replyToMyGrievance(grievanceId, message, attachments)),
    onSuccess: async (_data, variables): Promise<void> => {
      await queryClient.invalidateQueries({ queryKey: ['grievances'] })
      await queryClient.invalidateQueries({
        queryKey: ['grievance', 'mine', variables.grievanceId],
      })
    },
  })
}
