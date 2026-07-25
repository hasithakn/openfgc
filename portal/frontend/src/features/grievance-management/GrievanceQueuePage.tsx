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

import { Box, Stack, StatCard, Typography } from '@wso2/oxygen-ui'
import { AlertTriangle, CheckCircle2, Clock3, Inbox } from '@wso2/oxygen-ui-icons-react'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useNavigate } from 'react-router-dom'
import HeaderBreadcrumbs from '../../components/layout/main-layout/HeaderBreadcrumbs'
import { GRIEVANCE_QUEUE_ROWS_PER_PAGE_OPTIONS } from '../grievances/constants'
import GrievanceQueueFilters from './components/GrievanceQueueFilters'
import GrievanceQueueTable from './components/GrievanceQueueTable'
import {
  useGrievanceQueueQuery,
  useGrievanceQueueStatsQuery,
} from './hooks/useGrievanceQueueQueries'
import type { GrievanceQueueFiltersState } from './types'

const DEFAULT_FILTERS: GrievanceQueueFiltersState = {
  status: 'All',
  priority: 'All',
  search: '',
}

function GrievanceQueuePage(): React.JSX.Element {
  const { t } = useTranslation('common')
  const navigate = useNavigate()
  const [filters, setFilters] = useState<GrievanceQueueFiltersState>(DEFAULT_FILTERS)
  const [page, setPage] = useState<number>(0)
  const [rowsPerPage, setRowsPerPage] = useState<number>(GRIEVANCE_QUEUE_ROWS_PER_PAGE_OPTIONS[0])

  const statsQuery = useGrievanceQueueStatsQuery()
  const queueQuery = useGrievanceQueueQuery(filters, page, rowsPerPage)
  const rows = queueQuery.data?.rows ?? []
  const total = queueQuery.data?.total ?? 0
  const stats = {
    openCount: statsQuery.data?.open ?? 0,
    awaitingInfoCount: statsQuery.data?.awaitingInfo ?? 0,
    resolvedCount: statsQuery.data?.resolved ?? 0,
    slaBreachedCount: statsQuery.data?.slaBreached ?? 0,
  }

  return (
    <Box component="main" sx={{ p: { xs: 2, md: 4 } }}>
      <Stack spacing={3}>
        <Stack spacing={1}>
          <HeaderBreadcrumbs />
          <Typography variant="h4" fontWeight={700}>
            {t('grievances.management.queue.title')}
          </Typography>
          <Typography variant="body2" color="text.secondary">
            {t('grievances.management.queue.subtitle')}
          </Typography>
        </Stack>

        <Box
          sx={{
            display: 'grid',
            gridTemplateColumns: { xs: '1fr 1fr', sm: 'repeat(4, 1fr)' },
            gap: 2,
          }}
        >
          <StatCard
            value={stats.openCount}
            label={t('grievances.management.queue.stats.open')}
            icon={<Inbox size={22} />}
            iconColor="info"
          />
          <StatCard
            value={stats.awaitingInfoCount}
            label={t('grievances.management.queue.stats.awaitingInfo')}
            icon={<Clock3 size={22} />}
            iconColor="warning"
          />
          <StatCard
            value={stats.resolvedCount}
            label={t('grievances.management.queue.stats.resolved')}
            icon={<CheckCircle2 size={22} />}
            iconColor="success"
          />
          <StatCard
            value={stats.slaBreachedCount}
            label={t('grievances.management.queue.stats.slaBreached')}
            icon={<AlertTriangle size={22} />}
            iconColor="error"
          />
        </Box>

        <GrievanceQueueFilters
          filters={filters}
          onFilterChange={(nextFilters) => {
            setFilters(nextFilters)
            setPage(0)
          }}
          onClear={() => {
            setFilters(DEFAULT_FILTERS)
            setPage(0)
          }}
        />

        {queueQuery.isError ? (
          <Typography color="error.main">{t('grievances.management.queue.loadError')}</Typography>
        ) : null}

        {!queueQuery.isError && !queueQuery.isLoading && rows.length === 0 ? (
          <Typography>{t('grievances.management.queue.empty')}</Typography>
        ) : null}

        {!queueQuery.isError && (rows.length > 0 || queueQuery.isLoading) ? (
          <GrievanceQueueTable
            rows={rows}
            total={total}
            page={page}
            rowsPerPage={rowsPerPage}
            onPageChange={setPage}
            onRowsPerPageChange={(nextRowsPerPage) => {
              setRowsPerPage(nextRowsPerPage)
              setPage(0)
            }}
            onViewCase={(id) => navigate(`/grievance-management/${encodeURIComponent(id)}`)}
          />
        ) : null}
      </Stack>
    </Box>
  )
}

export default GrievanceQueuePage
