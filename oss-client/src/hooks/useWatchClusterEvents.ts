import { LogLine } from '@/gen/flyteidl2/logs/dataplane/payload_pb'
import { ActionDetails } from '@/gen/flyteidl2/workflow/run_definition_pb'
import {
  RunService,
  WatchClusterEventsRequestSchema,
  WatchClusterEventsResponse,
} from '@/gen/flyteidl2/workflow/run_service_pb'
import { create } from '@bufbuild/protobuf'
import { Code, ConnectError } from '@connectrpc/connect'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { useCallback, useEffect, useMemo, useRef } from 'react'
import { useConnectRpcClient } from './useConnectRpc'

interface UseWatchClusterEventsOptions {
  actionDetails?: ActionDetails
  attempt?: number | null
  enabled?: boolean
}

interface ClusterEventsState {
  lines: LogLine[]
}

export function useWatchClusterEvents({
  actionDetails,
  attempt = 0,
  enabled = false,
}: UseWatchClusterEventsOptions = {}) {
  const client = useConnectRpcClient(RunService)
  const queryClient = useQueryClient()

  const queryKey = useMemo(
    () => ['watchK8sEvents', { actionId: actionDetails?.id, attempt }],
    [actionDetails?.id, attempt],
  )

  const streamRef = useRef<AsyncIterable<WatchClusterEventsResponse>>(undefined)
  const abortControllerRef = useRef<AbortController>(undefined)

  // Cleanup function for the stream
  const cleanup = useCallback(() => {
    if (abortControllerRef.current) {
      abortControllerRef.current.abort()
      abortControllerRef.current = undefined
    }
  }, [])

  // Handle cleanup on unmount
  useEffect(() => {
    return () => {
      cleanup()
    }
  }, [cleanup])

  const query = useQuery<ClusterEventsState>({
    queryKey,
    queryFn: async () => {
      return new Promise<ClusterEventsState>(async (resolve, reject) => {
        const attemptNumber = attempt || 0
        // Create the tail logs request
        const tailRequest = create(WatchClusterEventsRequestSchema, {
          id: actionDetails!.id,
          attempt: attemptNumber,
        })

        // Start the watch stream for updates
        const abortController = new AbortController()
        abortControllerRef.current = abortController

        const stream = client.watchClusterEvents(tailRequest, {
          signal: abortController.signal,
        })
        streamRef.current = stream

        try {
          for await (const response of stream) {
            if (abortController.signal.aborted) {
              break
            }

            const newLines = response.clusterEvents.map(
              ({ occurredAt, message }) => ({
                timestamp: occurredAt,
                message,
              }),
            )

            queryClient.setQueryData(
              queryKey,
              (oldData: ClusterEventsState = { lines: [] }) => {
                const existingLines = oldData.lines ?? []
                return {
                  lines: [...existingLines, ...newLines],
                }
              },
            )
          }
        } catch (error) {
          if (!abortController.signal.aborted) {
            reject(error)
          }
        }
        const data = queryClient.getQueryData<ClusterEventsState>(queryKey)
        resolve(data || { lines: [] })
      })
    },
    enabled: !!actionDetails && enabled,
    refetchInterval: false,
    refetchOnWindowFocus: false,
    retry: (failureCount, error) => {
      // Disable retry for specific error types
      if (error instanceof ConnectError) {
        if (
          // Don't retry if err code is not found
          error.code === Code.NotFound
        ) {
          return false
        }
      }

      // For other errors, allow retry up to 3 times
      return failureCount < 3
    },
    gcTime: 0,
    staleTime: 0,
  })

  return {
    ...query,

    cleanup,
  }
}
