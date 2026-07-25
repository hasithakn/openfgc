/*
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
 * Licensed under the Apache License, Version 2.0.
 */

import { createContext, useContext, useEffect, useState } from 'react'
import { getAccessTokenPart1, isAuthEnabled } from '../utils/authClient'

const SCOPE_CONSENTS_READ_ANY = 'portal:consents:read:any'
const SCOPE_CONSENTS_WRITE_ANY = 'portal:consents:write:any'
const SCOPE_CONSENTS_READ_SELF = 'portal:consents:read:self'
const SCOPE_CONSENTS_WRITE_SELF = 'portal:consents:write:self'
const SCOPE_ELEMENTS_READ = 'portal:elements:read'
const SCOPE_ELEMENTS_WRITE = 'portal:elements:write'
const SCOPE_PURPOSES_READ = 'portal:purposes:read'
const SCOPE_PURPOSES_WRITE = 'portal:purposes:write'
const SCOPE_GRIEVANCES_MANAGE = 'portal:grievances:manage'
const SCOPE_GRIEVANCES_READ_SELF = 'portal:grievances:read:self'
const SCOPE_GRIEVANCES_WRITE_SELF = 'portal:grievances:write:self'

export const ALL_SCOPES = [
  SCOPE_CONSENTS_READ_ANY,
  SCOPE_CONSENTS_WRITE_ANY,
  SCOPE_CONSENTS_READ_SELF,
  SCOPE_CONSENTS_WRITE_SELF,
  SCOPE_ELEMENTS_READ,
  SCOPE_ELEMENTS_WRITE,
  SCOPE_PURPOSES_READ,
  SCOPE_PURPOSES_WRITE,
  SCOPE_GRIEVANCES_MANAGE,
  SCOPE_GRIEVANCES_READ_SELF,
  SCOPE_GRIEVANCES_WRITE_SELF,
]

interface UserInfo {
  userId: string
  orgId: string
  scopes: string[]
}

interface ScopeContextValue {
  scopes: Set<string>
  hasScope: (scope: string) => boolean
  isAdmin: boolean
  canReadElements: boolean
  canWriteElements: boolean
  canReadPurposes: boolean
  canWritePurposes: boolean
  isGrievanceOfficer: boolean
  canReadGrievancesSelf: boolean
  isLoading: boolean
}

const ScopeContext = createContext<ScopeContextValue | null>(null)

async function fetchUserInfo(): Promise<UserInfo> {
  const baseURL = import.meta.env.VITE_API_BASE_URL as string | undefined
  if (!baseURL) {
    throw new Error('VITE_API_BASE_URL is required')
  }
  const normalizedBase = baseURL.endsWith('/') ? baseURL.slice(0, -1) : baseURL
  const accessPart = getAccessTokenPart1()
  if (!accessPart) {
    throw new Error('access token is unavailable')
  }
  const response = await fetch(`${normalizedBase}/auth/userinfo`, {
    credentials: 'include',
    headers: {
      Authorization: `Bearer ${accessPart}`,
    },
  })
  if (!response.ok) {
    throw new Error('Failed to fetch user info')
  }
  return (await response.json()) as UserInfo
}

export function ScopeProvider({ children }: { children: React.ReactNode }): React.JSX.Element {
  const [scopes, setScopes] = useState<Set<string>>(new Set())
  const [isLoading, setIsLoading] = useState(true)

  useEffect(() => {
    if (!isAuthEnabled()) {
      // Auth disabled = placeholder/dev mode, grant all scopes
      setScopes(new Set(ALL_SCOPES))
      setIsLoading(false)
      return
    }

    fetchUserInfo()
      .then((info) => {
        setScopes(new Set(info.scopes))
      })
      .catch(() => {
        setScopes(new Set())
      })
      .finally(() => {
        setIsLoading(false)
      })
  }, [])

  const hasScope = (scope: string): boolean => scopes.has(scope)

  const value: ScopeContextValue = {
    scopes,
    hasScope,
    isAdmin: scopes.has(SCOPE_CONSENTS_READ_ANY),
    canReadElements: scopes.has(SCOPE_ELEMENTS_READ),
    canWriteElements: scopes.has(SCOPE_ELEMENTS_WRITE),
    canReadPurposes: scopes.has(SCOPE_PURPOSES_READ),
    canWritePurposes: scopes.has(SCOPE_PURPOSES_WRITE),
    isGrievanceOfficer: scopes.has(SCOPE_GRIEVANCES_MANAGE),
    canReadGrievancesSelf: scopes.has(SCOPE_GRIEVANCES_READ_SELF),
    isLoading,
  }

  return <ScopeContext.Provider value={value}>{children}</ScopeContext.Provider>
}

export function useScopes(): ScopeContextValue {
  const ctx = useContext(ScopeContext)
  if (!ctx) {
    throw new Error('useScopes must be used within ScopeProvider')
  }
  return ctx
}
