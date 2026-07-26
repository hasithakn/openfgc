/*
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
 * Licensed under the Apache License, Version 2.0.
 */

import { Alert, Button, Snackbar, UserMenu } from '@wso2/oxygen-ui'
import { User } from '@wso2/oxygen-ui-icons-react'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useNavigate } from 'react-router-dom'
import { getUserProfile, logout } from '../../../utils/authClient'

type UserClaims = Record<string, unknown>

function claim(claims: UserClaims, name: string): string | undefined {
  const value = claims[name]

  return typeof value === 'string' && value.trim() ? value.trim() : undefined
}

function displayName(claims: UserClaims, fallback: string): string {
  const name = claim(claims, 'name') ?? claim(claims, 'displayName')
  if (name) {
    return name
  }

  const fullName = [claim(claims, 'given_name'), claim(claims, 'family_name')]
    .filter(Boolean)
    .join(' ')

  return fullName || claim(claims, 'preferred_username') || claim(claims, 'username') || fallback
}

function UserProfileMenu(): React.JSX.Element {
  const { t } = useTranslation('common')
  const navigate = useNavigate()
  const [logoutFailed, setLogoutFailed] = useState(false)
  const [logoutPending, setLogoutPending] = useState(false)
  const [claims] = useState<UserClaims>(() => getUserProfile() ?? {})
  const name = displayName(claims, t('layout.userMenu.unknownUser'))
  const email = claim(claims, 'email') ?? claim(claims, 'sub') ?? t('layout.userMenu.noEmail')
  const avatar = claim(claims, 'picture')

  const handleLogout = async (): Promise<void> => {
    if (logoutPending) {
      return
    }

    setLogoutFailed(false)
    setLogoutPending(true)
    try {
      await logout()
    } catch {
      setLogoutFailed(true)
    } finally {
      setLogoutPending(false)
    }
  }

  return (
    <>
      <UserMenu>
        <UserMenu.Trigger name={name} avatar={avatar} />
        <UserMenu.Header name={name} email={email} avatar={avatar} />
        <UserMenu.Divider />
        <UserMenu.Item
          icon={<User />}
          label={t('layout.userMenu.profile')}
          onClick={() => navigate('/profile')}
        />
        <UserMenu.Divider />
        <UserMenu.Logout label={t('layout.userMenu.signOut')} onClick={handleLogout} />
      </UserMenu>
      <Snackbar
        open={logoutFailed}
        onClose={() => setLogoutFailed(false)}
        anchorOrigin={{ vertical: 'bottom', horizontal: 'center' }}
      >
        <Alert
          severity="error"
          variant="filled"
          action={
            <Button color="inherit" size="small" disabled={logoutPending} onClick={handleLogout}>
              {t('layout.userMenu.tryAgain')}
            </Button>
          }
        >
          {t('layout.userMenu.signOutError')}
        </Alert>
      </Snackbar>
    </>
  )
}

export default UserProfileMenu
