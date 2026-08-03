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
  type UseMutationResult,
  type UseQueryResult,
  useMutation,
  useQuery,
  useQueryClient,
} from '@tanstack/react-query'
import type {
  EventDeliveryHistory,
  EventDeliveryListResponse,
  SubscriptionCreateRequest,
  SubscriptionListFilters,
  SubscriptionListResponse,
  SubscriptionRecord,
  TopicListResponse,
} from '../../../types/eventSubscription'
import {
  createSubscription,
  deleteSubscription,
  fetchOrgEventHistory,
  fetchOrgEvents,
  fetchSubscription,
  fetchSubscriptionEventHistory,
  fetchSubscriptionEvents,
  fetchSubscriptions,
  fetchTopics,
} from '../api/eventSubscriptionsApi'

export function useTopicsQuery(): UseQueryResult<TopicListResponse> {
  return useQuery({
    queryKey: ['event-topics'],
    queryFn: fetchTopics,
  })
}

export function useSubscriptionListQuery(
  filters: SubscriptionListFilters,
  page: number,
  rowsPerPage: number,
): UseQueryResult<SubscriptionListResponse> {
  return useQuery({
    queryKey: ['event-subscriptions', filters, page, rowsPerPage],
    queryFn: () => fetchSubscriptions(filters, rowsPerPage, page * rowsPerPage),
    placeholderData: keepPreviousData,
  })
}

export function useSubscriptionQuery(subscriptionId?: string): UseQueryResult<SubscriptionRecord> {
  return useQuery({
    queryKey: ['event-subscription', subscriptionId],
    queryFn: () => fetchSubscription(String(subscriptionId)),
    enabled: Boolean(subscriptionId),
  })
}

export function useSubscriptionEventsQuery(
  subscriptionId: string | undefined,
  page: number,
  rowsPerPage: number,
): UseQueryResult<EventDeliveryListResponse> {
  return useQuery({
    queryKey: ['event-subscription', subscriptionId, 'events', page, rowsPerPage],
    queryFn: () => fetchSubscriptionEvents(String(subscriptionId), rowsPerPage, page * rowsPerPage),
    enabled: Boolean(subscriptionId),
    placeholderData: keepPreviousData,
  })
}

export function useSubscriptionEventHistoryQuery(
  subscriptionId: string | undefined,
  deliveryId: string | undefined,
): UseQueryResult<EventDeliveryHistory> {
  return useQuery({
    queryKey: ['event-subscription', subscriptionId, 'event-history', deliveryId],
    queryFn: () => fetchSubscriptionEventHistory(String(subscriptionId), String(deliveryId)),
    enabled: Boolean(subscriptionId) && Boolean(deliveryId),
  })
}

export function useOrgEventListQuery(
  status: string,
  subscriptionId: string,
  search: string,
  page: number,
  rowsPerPage: number,
): UseQueryResult<EventDeliveryListResponse> {
  return useQuery({
    queryKey: ['org-events', status, subscriptionId, search, page, rowsPerPage],
    queryFn: () => fetchOrgEvents(status, subscriptionId, search, rowsPerPage, page * rowsPerPage),
    placeholderData: keepPreviousData,
    // New deliveries can land at any time — unlike mostly-static list pages elsewhere in the
    // app, this feed needs a fresh fetch on every visit rather than reusing whatever was
    // cached from up to staleTime (30s) ago. Without this, revisiting /events within that
    // window via client-side navigation silently reused a stale (possibly empty) cached
    // result and fired no request at all; only a full page reload (which discards the whole
    // QueryClient) forced a real fetch.
    refetchOnMount: 'always',
  })
}

export function useOrgEventHistoryQuery(
  deliveryId: string | undefined,
): UseQueryResult<EventDeliveryHistory> {
  return useQuery({
    queryKey: ['org-event-history', deliveryId],
    queryFn: () => fetchOrgEventHistory(String(deliveryId)),
    enabled: Boolean(deliveryId),
  })
}

export function useCreateSubscriptionMutation(): UseMutationResult<
  SubscriptionRecord,
  Error,
  SubscriptionCreateRequest
> {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: createSubscription,
    onSuccess: async () => queryClient.invalidateQueries({ queryKey: ['event-subscriptions'] }),
  })
}

export function useDeleteSubscriptionMutation(): UseMutationResult<void, Error, string> {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: deleteSubscription,
    onSuccess: async (_data, subscriptionId) => {
      await queryClient.invalidateQueries({ queryKey: ['event-subscriptions'] })
      await queryClient.invalidateQueries({ queryKey: ['event-subscription', subscriptionId] })
    },
  })
}
