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
  Alert,
  Box,
  Button,
  FormControl,
  InputLabel,
  MenuItem,
  Select,
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
import { Plus } from '@wso2/oxygen-ui-icons-react'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useNavigate } from 'react-router-dom'
import HeaderBreadcrumbs from '../../components/layout/main-layout/HeaderBreadcrumbs'
import { GRIEVANCE_STATUSES, type GrievanceStatus } from '../../types/grievance'
import { formatIsoDateTime } from '../../utils/dateTime'
import GrievanceStatusChip from './components/GrievanceStatusChip'
import GrievanceSubmitDialog from './components/GrievanceSubmitDialog'
import { GRIEVANCE_LIST_ROWS_PER_PAGE_OPTIONS } from './constants'
import { useMyGrievanceListQuery } from './hooks/useGrievanceQueries'
import { getGrievanceStatusLabelKey } from './utils/grievanceDisplay'

type StatusFilter = GrievanceStatus | 'All'

const DATE_FORMAT_OPTIONS: Intl.DateTimeFormatOptions = {
  month: 'short',
  day: '2-digit',
  year: 'numeric',
}

function GrievanceListPage(): React.JSX.Element {
  const { t } = useTranslation('common')
  const navigate = useNavigate()
  const [isSubmitDialogOpen, setIsSubmitDialogOpen] = useState<boolean>(false)
  const [submittedReferenceId, setSubmittedReferenceId] = useState<string | null>(null)
  const [statusFilter, setStatusFilter] = useState<StatusFilter>('All')
  const [page, setPage] = useState<number>(0)
  const [rowsPerPage, setRowsPerPage] = useState<number>(GRIEVANCE_LIST_ROWS_PER_PAGE_OPTIONS[0])

  const listQuery = useMyGrievanceListQuery(statusFilter, page, rowsPerPage)
  const rows = listQuery.data?.rows ?? []
  const total = listQuery.data?.total ?? 0

  return (
    <Box component="main" sx={{ p: { xs: 2, md: 4 } }}>
      <Stack spacing={3}>
        <Stack
          direction={{ xs: 'column', sm: 'row' }}
          spacing={2}
          alignItems={{ sm: 'center' }}
          justifyContent="space-between"
        >
          <Stack spacing={1}>
            <HeaderBreadcrumbs />
            <Typography variant="h4" fontWeight={700}>
              {t('grievances.list.title')}
            </Typography>
            <Typography variant="body2" color="text.secondary">
              {t('grievances.list.subtitle')}
            </Typography>
          </Stack>
          <Button
            variant="contained"
            startIcon={<Plus size={16} />}
            onClick={() => {
              setIsSubmitDialogOpen(true)
            }}
          >
            {t('grievances.list.submitNew')}
          </Button>
        </Stack>

        {submittedReferenceId ? (
          <Alert severity="success" onClose={() => setSubmittedReferenceId(null)}>
            {t('grievances.submit.success.message', { referenceId: submittedReferenceId })}
          </Alert>
        ) : null}

        <FormControl size="small" sx={{ width: { xs: '100%', sm: 'auto' }, minWidth: { sm: 200 } }}>
          <InputLabel id="grievance-list-status-label">
            {t('grievances.list.filters.status')}
          </InputLabel>
          <Select
            labelId="grievance-list-status-label"
            id="grievance-list-status"
            value={statusFilter}
            label={t('grievances.list.filters.status')}
            onChange={(event) => {
              setStatusFilter(event.target.value as StatusFilter)
              setPage(0)
            }}
          >
            <MenuItem value="All">{t('grievances.list.filters.all')}</MenuItem>
            {GRIEVANCE_STATUSES.map((status) => (
              <MenuItem key={status} value={status}>
                {status === 'AwaitingInfo'
                  ? t('grievances.status.awaitingInfoDataPrincipal')
                  : t(`grievances.status.${getGrievanceStatusLabelKey(status)}`)}
              </MenuItem>
            ))}
          </Select>
        </FormControl>

        {listQuery.isError ? (
          <Alert severity="error">{t('grievances.list.loadError')}</Alert>
        ) : null}

        {!listQuery.isError && !listQuery.isLoading && rows.length === 0 ? (
          <Typography>
            {statusFilter === 'All'
              ? t('grievances.list.empty')
              : t('grievances.list.emptyFiltered')}
          </Typography>
        ) : null}

        {!listQuery.isError && (rows.length > 0 || listQuery.isLoading) ? (
          <TableContainer sx={{ border: 1, borderColor: 'divider', borderRadius: 1 }}>
            <Table aria-label={t('grievances.list.table.tableAriaLabel')}>
              <TableHead>
                <TableRow>
                  <TableCell>{t('grievances.list.table.headers.referenceId')}</TableCell>
                  <TableCell>{t('grievances.list.table.headers.category')}</TableCell>
                  <TableCell>{t('grievances.list.table.headers.status')}</TableCell>
                  <TableCell>{t('grievances.list.table.headers.submitted')}</TableCell>
                  <TableCell>{t('grievances.list.table.headers.updated')}</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {rows.map((row) => (
                  <TableRow
                    key={row.id}
                    hover
                    onClick={() => {
                      navigate(`/grievances/${encodeURIComponent(row.id)}`)
                    }}
                    sx={{ cursor: 'pointer' }}
                  >
                    <TableCell sx={{ fontWeight: 600 }}>{row.referenceId}</TableCell>
                    <TableCell>{t(`grievances.categories.${row.category}`)}</TableCell>
                    <TableCell>
                      <GrievanceStatusChip status={row.status} viewerRole="DataPrincipal" />
                    </TableCell>
                    <TableCell>{formatIsoDateTime(row.submittedAt, DATE_FORMAT_OPTIONS)}</TableCell>
                    <TableCell>{formatIsoDateTime(row.updatedAt, DATE_FORMAT_OPTIONS)}</TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
            <TablePagination
              component="div"
              count={total}
              page={page}
              rowsPerPage={rowsPerPage}
              rowsPerPageOptions={[...GRIEVANCE_LIST_ROWS_PER_PAGE_OPTIONS]}
              onPageChange={(_, nextPage) => setPage(nextPage)}
              onRowsPerPageChange={(event) => {
                setRowsPerPage(Number(event.target.value))
                setPage(0)
              }}
            />
          </TableContainer>
        ) : null}
      </Stack>

      <GrievanceSubmitDialog
        open={isSubmitDialogOpen}
        onClose={() => setIsSubmitDialogOpen(false)}
        onSubmitted={(referenceId) => {
          setIsSubmitDialogOpen(false)
          setSubmittedReferenceId(referenceId)
        }}
      />
    </Box>
  )
}

export default GrievanceListPage
