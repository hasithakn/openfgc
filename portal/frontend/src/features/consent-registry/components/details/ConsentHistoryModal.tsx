/*
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
 * Licensed under the Apache License, Version 2.0.
 */

import {
  Alert,
  Box,
  Button,
  Chip,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  Divider,
  Skeleton,
  Stack,
  Typography,
} from '@wso2/oxygen-ui'
import { ChevronDown, ChevronUp } from '@wso2/oxygen-ui-icons-react'
import { useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import type { ConsentHistoryEntry } from '../../../../types/consent'
import { formatEpochTimestamp } from '../../../../utils/dateTime'
import { useConsentHistoryQuery } from '../../hooks/useConsentQueries'
import { getConsentStatusChipColor, getConsentStatusLabelKey } from '../../utils/statusChip'

interface ConsentHistoryModalProps {
  open: boolean
  consentId: string
  onClose: () => void
}

function formatSnapshotJSON(entry: ConsentHistoryEntry): string {
  if (!entry.snapshot) {
    return '-'
  }
  try {
    return JSON.stringify(entry.snapshot, null, 2)
  } catch {
    return '-'
  }
}

function HistoryEntryRow({ entry }: { entry: ConsentHistoryEntry }): React.JSX.Element {
  const { t } = useTranslation('common')
  const [expanded, setExpanded] = useState(false)
  const { snapshot } = entry
  const purposeCount = snapshot?.purposes?.length ?? 0
  const authorizationCount = snapshot?.authorizations?.length ?? 0

  return (
    <Box sx={{ py: 1.5 }}>
      <Stack
        direction={{ xs: 'column', sm: 'row' }}
        justifyContent="space-between"
        alignItems={{ sm: 'center' }}
        spacing={1}
      >
        <Stack spacing={0.5}>
          <Typography variant="body2" fontWeight={600}>
            {entry.reason ??
              t('consentRegistry.details.historyModal.unknownReason', 'Consent updated')}
          </Typography>
          <Typography variant="caption" color="text.secondary">
            {formatEpochTimestamp(entry.actionTime)}
            {' · '}
            {entry.actionBy ?? t('consentRegistry.details.statusHistory.systemActor', 'System')}
          </Typography>
        </Stack>
        {snapshot ? (
          <Stack direction="row" spacing={0.75} alignItems="center" flexWrap="wrap">
            <Chip
              size="small"
              variant="outlined"
              color={getConsentStatusChipColor(snapshot.status)}
              label={t(`consentRegistry.status.${getConsentStatusLabelKey(snapshot.status)}`)}
            />
            <Chip
              size="small"
              variant="outlined"
              label={t('consentRegistry.details.historyModal.purposeCount', '{{count}} purposes', {
                count: purposeCount,
              })}
            />
            <Chip
              size="small"
              variant="outlined"
              label={t(
                'consentRegistry.details.historyModal.authorizationCount',
                '{{count}} authorizations',
                { count: authorizationCount },
              )}
            />
            <Button
              size="small"
              variant="text"
              endIcon={expanded ? <ChevronUp size={14} /> : <ChevronDown size={14} />}
              onClick={() => setExpanded((current) => !current)}
            >
              {expanded
                ? t('consentRegistry.details.historyModal.hideRaw', 'Hide raw')
                : t('consentRegistry.details.historyModal.viewRaw', 'View raw')}
            </Button>
          </Stack>
        ) : null}
      </Stack>
      {expanded ? (
        <Box
          component="pre"
          sx={{
            mt: 1.5,
            mb: 0,
            p: 2,
            border: 1,
            borderColor: 'divider',
            borderRadius: 1,
            bgcolor: 'background.default',
            color: 'text.primary',
            fontSize: '0.8125rem',
            fontFamily: 'monospace',
            whiteSpace: 'pre',
            overflow: 'auto',
            maxHeight: '40vh',
          }}
        >
          {formatSnapshotJSON(entry)}
        </Box>
      ) : null}
    </Box>
  )
}

function ConsentHistoryModal({
  open,
  consentId,
  onClose,
}: ConsentHistoryModalProps): React.JSX.Element {
  const { t } = useTranslation('common')
  const historyQuery = useConsentHistoryQuery(consentId, open)

  const sortedHistory = useMemo(
    () => [...(historyQuery.data ?? [])].sort((left, right) => right.actionTime - left.actionTime),
    [historyQuery.data],
  )

  return (
    <Dialog open={open} onClose={onClose} maxWidth="md" fullWidth>
      <DialogTitle sx={{ borderBottom: 1, borderColor: 'divider' }}>
        {t('consentRegistry.details.historyModal.title', 'Full Consent History')}
      </DialogTitle>
      <DialogContent sx={{ pt: 2.5, pb: 2 }}>
        {historyQuery.isLoading ? (
          <Stack spacing={2}>
            <Skeleton variant="rounded" height={56} />
            <Skeleton variant="rounded" height={56} />
            <Skeleton variant="rounded" height={56} />
          </Stack>
        ) : null}
        {historyQuery.isError ? (
          <Alert severity="error">
            {t(
              'consentRegistry.details.historyModal.loadFailed',
              'Unable to load history right now.',
            )}
          </Alert>
        ) : null}
        {!historyQuery.isLoading && !historyQuery.isError && sortedHistory.length === 0 ? (
          <Typography variant="body2" color="text.secondary">
            {t('consentRegistry.details.historyModal.empty', 'No history recorded yet.')}
          </Typography>
        ) : null}
        {!historyQuery.isLoading && !historyQuery.isError && sortedHistory.length > 0 ? (
          <Stack divider={<Divider />}>
            {sortedHistory.map((entry) => (
              <HistoryEntryRow key={entry.historyId} entry={entry} />
            ))}
          </Stack>
        ) : null}
      </DialogContent>
      <DialogActions sx={{ px: 3, pb: 3 }}>
        <Button variant="outlined" onClick={onClose}>
          {t('consentRegistry.details.historyModal.close', 'Close')}
        </Button>
      </DialogActions>
    </Dialog>
  )
}

export default ConsentHistoryModal
