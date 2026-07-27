/*
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
 * Licensed under the Apache License, Version 2.0.
 */

import {
  Box,
  Card,
  CardContent,
  CardHeader,
  Chip,
  Divider,
  Stack,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Typography,
} from '@wso2/oxygen-ui'
import { ArrowRight } from '@wso2/oxygen-ui-icons-react'
import { useMemo } from 'react'
import { useTranslation } from 'react-i18next'
import type { ConsentStatusAudit } from '../../../../types/consent'
import { formatEpochTimestamp } from '../../../../utils/dateTime'
import { getConsentStatusChipColor, getConsentStatusLabelKey } from '../../utils/statusChip'

interface ConsentStatusHistorySectionProps {
  statusHistory: ConsentStatusAudit[]
}

function StatusChip({ status }: { status: string }): React.JSX.Element {
  const { t } = useTranslation('common')
  return (
    <Chip
      size="small"
      variant="outlined"
      color={getConsentStatusChipColor(status)}
      label={t(`consentRegistry.status.${getConsentStatusLabelKey(status)}`)}
    />
  )
}

function ConsentStatusHistorySection({
  statusHistory,
}: ConsentStatusHistorySectionProps): React.JSX.Element {
  const { t } = useTranslation('common')

  const sortedHistory = useMemo(
    () => [...statusHistory].sort((left, right) => right.actionTime - left.actionTime),
    [statusHistory],
  )

  return (
    <Card sx={{ boxShadow: 1 }}>
      <CardHeader
        title={
          <Typography variant="subtitle1" fontWeight={600}>
            {t('consentRegistry.details.statusHistory.title', 'Status History')}
          </Typography>
        }
        sx={{ pb: sortedHistory.length === 0 ? 2 : 0 }}
      />
      <Divider />
      {sortedHistory.length === 0 ? (
        <CardContent>
          <Typography variant="body2" color="text.secondary">
            {t('consentRegistry.details.statusHistory.empty', 'No status changes recorded yet.')}
          </Typography>
        </CardContent>
      ) : (
        <TableContainer>
          <Table sx={{ '& tbody tr:hover': { bgcolor: 'action.hover' } }}>
            <TableHead>
              <TableRow sx={{ bgcolor: 'action.default' }}>
                <TableCell sx={{ fontWeight: 700 }}>
                  {t('consentRegistry.details.statusHistory.change', 'Change')}
                </TableCell>
                <TableCell sx={{ fontWeight: 700 }}>
                  {t('consentRegistry.details.statusHistory.when', 'When')}
                </TableCell>
                <TableCell sx={{ fontWeight: 700 }}>
                  {t('consentRegistry.details.statusHistory.by', 'By')}
                </TableCell>
                <TableCell sx={{ fontWeight: 700 }}>
                  {t('consentRegistry.details.statusHistory.reason', 'Reason')}
                </TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {sortedHistory.map((entry) => (
                <TableRow key={entry.statusAuditId}>
                  <TableCell>
                    <Stack direction="row" spacing={0.75} alignItems="center" flexWrap="wrap">
                      {entry.previousStatus ? (
                        <>
                          <StatusChip status={entry.previousStatus} />
                          <Box sx={{ display: 'flex', color: 'text.disabled' }}>
                            <ArrowRight size={14} />
                          </Box>
                        </>
                      ) : null}
                      <StatusChip status={entry.currentStatus} />
                    </Stack>
                  </TableCell>
                  <TableCell>
                    <Typography variant="body2">
                      {formatEpochTimestamp(entry.actionTime)}
                    </Typography>
                  </TableCell>
                  <TableCell>
                    <Typography variant="body2">
                      {entry.actionBy ??
                        t('consentRegistry.details.statusHistory.systemActor', 'System')}
                    </Typography>
                  </TableCell>
                  <TableCell>
                    <Typography
                      variant="body2"
                      color={entry.reason ? 'text.primary' : 'text.secondary'}
                    >
                      {entry.reason ?? '-'}
                    </Typography>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </TableContainer>
      )}
    </Card>
  )
}

export default ConsentStatusHistorySection
