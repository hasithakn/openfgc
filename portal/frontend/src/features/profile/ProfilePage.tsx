/*
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
 * Licensed under the Apache License, Version 2.0.
 */

import {
  Alert,
  Box,
  Button,
  Card,
  CardContent,
  CardHeader,
  Divider,
  Skeleton,
  Snackbar,
  Stack,
  TextField,
  Typography,
} from '@wso2/oxygen-ui'
import { Mail, Phone } from '@wso2/oxygen-ui-icons-react'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import HeaderBreadcrumbs from '../../components/layout/main-layout/HeaderBreadcrumbs'
import type { AddressAttr, EmailAttr, PhoneAttr } from '../../types/profile'
import { AddressesCard, MultiValueCard } from './components/ProfileContactCards'
import { useProfileQuery, useUpdateProfileMutation } from './hooks/useProfileQueries'

type EditableEmail = Required<EmailAttr> & { id: string }
type EditablePhone = Required<PhoneAttr> & { id: string }
type EditableAddress = Required<AddressAttr> & { id: string }

function emptyEmail(): EditableEmail {
  return { value: '', type: 'work', primary: false, id: crypto.randomUUID() }
}

function emptyPhone(): EditablePhone {
  return { value: '', type: 'work', primary: false, id: crypto.randomUUID() }
}

function emptyAddress(): EditableAddress {
  return {
    formatted: '',
    streetAddress: '',
    locality: '',
    region: '',
    postalCode: '',
    country: '',
    type: 'home',
    primary: false,
    id: crypto.randomUUID(),
  }
}

function normalizeEmails(emails: EmailAttr[]): EditableEmail[] {
  return emails.length > 0
    ? emails.map((email) => ({
        value: email.value,
        type: email.type ?? 'work',
        primary: Boolean(email.primary),
        id: crypto.randomUUID(),
      }))
    : [emptyEmail()]
}

function normalizePhones(phones: PhoneAttr[]): EditablePhone[] {
  return phones.length > 0
    ? phones.map((phone) => ({
        value: phone.value,
        type: phone.type ?? 'work',
        primary: Boolean(phone.primary),
        id: crypto.randomUUID(),
      }))
    : [emptyPhone()]
}

function normalizeAddresses(addresses: AddressAttr[]): EditableAddress[] {
  return addresses.length > 0
    ? addresses.map((address) => ({
        formatted: address.formatted ?? '',
        streetAddress: address.streetAddress ?? '',
        locality: address.locality ?? '',
        region: address.region ?? '',
        postalCode: address.postalCode ?? '',
        country: address.country ?? '',
        type: address.type ?? 'home',
        primary: Boolean(address.primary),
        id: crypto.randomUUID(),
      }))
    : [emptyAddress()]
}

// Read-only display keys don't need to survive edits/removals, so a content-derived id
// (rather than a stored uuid) is enough here and avoids regenerating one on every render.
function withEmailDisplayId(email: EmailAttr): EmailAttr & { id: string } {
  return { ...email, id: `${email.value}|${email.type ?? ''}` }
}

function withPhoneDisplayId(phone: PhoneAttr): PhoneAttr & { id: string } {
  return { ...phone, id: `${phone.value}|${phone.type ?? ''}` }
}

function withAddressDisplayId(address: AddressAttr): AddressAttr & { id: string } {
  return {
    ...address,
    id: [
      address.streetAddress,
      address.locality,
      address.region,
      address.postalCode,
      address.country,
      address.type,
    ].join('|'),
  }
}

function ProfilePage(): React.JSX.Element {
  const { t } = useTranslation('common')
  const profileQuery = useProfileQuery()
  const updateMutation = useUpdateProfileMutation()

  const [editing, setEditing] = useState(false)
  const [givenName, setGivenName] = useState('')
  const [familyName, setFamilyName] = useState('')
  const [emails, setEmails] = useState<EditableEmail[]>([])
  const [phones, setPhones] = useState<EditablePhone[]>([])
  const [addresses, setAddresses] = useState<EditableAddress[]>([])
  const [saveError, setSaveError] = useState<string>()
  const [saveSuccess, setSaveSuccess] = useState(false)

  const profile = profileQuery.data

  const startEditing = (): void => {
    if (!profile) return
    setGivenName(profile.givenName)
    setFamilyName(profile.familyName)
    setEmails(normalizeEmails(profile.emails))
    setPhones(normalizePhones(profile.phoneNumbers))
    setAddresses(normalizeAddresses(profile.addresses))
    setSaveError(undefined)
    setEditing(true)
  }

  const cancelEditing = (): void => {
    setEditing(false)
    setSaveError(undefined)
  }

  const handleSave = (): void => {
    setSaveError(undefined)
    updateMutation.mutate(
      {
        name: { givenName: givenName.trim(), familyName: familyName.trim() },
        emails: emails
          .filter((email) => email.value.trim())
          .map((email) => ({ value: email.value, type: email.type, primary: email.primary })),
        phoneNumbers: phones
          .filter((phone) => phone.value.trim())
          .map((phone) => ({ value: phone.value, type: phone.type, primary: phone.primary })),
        addresses: addresses
          .filter((address) =>
            [
              address.streetAddress,
              address.locality,
              address.region,
              address.postalCode,
              address.country,
            ].some((part) => part.trim()),
          )
          .map((address) => ({
            streetAddress: address.streetAddress,
            locality: address.locality,
            region: address.region,
            postalCode: address.postalCode,
            country: address.country,
            type: address.type,
            primary: address.primary,
            formatted: [
              address.streetAddress,
              address.locality,
              address.region,
              address.postalCode,
              address.country,
            ]
              .filter(Boolean)
              .join(', '),
          })),
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

  if (profileQuery.isLoading) {
    return (
      <Box component="main" sx={{ p: { xs: 2, md: 4 } }}>
        <Stack spacing={3}>
          <HeaderBreadcrumbs />
          <Skeleton width={300} height={48} />
          <Skeleton variant="rounded" height={160} />
          <Skeleton variant="rounded" height={160} />
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
      <Stack spacing={3}>
        <Stack
          direction={{ xs: 'column', md: 'row' }}
          justifyContent="space-between"
          alignItems={{ md: 'flex-end' }}
          spacing={2}
        >
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

        <Card sx={{ boxShadow: 1 }}>
          <CardHeader
            title={<Typography fontWeight={600}>{t('profile.fields.identity')}</Typography>}
          />
          <Divider />
          <CardContent>
            <Box
              sx={{
                display: 'grid',
                gridTemplateColumns: { xs: '1fr', sm: 'repeat(2, 1fr)', lg: 'repeat(3, 1fr)' },
                gap: 3,
              }}
            >
              <Stack spacing={0.5}>
                <Typography variant="caption" color="text.secondary">
                  {t('profile.fields.username')}
                </Typography>
                <Typography variant="body2">{profile.username || '-'}</Typography>
              </Stack>
              {editing ? (
                <>
                  <TextField
                    size="small"
                    label={t('profile.fields.givenName')}
                    value={givenName}
                    onChange={(event) => setGivenName(event.target.value)}
                  />
                  <TextField
                    size="small"
                    label={t('profile.fields.familyName')}
                    value={familyName}
                    onChange={(event) => setFamilyName(event.target.value)}
                  />
                </>
              ) : (
                <>
                  <Stack spacing={0.5}>
                    <Typography variant="caption" color="text.secondary">
                      {t('profile.fields.givenName')}
                    </Typography>
                    <Typography variant="body2">{profile.givenName || '-'}</Typography>
                  </Stack>
                  <Stack spacing={0.5}>
                    <Typography variant="caption" color="text.secondary">
                      {t('profile.fields.familyName')}
                    </Typography>
                    <Typography variant="body2">{profile.familyName || '-'}</Typography>
                  </Stack>
                </>
              )}
            </Box>
          </CardContent>
        </Card>

        <MultiValueCard
          icon={<Mail size={14} />}
          title={t('profile.fields.emails')}
          editing={editing}
          values={editing ? emails : profile.emails.map(withEmailDisplayId)}
          emptyLabel={t('profile.messages.noEmails')}
          addLabel={t('profile.actions.addEmail')}
          onAdd={() => setEmails((current) => [...current, emptyEmail()])}
          onRemove={(index) => setEmails((current) => current.filter((_, i) => i !== index))}
          onChange={(index, field, value) =>
            setEmails((current) =>
              current.map((item, i) => (i === index ? { ...item, [field]: value } : item)),
            )
          }
        />

        <MultiValueCard
          icon={<Phone size={14} />}
          title={t('profile.fields.phoneNumbers')}
          editing={editing}
          values={editing ? phones : profile.phoneNumbers.map(withPhoneDisplayId)}
          emptyLabel={t('profile.messages.noPhoneNumbers')}
          addLabel={t('profile.actions.addPhoneNumber')}
          onAdd={() => setPhones((current) => [...current, emptyPhone()])}
          onRemove={(index) => setPhones((current) => current.filter((_, i) => i !== index))}
          onChange={(index, field, value) =>
            setPhones((current) =>
              current.map((item, i) => (i === index ? { ...item, [field]: value } : item)),
            )
          }
        />

        <AddressesCard
          editing={editing}
          values={editing ? addresses : profile.addresses.map(withAddressDisplayId)}
          onAdd={() => setAddresses((current) => [...current, emptyAddress()])}
          onRemove={(index) => setAddresses((current) => current.filter((_, i) => i !== index))}
          onChange={(index, field, value) =>
            setAddresses((current) =>
              current.map((item, i) => (i === index ? { ...item, [field]: value } : item)),
            )
          }
        />
      </Stack>

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
