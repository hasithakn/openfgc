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
  EventDeliveryHistory,
  EventDeliveryListResponse,
  SubscriptionCreateRequest,
  SubscriptionListFilters,
  SubscriptionListResponse,
  SubscriptionRecord,
  TopicListResponse,
} from '../../../types/eventSubscription'
import { apiRequest, apiRequestNoContent } from '../../../utils/apiClient'

const jsonHeaders = { 'Content-Type': 'application/json' }

export function fetchSubscriptions(
  filters: SubscriptionListFilters,
  limit: number,
  offset: number,
): Promise<SubscriptionListResponse> {
  return apiRequest<SubscriptionListResponse>('/admin/event-subscriptions', {
    method: 'GET',
    query: {
      status: filters.status === 'All' ? undefined : filters.status,
      search: filters.search.trim() || undefined,
      limit,
      offset,
    },
  })
}

export function fetchSubscription(subscriptionId: string): Promise<SubscriptionRecord> {
  return apiRequest<SubscriptionRecord>(
    `/admin/event-subscriptions/${encodeURIComponent(subscriptionId)}`,
    { method: 'GET' },
  )
}

export function createSubscription(
  payload: SubscriptionCreateRequest,
): Promise<SubscriptionRecord> {
  return apiRequest<SubscriptionRecord>('/admin/event-subscriptions', {
    method: 'POST',
    headers: jsonHeaders,
    body: JSON.stringify(payload),
  })
}

export function deleteSubscription(subscriptionId: string): Promise<void> {
  return apiRequestNoContent(`/admin/event-subscriptions/${encodeURIComponent(subscriptionId)}`, {
    method: 'DELETE',
  })
}

export function fetchSubscriptionEvents(
  subscriptionId: string,
  limit: number,
  offset: number,
): Promise<EventDeliveryListResponse> {
  return apiRequest<EventDeliveryListResponse>(
    `/admin/event-subscriptions/${encodeURIComponent(subscriptionId)}/events`,
    { method: 'GET', query: { limit, offset } },
  )
}

export function fetchSubscriptionEventHistory(
  subscriptionId: string,
  deliveryId: string,
): Promise<EventDeliveryHistory> {
  return apiRequest<EventDeliveryHistory>(
    `/admin/event-subscriptions/${encodeURIComponent(subscriptionId)}/events/${encodeURIComponent(deliveryId)}/history`,
    { method: 'GET' },
  )
}

export function fetchOrgEvents(
  status: string,
  subscriptionId: string,
  search: string,
  limit: number,
  offset: number,
): Promise<EventDeliveryListResponse> {
  return apiRequest<EventDeliveryListResponse>('/admin/events', {
    method: 'GET',
    query: {
      status: status.trim() || undefined,
      subscriptionId: subscriptionId.trim() || undefined,
      search: search.trim() || undefined,
      limit,
      offset,
    },
  })
}

export function fetchOrgEventHistory(deliveryId: string): Promise<EventDeliveryHistory> {
  return apiRequest<EventDeliveryHistory>(
    `/admin/events/${encodeURIComponent(deliveryId)}/history`,
    { method: 'GET' },
  )
}

export function fetchTopics(): Promise<TopicListResponse> {
  return apiRequest<TopicListResponse>('/admin/topics', {
    method: 'GET',
    query: { status: 'active', limit: 100, offset: 0 },
  })
}
