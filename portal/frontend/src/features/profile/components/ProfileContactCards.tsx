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
import type { EmailAttr } from '../../../types/profile'

const CONTACT_TYPES = ['work', 'home', 'other']

export type KeyedEmail = EmailAttr & { id: string }

interface MultiValueCardProps {
  icon: React.ReactNode
  title: string
  editing: boolean
  values: KeyedEmail[]
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
