'use client'

import { ActionDetails } from '@/gen/flyteidl2/workflow/run_definition_pb'
import {
  RunService,
  WatchActionDetailsRequestSchema,
  WatchActionDetailsResponse,
} from '@/gen/flyteidl2/workflow/run_service_pb'
import { isActionTerminal } from '@/lib/actionUtils'
import { create } from '@bufbuild/protobuf'
import { Code, ConnectError } from '@connectrpc/connect'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { useCallback, useEffect, useRef } from 'react'
import { useActionIdentifier } from './useActionIdentifier'
import { useConnectRpcClient } from './useConnectRpc'

const isNetworkError = (error: unknown) => {
  if (!(error instanceof Error)) {
    return false
  }

  if (error instanceof ConnectError) {
    return (
      error.code === Code.Unavailable ||
      error.code === Code.DeadlineExceeded ||
      error.code === Code.Unknown
    )
  }

  if (error instanceof TypeError) {
    const message = error.message.toLowerCase()
    // Typical fetch transport network failures
    return (
      message.includes('failed to fetch') ||
      message.includes('networkerror') ||
      message.includes('load failed')
    )
  }

  return false
}

export function useWatchActionDetails(actionId?: string | null) {
  const client = useConnectRpcClient(RunService)
  const queryClient = useQueryClient()
  const streamRef = useRef<AsyncIterable<WatchActionDetailsResponse>>(undefined)
  const abortControllerRef = useRef<AbortController>(undefined)
  const isInitializedRef = useRef(false)

  const actionIdentifier = useActionIdentifier(actionId)

  // Helper function to check if we should reconnect
  const checkAndReconnectIfNeeded = useCallback(
    (reason: string) => {
      const action = queryClient.getQueryData([
        'actionDetails',
        actionIdentifier,
      ]) as ActionDetails | undefined
      const isTerminal = isActionTerminal(action)

      if (!isTerminal) {
        console.log(
          `%c watchActionDetails RECONNECTING after ${reason} - action not terminal`,
          'background: #FF9800; color: white; padding: 2px 5px; border-radius: 2px',
          'actionId:',
          actionId,
        )
        // Reset initialization flag to allow reconnection
        isInitializedRef.current = false
        // Trigger reconnection by invalidating the query
        queryClient.invalidateQueries({
          queryKey: ['actionDetails', actionIdentifier],
        })
      }
    },
    [actionIdentifier, actionId, queryClient],
  )

  // Cleanup function for the stream
  const cleanup = useCallback(() => {
    if (abortControllerRef.current) {
      abortControllerRef.current.abort()
      abortControllerRef.current = undefined
    }
    isInitializedRef.current = false
  }, [])

  // Handle cleanup on unmount or actionId change
  useEffect(() => {
    return () => cleanup()
  }, [cleanup, actionId])

  // Invalidate query when actionIdentifier changes
  useEffect(() => {
    if (actionIdentifier) {
      queryClient.invalidateQueries({
        queryKey: ['actionDetails', actionIdentifier],
      })
    }
  }, [actionIdentifier, queryClient])

  useEffect(() => {
    if (!actionId) {
      cleanup()
      queryClient.removeQueries({ queryKey: ['actionDetails'] })
    }
  }, [actionId, cleanup, queryClient])

  const startStream = useCallback(async () => {
    if (!actionIdentifier) {
      cleanup()
      return null
    }

    if (isInitializedRef.current) {
      // Return the current data from the cache if we're already initialized
      return (
        queryClient.getQueryData(['actionDetails', actionIdentifier]) ?? null
      )
    }

    console.log(
      '%c watchActionDetails STARTING',
      'background: #4CAF50; color: white; padding: 2px 5px; border-radius: 2px',
      'actionId:',
      actionId,
    )

    isInitializedRef.current = true

    try {
      // Create request
      const request = create(WatchActionDetailsRequestSchema, {
        actionId: actionIdentifier,
      })

      // Start the watch stream for updates
      const abortController = new AbortController()
      abortControllerRef.current = abortController

      const stream = client.watchActionDetails(request, {
        signal: abortController.signal,
      })
      streamRef.current = stream

      // Start processing the stream
      const processStream = async () => {
        try {
          for await (const event of stream) {
            if (abortController.signal.aborted) {
              console.log(
                '%c watchActionDetails ABORTED',
                'background: #ff0000; color: white; padding: 2px 5px; border-radius: 2px',
                'actionId:',
                actionId,
              )
              break
            }

            if (event.details) {
              console.log(
                '%c watchActionDetails DATA received',
                'background: #2196F3; color: white; padding: 2px 5px; border-radius: 2px',
                'actionId:',
                actionId,
                event.details,
              )
              // Update the query cache with the new data
              queryClient.setQueryData(
                ['actionDetails', actionIdentifier],
                event.details,
              )
            }
          }
        } catch (error) {
          if (!abortController.signal.aborted) {
            console.error(
              '%c Stream ERROR',
              'background: #ff0000; color: white; padding: 2px 5px; border-radius: 2px',
              'actionId:',
              actionId,
              '\n error:',
              error,
            )

            // Only reconnect on transient network errors; otherwise surface the error
            if (isNetworkError(error)) {
              checkAndReconnectIfNeeded('ERROR')
            } else {
              cleanup()
            }

            throw error
          }
        }

        // If the action is not terminal and the stream ended normally, reconnect
        checkAndReconnectIfNeeded('normal end')
      }

      // Start processing the stream without awaiting it
      processStream().catch((error) => {
        console.error('Stream processing error:', error)
        cleanup()
      })

      // Return null initially - the data will be updated through the stream
      return null
    } catch (error) {
      cleanup()
      throw error
    }
  }, [
    client,
    actionIdentifier,
    queryClient,
    cleanup,
    actionId,
    checkAndReconnectIfNeeded,
  ])

  const query = useQuery({
    queryKey: ['actionDetails', actionIdentifier],
    queryFn: startStream,
    refetchInterval: false,
    refetchOnWindowFocus: false,
    enabled: !!actionIdentifier,
    experimental_prefetchInRender: true,
    retry: (failureCount, error) => {
      // Don't retry if the query was cancelled
      if (error instanceof Error && error.message.includes('aborted')) {
        return false
      }
      // Only retry on transient network errors
      if (!isNetworkError(error)) {
        return false
      }
      // Retry network errors up to 3 times with exponential backoff
      return failureCount < 3
    },
    retryDelay: (attemptIndex) =>
      Math.min(1000 * Math.pow(2, attemptIndex), 30000),
  })

  return {
    ...query,
    data: query.data as ActionDetails | undefined,
    cleanup,
  }
}
