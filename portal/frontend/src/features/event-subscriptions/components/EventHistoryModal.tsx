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

import {
  Box,
  Button,
  Chip,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  IconButton,
  Skeleton,
  Stack,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableRow,
  Typography,
} from '@wso2/oxygen-ui'
import { Activity, ChevronDown, ChevronUp, X } from '@wso2/oxygen-ui-icons-react'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import type { EventDeliveryHistory, FilterType } from '../../../types/eventSubscription'
import { formatEpochTimestamp } from '../../../utils/dateTime'
import { getDeliveryStatusChipColor } from '../utils/subscriptionStatusChip'

interface EventHistoryModalProps {
  open: boolean
  loading: boolean
  history: EventDeliveryHistory | undefined
  subscriptionFilter: { type: FilterType; purposes?: string[] } | undefined
  onClose: () => void
}

interface InfoFieldProps {
  label: string
  value: React.ReactNode
}

function InfoField({ label, value }: InfoFieldProps): React.JSX.Element {
  return (
    <Stack spacing={0.5} minWidth={0}>
      <Typography variant="caption" color="text.secondary" fontWeight={700}>
        {label}
      </Typography>
      <Typography component="div" variant="body2" sx={{ overflowWrap: 'anywhere' }}>
        {value || '-'}
      </Typography>
    </Stack>
  )
}

interface CollapsibleSectionProps {
  title: string
  children: React.ReactNode
}

function CollapsibleSection({ title, children }: CollapsibleSectionProps): React.JSX.Element {
  const [open, setOpen] = useState(false)

  return (
    <Box sx={{ border: 1, borderColor: 'divider', borderRadius: 1 }}>
      <Box
        role="button"
        tabIndex={0}
        onClick={() => setOpen((value) => !value)}
        onKeyDown={(event) => {
          if (event.key === 'Enter') setOpen((value) => !value)
        }}
        sx={(theme) => ({
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          px: 2,
          py: 1.25,
          cursor: 'pointer',
          ...theme.applyStyles('light', { bgcolor: theme.palette.grey[50] }),
          ...theme.applyStyles('dark', { bgcolor: 'rgba(255, 255, 255, 0.04)' }),
        })}
      >
        <Typography variant="subtitle2" fontWeight={600}>
          {title}
        </Typography>
        {open ? <ChevronUp size={16} /> : <ChevronDown size={16} />}
      </Box>
      {open ? <Box sx={{ p: 2, borderTop: 1, borderColor: 'divider' }}>{children}</Box> : null}
    </Box>
  )
}

function formatPayloadJson(payload: unknown): string {
  if (!payload) {
    return ''
  }
  if (typeof payload === 'string') {
    try {
      return JSON.stringify(JSON.parse(payload), null, 2)
    } catch {
      return payload
    }
  }
  try {
    return JSON.stringify(payload, null, 2)
  } catch {
    return String(payload)
  }
}

function EventHistoryModal({
  open,
  loading,
  history,
  subscriptionFilter,
  onClose,
}: EventHistoryModalProps): React.JSX.Element {
  const { t } = useTranslation('common')
  const payloadJson = formatPayloadJson(history?.payload)

  return (
    <Dialog open={open} onClose={onClose} maxWidth="md" fullWidth>
      <DialogTitle sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
        <Stack direction="row" spacing={1} alignItems="center">
          <Activity size={18} />
          <Typography component="span" fontWeight={700}>
            {t('eventSubscriptions.history.title')}
          </Typography>
        </Stack>
        <IconButton
          size="small"
          onClick={onClose}
          aria-label={t('eventSubscriptions.actions.close')}
        >
          <X size={18} />
        </IconButton>
      </DialogTitle>
      <DialogContent dividers>
        {loading || !history ? (
          <Stack spacing={2}>
            <Skeleton variant="rounded" height={90} />
            <Skeleton variant="rounded" height={48} />
            <Skeleton variant="rounded" height={48} />
          </Stack>
        ) : (
          <Stack spacing={3}>
            <Box
              sx={{
                display: 'grid',
                gridTemplateColumns: { xs: '1fr', sm: 'repeat(3, 1fr)' },
                gap: 2,
              }}
            >
              <InfoField
                label={t('eventSubscriptions.history.currentStatus')}
                value={
                  <Chip
                    size="small"
                    variant="outlined"
                    color={getDeliveryStatusChipColor(history.currentStatus)}
                    label={history.currentStatus}
                  />
                }
              />
              <InfoField label={t('eventSubscriptions.fields.topic')} value={history.topic} />
              {subscriptionFilter ? (
                <InfoField
                  label={t('eventSubscriptions.fields.purposeFilterMode')}
                  value={t(`eventSubscriptions.filterType.${subscriptionFilter.type}`)}
                />
              ) : null}
            </Box>

            <Box
              sx={{
                display: 'grid',
                gridTemplateColumns: { xs: '1fr', sm: 'repeat(3, 1fr)' },
                gap: 2,
              }}
            >
              <InfoField
                label={t('eventSubscriptions.fields.deliveryMode')}
                value={<Chip size="small" variant="outlined" label={history.deliveryMode} />}
              />
              <InfoField
                label={t('eventSubscriptions.history.occurredAt')}
                value={formatEpochTimestamp(history.occurredAt)}
              />
              {history.nextRetryAt ? (
                <InfoField
                  label={t('eventSubscriptions.history.nextRetryAt')}
                  value={formatEpochTimestamp(history.nextRetryAt)}
                />
              ) : null}
              {history.completionStatus ? (
                <InfoField
                  label={t('eventSubscriptions.history.completionStatus')}
                  value={history.completionStatus}
                />
              ) : null}
              {history.completionEvidence ? (
                <InfoField
                  label={t('eventSubscriptions.history.completionEvidence')}
                  value={history.completionEvidence}
                />
              ) : null}
            </Box>

            <Box
              sx={{
                display: 'grid',
                gridTemplateColumns: { xs: '1fr', sm: 'repeat(2, 1fr)' },
                gap: 2,
              }}
            >
              <InfoField label={t('eventSubscriptions.fields.eventId')} value={history.eventId} />
              <InfoField
                label={t('eventSubscriptions.fields.deliveryId')}
                value={history.deliveryId}
              />
            </Box>

            <CollapsibleSection title={t('eventSubscriptions.history.retryHistory')}>
              {history.history.length === 0 ? (
                <Typography variant="body2" color="text.secondary">
                  {t('eventSubscriptions.history.noRetries')}
                </Typography>
              ) : (
                <Table size="small">
                  <TableHead>
                    <TableRow>
                      <TableCell>{t('eventSubscriptions.history.attempt')}</TableCell>
                      <TableCell>{t('eventSubscriptions.history.status')}</TableCell>
                      <TableCell>{t('eventSubscriptions.history.timestamp')}</TableCell>
                      <TableCell>{t('eventSubscriptions.history.httpStatus')}</TableCell>
                      <TableCell>{t('eventSubscriptions.history.error')}</TableCell>
                    </TableRow>
                  </TableHead>
                  <TableBody>
                    {history.history.map((attempt) => (
                      <TableRow key={attempt.attempt}>
                        <TableCell>{attempt.attempt}</TableCell>
                        <TableCell>
                          <Chip
                            size="small"
                            variant="outlined"
                            color={getDeliveryStatusChipColor(attempt.status)}
                            label={attempt.status}
                          />
                        </TableCell>
                        <TableCell>{formatEpochTimestamp(attempt.timestamp)}</TableCell>
                        <TableCell>{attempt.httpStatus ?? '-'}</TableCell>
                        <TableCell>{attempt.error ?? '-'}</TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              )}
            </CollapsibleSection>

            <CollapsibleSection title={t('eventSubscriptions.history.eventPayload')}>
              {payloadJson ? (
                <Box
                  component="pre"
                  sx={{
                    m: 0,
                    p: 2,
                    borderRadius: 1,
                    bgcolor: 'action.hover',
                    overflow: 'auto',
                    maxHeight: '40vh',
                    fontSize: '0.8125rem',
                    fontFamily: 'monospace',
                  }}
                >
                  {payloadJson}
                </Box>
              ) : (
                <Typography variant="body2" color="text.secondary">
                  {t('eventSubscriptions.history.noPayload')}
                </Typography>
              )}
            </CollapsibleSection>
          </Stack>
        )}
      </DialogContent>
      <DialogActions sx={{ p: 2 }}>
        <Button variant="contained" onClick={onClose}>
          {t('eventSubscriptions.actions.close')}
        </Button>
      </DialogActions>
    </Dialog>
  )
}

export default EventHistoryModal
