'use client'

import { create } from '@bufbuild/protobuf'
import { useCallback } from 'react'
import SparkMD5 from 'spark-md5'
import { useConnectRpcClient } from './useConnectRpc'
import {
  CreateUploadLocationRequestSchema,
  DataProxyService,
} from '@/gen/flyteidl2/dataproxy/dataproxy_service_pb'

function calculateMD5(file: File): Promise<Uint8Array> {
  return new Promise((resolve, reject) => {
    const chunkSize = 2 * 1024 * 1024
    const chunks = Math.ceil(file.size / chunkSize)
    let currentChunk = 0
    const spark = new SparkMD5.ArrayBuffer()
    const fileReader = new FileReader()
    const blobSlice =
      File.prototype.slice ||
      (File.prototype as { mozSlice?: typeof File.prototype.slice }).mozSlice ||
      (File.prototype as { webkitSlice?: typeof File.prototype.slice })
        .webkitSlice ||
      File.prototype.slice

    const loadNext = () => {
      const start = currentChunk * chunkSize
      const end = start + chunkSize >= file.size ? file.size : start + chunkSize
      fileReader.readAsArrayBuffer(blobSlice.call(file, start, end))
    }

    fileReader.onload = (e) => {
      const result = e.target?.result
      if (result instanceof ArrayBuffer) {
        spark.append(result)
        currentChunk++
        if (currentChunk < chunks) {
          loadNext()
        } else {
          const hashHex = spark.end()
          const bytes = new Uint8Array(hashHex.length / 2)
          for (let i = 0; i < hashHex.length; i += 2) {
            bytes[i / 2] = parseInt(hashHex.substring(i, i + 2), 16)
          }
          resolve(bytes)
        }
      } else {
        reject(
          new Error(
            'MD5 calculation failed: file read did not return an ArrayBuffer (read may have been aborted)',
          ),
        )
      }
    }
    fileReader.onabort = () =>
      reject(new Error('MD5 calculation aborted while reading file'))
    fileReader.onerror = () =>
      reject(new Error('Failed to read file for MD5 calculation'))
    loadNext()
  })
}

async function uploadToSignedUrl(
  file: File,
  signedUrl: string,
  headers: Record<string, string>,
): Promise<void> {
  // Do not send Content-Type: GCS signed URLs are validated against the exact
  // headers that were signed; Content-Type can cause 403 Forbidden.
  const headersWithoutContentType = Object.fromEntries(
    Object.entries(headers).filter(
      ([key]) => key.toLowerCase() !== 'content-type',
    ),
  )
  const body = await file.arrayBuffer()
  const response = await fetch(signedUrl, {
    method: 'PUT',
    body,
    headers: headersWithoutContentType,
  })
  if (!response.ok) {
    throw new Error(`Upload failed: ${response.statusText}`)
  }
}

export type UseCreateUploadLocationParams = {
  project: string
  domain: string
}

/**
 * Uploads a file via DataProxy createUploadLocation and PUT to the signed URL.
 * Returns the native URL (storage path) for the uploaded file.
 */
export function useCreateUploadLocation() {
  const client = useConnectRpcClient(DataProxyService)

  const uploadFile = useCallback(
    async (
      file: File,
      { project, domain }: UseCreateUploadLocationParams,
    ): Promise<string> => {
      const contentMd5 = await calculateMD5(file)
      const uploadRequest = create(CreateUploadLocationRequestSchema, {
        project,
        domain,
        filename: file.name,
        contentMd5,
        addContentMd5Metadata: true,
      })

      const uploadResponse = await client.createUploadLocation(uploadRequest)
      await uploadToSignedUrl(
        file,
        uploadResponse.signedUrl,
        uploadResponse.headers ?? {},
      )
      return uploadResponse.nativeUrl
    },
    [client],
  )

  return { uploadFile }
}
