/*
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
 * Licensed under the Apache License, Version 2.0.
 */

import {
  type UseMutationResult,
  type UseQueryResult,
  useMutation,
  useQuery,
  useQueryClient,
} from '@tanstack/react-query'
import type { ProfileResponse, ProfileUpdateRequest } from '../../../types/profile'
import { deleteAccount, getProfile, updateProfile } from '../api/profileApi'

const PROFILE_QUERY_KEY = ['profile', 'mine']

export function useProfileQuery(): UseQueryResult<ProfileResponse> {
  return useQuery({
    queryKey: PROFILE_QUERY_KEY,
    queryFn: getProfile,
  })
}

export function useUpdateProfileMutation(): UseMutationResult<
  ProfileResponse,
  Error,
  ProfileUpdateRequest
> {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (payload: ProfileUpdateRequest): Promise<ProfileResponse> => updateProfile(payload),
    onSuccess: (data): void => {
      queryClient.setQueryData(PROFILE_QUERY_KEY, data)
    },
  })
}

export function useDeleteAccountMutation(): UseMutationResult<void, Error, void> {
  return useMutation({
    mutationFn: (): Promise<void> => deleteAccount(),
  })
}
