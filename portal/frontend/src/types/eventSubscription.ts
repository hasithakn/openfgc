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

export type SubscriptionStatus = 'active' | 'pending' | 'stale' | 'deleted'

export interface TopicRecord {
  topicId: string
  name: string
  description?: string
  status: string
}

export interface TopicListResponse {
  topics: TopicRecord[]
  total: number
  limit: number
  offset: number
  count: number
}

export type FilterType = 'all' | 'specific' | 'all_except'

export type DeliveryMode = 'webhook' | 'poll' | 'pull'

export interface SubscriptionFilter {
  type: FilterType
  purposes?: string[]
}

export interface SubscriptionDeliveryConfig {
  mode: DeliveryMode
  callbackUrl?: string
}

export interface SubscriptionRecord {
  subscriptionId: string
  topic: string
  filter: SubscriptionFilter
  delivery: SubscriptionDeliveryConfig
  status: SubscriptionStatus
  createdAt?: number
  updatedAt?: number
  alreadyExists?: boolean
  message?: string
}

export interface SubscriptionListResponse {
  subscriptions: SubscriptionRecord[]
  total: number
  limit: number
  offset: number
  count: number
}

export interface SubscriptionListFilters {
  status: SubscriptionStatus | 'All'
  search: string
}

export interface SubscriptionCreateRequest {
  topic: string
  filter: SubscriptionFilter
  delivery: SubscriptionDeliveryConfig & { sharedSecret: string }
}

export interface EventDeliveryItem {
  deliveryId: string
  eventId: string
  topic: string
  currentStatus: string
  deliveryMode: DeliveryMode
  occurredAt: number
  payload?: unknown
}

export interface EventDeliveryListResponse {
  content: EventDeliveryItem[]
  totalElements: number
  limit: number
  offset: number
  count: number
  page: number
}

export interface DeliveryAttempt {
  attempt: number
  status: string
  timestamp: number
  httpStatus?: number
  error?: string
}

export interface EventDeliveryHistory {
  deliveryId: string
  eventId: string
  topic: string
  deliveryMode: DeliveryMode
  currentStatus: string
  occurredAt: number
  nextRetryAt?: number
  completionStatus?: string
  completionEvidence?: string
  payload?: unknown
  history: DeliveryAttempt[]
}

export interface OrgEventListFilters {
  status: string
  subscriptionId: string
  search: string
}
