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
  Card,
  CardContent,
  CardHeader,
  Chip,
  Divider,
  Stack,
  Typography,
} from '@wso2/oxygen-ui'
import { Paperclip } from '@wso2/oxygen-ui-icons-react'
import { useTranslation } from 'react-i18next'
import { useNavigate, useParams } from 'react-router-dom'
import HeaderBreadcrumbs from '../../components/layout/main-layout/HeaderBreadcrumbs'
import { formatIsoDateTime } from '../../utils/dateTime'
import GrievanceActivityFeed from './components/GrievanceActivityFeed'
import GrievanceReplyComposer from './components/GrievanceReplyComposer'
import GrievanceSlaIndicator from './components/GrievanceSlaIndicator'
import GrievanceStatusChip from './components/GrievanceStatusChip'
import { useMyGrievanceDetailQuery, useReplyToGrievanceMutation } from './hooks/useGrievanceQueries'

const DATE_FORMAT_OPTIONS: Intl.DateTimeFormatOptions = {
  month: 'short',
  day: '2-digit',
  year: 'numeric',
  hour: '2-digit',
  minute: '2-digit',
  hour12: false,
}

function GrievanceDetailPage(): React.JSX.Element {
  const { t } = useTranslation('common')
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const detailQuery = useMyGrievanceDetailQuery(id)
  const replyMutation = useReplyToGrievanceMutation()
  const grievance = detailQuery.data

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
        <Typography variant="h5">{t('grievances.detail.notFound')}</Typography>
        <Box>
          <Button variant="outlined" onClick={() => navigate('/grievances')}>
            {t('grievances.detail.back')}
          </Button>
        </Box>
      </Box>
    )
  }

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
          <GrievanceStatusChip status={grievance.status} viewerRole="DataPrincipal" />
          {grievance.status !== 'Resolved' ? (
            <GrievanceSlaIndicator
              statutoryDueDate={grievance.statutoryDueDate}
              status={grievance.status}
            />
          ) : null}
        </Stack>
      </Stack>

      <Card sx={{ boxShadow: 1 }}>
        <CardHeader title={t(`grievances.categories.${grievance.category}`)} sx={{ pb: 1 }} />
        <Divider />
        <CardContent>
          <Stack spacing={2}>
            <Box>
              <Typography variant="caption" color="text.secondary" sx={{ display: 'block' }}>
                {t('grievances.detail.description')}
              </Typography>
              <Typography variant="body2" sx={{ mt: 0.5 }}>
                {grievance.description}
              </Typography>
            </Box>

            <Box>
              <Typography variant="caption" color="text.secondary" sx={{ display: 'block' }}>
                {t('grievances.detail.submittedOn', {
                  date: formatIsoDateTime(grievance.submittedAt, DATE_FORMAT_OPTIONS),
                })}
              </Typography>
            </Box>

            {grievance.attachments.length > 0 ? (
              <Box>
                <Typography
                  variant="caption"
                  color="text.secondary"
                  sx={{ display: 'block', mb: 0.75 }}
                >
                  {t('grievances.detail.attachments')}
                </Typography>
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
              </Box>
            ) : null}
          </Stack>
        </CardContent>
      </Card>

      {grievance.status === 'AwaitingInfo' ? (
        <Alert severity="warning">{t('grievances.detail.awaitingInfo.note')}</Alert>
      ) : null}

      {grievance.status === 'Resolved' ? (
        <Alert severity="success">{t('grievances.detail.resolved.note')}</Alert>
      ) : null}

      <Card sx={{ boxShadow: 1 }}>
        <CardHeader title={t('grievances.activity.messagesTitle')} sx={{ pb: 1 }} />
        <Divider />
        <CardContent>
          <Stack spacing={3}>
            <GrievanceReplyComposer
              canPostInternalNote={false}
              onSend={(message, attachments) => {
                replyMutation.mutate({ grievanceId: grievance.id, message, attachments })
              }}
            />
            <GrievanceActivityFeed entries={grievance.timeline} viewerRole="DataPrincipal" />
          </Stack>
        </CardContent>
      </Card>

      <Box>
        <Button variant="outlined" onClick={() => navigate('/grievances')}>
          {t('grievances.detail.back')}
        </Button>
      </Box>
    </Box>
  )
}

export default GrievanceDetailPage
