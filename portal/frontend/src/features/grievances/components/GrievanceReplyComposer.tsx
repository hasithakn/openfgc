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
  Chip,
  Stack,
  TextField,
  ToggleButton,
  ToggleButtonGroup,
  Typography,
} from '@wso2/oxygen-ui'
import { Lock, Paperclip, Send, X } from '@wso2/oxygen-ui-icons-react'
import { useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import type { GrievanceTimelineVisibility } from '../../../types/grievance'
import { MAX_ATTACHMENT_SIZE_BYTES, MAX_ATTACHMENT_SIZE_LABEL } from '../constants'

interface GrievanceReplyComposerProps {
  canPostInternalNote: boolean
  onSend: (message: string, attachments: File[], visibility: GrievanceTimelineVisibility) => void
}

function GrievanceReplyComposer({
  canPostInternalNote,
  onSend,
}: GrievanceReplyComposerProps): React.JSX.Element {
  const { t } = useTranslation('common')
  const fileInputRef = useRef<HTMLInputElement>(null)
  const [draft, setDraft] = useState<string>('')
  const [draftAttachments, setDraftAttachments] = useState<File[]>([])
  const [attachmentSizeError, setAttachmentSizeError] = useState<string | null>(null)
  const [composerVisibility, setComposerVisibility] =
    useState<GrievanceTimelineVisibility>('shared')

  const isInternalDraft = canPostInternalNote && composerVisibility === 'internal'

  return (
    <Stack spacing={1}>
      {canPostInternalNote ? (
        <ToggleButtonGroup
          size="small"
          exclusive
          value={composerVisibility}
          onChange={(_, nextValue: GrievanceTimelineVisibility | null) => {
            if (nextValue) {
              setComposerVisibility(nextValue)
            }
          }}
        >
          <ToggleButton value="shared">{t('grievances.activity.publicReply')}</ToggleButton>
          <ToggleButton value="internal">{t('grievances.activity.internalNote')}</ToggleButton>
        </ToggleButtonGroup>
      ) : null}

      <TextField
        fullWidth
        multiline
        minRows={2}
        placeholder={
          isInternalDraft
            ? t('grievances.activity.composerPlaceholderInternal')
            : t('grievances.activity.composerPlaceholderPublic')
        }
        value={draft}
        onChange={(event) => {
          setDraft(event.target.value)
        }}
      />

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
              ? t('grievances.activity.attachmentTooLarge', {
                  fileNames: oversizedFiles.map((file) => file.name).join(', '),
                  maxSize: MAX_ATTACHMENT_SIZE_LABEL,
                })
              : null,
          )
          setDraftAttachments((previousFiles) => [...previousFiles, ...acceptedFiles])
          input.value = ''
        }}
      />

      {attachmentSizeError ? (
        <Typography variant="caption" color="error.main">
          {attachmentSizeError}
        </Typography>
      ) : null}

      {draftAttachments.length > 0 ? (
        <Stack direction="row" spacing={1} flexWrap="wrap" useFlexGap>
          {draftAttachments.map((file, index) => (
            <Chip
              key={`${file.name}-${String(index)}`}
              size="small"
              variant="outlined"
              icon={<Paperclip size={14} />}
              label={file.name}
              onDelete={() => {
                setDraftAttachments((previousFiles) =>
                  previousFiles.filter((_, fileIndex) => fileIndex !== index),
                )
              }}
              deleteIcon={<X size={14} />}
            />
          ))}
        </Stack>
      ) : null}

      <Stack
        direction="row"
        spacing={1.5}
        justifyContent="space-between"
        alignItems="center"
        flexWrap="wrap"
        useFlexGap
      >
        <Stack direction="row" spacing={1} alignItems="center">
          <Button
            variant="outlined"
            size="small"
            startIcon={<Paperclip size={16} />}
            onClick={() => fileInputRef.current?.click()}
          >
            {t('grievances.activity.attach')}
          </Button>
          <Typography variant="caption" color="text.secondary">
            {t('grievances.activity.attachmentsHelp', { maxSize: MAX_ATTACHMENT_SIZE_LABEL })}
          </Typography>
        </Stack>
        <Button
          variant="contained"
          color={isInternalDraft ? 'warning' : 'primary'}
          size="small"
          startIcon={isInternalDraft ? <Lock size={16} /> : <Send size={16} />}
          disabled={!draft.trim()}
          onClick={() => {
            onSend(draft.trim(), draftAttachments, isInternalDraft ? 'internal' : 'shared')
            setDraft('')
            setDraftAttachments([])
            setAttachmentSizeError(null)
          }}
        >
          {t('grievances.activity.send')}
        </Button>
      </Stack>
    </Stack>
  )
}

export default GrievanceReplyComposer
