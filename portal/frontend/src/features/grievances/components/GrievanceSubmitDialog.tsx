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
  Chip,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  FormControl,
  FormHelperText,
  MenuItem,
  Select,
  Stack,
  TextField,
  Typography,
} from '@wso2/oxygen-ui'
import { Paperclip, UploadCloud, X } from '@wso2/oxygen-ui-icons-react'
import { useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import type { GrievanceCategory } from '../../../types/grievance'
import { GRIEVANCE_CATEGORIES } from '../../../types/grievance'
import { MAX_ATTACHMENT_SIZE_BYTES, MAX_ATTACHMENT_SIZE_LABEL } from '../constants'
import { useSubmitGrievanceMutation } from '../hooks/useGrievanceQueries'

interface GrievanceSubmitDialogProps {
  open: boolean
  onClose: () => void
  onSubmitted: (referenceId: string) => void
}

function GrievanceSubmitDialog({
  open,
  onClose,
  onSubmitted,
}: GrievanceSubmitDialogProps): React.JSX.Element {
  const { t } = useTranslation('common')
  const fileInputRef = useRef<HTMLInputElement>(null)
  const submitGrievanceMutation = useSubmitGrievanceMutation()

  const [category, setCategory] = useState<GrievanceCategory | ''>('')
  const [description, setDescription] = useState<string>('')
  const [attachments, setAttachments] = useState<File[]>([])
  const [attachmentSizeError, setAttachmentSizeError] = useState<string | null>(null)
  const [showValidation, setShowValidation] = useState<boolean>(false)

  const resetForm = (): void => {
    setCategory('')
    setDescription('')
    setAttachments([])
    setAttachmentSizeError(null)
    setShowValidation(false)
  }

  const handleClose = (): void => {
    resetForm()
    onClose()
  }

  const categoryError = showValidation && !category
  const descriptionError = showValidation && !description.trim()

  return (
    <Dialog
      open={open}
      onClose={handleClose}
      maxWidth="sm"
      fullWidth
      scroll="paper"
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
          bgcolor: 'background.default',
        }}
      >
        <Stack spacing={1}>
          <Typography variant="h4" fontWeight={700}>
            {t('grievances.submit.title')}
          </Typography>
          <Typography variant="body2" color="text.secondary">
            {t('grievances.submit.subtitle')}
          </Typography>
        </Stack>
      </DialogTitle>

      <DialogContent sx={{ px: 3, mt: 3, pb: 3 }}>
        <Stack spacing={3}>
          <Box>
            <Typography variant="body2" sx={{ mb: 0.5 }}>
              {t('grievances.submit.fields.category')}
              <Box component="span" sx={{ color: 'error.main' }}>
                {' '}
                *
              </Box>
            </Typography>
            <FormControl fullWidth error={categoryError}>
              <Select
                id="grievance-category"
                value={category}
                displayEmpty
                onChange={(event) => {
                  setCategory(event.target.value as GrievanceCategory)
                }}
                renderValue={(selected) =>
                  selected ? (
                    t(`grievances.categories.${selected}`)
                  ) : (
                    <Typography component="span" color="text.secondary">
                      {t('grievances.submit.fields.categoryPlaceholder')}
                    </Typography>
                  )
                }
              >
                {GRIEVANCE_CATEGORIES.map((categoryOption) => (
                  <MenuItem key={categoryOption} value={categoryOption}>
                    {t(`grievances.categories.${categoryOption}`)}
                  </MenuItem>
                ))}
              </Select>
              <FormHelperText>
                {categoryError
                  ? t('grievances.submit.validation.categoryRequired')
                  : t('grievances.submit.fields.categoryHelp')}
              </FormHelperText>
            </FormControl>
          </Box>

          <Box>
            <Typography variant="body2" sx={{ mb: 0.5 }}>
              {t('grievances.submit.fields.description')}
              <Box component="span" sx={{ color: 'error.main' }}>
                {' '}
                *
              </Box>
            </Typography>
            <TextField
              placeholder={t('grievances.submit.fields.descriptionPlaceholder')}
              multiline
              minRows={4}
              fullWidth
              value={description}
              error={descriptionError}
              helperText={
                descriptionError
                  ? t('grievances.submit.validation.descriptionRequired')
                  : t('grievances.submit.fields.descriptionHelp')
              }
              onChange={(event) => {
                setDescription(event.target.value)
              }}
            />
          </Box>

          <Box>
            <Typography variant="body2" sx={{ mb: 0.5 }}>
              {t('grievances.submit.fields.attachments')}
            </Typography>
            <Typography variant="caption" color="text.secondary" sx={{ display: 'block', mb: 1 }}>
              {t('grievances.submit.fields.attachmentsHelp', {
                maxSize: MAX_ATTACHMENT_SIZE_LABEL,
              })}
            </Typography>

            <input
              ref={fileInputRef}
              type="file"
              multiple
              accept=".pdf,.docx,.png,.jpg,.jpeg"
              hidden
              onChange={(event) => {
                const input = event.target
                const files = Array.from(input.files ?? [])
                const acceptedFiles = files.filter((file) => file.size <= MAX_ATTACHMENT_SIZE_BYTES)
                const oversizedFiles = files.filter((file) => file.size > MAX_ATTACHMENT_SIZE_BYTES)

                setAttachmentSizeError(
                  oversizedFiles.length > 0
                    ? t('grievances.submit.validation.attachmentTooLarge', {
                        fileNames: oversizedFiles.map((file) => file.name).join(', '),
                        maxSize: MAX_ATTACHMENT_SIZE_LABEL,
                      })
                    : null,
                )
                setAttachments((previousFiles) => [...previousFiles, ...acceptedFiles])
                input.value = ''
              }}
            />
            <Button
              variant="outlined"
              size="small"
              startIcon={<UploadCloud size={16} />}
              onClick={() => fileInputRef.current?.click()}
            >
              {t('grievances.submit.fields.uploadButton')}
            </Button>

            {attachmentSizeError ? (
              <Typography variant="caption" color="error.main" sx={{ display: 'block', mt: 1 }}>
                {attachmentSizeError}
              </Typography>
            ) : null}

            {attachments.length > 0 ? (
              <Stack direction="row" spacing={1} flexWrap="wrap" useFlexGap sx={{ mt: 1.5 }}>
                {attachments.map((file, index) => (
                  <Chip
                    key={`${file.name}-${String(index)}`}
                    size="small"
                    variant="outlined"
                    icon={<Paperclip size={14} />}
                    label={file.name}
                    onDelete={() => {
                      setAttachments((previousFiles) =>
                        previousFiles.filter((_, fileIndex) => fileIndex !== index),
                      )
                    }}
                    deleteIcon={<X size={14} />}
                  />
                ))}
              </Stack>
            ) : null}
          </Box>
        </Stack>
      </DialogContent>

      <DialogActions
        sx={{
          px: 3,
          py: 2.5,
          borderTop: 1,
          borderColor: 'divider',
          bgcolor: 'background.default',
          flexDirection: { xs: 'column-reverse', sm: 'row' },
          gap: 1.25,
        }}
      >
        <Button fullWidth variant="outlined" onClick={handleClose}>
          {t('grievances.submit.actions.cancel')}
        </Button>
        <Button
          autoFocus
          fullWidth
          variant="contained"
          disabled={submitGrievanceMutation.isPending}
          onClick={() => {
            if (!category || !description.trim()) {
              setShowValidation(true)
              return
            }

            submitGrievanceMutation.mutate(
              {
                category,
                description: description.trim(),
                attachments,
              },
              {
                onSuccess: (newGrievance) => {
                  resetForm()
                  onSubmitted(newGrievance.referenceId)
                },
              },
            )
          }}
        >
          {t('grievances.submit.actions.submit')}
        </Button>
      </DialogActions>
    </Dialog>
  )
}

export default GrievanceSubmitDialog
