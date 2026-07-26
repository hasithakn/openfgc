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
  Chip,
  Paper,
  Skeleton,
  Stack,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TablePagination,
  TableRow,
  TextField,
  Typography,
} from '@wso2/oxygen-ui'
import { Search } from '@wso2/oxygen-ui-icons-react'
import { useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useSearchParams } from 'react-router-dom'
import HeaderBreadcrumbs from '../../components/layout/main-layout/HeaderBreadcrumbs'
import { formatEpochTimestamp } from '../../utils/dateTime'
import EventHistoryModal from './components/EventHistoryModal'
import { useOrgEventHistoryQuery, useOrgEventListQuery } from './hooks/useEventSubscriptionQueries'
import { getDeliveryStatusChipColor } from './utils/subscriptionStatusChip'

const ROW_OPTIONS = [10, 25, 50]

function getPage(searchParams: URLSearchParams): number {
  const page = Number(searchParams.get('page') ?? '1')
  return Number.isInteger(page) && page > 0 ? page - 1 : 0
}

function getRowsPerPage(searchParams: URLSearchParams): number {
  const rows = Number(searchParams.get('rowsPerPage') ?? '10')
  return ROW_OPTIONS.includes(rows) ? rows : 10
}

function EventListPage(): React.JSX.Element {
  const { t } = useTranslation('common')
  const [searchParams, setSearchParams] = useSearchParams()
  const [selectedDeliveryId, setSelectedDeliveryId] = useState<string>()
  const page = useMemo(() => getPage(searchParams), [searchParams])
  const rowsPerPage = useMemo(() => getRowsPerPage(searchParams), [searchParams])
  const status = searchParams.get('status') ?? ''
  const subscriptionId = searchParams.get('subscriptionId') ?? ''
  const search = searchParams.get('search') ?? ''
  const query = useOrgEventListQuery(status, subscriptionId, search, page, rowsPerPage)
  const historyQuery = useOrgEventHistoryQuery(selectedDeliveryId)

  const updateParams = (
    next: { status: string; subscriptionId: string; search: string },
    nextPage = 0,
    nextRowsPerPage = rowsPerPage,
  ): void => {
    const params = new URLSearchParams()
    if (next.status.trim()) params.set('status', next.status.trim())
    if (next.subscriptionId.trim()) params.set('subscriptionId', next.subscriptionId.trim())
    if (next.search.trim()) params.set('search', next.search.trim())
    if (nextPage > 0) params.set('page', String(nextPage + 1))
    if (nextRowsPerPage !== 10) params.set('rowsPerPage', String(nextRowsPerPage))
    setSearchParams(params, { replace: true })
  }

  const rows = query.data?.content ?? []

  return (
    <Box component="main" sx={{ p: { xs: 2, md: 4 } }}>
      <Stack spacing={3}>
        <Stack spacing={1}>
          <HeaderBreadcrumbs />
          <Typography variant="h4" fontWeight={700}>
            {t('eventSubscriptions.eventList.title')}
          </Typography>
          <Typography variant="body2" color="text.secondary">
            {t('eventSubscriptions.eventList.subtitle')}
          </Typography>
        </Stack>

        <Stack direction={{ xs: 'column', lg: 'row' }} spacing={2}>
          <TextField
            size="small"
            label={t('eventSubscriptions.filters.status')}
            value={status}
            onChange={(event) =>
              updateParams({ status: event.target.value, subscriptionId, search })
            }
            sx={{ minWidth: { lg: 160 } }}
          />
          <TextField
            size="small"
            label={t('eventSubscriptions.fields.subscriptionId')}
            value={subscriptionId}
            onChange={(event) =>
              updateParams({ status, subscriptionId: event.target.value, search })
            }
            sx={{ minWidth: { lg: 220 } }}
          />
          <TextField
            size="small"
            fullWidth
            label={t('eventSubscriptions.filters.search')}
            value={search}
            onChange={(event) =>
              updateParams({ status, subscriptionId, search: event.target.value })
            }
          />
        </Stack>

        {query.isError ? (
          <Paper variant="outlined" sx={{ p: 2, borderColor: 'error.main' }}>
            <Typography color="error.main">
              {t('eventSubscriptions.messages.loadFailed')}
            </Typography>
          </Paper>
        ) : null}

        <TableContainer component={Paper} elevation={1}>
          <Table aria-label={t('eventSubscriptions.eventList.tableLabel')}>
            <TableHead
              sx={(theme) => ({
                '& .MuiTableCell-head': {
                  fontWeight: 600,
                  ...theme.applyStyles('light', { backgroundColor: theme.palette.grey[50] }),
                  ...theme.applyStyles('dark', { backgroundColor: 'rgba(255, 255, 255, 0.04)' }),
                },
              })}
            >
              <TableRow>
                <TableCell>{t('eventSubscriptions.fields.eventId')}</TableCell>
                <TableCell>{t('eventSubscriptions.fields.topic')}</TableCell>
                <TableCell>{t('eventSubscriptions.fields.status')}</TableCell>
                <TableCell>{t('eventSubscriptions.fields.deliveryMode')}</TableCell>
                <TableCell>{t('eventSubscriptions.history.occurredAt')}</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {query.isPending
                ? Array.from({ length: 5 }).map((_, index) => (
                    <TableRow key={`org-event-skeleton-${String(index)}`}>
                      {Array.from({ length: 5 }).map((__, cell) => (
                        <TableCell key={`org-event-skeleton-${String(index)}-${String(cell)}`}>
                          <Skeleton />
                        </TableCell>
                      ))}
                    </TableRow>
                  ))
                : rows.map((event) => (
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
                      <TableCell>{event.topic}</TableCell>
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
              {!query.isPending && !query.isError && rows.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={5} align="center" sx={{ py: 8 }}>
                    <Stack spacing={1} alignItems="center">
                      <Search size={28} aria-hidden="true" />
                      <Typography variant="body2" color="text.secondary">
                        {t('eventSubscriptions.messages.empty')}
                      </Typography>
                    </Stack>
                  </TableCell>
                </TableRow>
              ) : null}
            </TableBody>
          </Table>
          <TablePagination
            component="div"
            count={query.data?.totalElements ?? 0}
            page={page}
            rowsPerPage={rowsPerPage}
            rowsPerPageOptions={ROW_OPTIONS}
            onPageChange={(_, nextPage) =>
              updateParams({ status, subscriptionId, search }, nextPage)
            }
            onRowsPerPageChange={(event) =>
              updateParams({ status, subscriptionId, search }, 0, Number(event.target.value))
            }
          />
        </TableContainer>
      </Stack>

      <EventHistoryModal
        open={Boolean(selectedDeliveryId)}
        loading={historyQuery.isLoading}
        history={historyQuery.data}
        subscriptionFilter={undefined}
        onClose={() => setSelectedDeliveryId(undefined)}
      />
    </Box>
  )
}

export default EventListPage
