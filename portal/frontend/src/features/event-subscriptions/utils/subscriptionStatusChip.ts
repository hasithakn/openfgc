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

type ChipColor = 'success' | 'warning' | 'info' | 'error' | 'default'

export function normalizeStatus(status: string): string {
  return status.trim().toLowerCase()
}

export function getSubscriptionStatusChipColor(status: string): ChipColor {
  switch (normalizeStatus(status)) {
    case 'active':
      return 'success'
    case 'pending':
      return 'warning'
    case 'stale':
      return 'info'
    case 'deleted':
      return 'error'
    default:
      return 'default'
  }
}

export function getDeliveryStatusChipColor(status: string): ChipColor {
  switch (normalizeStatus(status)) {
    case 'delivered':
    case 'completed':
    case 'acknowledged':
      return 'success'
    case 'pending':
      return 'warning'
    case 'failed':
    case 'err':
    case 'disputed':
      return 'error'
    default:
      return 'default'
  }
}
