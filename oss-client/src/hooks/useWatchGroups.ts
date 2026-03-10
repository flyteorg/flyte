import { ProjectIdentifier } from '@/gen/flyteidl2/common/identifier_pb'
import {
  Filter,
  ListRequestSchema,
  Sort_Direction,
  SortSchema,
} from '@/gen/flyteidl2/common/list_pb'
import { TaskGroup } from '@/gen/flyteidl2/workflow/run_definition_pb'
import {
  RunService,
  WatchGroupsRequestSchema,
  WatchGroupsResponse,
} from '@/gen/flyteidl2/workflow/run_service_pb'
import { create } from '@bufbuild/protobuf'
import { timestampFromDate } from '@bufbuild/protobuf/wkt'
// NOTE: `experimental_streamedQuery` is an experimental TanStack Query API.
// Future releases of `@tanstack/react-query` may introduce breaking changes
// to this API. When upgrading the dependency, verify that this import and
// any related usage in this hook are still valid.
import {
  experimental_streamedQuery as streamedQuery,
  useQuery,
} from '@tanstack/react-query'
import { useMemo } from 'react'
import { getErrorInfo, isNonRetryableError } from '@/lib/errorUtils'
import { useConnectRpcClient } from './useConnectRpc'

const RETRY_EVERY_MS = 1000

export interface UseWatchGroupsOptions {
  startTime: Date
  projectId?: ProjectIdentifier
  filters?: Filter[]
  enabled?: boolean
}

type Groups = WatchGroupsResponse['taskGroups']

/**
 * Merges streamed group responses into the accumulated state.
 * Handles sentinel responses (full replacement) and incremental updates.
 */
function mergeGroups(prev: Groups, response: WatchGroupsResponse): Groups {
  const existing = prev ?? []
  const next = response.taskGroups ?? []

  const groupKey = (g: TaskGroup) => `${g.taskName}:${g.environmentName}`

  if (response.sentinel) {
    return next
  }

  if (next.length === 0) {
    return existing
  }

  const updatedKeys = new Set(next.map(groupKey))
  const remaining = existing.filter((g) => !updatedKeys.has(groupKey(g)))
  return [...next, ...remaining]
}

export function useWatchGroups({
  projectId,
  filters,
  startTime,
  enabled = true,
}: UseWatchGroupsOptions) {
  const client = useConnectRpcClient(RunService)

  const hasValidProjectId =
    !!projectId?.domain && !!projectId?.name && !!projectId?.organization

  // Serialize filters to a stable string to prevent key changes from reference equality
  const filtersKey = useMemo(
    () =>
      filters
        ? JSON.stringify(
            filters.map((f) => ({
              function: f.function,
              field: f.field,
              values: f.values,
            })),
          )
        : null,
    [filters],
  )

  // Convert startTime to ISO string for stable queryKey comparison
  // startTime is already stabilized via ref in ListRunsGroupedView
  const queryKey = useMemo(
    () =>
      [
        'watchGroups',
        {
          organization: projectId?.organization ?? null,
          domain: projectId?.domain ?? null,
          name: projectId?.name ?? null,
        },
        startTime.toISOString(),
        filtersKey,
      ] as const,
    [
      projectId?.organization,
      projectId?.domain,
      projectId?.name,
      startTime,
      filtersKey,
    ],
  )

  const query = useQuery({
    queryKey,
    enabled: hasValidProjectId && enabled,
    retry: (failureCount, error) => {
      const errorInfo = getErrorInfo(error)
      const isNonRetryable = isNonRetryableError(error)

      // Log the error when it occurs (this catches stream consumption errors)
      console.error('[useWatchGroups] Query error:', {
        error,
        ...errorInfo,
        failureCount,
        projectId: {
          organization: projectId?.organization,
          domain: projectId?.domain,
          name: projectId?.name,
        },
        startTime: startTime.toISOString(),
        filters: filtersKey,
        isNonRetryable,
      })

      // Don't retry non-retryable errors
      if (isNonRetryable) {
        console.info(
          `[useWatchGroups] Non-retryable error (${errorInfo.type}${errorInfo.code ? `, code: ${errorInfo.code}` : ''}), not retrying`,
        )
        return false
      }

      // Limit retries to prevent infinite loops
      const maxRetries = 5
      if (failureCount >= maxRetries) {
        console.error(
          `[useWatchGroups] Max retries (${maxRetries}) reached, stopping retries`,
        )
        return false
      }

      console.info(
        `[useWatchGroups] Retrying (attempt ${failureCount + 1}/${maxRetries})`,
      )
      return true
    },
    retryDelay: RETRY_EVERY_MS,
    queryFn: streamedQuery<WatchGroupsResponse, Groups>({
      streamFn: async ({ signal }) => {
        // signal is automatically aborted when query is cancelled or component unmounts
        const startDate = timestampFromDate(startTime)

        const watchRequest = create(WatchGroupsRequestSchema, {
          scopeBy: {
            case: 'projectId',
            value: projectId!,
          },
          startDate,
          request: create(ListRequestSchema, {
            limit: 100,
            token: '',
            filters: filters,
            sortBy: create(SortSchema, {
              key: 'created_at',
              direction: Sort_Direction.DESCENDING,
            }),
          }),
        })

        return client.watchGroups(watchRequest, { signal })
      },
      reducer: (acc, chunk) => {
        try {
          return mergeGroups(acc, chunk)
        } catch (error) {
          const errorInfo = getErrorInfo(error)
          console.error('[useWatchGroups] Reducer error:', {
            error,
            ...errorInfo,
            accumulatedGroupsCount: acc?.length ?? 0,
            chunkSentinel: chunk.sentinel,
            chunkGroupsCount: chunk.taskGroups?.length ?? 0,
          })
          // Return accumulated state on reducer error to prevent data loss
          return acc ?? []
        }
      },

      initialValue: [],
      refetchMode: 'reset', // Start fresh on refetch rather than appending
    }),
  })

  return {
    groups: query.data ?? [],
    isLoading: query.isPending, // true until first chunk arrives
    isError: query.isError,
    error: query.error,
    isReconnecting: query.isFetching && !query.isPending, // stream is active but not initial load
  }
}
