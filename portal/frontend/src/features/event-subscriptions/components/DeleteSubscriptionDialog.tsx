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
  Stack,
  Typography,
} from '@wso2/oxygen-ui'
import { useTranslation } from 'react-i18next'

interface DeleteSubscriptionDialogProps {
  open: boolean
  subscriptionId: string
  loading: boolean
  error: string | undefined
  onClose: () => void
  onConfirm: () => void
}

function DeleteSubscriptionDialog({
  open,
  subscriptionId,
  loading,
  error,
  onClose,
  onConfirm,
}: DeleteSubscriptionDialogProps): React.JSX.Element {
  const { t } = useTranslation('common')

  return (
    <Dialog open={open} onClose={loading ? undefined : onClose} maxWidth="xs" fullWidth>
      <DialogTitle>{t('eventSubscriptions.delete.title')}</DialogTitle>
      <DialogContent dividers>
        <Stack spacing={2}>
          <Typography>{t('eventSubscriptions.delete.message', { subscriptionId })}</Typography>
          {error ? (
            <Typography variant="body2" color="error.main">
              {error}
            </Typography>
          ) : null}
        </Stack>
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose} disabled={loading}>
          {t('eventSubscriptions.actions.cancel')}
        </Button>
        <Button color="error" variant="contained" onClick={onConfirm} disabled={loading}>
          {loading
            ? t('eventSubscriptions.actions.deleting')
            : t('eventSubscriptions.actions.delete')}
        </Button>
      </DialogActions>
    </Dialog>
  )
}

export default DeleteSubscriptionDialog
