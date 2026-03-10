import { ProjectIdentifierSchema } from '@/gen/flyteidl2/common/identifier_pb'
import { Filter, Filter_Function } from '@/gen/flyteidl2/common/list_pb'
import {
  ListTasksRequestSchema,
  ListTasksResponse,
  TaskService,
} from '@/gen/flyteidl2/task/task_service_pb'

import { create } from '@bufbuild/protobuf'
import { useInfiniteQuery } from '@tanstack/react-query'
import { useParams } from 'next/navigation'
import { useCallback, useMemo } from 'react'
import { ProjectDomainParams } from '../components/pages/RunDetails/types'
import { useConnectRpcClient } from './useConnectRpc'
import { useOrg } from './useOrg'
import { QuerySort } from './useQueryParamSort'

export const useListTasks = ({
  enabled = true,
  name,
  limit,
  filterFunction = Filter_Function.CONTAINS_CASE_INSENSITIVE,
  sort: { sortBy, sortForQueryKey },
  filters: additionalFilters = [],
}: {
  enabled?: boolean
  name?: string
  limit?: number
  filterFunction?: Filter_Function
  sort: QuerySort
  filters?: Filter[]
}) => {
  const params = useParams<ProjectDomainParams>()
  const client = useConnectRpcClient(TaskService)
  const org = useOrg()

  const projectId = useMemo(
    () =>
      org
        ? create(ProjectIdentifierSchema, {
            domain: params.domain,
            name: params.project,
            organization: org,
          })
        : undefined,
    [params.domain, params.project, org],
  )

  const filters = useMemo(() => {
    const nameFilter = name
      ? [
          {
            function: filterFunction,
            field: 'name',
            values: [name],
          },
        ]
      : []
    return [...nameFilter, ...additionalFilters]
  }, [name, filterFunction, additionalFilters])

  // Memoize the query key to prevent unnecessary refetches
  const queryKey = useMemo(
    () => [
      'tasks',
      params.domain,
      params.project,
      org,
      name,
      sortForQueryKey,
      additionalFilters.length > 0 ? JSON.stringify(additionalFilters) : null,
    ],
    [
      params.domain,
      params.project,
      org,
      name,
      sortForQueryKey,
      additionalFilters,
    ],
  )

  const getTasksPage = useCallback(
    async ({ pageParam: token = '' }: { pageParam?: string }) => {
      const request = create(ListTasksRequestSchema, {
        request: {
          filters,
          sortByFields: [sortBy],
          limit,
          token,
        },
        scopeBy: {
          case: 'projectId',
          value: projectId!,
        },
      })
      const tasks = await client.listTasks(request)
      return tasks
    },
    [filters, sortBy, limit, projectId, client],
  )

  return useInfiniteQuery({
    queryKey,
    queryFn: getTasksPage,
    initialPageParam: '',
    getNextPageParam: (lastPage: ListTasksResponse) => {
      // Only return the next token if the current page has data and a valid token
      if (
        lastPage.tasks &&
        lastPage.tasks.length > 0 &&
        lastPage.token &&
        lastPage.token.trim() !== ''
      ) {
        return lastPage.token
      }
      return undefined
    },
    enabled: enabled && Boolean(projectId),
    // ensure we don't refetch too often
    staleTime: 5 * 60 * 1000, // 5 minutes
    gcTime: 10 * 60 * 1000, // 10 minutes (formerly cacheTime)
  })
}
