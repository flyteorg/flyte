import { ProjectIdentifier } from '@/gen/flyteidl2/common/identifier_pb'
import { Filter, Sort_Direction } from '@/gen/flyteidl2/common/list_pb'
import {
  ListRunsRequestSchema,
  ListRunsResponse,
  RunService,
  WatchRunsResponse,
} from '@/gen/flyteidl2/workflow/run_service_pb'
import { create } from '@bufbuild/protobuf'
import { useInfiniteQuery, useQueryClient } from '@tanstack/react-query'
import { useCallback, useEffect, useRef } from 'react'
import { useConnectRpcClient } from './useConnectRpc'

interface UseWatchRunsOptions {
  limit?: number
  projectId?: ProjectIdentifier
  filters?: Filter[]
  enabled?: boolean
  /** When true, enables the watch stream for live updates even when filters are present.
   *  Useful when filters are "open-ended" (e.g., created_at >= X without an upper bound). */
  enableLiveUpdates?: boolean
}

// Retry every 1000ms
const RETRY_DELAY = 1000
// Buffer configuration
const BUFFER_FLUSH_INTERVAL_MS = 200 // Flush every 200ms
const BUFFER_MAX_SIZE = 10 // Flush immediately if buffer exceeds this
// Polling interval when filters are present
const POLLING_INTERVAL_MS = 2000 // Poll every 2 seconds

// Helper function to extract created_at timestamp from a run
// Uses startTime as a proxy for created_at since runs are sorted by created_at
const getRunTimestamp = (run: {
  action?: {
    status?: { startTime?: { seconds?: bigint | number; nanos?: number } }
  }
}): Date | undefined => {
  const startTime = run.action?.status?.startTime
  if (!startTime?.seconds) return undefined

  // Convert protobuf Timestamp to Date
  const seconds = Number(startTime.seconds)
  const nanos = Number(startTime.nanos || 0)
  return new Date(seconds * 1000 + nanos / 1_000_000)
}

export function useWatchRuns({
  limit = 100,
  projectId,
  filters,
  enabled = true,
  enableLiveUpdates = false,
}: UseWatchRunsOptions = {}) {
  const client = useConnectRpcClient(RunService)
  const queryClient = useQueryClient()
  const streamRef = useRef<AsyncIterable<WatchRunsResponse>>(undefined)
  const abortControllerRef = useRef<AbortController>(undefined)
  const watchStartedRef = useRef<boolean>(false)
  const retryCountRef = useRef<number>(0)
  const retryTimeoutRef = useRef<NodeJS.Timeout>(undefined)
  const isReconnectingRef = useRef<boolean>(false)
  const updateBufferRef = useRef<WatchRunsResponse[]>([])
  const flushTimeoutRef = useRef<NodeJS.Timeout>(undefined)
  const newRunsCountRef = useRef<number>(0) // Track new runs added to first page
  const pollingIntervalRef = useRef<NodeJS.Timeout>(undefined)

  // Check if any filters are present that should disable live updates
  // When enableLiveUpdates is true, we use the watch stream even with filters
  const hasFilters = filters && filters.length > 0 && !enableLiveUpdates

  // When task name is in filters (e.g. task details page), disable the watch stream.
  // Polling is still used to refresh the list (i.e., list via polling instead of watch).
  const hasTaskNameFilter =
    filters?.some((f) => f.field === 'task_name') ?? false

  // Flush buffered updates to query data
  const flushBuffer = useCallback(() => {
    if (updateBufferRef.current.length === 0) return

    // Merge all buffered responses into a single update
    const allRuns = updateBufferRef.current.flatMap(
      (response) => response.runs ?? [],
    )

    // Clear the buffer
    updateBufferRef.current = []

    // Update the query cache with all buffered data
    queryClient.setQueryData(
      ['watchRuns', { projectId, limit, filters }],
      (
        oldData: { pages: ListRunsResponse[] } | undefined = {
          pages: [],
        },
      ) => {
        // Create a map of all runs from buffered updates, keeping the latest version of each
        const runsMap = new Map<string, (typeof allRuns)[0]>()
        for (const run of allRuns) {
          const runId = run.action?.id?.run?.name
          if (runId) {
            runsMap.set(runId, run)
          }
        }

        // Collect all existing run IDs from ALL pages to determine which runs are new
        const allExistingRunIds = new Set<string>()
        for (const page of oldData.pages) {
          for (const run of page.runs ?? []) {
            const runId = run.action?.id?.run?.name
            if (runId) {
              allExistingRunIds.add(runId)
            }
          }
        }

        // Find runs that are truly new
        const newRuns = Array.from(runsMap.values()).filter((run) => {
          const runId = run.action?.id?.run?.name
          return runId && !allExistingRunIds.has(runId)
        })

        // Filter out new runs that are older than the top run in the list (within 1s threshold)
        let filteredNewRuns = newRuns
        if (newRuns.length > 0 && oldData.pages.length > 0) {
          const firstPage = oldData.pages[0]
          const topRun = firstPage.runs?.[0]

          if (topRun) {
            const topRunTimestamp = getRunTimestamp(topRun)

            if (topRunTimestamp) {
              // Calculate threshold: top run timestamp minus 1 second
              const thresholdTimestamp = new Date(
                topRunTimestamp.getTime() - 1000,
              )

              // Filter new runs: only keep those that are newer than or equal to the threshold
              filteredNewRuns = newRuns.filter((run) => {
                const runTimestamp = getRunTimestamp(run)
                if (!runTimestamp) return true // Keep runs without timestamps (shouldn't happen, but be safe)
                // Only keep runs that are newer than or equal to the threshold
                return runTimestamp >= thresholdTimestamp
              })
            }
          }
        }

        // Track new runs added to first page for pagination offset adjustment
        if (filteredNewRuns.length > 0 && oldData.pages.length > 0) {
          newRunsCountRef.current += filteredNewRuns.length
        }

        // Update all pages with new data
        const updatedPages = oldData.pages.map(
          (page: ListRunsResponse, pageIndex) => {
            const existingRuns = page.runs ?? []

            // Update existing runs in place (maintain their position)
            const updatedRuns = existingRuns.map((existingRun) => {
              const runId = existingRun.action?.id?.run?.name
              const matchingNewRun = runId ? runsMap.get(runId) : undefined
              return matchingNewRun || existingRun
            })

            // Only add new runs to the first page (they go at the top since sorting is DESCENDING by created_at)
            // Existing runs stay in their original positions across all pages
            if (pageIndex === 0) {
              return {
                ...page,
                runs: [...filteredNewRuns, ...updatedRuns],
              }
            } else {
              return {
                ...page,
                runs: updatedRuns,
              }
            }
          },
        )

        return {
          ...oldData,
          pages: updatedPages,
        }
      },
    )
  }, [queryClient, projectId, limit, filters])

  // Add response to buffer and schedule flush if needed
  const addToBuffer = useCallback(
    (response: WatchRunsResponse) => {
      updateBufferRef.current.push(response)

      // Flush immediately if buffer is too large
      if (updateBufferRef.current.length >= BUFFER_MAX_SIZE) {
        if (flushTimeoutRef.current) {
          clearTimeout(flushTimeoutRef.current)
          flushTimeoutRef.current = undefined
        }
        flushBuffer()
        return
      }

      // Schedule flush if not already scheduled
      if (!flushTimeoutRef.current) {
        flushTimeoutRef.current = setTimeout(() => {
          flushTimeoutRef.current = undefined
          flushBuffer()
        }, BUFFER_FLUSH_INTERVAL_MS)
      }
    },
    [flushBuffer],
  )

  // Stop polling
  const stopPolling = useCallback(() => {
    if (pollingIntervalRef.current) {
      clearInterval(pollingIntervalRef.current)
      pollingIntervalRef.current = undefined
    }
  }, [])

  // Start polling when filters are present
  const startPolling = useCallback(() => {
    if (!projectId?.organization || !projectId?.domain || !projectId?.name) {
      return
    }

    // Stop any existing polling
    stopPolling()

    // Start polling by refetching the query
    const poll = () => {
      queryClient.invalidateQueries({
        queryKey: ['watchRuns', { projectId, limit, filters }],
      })
    }

    // Poll immediately, then set up interval
    poll()
    pollingIntervalRef.current = setInterval(poll, POLLING_INTERVAL_MS)
  }, [projectId, limit, filters, queryClient, stopPolling])

  // Cleanup function for the stream
  const cleanup = useCallback(() => {
    if (abortControllerRef.current) {
      abortControllerRef.current.abort()
      abortControllerRef.current = undefined
    }
    if (retryTimeoutRef.current) {
      clearTimeout(retryTimeoutRef.current)
      retryTimeoutRef.current = undefined
    }
    if (flushTimeoutRef.current) {
      clearTimeout(flushTimeoutRef.current)
      flushTimeoutRef.current = undefined
    }
    // Stop polling if running
    stopPolling()
    // Flush any remaining buffered updates before cleanup
    if (updateBufferRef.current.length > 0) {
      flushBuffer()
    }
    updateBufferRef.current = []
    streamRef.current = undefined
    watchStartedRef.current = false
    isReconnectingRef.current = false
    newRunsCountRef.current = 0
  }, [flushBuffer, stopPolling])

  // Handle cleanup on unmount
  useEffect(() => {
    return () => cleanup()
  }, [cleanup])

  // Handle projectId and filter changes - switch between watch and polling
  useEffect(() => {
    if (!projectId?.organization || !projectId?.domain || !projectId?.name) {
      return
    }

    // Trigger a refetch when projectId or filters change
    queryClient.invalidateQueries({
      queryKey: ['watchRuns', { projectId, limit, filters }],
    })

    if (hasTaskNameFilter || hasFilters) {
      // Watch doesn't support task_name filter; other filters also disable watch. Use list + polling.
      cleanup()
      startPolling()
    } else {
      // If no filters, stop polling and reset watch stream state
      stopPolling()
      watchStartedRef.current = false
      retryCountRef.current = 0
      isReconnectingRef.current = false
      newRunsCountRef.current = 0
      // Clear any existing retry timeout
      if (retryTimeoutRef.current) {
        clearTimeout(retryTimeoutRef.current)
        retryTimeoutRef.current = undefined
      }
    }
  }, [
    projectId,
    queryClient,
    limit,
    filters,
    hasFilters,
    hasTaskNameFilter,
    cleanup,
    startPolling,
    stopPolling,
  ])

  const fetchRunsPage = useCallback(
    async ({
      pageParam = '',
    }: {
      pageParam?: string
    }): Promise<ListRunsResponse> => {
      if (!projectId?.organization || !projectId?.domain || !projectId?.name) {
        // Return empty response with proper structure
        return {
          $typeName: 'flyteidl2.workflow.ListRunsResponse',
          runs: [],
          token: '',
        } as ListRunsResponse
      }

      const listRequest = create(ListRunsRequestSchema, {
        request: {
          limit,
          token: pageParam,
          sortBy: {
            direction: Sort_Direction.DESCENDING,
            key: 'created_at',
          },
          filters,
        },
        scopeBy: {
          case: 'projectId',
          value: projectId,
        },
      })

      const response = await client.listRuns(listRequest)
      return response
    },
    [client, limit, projectId, filters],
  )

  const infiniteQuery = useInfiniteQuery({
    queryKey: ['watchRuns', { projectId, limit, filters }],
    queryFn: fetchRunsPage,
    initialPageParam: '',
    getNextPageParam: (
      lastPage: ListRunsResponse,
      allPages: ListRunsResponse[],
    ) => {
      // Only return the next token if the current page has data and a valid token
      if (
        lastPage.runs &&
        lastPage.runs.length > 0 &&
        lastPage.token &&
        lastPage.token.trim() !== ''
      ) {
        // If we're on the first page and new runs were added, adjust the token
        if (allPages.length === 1 && newRunsCountRef.current > 0) {
          const baseOffset = parseInt(lastPage.token, 10)
          if (!isNaN(baseOffset)) {
            // Adjust token to account for new runs added to the first page
            const adjustedToken = (
              baseOffset + newRunsCountRef.current
            ).toString()
            return adjustedToken
          }
        }
        return lastPage.token
      }
      return undefined
    },
    enabled:
      enabled &&
      !!projectId?.domain &&
      !!projectId?.name &&
      !!projectId?.organization,
    gcTime: 0,
    staleTime: 0,
    refetchOnWindowFocus: false,
  })

  // Function to start the watch stream
  const startWatchStream = useCallback(async () => {
    if (!projectId?.organization || !projectId?.domain || !projectId?.name) {
      return
    }

    // Clean up any existing stream
    if (abortControllerRef.current) {
      abortControllerRef.current.abort()
    }

    // Start the watch stream for updates
    const abortController = new AbortController()
    abortControllerRef.current = abortController
    watchStartedRef.current = true
    isReconnectingRef.current = false

    const stream = client.watchRuns(
      {
        target: {
          case: 'projectId',
          value: projectId,
        },
      },
      {
        signal: abortController.signal,
      },
    )
    streamRef.current = stream

    // Start processing the stream
    const processStream = async () => {
      try {
        for await (const response of stream) {
          if (abortController.signal.aborted) {
            break
          }

          // Add to buffer instead of immediately updating query data
          addToBuffer(response)
        }

        // Flush any remaining buffered updates when stream ends
        flushBuffer()
      } catch (error) {
        if (!abortController.signal.aborted) {
          console.error('Error watching runs:', error)

          // Attempt to reconnect indefinitely
          if (!isReconnectingRef.current) {
            isReconnectingRef.current = true
            retryCountRef.current++

            console.log(
              `Attempting to reconnect in ${RETRY_DELAY}ms (attempt ${retryCountRef.current})`,
            )

            retryTimeoutRef.current = setTimeout(() => {
              watchStartedRef.current = false
              startWatchStream()
            }, RETRY_DELAY)
          }
        }
      }
    }

    // Start processing the stream without awaiting it
    processStream().catch(console.error)
  }, [client, projectId, addToBuffer, flushBuffer])

  // Start the watch stream only when we have a valid projectId and no filters.
  // Never start watch when task_name filter is present (task details page).
  useEffect(() => {
    if (
      !projectId?.organization ||
      !projectId?.domain ||
      !projectId?.name ||
      watchStartedRef.current ||
      hasFilters ||
      hasTaskNameFilter
    ) {
      return
    }

    startWatchStream()

    return () => {
      cleanup()
    }
  }, [projectId, startWatchStream, cleanup, hasFilters, hasTaskNameFilter])

  return infiniteQuery
}
