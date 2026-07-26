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
  FormControl,
  IconButton,
  InputLabel,
  MenuItem,
  Paper,
  Select,
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
  Tooltip,
  Typography,
} from '@wso2/oxygen-ui'
import { Eye, Plus, Search } from '@wso2/oxygen-ui-icons-react'
import { useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useNavigate, useSearchParams } from 'react-router-dom'
import HeaderBreadcrumbs from '../../components/layout/main-layout/HeaderBreadcrumbs'
import type { SubscriptionListFilters, SubscriptionStatus } from '../../types/eventSubscription'
import { formatEpochTimestamp } from '../../utils/dateTime'
import SubscriptionFormDialog from './components/SubscriptionFormDialog'
import {
  useCreateSubscriptionMutation,
  useSubscriptionListQuery,
} from './hooks/useEventSubscriptionQueries'
import { getSubscriptionStatusChipColor } from './utils/subscriptionStatusChip'

const ROW_OPTIONS = [10, 25, 50]
const STATUS_VALUES: SubscriptionStatus[] = ['active', 'pending', 'stale', 'deleted']

function getPage(searchParams: URLSearchParams): number {
  const page = Number(searchParams.get('page') ?? '1')
  return Number.isInteger(page) && page > 0 ? page - 1 : 0
}

function getRowsPerPage(searchParams: URLSearchParams): number {
  const rows = Number(searchParams.get('rowsPerPage') ?? '10')
  return ROW_OPTIONS.includes(rows) ? rows : 10
}

function getFilters(searchParams: URLSearchParams): SubscriptionListFilters {
  const status = searchParams.get('status')
  return {
    status:
      status && STATUS_VALUES.includes(status as SubscriptionStatus)
        ? (status as SubscriptionStatus)
        : 'All',
    search: searchParams.get('search') ?? '',
  }
}

function SubscriptionListPage(): React.JSX.Element {
  const { t } = useTranslation('common')
  const navigate = useNavigate()
  const [searchParams, setSearchParams] = useSearchParams()
  const [createOpen, setCreateOpen] = useState(false)
  const filters = useMemo(() => getFilters(searchParams), [searchParams])
  const page = useMemo(() => getPage(searchParams), [searchParams])
  const rowsPerPage = useMemo(() => getRowsPerPage(searchParams), [searchParams])
  const query = useSubscriptionListQuery(filters, page, rowsPerPage)
  const createMutation = useCreateSubscriptionMutation()

  const updateParams = (
    nextFilters: SubscriptionListFilters,
    nextPage = 0,
    nextRowsPerPage = rowsPerPage,
  ): void => {
    const params = new URLSearchParams()
    if (nextFilters.status !== 'All') params.set('status', nextFilters.status)
    if (nextFilters.search.trim()) params.set('search', nextFilters.search.trim())
    if (nextPage > 0) params.set('page', String(nextPage + 1))
    if (nextRowsPerPage !== 10) params.set('rowsPerPage', String(nextRowsPerPage))
    setSearchParams(params, { replace: true })
  }

  const rows = query.data?.subscriptions ?? []

  return (
    <Box component="main" sx={{ p: { xs: 2, md: 4 } }}>
      <Stack spacing={3}>
        <Stack
          direction={{ xs: 'column', sm: 'row' }}
          justifyContent="space-between"
          alignItems={{ sm: 'flex-end' }}
          spacing={2}
        >
          <Stack spacing={1}>
            <HeaderBreadcrumbs />
            <Typography variant="h4" fontWeight={700}>
              {t('eventSubscriptions.list.title')}
            </Typography>
            <Typography variant="body2" color="text.secondary">
              {t('eventSubscriptions.list.subtitle')}
            </Typography>
          </Stack>
          <Button
            variant="contained"
            startIcon={<Plus size={18} />}
            onClick={() => setCreateOpen(true)}
          >
            {t('eventSubscriptions.actions.addSubscription')}
          </Button>
        </Stack>

        <Stack direction={{ xs: 'column', lg: 'row' }} spacing={2} alignItems={{ lg: 'center' }}>
          <FormControl size="small" sx={{ minWidth: { lg: 180 } }}>
            <InputLabel id="subscription-status-label">
              {t('eventSubscriptions.filters.status')}
            </InputLabel>
            <Select
              labelId="subscription-status-label"
              value={filters.status}
              label={t('eventSubscriptions.filters.status')}
              onChange={(event) =>
                updateParams({
                  ...filters,
                  status: event.target.value as SubscriptionStatus | 'All',
                })
              }
            >
              <MenuItem value="All">{t('eventSubscriptions.filters.allStatuses')}</MenuItem>
              {STATUS_VALUES.map((status) => (
                <MenuItem key={status} value={status}>
                  {status}
                </MenuItem>
              ))}
            </Select>
          </FormControl>
          <TextField
            size="small"
            fullWidth
            label={t('eventSubscriptions.filters.search')}
            helperText={t('eventSubscriptions.filters.searchHelp')}
            value={filters.search}
            onChange={(event) => updateParams({ ...filters, search: event.target.value })}
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
          <Table aria-label={t('eventSubscriptions.list.tableLabel')}>
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
                <TableCell>{t('eventSubscriptions.fields.subscriptionId')}</TableCell>
                <TableCell>{t('eventSubscriptions.fields.topic')}</TableCell>
                <TableCell>{t('eventSubscriptions.fields.purposeFilterMode')}</TableCell>
                <TableCell>{t('eventSubscriptions.fields.callbackUrl')}</TableCell>
                <TableCell>{t('eventSubscriptions.fields.status')}</TableCell>
                <TableCell>{t('eventSubscriptions.fields.deliveryMode')}</TableCell>
                <TableCell>{t('eventSubscriptions.fields.created')}</TableCell>
                <TableCell align="right">{t('eventSubscriptions.fields.actions')}</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {query.isPending
                ? Array.from({ length: 5 }).map((_, index) => (
                    <TableRow key={`subscription-skeleton-${String(index)}`}>
                      {Array.from({ length: 8 }).map((__, cell) => (
                        <TableCell key={`subscription-skeleton-${String(index)}-${String(cell)}`}>
                          <Skeleton />
                        </TableCell>
                      ))}
                    </TableRow>
                  ))
                : rows.map((subscription) => (
                    <TableRow
                      hover
                      key={subscription.subscriptionId}
                      tabIndex={0}
                      sx={{ cursor: 'pointer' }}
                      onClick={() =>
                        navigate(
                          `/subscriptions/${encodeURIComponent(subscription.subscriptionId)}`,
                        )
                      }
                    >
                      <TableCell>
                        <Typography component="code" variant="body2">
                          {subscription.subscriptionId}
                        </Typography>
                      </TableCell>
                      <TableCell>{subscription.topic}</TableCell>
                      <TableCell>{subscription.filter.type}</TableCell>
                      <TableCell>{subscription.delivery.callbackUrl ?? '-'}</TableCell>
                      <TableCell>
                        <Chip
                          size="small"
                          variant="outlined"
                          color={getSubscriptionStatusChipColor(subscription.status)}
                          label={subscription.status}
                        />
                      </TableCell>
                      <TableCell>
                        <Chip size="small" variant="outlined" label={subscription.delivery.mode} />
                      </TableCell>
                      <TableCell>{formatEpochTimestamp(subscription.createdAt)}</TableCell>
                      <TableCell align="right">
                        <Tooltip title={t('eventSubscriptions.actions.view')}>
                          <IconButton
                            size="small"
                            aria-label={t('eventSubscriptions.actions.view')}
                            onClick={(event) => {
                              event.stopPropagation()
                              navigate(
                                `/subscriptions/${encodeURIComponent(subscription.subscriptionId)}`,
                              )
                            }}
                          >
                            <Eye size={17} />
                          </IconButton>
                        </Tooltip>
                      </TableCell>
                    </TableRow>
                  ))}
              {!query.isPending && !query.isError && rows.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={8} align="center" sx={{ py: 8 }}>
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
            count={query.data?.total ?? 0}
            page={page}
            rowsPerPage={rowsPerPage}
            rowsPerPageOptions={ROW_OPTIONS}
            onPageChange={(_, nextPage) => updateParams(filters, nextPage)}
            onRowsPerPageChange={(event) => updateParams(filters, 0, Number(event.target.value))}
          />
        </TableContainer>
      </Stack>

      <SubscriptionFormDialog
        key={`create-subscription-${String(createOpen)}`}
        open={createOpen}
        loading={createMutation.isPending}
        error={createMutation.error?.message}
        onClose={() => {
          setCreateOpen(false)
          createMutation.reset()
        }}
        onCreate={(payload) => {
          createMutation.mutate(payload, {
            onSuccess: (subscription) => {
              setCreateOpen(false)
              navigate(`/subscriptions/${encodeURIComponent(subscription.subscriptionId)}`)
            },
          })
        }}
      />
    </Box>
  )
}

export default SubscriptionListPage
