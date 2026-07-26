/*
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
 * Licensed under the Apache License, Version 2.0.
 */

import {
  Box,
  Card,
  CardContent,
  CardHeader,
  Divider,
  IconButton,
  MenuItem,
  Stack,
  TextField,
  Typography,
} from '@wso2/oxygen-ui'
import { Plus, Trash2 } from '@wso2/oxygen-ui-icons-react'
import { useTranslation } from 'react-i18next'
import type { AddressAttr, EmailAttr, PhoneAttr } from '../../../types/profile'

const CONTACT_TYPES = ['work', 'home', 'other']

export type KeyedEmail = EmailAttr & { id: string }
export type KeyedPhone = PhoneAttr & { id: string }
export type KeyedAddress = AddressAttr & { id: string }

interface MultiValueCardProps {
  icon: React.ReactNode
  title: string
  editing: boolean
  values: (KeyedEmail | KeyedPhone)[]
  emptyLabel: string
  addLabel: string
  onAdd: () => void
  onRemove: (index: number) => void
  onChange: (index: number, field: 'value' | 'type', value: string) => void
}

export function MultiValueCard({
  icon,
  title,
  editing,
  values,
  emptyLabel,
  addLabel,
  onAdd,
  onRemove,
  onChange,
}: MultiValueCardProps): React.JSX.Element {
  const { t } = useTranslation('common')

  return (
    <Card sx={{ boxShadow: 1 }}>
      <CardHeader
        title={
          <Stack direction="row" spacing={1} alignItems="center">
            {icon}
            <Typography fontWeight={600}>{title}</Typography>
          </Stack>
        }
        action={
          editing ? (
            <IconButton size="small" aria-label={addLabel} onClick={onAdd}>
              <Plus size={18} />
            </IconButton>
          ) : null
        }
      />
      <Divider />
      <CardContent>
        {values.length === 0 ? (
          <Typography variant="body2" color="text.secondary">
            {emptyLabel}
          </Typography>
        ) : (
          <Stack spacing={2}>
            {values.map((item, index) =>
              editing ? (
                <Stack
                  key={item.id}
                  direction={{ xs: 'column', sm: 'row' }}
                  spacing={1}
                  alignItems={{ sm: 'center' }}
                >
                  <TextField
                    size="small"
                    fullWidth
                    label={t('profile.fields.value')}
                    value={item.value}
                    onChange={(event) => onChange(index, 'value', event.target.value)}
                  />
                  <TextField
                    select
                    size="small"
                    label={t('profile.fields.type')}
                    value={item.type ?? 'work'}
                    sx={{ minWidth: 130 }}
                    onChange={(event) => onChange(index, 'type', event.target.value)}
                  >
                    {CONTACT_TYPES.map((type) => (
                      <MenuItem key={type} value={type}>
                        {type}
                      </MenuItem>
                    ))}
                  </TextField>
                  <IconButton
                    size="small"
                    aria-label={t('profile.actions.remove')}
                    onClick={() => onRemove(index)}
                  >
                    <Trash2 size={16} />
                  </IconButton>
                </Stack>
              ) : (
                <Box key={item.id}>
                  <Typography variant="body2">
                    {item.value}
                    {item.type ? ` (${item.type})` : ''}
                    {item.primary ? ` · ${t('profile.values.primary')}` : ''}
                  </Typography>
                </Box>
              ),
            )}
          </Stack>
        )}
      </CardContent>
    </Card>
  )
}

interface AddressesCardProps {
  editing: boolean
  values: KeyedAddress[]
  onAdd: () => void
  onRemove: (index: number) => void
  onChange: (index: number, field: keyof AddressAttr, value: string) => void
}

export function AddressesCard({
  editing,
  values,
  onAdd,
  onRemove,
  onChange,
}: AddressesCardProps): React.JSX.Element {
  const { t } = useTranslation('common')

  return (
    <Card sx={{ boxShadow: 1 }}>
      <CardHeader
        title={<Typography fontWeight={600}>{t('profile.fields.addresses')}</Typography>}
        action={
          editing ? (
            <IconButton size="small" aria-label={t('profile.actions.addAddress')} onClick={onAdd}>
              <Plus size={18} />
            </IconButton>
          ) : null
        }
      />
      <Divider />
      <CardContent>
        {values.length === 0 ? (
          <Typography variant="body2" color="text.secondary">
            {t('profile.messages.noAddresses')}
          </Typography>
        ) : (
          <Stack spacing={3}>
            {values.map((address, index) =>
              editing ? (
                <Stack key={address.id} spacing={1}>
                  <Box
                    sx={{
                      display: 'grid',
                      gridTemplateColumns: { xs: '1fr', sm: 'repeat(2, 1fr)' },
                      gap: 1.5,
                    }}
                  >
                    <TextField
                      size="small"
                      label={t('profile.fields.streetAddress')}
                      value={address.streetAddress ?? ''}
                      onChange={(event) => onChange(index, 'streetAddress', event.target.value)}
                    />
                    <TextField
                      size="small"
                      label={t('profile.fields.locality')}
                      value={address.locality ?? ''}
                      onChange={(event) => onChange(index, 'locality', event.target.value)}
                    />
                    <TextField
                      size="small"
                      label={t('profile.fields.region')}
                      value={address.region ?? ''}
                      onChange={(event) => onChange(index, 'region', event.target.value)}
                    />
                    <TextField
                      size="small"
                      label={t('profile.fields.postalCode')}
                      value={address.postalCode ?? ''}
                      onChange={(event) => onChange(index, 'postalCode', event.target.value)}
                    />
                    <TextField
                      size="small"
                      label={t('profile.fields.country')}
                      value={address.country ?? ''}
                      onChange={(event) => onChange(index, 'country', event.target.value)}
                    />
                  </Box>
                  <Stack direction="row" justifyContent="flex-end">
                    <IconButton
                      size="small"
                      aria-label={t('profile.actions.remove')}
                      onClick={() => onRemove(index)}
                    >
                      <Trash2 size={16} />
                    </IconButton>
                  </Stack>
                </Stack>
              ) : (
                <Stack key={address.id} spacing={0.5}>
                  <Typography variant="body2">
                    {[
                      address.streetAddress,
                      address.locality,
                      address.region,
                      address.postalCode,
                      address.country,
                    ]
                      .filter(Boolean)
                      .join(', ') ||
                      address.formatted ||
                      '-'}
                  </Typography>
                </Stack>
              ),
            )}
          </Stack>
        )}
      </CardContent>
    </Card>
  )
}
