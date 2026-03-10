'use client'

import {
  ActionIdentifierSchema,
  RunIdentifierSchema,
} from '@/gen/flyteidl2/common/identifier_pb'
import {
  AssistantService,
  ExplainErrorRequestSchema,
  ExplainErrorResponse,
} from '@/gen/workflow/assistant_service_pb'
import { create } from '@bufbuild/protobuf'
import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { flushSync } from 'react-dom'
import { useConnectRpcClient } from './useConnectRpc'

export interface UseAIAssistantProps {
  org: string
  domain: string
  projectId: string
  runName: string
  actionId: string
  enabled?: boolean
}

interface StreamState {
  data: string
  isLoading: boolean
  isError: boolean
  error: Error | null
  isCompleted: boolean
}

export function useAIAssistant({
  org,
  domain,
  projectId,
  runName,
  actionId,
  enabled = true,
}: UseAIAssistantProps) {
  const client = useConnectRpcClient(AssistantService)

  const [streamState, setStreamState] = useState<StreamState>({
    data: '',
    isLoading: false,
    isError: false,
    error: null,
    isCompleted: false,
  })

  const abortControllerRef = useRef<AbortController | null>(null)
  const streamRef = useRef<AsyncIterable<ExplainErrorResponse> | null>(null)
  const accumulatedTextRef = useRef<string>('')

  const runIdentifier = useMemo(() => {
    if (!org || !domain || !projectId || !runName) return null
    return create(RunIdentifierSchema, {
      org,
      domain,
      project: projectId,
      name: runName,
    })
  }, [org, domain, projectId, runName])

  const shouldStream = useMemo(
    () =>
      enabled &&
      !!org &&
      !!domain &&
      !!projectId &&
      !!runName &&
      !!actionId &&
      !!runIdentifier,
    [enabled, org, domain, projectId, runName, actionId, runIdentifier],
  )

  const cleanup = useCallback(() => {
    if (abortControllerRef.current) {
      abortControllerRef.current.abort()
      abortControllerRef.current = null
    }
    streamRef.current = null
    accumulatedTextRef.current = ''
  }, [])

  useEffect(() => {
    if (!shouldStream || !runIdentifier) {
      setStreamState({
        data: '',
        isLoading: false,
        isError: false,
        error: null,
        isCompleted: false,
      })
      return
    }

    // Cleanup any existing stream
    cleanup()

    // Reset state
    setStreamState({
      data: '',
      isLoading: true,
      isError: false,
      error: null,
      isCompleted: false,
    })

    const abortController = new AbortController()
    abortControllerRef.current = abortController

    const startStream = async () => {
      try {
        const actionIdentifier = create(ActionIdentifierSchema, {
          run: runIdentifier,
          name: actionId,
        })
        const request = create(ExplainErrorRequestSchema, {
          actionId: actionIdentifier,
        })

        const stream = client.explainError(request, {
          signal: abortController.signal,
        })
        streamRef.current = stream

        accumulatedTextRef.current = ''
        let hasReceivedData = false

        const updateState = (isFirstData = false) => {
          flushSync(() => {
            setStreamState((prev) => ({
              ...prev,
              data: accumulatedTextRef.current,
              isLoading: isFirstData ? false : prev.isLoading,
            }))
          })
        }

        for await (const response of stream) {
          if (abortController.signal.aborted) {
            break
          }

          if (response.explanation) {
            const isFirstData = !hasReceivedData
            hasReceivedData = true
            accumulatedTextRef.current += response.explanation
            updateState(isFirstData)
          }
        }

        setStreamState((prev) => ({
          ...prev,
          data: accumulatedTextRef.current,
          isLoading: false,
          isCompleted: true,
        }))
      } catch (error) {
        if (!abortController.signal.aborted) {
          setStreamState({
            data: accumulatedTextRef.current,
            isLoading: false,
            isError: true,
            error: error instanceof Error ? error : new Error(String(error)),
            isCompleted: false,
          })
        }
      } finally {
        streamRef.current = null
      }
    }

    startStream()

    return () => {
      cleanup()
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [
    shouldStream,
    runIdentifier,
    org,
    domain,
    projectId,
    runName,
    actionId,
    client,
    // cleanup is intentionally omitted - it's a stable useCallback with no dependencies
  ])

  let status: 'pending' | 'error' | 'success'
  if (streamState.isLoading) {
    status = 'pending'
  } else if (streamState.isError) {
    status = 'error'
  } else if (streamState.isCompleted) {
    status = 'success'
  } else {
    // Initial state or when streaming has not started; treated as 'pending' by design.
    status = 'pending'
  }

  return {
    data: streamState.data,
    isLoading: streamState.isLoading,
    isError: streamState.isError,
    error: streamState.error,
    isCompleted: streamState.isCompleted,
    status,
    isSuccess: streamState.isCompleted && !streamState.isError,
    isPending: streamState.isLoading,
  }
}
