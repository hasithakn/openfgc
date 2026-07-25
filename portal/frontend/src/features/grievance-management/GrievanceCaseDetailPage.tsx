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
  FormControl,
  InputLabel,
  MenuItem,
  Select,
  Stack,
  Typography,
} from '@wso2/oxygen-ui'
import { Paperclip } from '@wso2/oxygen-ui-icons-react'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useNavigate, useParams } from 'react-router-dom'
import HeaderBreadcrumbs from '../../components/layout/main-layout/HeaderBreadcrumbs'
import type { GrievanceStatus } from '../../types/grievance'
import { formatIsoDateTime } from '../../utils/dateTime'
import GrievanceActivityFeed from '../grievances/components/GrievanceActivityFeed'
import GrievancePriorityChip from '../grievances/components/GrievancePriorityChip'
import GrievanceReplyComposer from '../grievances/components/GrievanceReplyComposer'
import GrievanceSlaIndicator from '../grievances/components/GrievanceSlaIndicator'
import GrievanceStatusChip from '../grievances/components/GrievanceStatusChip'
import { GRIEVANCE_NEXT_STATUSES } from '../grievances/constants'
import { getGrievanceStatusLabelKey } from '../grievances/utils/grievanceDisplay'
import {
  useCaseDetailQuery,
  useReplyToCaseMutation,
  useTransitionStatusMutation,
} from './hooks/useGrievanceQueueQueries'

const DATE_FORMAT_OPTIONS: Intl.DateTimeFormatOptions = {
  month: 'short',
  day: '2-digit',
  year: 'numeric',
  hour: '2-digit',
  minute: '2-digit',
  hour12: false,
}

function GrievanceCaseDetailPage(): React.JSX.Element {
  const { t } = useTranslation('common')
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const detailQuery = useCaseDetailQuery(id)
  const replyMutation = useReplyToCaseMutation()
  const transitionStatusMutation = useTransitionStatusMutation()
  const grievance = detailQuery.data
  const [nextStatus, setNextStatus] = useState<GrievanceStatus | ''>('')

  if (detailQuery.isLoading) {
    return (
      <Box component="main" sx={{ p: { xs: 2, md: 4 } }}>
        <HeaderBreadcrumbs />
      </Box>
    )
  }

  if (!grievance) {
    return (
      <Box
        component="main"
        sx={{ p: { xs: 2, md: 4 }, display: 'flex', flexDirection: 'column', gap: 2 }}
      >
        <Typography variant="h5">{t('grievances.management.case.notFound')}</Typography>
        <Box>
          <Button variant="outlined" onClick={() => navigate('/grievance-management')}>
            {t('grievances.management.case.back')}
          </Button>
        </Box>
      </Box>
    )
  }

  const allowedNextStatuses = GRIEVANCE_NEXT_STATUSES[grievance.status]

  return (
    <Box
      component="main"
      sx={{ p: { xs: 2, md: 4 }, display: 'flex', flexDirection: 'column', gap: 3 }}
    >
      <Stack spacing={1}>
        <HeaderBreadcrumbs />
        <Typography variant="h4" fontWeight={700}>
          {grievance.referenceId}
        </Typography>
        <Stack direction="row" spacing={1.5} alignItems="center" flexWrap="wrap" useFlexGap>
          <GrievancePriorityChip priority={grievance.priority} />
          <GrievanceStatusChip status={grievance.status} viewerRole="GrievanceOfficer" />
          <GrievanceSlaIndicator
            statutoryDueDate={grievance.statutoryDueDate}
            status={grievance.status}
          />
        </Stack>
      </Stack>

      <Card sx={{ boxShadow: 1 }}>
        <CardHeader title={t(`grievances.categories.${grievance.category}`)} sx={{ pb: 1 }} />
        <Divider />
        <CardContent>
          <Stack spacing={2}>
            <Box
              sx={{
                display: 'grid',
                gridTemplateColumns: { xs: '1fr', sm: '1fr 1fr', lg: 'repeat(3, 1fr)' },
                gap: 2,
              }}
            >
              <Box>
                <Typography
                  variant="caption"
                  color="text.secondary"
                  fontWeight={600}
                  sx={{ display: 'block', textTransform: 'uppercase' }}
                >
                  {t('grievances.management.case.dataPrincipal')}
                </Typography>
                <Typography variant="body2">{grievance.dataPrincipalName}</Typography>
              </Box>
              <Box>
                <Typography
                  variant="caption"
                  color="text.secondary"
                  fontWeight={600}
                  sx={{ display: 'block', textTransform: 'uppercase' }}
                >
                  {t('grievances.detail.submittedOn', {
                    date: formatIsoDateTime(grievance.submittedAt, DATE_FORMAT_OPTIONS),
                  })}
                </Typography>
              </Box>
              <Box>
                <Typography
                  variant="caption"
                  color="text.secondary"
                  fontWeight={600}
                  sx={{ display: 'block', textTransform: 'uppercase' }}
                >
                  {t('grievances.sla.dueDate', {
                    date: formatIsoDateTime(grievance.statutoryDueDate, DATE_FORMAT_OPTIONS),
                  })}
                </Typography>
              </Box>
            </Box>

            <Box>
              <Typography
                variant="caption"
                color="text.secondary"
                fontWeight={600}
                sx={{ display: 'block', textTransform: 'uppercase' }}
              >
                {t('grievances.detail.description')}
              </Typography>
              <Typography variant="body2" sx={{ mt: 0.5 }}>
                {grievance.description}
              </Typography>
            </Box>

            {grievance.attachments.length > 0 ? (
              <Stack direction="row" spacing={1} flexWrap="wrap" useFlexGap>
                {grievance.attachments.map((attachment) => (
                  <Chip
                    key={attachment.id}
                    size="small"
                    variant="outlined"
                    icon={<Paperclip size={14} />}
                    label={attachment.fileName}
                  />
                ))}
              </Stack>
            ) : null}
          </Stack>
        </CardContent>
      </Card>

      {allowedNextStatuses.length > 0 ? (
        <Card sx={{ boxShadow: 1 }}>
          <CardHeader title={t('grievances.management.case.statusUpdate.title')} sx={{ pb: 1 }} />
          <Divider />
          <CardContent>
            <Stack spacing={2}>
              <FormControl fullWidth>
                <InputLabel id="grievance-case-next-status-label">
                  {t('grievances.management.case.statusUpdate.newStatus')}
                </InputLabel>
                <Select
                  labelId="grievance-case-next-status-label"
                  id="grievance-case-next-status"
                  value={nextStatus}
                  label={t('grievances.management.case.statusUpdate.newStatus')}
                  onChange={(event) => {
                    setNextStatus(event.target.value as GrievanceStatus)
                  }}
                >
                  {allowedNextStatuses.map((status) => (
                    <MenuItem key={status} value={status}>
                      {t(`grievances.status.${getGrievanceStatusLabelKey(status)}`)}
                    </MenuItem>
                  ))}
                </Select>
              </FormControl>

              <Box>
                <Button
                  variant="contained"
                  disabled={!nextStatus || transitionStatusMutation.isPending}
                  onClick={() => {
                    if (!nextStatus) {
                      return
                    }

                    transitionStatusMutation.mutate(
                      { grievanceId: grievance.id, nextStatus },
                      { onSuccess: () => setNextStatus('') },
                    )
                  }}
                >
                  {t('grievances.management.case.statusUpdate.apply')}
                </Button>
              </Box>
            </Stack>
          </CardContent>
        </Card>
      ) : null}

      <Card sx={{ boxShadow: 1 }}>
        <CardContent>
          <GrievanceReplyComposer
            canPostInternalNote
            onSend={(message, attachments, visibility) => {
              replyMutation.mutate({ grievanceId: grievance.id, message, visibility, attachments })
            }}
          />
        </CardContent>
      </Card>

      <Card sx={{ boxShadow: 1 }}>
        <CardHeader title={t('grievances.activity.title')} sx={{ pb: 1 }} />
        <Divider />
        <CardContent>
          <GrievanceActivityFeed entries={grievance.timeline} viewerRole="GrievanceOfficer" />
        </CardContent>
      </Card>

      <Box>
        <Button variant="outlined" onClick={() => navigate('/grievance-management')}>
          {t('grievances.management.case.back')}
        </Button>
      </Box>
    </Box>
  )
}

export default GrievanceCaseDetailPage
