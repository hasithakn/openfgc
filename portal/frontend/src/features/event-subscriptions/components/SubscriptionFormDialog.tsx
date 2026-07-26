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
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  MenuItem,
  Stack,
  TextField,
  Typography,
} from '@wso2/oxygen-ui'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import type {
  DeliveryMode,
  FilterType,
  SubscriptionCreateRequest,
} from '../../../types/eventSubscription'
import { useTopicsQuery } from '../hooks/useEventSubscriptionQueries'

interface SubscriptionFormDialogProps {
  open: boolean
  loading: boolean
  error: string | undefined
  onClose: () => void
  onCreate: (payload: SubscriptionCreateRequest) => void
}

function SubscriptionFormDialog({
  open,
  loading,
  error,
  onClose,
  onCreate,
}: SubscriptionFormDialogProps): React.JSX.Element {
  const { t } = useTranslation('common')
  const topicsQuery = useTopicsQuery()
  const topics = topicsQuery.data?.topics ?? []
  const [topic, setTopic] = useState('')
  const [filterType, setFilterType] = useState<FilterType>('all')
  const [purposes, setPurposes] = useState('')
  const [deliveryMode, setDeliveryMode] = useState<DeliveryMode>('webhook')
  const [callbackUrl, setCallbackUrl] = useState('')
  const [sharedSecret, setSharedSecret] = useState('')
  const [validationError, setValidationError] = useState('')

  const handleSubmit = (): void => {
    if (!topic.trim()) {
      setValidationError(t('eventSubscriptions.validation.topicRequired'))
      return
    }
    const purposeList = purposes
      .split(',')
      .map((purpose) => purpose.trim())
      .filter(Boolean)
    if (filterType !== 'all' && purposeList.length === 0) {
      setValidationError(t('eventSubscriptions.validation.purposesRequired'))
      return
    }
    if (deliveryMode === 'webhook' && !callbackUrl.trim()) {
      setValidationError(t('eventSubscriptions.validation.callbackUrlRequired'))
      return
    }
    if (!sharedSecret.trim()) {
      setValidationError(t('eventSubscriptions.validation.sharedSecretRequired'))
      return
    }

    onCreate({
      topic: topic.trim(),
      filter: {
        type: filterType,
        purposes: filterType === 'all' ? undefined : purposeList,
      },
      delivery: {
        mode: deliveryMode,
        callbackUrl: deliveryMode === 'webhook' ? callbackUrl.trim() : undefined,
        sharedSecret: sharedSecret.trim(),
      },
    })
  }

  return (
    <Dialog open={open} onClose={loading ? undefined : onClose} maxWidth="sm" fullWidth>
      <DialogTitle>{t('eventSubscriptions.form.createTitle')}</DialogTitle>
      <DialogContent dividers>
        <Stack spacing={2.5} sx={{ pt: 0.5 }}>
          <TextField
            select
            required
            fullWidth
            label={t('eventSubscriptions.fields.topic')}
            helperText={
              topicsQuery.isError
                ? t('eventSubscriptions.form.topicsLoadFailed')
                : t('eventSubscriptions.form.topicHelp')
            }
            error={topicsQuery.isError}
            disabled={topicsQuery.isLoading}
            value={topic}
            onChange={(event) => setTopic(event.target.value)}
          >
            {topics.map((topicOption) => (
              <MenuItem key={topicOption.topicId} value={topicOption.name}>
                {topicOption.name}
              </MenuItem>
            ))}
          </TextField>
          <TextField
            select
            fullWidth
            label={t('eventSubscriptions.fields.purposeFilterMode')}
            value={filterType}
            onChange={(event) => setFilterType(event.target.value as FilterType)}
          >
            <MenuItem value="all">{t('eventSubscriptions.filterType.all')}</MenuItem>
            <MenuItem value="specific">{t('eventSubscriptions.filterType.specific')}</MenuItem>
            <MenuItem value="all_except">{t('eventSubscriptions.filterType.all_except')}</MenuItem>
          </TextField>
          {filterType !== 'all' ? (
            <TextField
              required
              fullWidth
              label={t('eventSubscriptions.fields.purposes')}
              helperText={t('eventSubscriptions.form.purposesHelp')}
              value={purposes}
              onChange={(event) => setPurposes(event.target.value)}
            />
          ) : null}
          <TextField
            select
            fullWidth
            label={t('eventSubscriptions.fields.deliveryMode')}
            value={deliveryMode}
            onChange={(event) => setDeliveryMode(event.target.value as DeliveryMode)}
          >
            <MenuItem value="webhook">webhook</MenuItem>
            <MenuItem value="poll">poll</MenuItem>
            <MenuItem value="pull">pull</MenuItem>
          </TextField>
          {deliveryMode === 'webhook' ? (
            <TextField
              required
              fullWidth
              label={t('eventSubscriptions.fields.callbackUrl')}
              value={callbackUrl}
              onChange={(event) => setCallbackUrl(event.target.value)}
            />
          ) : null}
          <TextField
            required
            fullWidth
            label={t('eventSubscriptions.fields.sharedSecret')}
            value={sharedSecret}
            onChange={(event) => setSharedSecret(event.target.value)}
          />
          {validationError || error ? (
            <Typography color="error.main" variant="body2">
              {validationError || error}
            </Typography>
          ) : null}
        </Stack>
      </DialogContent>
      <DialogActions sx={{ p: 2 }}>
        <Button onClick={onClose} disabled={loading}>
          {t('eventSubscriptions.actions.cancel')}
        </Button>
        <Button variant="contained" onClick={handleSubmit} disabled={loading}>
          {loading
            ? t('eventSubscriptions.actions.saving')
            : t('eventSubscriptions.actions.create')}
        </Button>
      </DialogActions>
    </Dialog>
  )
}

export default SubscriptionFormDialog
