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
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TablePagination,
  TableRow,
  TableSortLabel,
} from '@wso2/oxygen-ui'
import { useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { GRIEVANCE_PRIORITIES, type GrievanceRecord } from '../../../types/grievance'
import { formatIsoDateTime } from '../../../utils/dateTime'
import GrievancePriorityChip from '../../grievances/components/GrievancePriorityChip'
import { GRIEVANCE_QUEUE_ROWS_PER_PAGE_OPTIONS } from '../../grievances/constants'
import GrievanceSlaIndicator from '../../grievances/components/GrievanceSlaIndicator'
import GrievanceStatusChip from '../../grievances/components/GrievanceStatusChip'

interface GrievanceQueueTableProps {
  rows: GrievanceRecord[]
  total: number
  page: number
  rowsPerPage: number
  onPageChange: (page: number) => void
  onRowsPerPageChange: (rowsPerPage: number) => void
  onViewCase: (id: string) => void
}

type SortableField = 'priority' | 'sla'
type SortDirection = 'asc' | 'desc'

const DATE_FORMAT_OPTIONS: Intl.DateTimeFormatOptions = {
  month: 'short',
  day: '2-digit',
  year: 'numeric',
}

function sortRows(
  rows: GrievanceRecord[],
  sortField: SortableField | null,
  sortDirection: SortDirection,
): GrievanceRecord[] {
  if (!sortField) {
    return rows
  }

  const sortedRows = [...rows].sort((left, right) => {
    if (sortField === 'priority') {
      return (
        GRIEVANCE_PRIORITIES.indexOf(left.priority) - GRIEVANCE_PRIORITIES.indexOf(right.priority)
      )
    }

    return new Date(left.statutoryDueDate).getTime() - new Date(right.statutoryDueDate).getTime()
  })

  return sortDirection === 'asc' ? sortedRows : sortedRows.reverse()
}

function GrievanceQueueTable({
  rows,
  total,
  page,
  rowsPerPage,
  onPageChange,
  onRowsPerPageChange,
  onViewCase,
}: GrievanceQueueTableProps): React.JSX.Element {
  const { t } = useTranslation('common')
  const [sortField, setSortField] = useState<SortableField | null>(null)
  const [sortDirection, setSortDirection] = useState<SortDirection>('asc')

  const sortedRows = useMemo(
    () => sortRows(rows, sortField, sortDirection),
    [rows, sortField, sortDirection],
  )

  const handleSortClick = (field: SortableField): void => {
    if (sortField !== field) {
      setSortField(field)
      setSortDirection('asc')
      return
    }

    setSortDirection((previousDirection) => (previousDirection === 'asc' ? 'desc' : 'asc'))
  }

  return (
    <TableContainer sx={{ border: 1, borderColor: 'divider', borderRadius: 1 }}>
      <Table aria-label={t('grievances.management.queue.table.tableAriaLabel')}>
        <TableHead>
          <TableRow>
            <TableCell>{t('grievances.management.queue.table.headers.referenceId')}</TableCell>
            <TableCell>{t('grievances.management.queue.table.headers.dataPrincipal')}</TableCell>
            <TableCell>{t('grievances.management.queue.table.headers.category')}</TableCell>
            <TableCell sortDirection={sortField === 'priority' ? sortDirection : false}>
              <TableSortLabel
                active={sortField === 'priority'}
                direction={sortField === 'priority' ? sortDirection : 'asc'}
                onClick={() => handleSortClick('priority')}
                sx={{ '& .MuiTableSortLabel-icon': { opacity: 1 } }}
              >
                {t('grievances.management.queue.table.headers.priority')}
              </TableSortLabel>
            </TableCell>
            <TableCell>{t('grievances.management.queue.table.headers.status')}</TableCell>
            <TableCell sortDirection={sortField === 'sla' ? sortDirection : false}>
              <TableSortLabel
                active={sortField === 'sla'}
                direction={sortField === 'sla' ? sortDirection : 'asc'}
                onClick={() => handleSortClick('sla')}
                sx={{ '& .MuiTableSortLabel-icon': { opacity: 1 } }}
              >
                {t('grievances.management.queue.table.headers.sla')}
              </TableSortLabel>
            </TableCell>
            <TableCell>{t('grievances.management.queue.table.headers.updated')}</TableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {sortedRows.map((row) => (
            <TableRow
              key={row.id}
              hover
              onClick={() => onViewCase(row.id)}
              sx={{ cursor: 'pointer' }}
            >
              <TableCell sx={{ fontWeight: 600 }}>{row.referenceId}</TableCell>
              <TableCell>{row.dataPrincipalName}</TableCell>
              <TableCell>{t(`grievances.categories.${row.category}`)}</TableCell>
              <TableCell>
                <GrievancePriorityChip priority={row.priority} />
              </TableCell>
              <TableCell>
                <GrievanceStatusChip status={row.status} viewerRole="GrievanceOfficer" />
              </TableCell>
              <TableCell>
                <GrievanceSlaIndicator
                  statutoryDueDate={row.statutoryDueDate}
                  status={row.status}
                />
              </TableCell>
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
        rowsPerPageOptions={[...GRIEVANCE_QUEUE_ROWS_PER_PAGE_OPTIONS]}
        onPageChange={(_, nextPage) => onPageChange(nextPage)}
        onRowsPerPageChange={(event) => onRowsPerPageChange(Number(event.target.value))}
      />
    </TableContainer>
  )
}

export default GrievanceQueueTable
