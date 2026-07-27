/*
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
 * Licensed under the Apache License, Version 2.0.
 */

import {
  Alert,
  Box,
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  Stack,
  Typography,
} from '@wso2/oxygen-ui'
import { AlertTriangle } from '@wso2/oxygen-ui-icons-react'
import { useTranslation } from 'react-i18next'

interface DeleteAccountDialogProps {
  open: boolean
  loading: boolean
  error: string | undefined
  onClose: () => void
  onConfirm: () => void
}

function DeleteAccountDialog({
  open,
  loading,
  error,
  onClose,
  onConfirm,
}: DeleteAccountDialogProps): React.JSX.Element {
  const { t } = useTranslation('common')

  return (
    <Dialog
      open={open}
      onClose={loading ? undefined : onClose}
      maxWidth="xs"
      fullWidth
      PaperProps={{
        sx: (theme) => ({
          borderRadius: 1,
          ...theme.applyStyles('light', { bgcolor: theme.palette.grey[50] }),
          ...theme.applyStyles('dark', { bgcolor: 'rgba(255, 255, 255, 0.06)' }),
        }),
      }}
    >
      <DialogTitle
        sx={{
          p: 3,
          borderBottom: 1,
          borderColor: 'divider',
          textAlign: 'center',
        }}
      >
        <Stack spacing={0.75}>
          <Typography variant="h6" fontWeight={700}>
            {t('profile.deleteAccount.title', 'Delete your account?')}
          </Typography>
          <Typography variant="body2" color="text.secondary">
            {t('profile.deleteAccount.message', 'This action is permanent and cannot be undone.')}
          </Typography>
        </Stack>
      </DialogTitle>

      <DialogContent sx={{ px: 3, pt: 3.5, pb: 3 }}>
        <Stack spacing={2}>
          <Box
            sx={{
              width: '100%',
              p: 2,
              border: 1,
              borderColor: 'error.light',
              borderRadius: 1,
              bgcolor: 'error.lighter',
              display: 'flex',
              alignItems: 'center',
              gap: 1.5,
            }}
          >
            <Box
              sx={{
                width: 36,
                height: 36,
                borderRadius: '50%',
                display: 'grid',
                placeItems: 'center',
                color: 'error.main',
                bgcolor: 'background.paper',
                flexShrink: 0,
              }}
            >
              <AlertTriangle size={20} />
            </Box>
            <Typography variant="body2" color="text.secondary">
              {t(
                'profile.deleteAccount.note',
                'Your consent records and complaints will be anonymized and your account will be permanently deleted. You will be signed out immediately.',
              )}
            </Typography>
          </Box>
          {error ? <Alert severity="error">{error}</Alert> : null}
        </Stack>
      </DialogContent>

      <DialogActions
        sx={{
          p: 3,
          pt: 2,
          borderTop: 1,
          borderColor: 'divider',
          bgcolor: 'background.default',
          flexDirection: 'column',
          gap: 1.25,
        }}
      >
        <Button fullWidth color="error" variant="contained" disabled={loading} onClick={onConfirm}>
          {loading
            ? t('profile.deleteAccount.processing', 'Deleting...')
            : t('profile.deleteAccount.confirm', 'Delete My Account')}
        </Button>
        <Button fullWidth variant="outlined" disabled={loading} onClick={onClose}>
          {t('profile.deleteAccount.cancel', 'Cancel')}
        </Button>
      </DialogActions>
    </Dialog>
  )
}

export default DeleteAccountDialog
