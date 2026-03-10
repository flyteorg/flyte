'use client'

import { useCreateUploadLocation } from '@/hooks/useCreateUploadLocation'
import type { FieldProps } from '@rjsf/utils'
import React, { useCallback, useEffect, useRef, useState } from 'react'
import { decodeBlobUri } from './blobUriUtils'

/** Form context for blob uploads: project and domain are required to get signed URLs. */
export type BlobFieldContext = {
  project?: string
  domain?: string
  /** Called when upload starts (true) or ends (false). Used to disable submit until upload completes. */
  onBlobUploadingChange?: (uploading: boolean) => void
}

/** Form data shape for blob object schema (dimensionality, format, uri). */
export type BlobObjectFormData = {
  dimensionality?: string
  format?: string
  uri?: string
}

const DEFAULT_PLACEHOLDER = 'Enter URI or upload a file'
const ERROR_MISSING_CONTEXT = 'Project and domain are required for file uploads'

/**
 * Blob field: single control for URI input or file upload.
 * Used when the launch form schema is object with format "blob" (dimensionality, format, uri).
 * Styled to match other launch form inputs (8px radius, gray-4 border, 14px text).
 */
export default function CustomBlobField(props: FieldProps<BlobObjectFormData>) {
  const { formData, onChange, readonly, schema, idSchema, name, required } =
    props

  // --- State ---
  const [isUploading, setIsUploading] = useState(false)
  const [uploadError, setUploadError] = useState<string | null>(null)
  const fileInputRef = useRef<HTMLInputElement>(null)

  // --- Context & derived values ---
  const { uploadFile } = useCreateUploadLocation()
  const { project = '', domain = '' } = (props.formContext ||
    {}) as BlobFieldContext
  const data = formData ?? {}
  const rawUri = typeof data.uri === 'string' ? data.uri : ''
  const uri = decodeBlobUri(rawUri)
  const dimensionality = data.dimensionality ?? 'SINGLE'
  const format = data.format ?? ''
  const placeholder = schema.description ?? DEFAULT_PLACEHOLDER
  const fieldId = idSchema?.$id ?? name

  // --- One-time normalize: decode URI if form was loaded with encoded value ---
  useEffect(() => {
    if (rawUri && decodeBlobUri(rawUri) !== rawUri) {
      onChange({ ...data, dimensionality, format, uri: decodeBlobUri(rawUri) })
    }
  }, []) // eslint-disable-line react-hooks/exhaustive-deps -- run once on mount to fix encoded initial URI

  // --- Handlers ---
  const updateUri = useCallback(
    (newUri: string) => {
      onChange({
        ...data,
        dimensionality,
        format,
        uri: newUri ? decodeBlobUri(newUri) : '',
      })
      setUploadError(null)
    },
    [data, dimensionality, format, onChange],
  )

  const handleFileChange = useCallback(
    async (e: React.ChangeEvent<HTMLInputElement>) => {
      const file = e.target.files?.[0]
      if (!file) return
      e.target.value = ''

      if (!project || !domain) {
        setUploadError(ERROR_MISSING_CONTEXT)
        return
      }

      const ctx = (props.formContext || {}) as BlobFieldContext
      setIsUploading(true)
      setUploadError(null)
      // Wipe previous value as soon as user picks a new file
      onChange({
        ...data,
        dimensionality,
        format,
        uri: '',
      })
      ctx.onBlobUploadingChange?.(true)
      try {
        const nativeUrl = await uploadFile(file, { project, domain })
        onChange({
          ...data,
          dimensionality,
          format,
          uri: decodeBlobUri(nativeUrl),
        })
      } catch (err) {
        setUploadError(
          err instanceof Error ? err.message : 'Failed to upload file',
        )
      } finally {
        setIsUploading(false)
        ctx.onBlobUploadingChange?.(false)
      }
    },
    [
      uploadFile,
      project,
      domain,
      data,
      dimensionality,
      format,
      onChange,
      props.formContext,
    ],
  )

  const openFilePicker = useCallback(() => {
    fileInputRef.current?.click()
  }, [])

  // --- Readonly: plain input, no upload UI ---
  if (readonly) {
    return (
      <input
        type="text"
        value={uri}
        readOnly
        id={fieldId}
        name={name}
        aria-required={required}
        data-blob-field="readonly"
      />
    )
  }

  // --- Editable: URI input + upload spinner + file input + button ---
  return (
    <div
      className="flex w-full min-w-0 flex-col gap-1"
      data-blob-field
      data-upload-state={isUploading ? 'uploading' : 'idle'}
    >
      <div
        className="flex w-full min-w-0 items-center gap-2 rounded-[8px] border border-(--system-gray-4) bg-white px-3 py-1.5 focus-within:border-transparent focus-within:ring-2 focus-within:ring-blue-500 focus-within:outline-none dark:bg-zinc-900"
        data-blob-field-container
      >
        {/* Upload in progress: spinner on the left */}
        {isUploading && (
          <span
            className="h-4 w-4 shrink-0 animate-spin rounded-full border-2 border-zinc-300 border-t-transparent dark:border-zinc-600"
            aria-hidden
            data-blob-field-spinner
          />
        )}

        <input
          type="text"
          value={uri}
          onChange={(e) => updateUri(e.target.value)}
          placeholder={placeholder}
          className="min-w-0 flex-1 border-0 bg-transparent p-0 text-[14px] text-zinc-900 placeholder:text-zinc-500 focus:ring-0 focus:outline-none dark:text-white dark:placeholder:text-zinc-400"
          id={fieldId}
          name={name}
          data-testid={fieldId ? `input-${fieldId}` : undefined}
          data-blob-field-input
          aria-required={required}
        />

        <input
          ref={fileInputRef}
          type="file"
          className="hidden"
          aria-hidden
          onChange={handleFileChange}
          disabled={isUploading}
          data-blob-field-file-input
        />

        <button
          type="button"
          onClick={openFilePicker}
          disabled={isUploading || !project || !domain}
          className="flex shrink-0 items-center justify-center gap-1 rounded border border-zinc-300 bg-transparent px-1.5 py-0.5 text-[13px] font-medium text-zinc-500 hover:bg-zinc-100 disabled:opacity-50 dark:border-(--system-gray-3) dark:text-(--system-gray-5) dark:hover:bg-zinc-800"
          data-blob-field-upload-button
        >
          Upload file
        </button>
      </div>

      {uploadError && (
        <span
          className="text-sm text-red-500"
          role="alert"
          data-blob-field-error
        >
          {uploadError}
        </span>
      )}
    </div>
  )
}
