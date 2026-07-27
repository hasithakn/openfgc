/*
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
 * Licensed under the Apache License, Version 2.0.
 */

import type { ProfileResponse, ProfileUpdateRequest } from '../../../types/profile'
import { apiRequest, apiRequestNoContent } from '../../../utils/apiClient'

const jsonHeaders = { 'Content-Type': 'application/json' }

export function getProfile(): Promise<ProfileResponse> {
  return apiRequest<ProfileResponse>('/profile', { method: 'GET' })
}

export function updateProfile(payload: ProfileUpdateRequest): Promise<ProfileResponse> {
  return apiRequest<ProfileResponse>('/profile', {
    method: 'PATCH',
    headers: jsonHeaders,
    body: JSON.stringify(payload),
  })
}

export function deleteAccount(): Promise<void> {
  return apiRequestNoContent('/profile', { method: 'DELETE' })
}
