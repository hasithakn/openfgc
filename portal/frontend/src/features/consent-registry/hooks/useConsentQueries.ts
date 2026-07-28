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

import {
  keepPreviousData,
  queryOptions,
  type UseMutationResult,
  type UseQueryResult,
  useMutation,
  useQuery,
  useQueryClient,
} from '@tanstack/react-query'
import { useEffect } from 'react'
import {
  approveMyConsent,
  fetchAllConsents,
  fetchConsentByID,
  fetchConsentHistoryByID,
  fetchMyConsentByID,
  fetchMyConsentHistory,
  fetchMyConsents,
  revokeMyConsent,
} from '../api/consentsApi'
import {
  isConsentApprovableStatus,
  isConsentRevokableStatus,
  normalizeConsentStatus,
} from '../utils/statusChip'
import type {
  ConsentApprovalSelection,
  ConsentDetailAPI,
  ConsentHistoryEntry,
  ConsentListQueryParams,
  ConsentRecord,
  ConsentRegistryFilters,
} from '../../../types/consent'
import { isConsentAPIStatus } from '../../../types/consent'
import {
  toEndOfDayEpochMilliseconds,
  toEpochMilliseconds,
  toStartOfDayEpochMilliseconds,
} from '../../../utils/dateTime'
import { useScopes } from '../../../context/ScopeContext'

interface ConsentListResult {
  rows: ConsentRecord[]
  total: number
}

interface ApproveConsentVariables {
  consentID: string
  selectedOptionalElements: ConsentApprovalSelection[]
}

function toListParams(
  filters: ConsentRegistryFilters,
  page: number,
  rowsPerPage: number,
): ConsentListQueryParams {
  const statusFilterMap: Record<Exclude<ConsentRegistryFilters['status'], 'All'>, string> = {
    Active: 'ACTIVE',
    Pending: 'CREATED',
    Rejected: 'REJECTED',
    Revoked: 'REVOKED',
    Expired: 'EXPIRED',
  }

  return {
    consentStatuses: filters.status === 'All' ? undefined : statusFilterMap[filters.status],
    purposeName: filters.purpose.trim() || undefined,
    fromTime: toStartOfDayEpochMilliseconds(filters.startDate),
    toTime: toEndOfDayEpochMilliseconds(filters.endDate),
    userId: filters.userId?.trim() || undefined,
    limit: rowsPerPage,
    offset: page * rowsPerPage,
  }
}

function toConsentRow(consent: ConsentDetailAPI, isAdmin: boolean): ConsentRecord {
  const normalizedStatus = normalizeConsentStatus(consent.status)

  if (!isConsentAPIStatus(normalizedStatus)) {
    throw new Error(`Unsupported consent status received from API: ${consent.status}`)
  }

  return {
    id: consent.id,
    groupId: consent.groupId,
    status: normalizedStatus,
    purposes: consent.purposes.map((purpose) => purpose.displayName ?? purpose.name),
    updatedAt: new Date(toEpochMilliseconds(consent.updatedTime) ?? 0).toISOString(),
    expirationTime: consent.expirationTime ?? 0,
    canRevoke: isConsentRevokableStatus(normalizedStatus),
    canApprove: !isAdmin && isConsentApprovableStatus(normalizedStatus),
    userId: consent.authorizations?.[0]?.userId,
  }
}

function consentListQueryOptions(
  filters: ConsentRegistryFilters,
  page: number,
  rowsPerPage: number,
  isAdmin: boolean,
  enabled: boolean,
) {
  const params = toListParams(filters, page, rowsPerPage)
  const fetchFn = isAdmin ? fetchAllConsents : fetchMyConsents

  return queryOptions({
    queryKey: ['consents', isAdmin ? 'all' : 'mine', params],
    queryFn: async (): Promise<ConsentListResult> => {
      const response = await fetchFn(params)
      return {
        rows: response.data.map((consent) => toConsentRow(consent, isAdmin)),
        total: response.metadata.total,
      }
    },
    placeholderData: keepPreviousData,
    enabled,
  })
}

export function useConsentListQuery(
  filters: ConsentRegistryFilters,
  page: number,
  rowsPerPage: number,
): UseQueryResult<ConsentListResult> {
  const { isAdmin, isLoading: scopesLoading } = useScopes()
  const queryClient = useQueryClient()
  // Deferred until scopes finish loading — otherwise this would fetch (and compute canApprove)
  // against the default isAdmin=false before the real scope set is known, which could briefly
  // show an admin/case-worker a self-service Approve action they aren't entitled to.
  const query = useQuery(
    consentListQueryOptions(filters, page, rowsPerPage, isAdmin, !scopesLoading),
  )

  useEffect(() => {
    const nextPage = page + 1
    const hasNextPage = nextPage * rowsPerPage < (query.data?.total ?? 0)

    if (!scopesLoading && !query.isPlaceholderData && hasNextPage) {
      queryClient
        .prefetchQuery(consentListQueryOptions(filters, nextPage, rowsPerPage, isAdmin, true))
        .catch(() => undefined)
    }
  }, [
    filters,
    isAdmin,
    page,
    query.data?.total,
    query.isPlaceholderData,
    queryClient,
    rowsPerPage,
    scopesLoading,
  ])

  return query
}

export function useConsentDetailQuery(
  consentID: string | undefined,
): UseQueryResult<ConsentDetailAPI> {
  const { isAdmin } = useScopes()

  return useQuery<ConsentDetailAPI>({
    queryKey: ['consent', isAdmin ? 'all' : 'mine', consentID],
    queryFn: async (): Promise<ConsentDetailAPI> =>
      isAdmin ? fetchConsentByID(String(consentID)) : fetchMyConsentByID(String(consentID)),
    enabled: Boolean(consentID),
  })
}

// Lazy: full snapshot history can be a heavier payload than the rest of the detail page, so
// this is only meant to be enabled while the history modal is actually open.
export function useConsentHistoryQuery(
  consentID: string | undefined,
  enabled: boolean,
): UseQueryResult<ConsentHistoryEntry[]> {
  const { isAdmin } = useScopes()

  return useQuery<ConsentHistoryEntry[]>({
    queryKey: ['consent-history', isAdmin ? 'all' : 'mine', consentID],
    queryFn: async (): Promise<ConsentHistoryEntry[]> => {
      const response = isAdmin
        ? await fetchConsentHistoryByID(String(consentID))
        : await fetchMyConsentHistory(String(consentID))
      return response.history
    },
    enabled: enabled && Boolean(consentID),
  })
}

export function useApproveConsentMutation(): UseMutationResult<
  unknown,
  Error,
  ApproveConsentVariables
> {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({
      consentID,
      selectedOptionalElements,
    }: ApproveConsentVariables): Promise<unknown> =>
      approveMyConsent(consentID, selectedOptionalElements),
    onSuccess: async (_data, variables): Promise<void> => {
      await queryClient.invalidateQueries({ queryKey: ['consents'] })
      await queryClient.invalidateQueries({
        predicate: (query) =>
          (query.queryKey[0] === 'consent' || query.queryKey[0] === 'consent-history') &&
          query.queryKey.includes(variables.consentID),
      })
    },
  })
}

export function useRevokeConsentMutation(): UseMutationResult<unknown, Error, string> {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (consentID: string): Promise<unknown> => revokeMyConsent(consentID),
    onSuccess: async (_data, consentID): Promise<void> => {
      await queryClient.invalidateQueries({ queryKey: ['consents'] })
      await queryClient.invalidateQueries({
        predicate: (query) =>
          (query.queryKey[0] === 'consent' || query.queryKey[0] === 'consent-history') &&
          query.queryKey.includes(consentID),
      })
    },
  })
}
