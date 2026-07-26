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
  Card,
  CardContent,
  CardHeader,
  Chip,
  Divider,
  Skeleton,
  Stack,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TablePagination,
  TableRow,
  Typography,
} from '@wso2/oxygen-ui'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useNavigate, useParams } from 'react-router-dom'
import HeaderBreadcrumbs from '../../components/layout/main-layout/HeaderBreadcrumbs'
import { formatEpochTimestamp } from '../../utils/dateTime'
import DeleteSubscriptionDialog from './components/DeleteSubscriptionDialog'
import EventHistoryModal from './components/EventHistoryModal'
import {
  useDeleteSubscriptionMutation,
  useSubscriptionEventHistoryQuery,
  useSubscriptionEventsQuery,
  useSubscriptionQuery,
} from './hooks/useEventSubscriptionQueries'
import {
  getDeliveryStatusChipColor,
  getSubscriptionStatusChipColor,
} from './utils/subscriptionStatusChip'

const ROW_OPTIONS = [10, 25, 50]

function SubscriptionDetailsPage(): React.JSX.Element {
  const { t } = useTranslation('common')
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const detailQuery = useSubscriptionQuery(id)
  const [page, setPage] = useState(0)
  const [rowsPerPage, setRowsPerPage] = useState(10)
  const eventsQuery = useSubscriptionEventsQuery(id, page, rowsPerPage)
  const deleteMutation = useDeleteSubscriptionMutation()
  const [deleteOpen, setDeleteOpen] = useState(false)
  const [selectedDeliveryId, setSelectedDeliveryId] = useState<string>()
  const historyQuery = useSubscriptionEventHistoryQuery(id, selectedDeliveryId)

  const detail = detailQuery.data
  const events = eventsQuery.data?.content ?? []

  if (detailQuery.isLoading) {
    return (
      <Box component="main" sx={{ p: { xs: 2, md: 4 } }}>
        <Stack spacing={3}>
          <HeaderBreadcrumbs />
          <Skeleton width={300} height={48} />
          <Skeleton variant="rounded" height={190} />
          <Skeleton variant="rounded" height={260} />
        </Stack>
      </Box>
    )
  }

  if (!id || detailQuery.isError || !detail) {
    return (
      <Box component="main" sx={{ p: { xs: 2, md: 4 } }}>
        <Stack spacing={2}>
          <Typography color="error.main">{t('eventSubscriptions.messages.loadFailed')}</Typography>
          <Button variant="outlined" onClick={() => navigate('/subscriptions')}>
            {t('eventSubscriptions.details.back')}
          </Button>
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
            <Typography
              variant="h4"
              fontWeight={700}
              component="code"
              sx={{ overflowWrap: 'anywhere' }}
            >
              {detail.subscriptionId}
            </Typography>
          </Stack>
          <Button variant="contained" color="error" onClick={() => setDeleteOpen(true)}>
            {t('eventSubscriptions.actions.delete')}
          </Button>
        </Stack>

        <Card sx={{ boxShadow: 1 }}>
          <CardHeader
            title={
              <Typography fontWeight={600}>{t('eventSubscriptions.details.title')}</Typography>
            }
          />
          <Divider />
          <CardContent>
            <Box
              sx={{
                display: 'grid',
                gridTemplateColumns: { xs: '1fr', sm: 'repeat(2, 1fr)', lg: 'repeat(3, 1fr)' },
                gap: { xs: 2, md: 3 },
              }}
            >
              <Stack spacing={0.5}>
                <Typography variant="caption" color="text.secondary" fontWeight={700}>
                  {t('eventSubscriptions.fields.topic')}
                </Typography>
                <Typography variant="body2">{detail.topic}</Typography>
              </Stack>
              <Stack spacing={0.5}>
                <Typography variant="caption" color="text.secondary" fontWeight={700}>
                  {t('eventSubscriptions.fields.callbackUrl')}
                </Typography>
                <Typography variant="body2" sx={{ overflowWrap: 'anywhere' }}>
                  {detail.delivery.callbackUrl ?? '-'}
                </Typography>
              </Stack>
              <Stack spacing={0.5}>
                <Typography variant="caption" color="text.secondary" fontWeight={700}>
                  {t('eventSubscriptions.fields.purposeFilterMode')}
                </Typography>
                <Typography variant="body2">{detail.filter.type}</Typography>
              </Stack>
              <Stack spacing={0.5}>
                <Typography variant="caption" color="text.secondary" fontWeight={700}>
                  {t('eventSubscriptions.fields.status')}
                </Typography>
                <Chip
                  size="small"
                  variant="outlined"
                  color={getSubscriptionStatusChipColor(detail.status)}
                  label={detail.status}
                />
              </Stack>
              <Stack spacing={0.5}>
                <Typography variant="caption" color="text.secondary" fontWeight={700}>
                  {t('eventSubscriptions.fields.deliveryMode')}
                </Typography>
                <Chip size="small" variant="outlined" label={detail.delivery.mode} />
              </Stack>
              <Stack spacing={0.5}>
                <Typography variant="caption" color="text.secondary" fontWeight={700}>
                  {t('eventSubscriptions.fields.updated')}
                </Typography>
                <Typography variant="body2">{formatEpochTimestamp(detail.updatedAt)}</Typography>
              </Stack>
            </Box>
          </CardContent>
        </Card>

        <Card sx={{ boxShadow: 1 }}>
          <CardHeader
            title={
              <Typography fontWeight={600}>{t('eventSubscriptions.details.events')}</Typography>
            }
          />
          <Divider />
          {eventsQuery.isError ? (
            <CardContent>
              <Typography color="error.main">
                {t('eventSubscriptions.messages.eventsLoadFailed')}
              </Typography>
            </CardContent>
          ) : (
            <TableContainer>
              <Table size="small">
                <TableHead>
                  <TableRow>
                    <TableCell>{t('eventSubscriptions.fields.eventId')}</TableCell>
                    <TableCell>{t('eventSubscriptions.fields.status')}</TableCell>
                    <TableCell>{t('eventSubscriptions.fields.deliveryMode')}</TableCell>
                    <TableCell>{t('eventSubscriptions.history.occurredAt')}</TableCell>
                  </TableRow>
                </TableHead>
                <TableBody>
                  {eventsQuery.isPending
                    ? Array.from({ length: 3 }).map((_, index) => (
                        <TableRow key={`event-skeleton-${String(index)}`}>
                          {Array.from({ length: 4 }).map((__, cell) => (
                            <TableCell key={`event-skeleton-${String(index)}-${String(cell)}`}>
                              <Skeleton />
                            </TableCell>
                          ))}
                        </TableRow>
                      ))
                    : events.map((event) => (
                        <TableRow
                          hover
                          key={event.deliveryId}
                          sx={{ cursor: 'pointer' }}
                          onClick={() => setSelectedDeliveryId(event.deliveryId)}
                        >
                          <TableCell>
                            <Typography component="code" variant="body2">
                              {event.eventId}
                            </Typography>
                          </TableCell>
                          <TableCell>
                            <Chip
                              size="small"
                              variant="outlined"
                              color={getDeliveryStatusChipColor(event.currentStatus)}
                              label={event.currentStatus}
                            />
                          </TableCell>
                          <TableCell>
                            <Chip size="small" variant="outlined" label={event.deliveryMode} />
                          </TableCell>
                          <TableCell>{formatEpochTimestamp(event.occurredAt)}</TableCell>
                        </TableRow>
                      ))}
                  {!eventsQuery.isPending && events.length === 0 ? (
                    <TableRow>
                      <TableCell colSpan={4} align="center" sx={{ py: 4, color: 'text.secondary' }}>
                        {t('eventSubscriptions.messages.noEvents')}
                      </TableCell>
                    </TableRow>
                  ) : null}
                </TableBody>
              </Table>
              <TablePagination
                component="div"
                count={eventsQuery.data?.totalElements ?? 0}
                page={page}
                rowsPerPage={rowsPerPage}
                rowsPerPageOptions={ROW_OPTIONS}
                onPageChange={(_, nextPage) => setPage(nextPage)}
                onRowsPerPageChange={(event) => {
                  setRowsPerPage(Number(event.target.value))
                  setPage(0)
                }}
              />
            </TableContainer>
          )}
        </Card>
      </Stack>

      <DeleteSubscriptionDialog
        open={deleteOpen}
        subscriptionId={detail.subscriptionId}
        loading={deleteMutation.isPending}
        error={deleteMutation.error?.message}
        onClose={() => {
          setDeleteOpen(false)
          deleteMutation.reset()
        }}
        onConfirm={() => {
          deleteMutation.mutate(detail.subscriptionId, {
            onSuccess: () => navigate('/subscriptions'),
          })
        }}
      />

      <EventHistoryModal
        open={Boolean(selectedDeliveryId)}
        loading={historyQuery.isLoading}
        history={historyQuery.data}
        subscriptionFilter={detail.filter}
        onClose={() => setSelectedDeliveryId(undefined)}
      />
    </Box>
  )
}

export default SubscriptionDetailsPage
