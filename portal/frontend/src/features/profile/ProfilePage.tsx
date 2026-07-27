/*
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
 * Licensed under the Apache License, Version 2.0.
 */

import {
  Alert,
  Box,
  Button,
  Divider,
  Skeleton,
  Snackbar,
  Stack,
  TextField,
  Typography,
} from '@wso2/oxygen-ui'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import HeaderBreadcrumbs from '../../components/layout/main-layout/HeaderBreadcrumbs'
import { logout } from '../../utils/authClient'
import DeleteAccountDialog from './components/DeleteAccountDialog'
import {
  useDeleteAccountMutation,
  useProfileQuery,
  useUpdateProfileMutation,
} from './hooks/useProfileQueries'

function pickPrimary<T extends { primary?: boolean }>(
  items: T[],
  pick: (item: T) => string,
): string {
  const item = items.find((candidate) => candidate.primary) ?? items[0]
  return item ? pick(item) : ''
}

const FIELD_LABEL_WIDTH = 110

function FieldRow({
  label,
  children,
}: {
  label: string
  children: React.ReactNode
}): React.JSX.Element {
  return (
    <Stack direction="row" spacing={2} alignItems="center">
      <Typography
        variant="body2"
        fontWeight={600}
        color="text.secondary"
        sx={{ minWidth: FIELD_LABEL_WIDTH }}
      >
        {label}:
      </Typography>
      <Box sx={{ flex: 1, minWidth: 0 }}>{children}</Box>
    </Stack>
  )
}

function ProfilePage(): React.JSX.Element {
  const { t } = useTranslation('common')
  const profileQuery = useProfileQuery()
  const updateMutation = useUpdateProfileMutation()
  const deleteAccountMutation = useDeleteAccountMutation()

  const [editing, setEditing] = useState(false)
  const [name, setName] = useState('')
  const [nickName, setNickName] = useState('')
  const [birthday, setBirthday] = useState('')
  const [phone, setPhone] = useState('')
  const [address, setAddress] = useState('')
  const [saveError, setSaveError] = useState<string>()
  const [saveSuccess, setSaveSuccess] = useState(false)
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)
  const [deleteError, setDeleteError] = useState<string>()

  const profile = profileQuery.data

  const startEditing = (): void => {
    if (!profile) return
    setName(profile.formattedName)
    setNickName(profile.nickName ?? '')
    setBirthday(profile.birthday ?? '')
    setPhone(pickPrimary(profile.phoneNumbers, (phoneNumber) => phoneNumber.value))
    setAddress(pickPrimary(profile.addresses, (item) => item.formatted))
    setSaveError(undefined)
    setEditing(true)
  }

  const cancelEditing = (): void => {
    setEditing(false)
    setSaveError(undefined)
  }

  const handleSave = (): void => {
    setSaveError(undefined)
    const trimmedPhone = phone.trim()
    const trimmedAddress = address.trim()
    updateMutation.mutate(
      {
        name: { givenName: name.trim(), familyName: '' },
        nickName: nickName.trim() || undefined,
        birthday: birthday.trim() || undefined,
        phoneNumbers: trimmedPhone ? [{ value: trimmedPhone, primary: true }] : undefined,
        addresses: trimmedAddress ? [{ formatted: trimmedAddress, primary: true }] : undefined,
      },
      {
        onSuccess: (): void => {
          setEditing(false)
          setSaveSuccess(true)
        },
        onError: (error): void => {
          setSaveError(error.message)
        },
      },
    )
  }

  const handleDeleteAccount = (): void => {
    setDeleteError(undefined)
    deleteAccountMutation.mutate(undefined, {
      onSuccess: (): void => {
        logout().catch(() => undefined)
      },
      onError: (error): void => {
        setDeleteError(error.message)
      },
    })
  }

  if (profileQuery.isLoading) {
    return (
      <Box component="main" sx={{ p: { xs: 2, md: 4 } }}>
        <Stack spacing={3}>
          <HeaderBreadcrumbs />
          <Skeleton width={300} height={48} />
          <Skeleton variant="rounded" height={160} />
        </Stack>
      </Box>
    )
  }

  if (profileQuery.isError || !profile) {
    return (
      <Box component="main" sx={{ p: { xs: 2, md: 4 } }}>
        <Stack spacing={2}>
          <HeaderBreadcrumbs />
          <Typography color="error.main">{t('profile.messages.loadFailed')}</Typography>
        </Stack>
      </Box>
    )
  }

  return (
    <Box component="main" sx={{ p: { xs: 2, md: 4 } }}>
      <Stack spacing={3} sx={{ maxWidth: 480 }}>
        <Stack direction="row" justifyContent="space-between" alignItems="flex-end" spacing={2}>
          <Stack spacing={0.75} minWidth={0}>
            <HeaderBreadcrumbs />
            <Typography variant="h4" fontWeight={700}>
              {t('profile.title')}
            </Typography>
          </Stack>
          {!editing ? (
            <Button variant="contained" onClick={startEditing}>
              {t('profile.actions.edit')}
            </Button>
          ) : (
            <Stack direction="row" spacing={1}>
              <Button
                variant="outlined"
                onClick={cancelEditing}
                disabled={updateMutation.isPending}
              >
                {t('profile.actions.cancel')}
              </Button>
              <Button variant="contained" onClick={handleSave} disabled={updateMutation.isPending}>
                {t('profile.actions.save')}
              </Button>
            </Stack>
          )}
        </Stack>

        {saveError ? <Alert severity="error">{saveError}</Alert> : null}

        <Stack spacing={2}>
          <FieldRow label={t('profile.fields.name')}>
            {editing ? (
              <TextField
                size="small"
                fullWidth
                value={name}
                onChange={(event) => setName(event.target.value)}
              />
            ) : (
              <Typography variant="body2">{profile.formattedName || '-'}</Typography>
            )}
          </FieldRow>

          <FieldRow label={t('profile.fields.nickName')}>
            {editing ? (
              <TextField
                size="small"
                fullWidth
                value={nickName}
                onChange={(event) => setNickName(event.target.value)}
              />
            ) : (
              <Typography variant="body2">{profile.nickName || '-'}</Typography>
            )}
          </FieldRow>

          <FieldRow label={t('profile.fields.birthday')}>
            {editing ? (
              <TextField
                size="small"
                fullWidth
                type="date"
                value={birthday}
                onChange={(event) => setBirthday(event.target.value)}
                slotProps={{ inputLabel: { shrink: true } }}
              />
            ) : (
              <Typography variant="body2">{profile.birthday || '-'}</Typography>
            )}
          </FieldRow>

          <FieldRow label={t('profile.fields.phone')}>
            {editing ? (
              <TextField
                size="small"
                fullWidth
                value={phone}
                onChange={(event) => setPhone(event.target.value)}
              />
            ) : (
              <Typography variant="body2">
                {pickPrimary(profile.phoneNumbers, (phoneNumber) => phoneNumber.value) || '-'}
              </Typography>
            )}
          </FieldRow>

          <FieldRow label={t('profile.fields.address')}>
            {editing ? (
              <TextField
                size="small"
                fullWidth
                value={address}
                onChange={(event) => setAddress(event.target.value)}
              />
            ) : (
              <Typography variant="body2">
                {pickPrimary(profile.addresses, (item) => item.formatted) || '-'}
              </Typography>
            )}
          </FieldRow>

          <FieldRow label={t('profile.fields.email')}>
            <Typography variant="body2">
              {pickPrimary(profile.emails, (email) => email.value) || '-'}
            </Typography>
          </FieldRow>
        </Stack>

        <Divider />

        <Stack spacing={1} alignItems="flex-start">
          <Typography variant="body2" color="text.secondary">
            {t(
              'profile.deleteAccount.description',
              'Permanently delete your account and anonymize your data.',
            )}
          </Typography>
          <Button color="error" variant="outlined" onClick={() => setDeleteDialogOpen(true)}>
            {t('profile.deleteAccount.button', 'Delete Account')}
          </Button>
        </Stack>
      </Stack>

      <DeleteAccountDialog
        open={deleteDialogOpen}
        loading={deleteAccountMutation.isPending}
        error={deleteError}
        onClose={() => setDeleteDialogOpen(false)}
        onConfirm={handleDeleteAccount}
      />

      <Snackbar
        open={saveSuccess}
        autoHideDuration={4000}
        onClose={() => setSaveSuccess(false)}
        anchorOrigin={{ vertical: 'bottom', horizontal: 'center' }}
      >
        <Alert severity="success" variant="filled">
          {t('profile.messages.updateSuccess')}
        </Alert>
      </Snackbar>
    </Box>
  )
}

export default ProfilePage
