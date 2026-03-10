import { TaskNameSchema } from '@/gen/flyteidl2/task/task_definition_pb'
import {
  ListVersionsRequestSchema,
  ListVersionsResponse,
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

export const useListTaskVersions = ({
  taskName,
  limit,
  sort: { sortBy, sortForQueryKey },
  project,
  domain,
}: {
  taskName: string
  limit?: number
  sort: QuerySort
  project?: string
  domain?: string
}) => {
  const params = useParams<ProjectDomainParams>()
  const client = useConnectRpcClient(TaskService)
  const org = useOrg()

  // Use provided project/domain if available, otherwise fall back to params
  const projectToUse = project ?? params.project
  const domainToUse = domain ?? params.domain

  const taskNameObj = useMemo(
    () =>
      org && projectToUse && domainToUse && taskName
        ? create(TaskNameSchema, {
            org,
            project: projectToUse,
            domain: domainToUse,
            name: taskName,
          })
        : undefined,
    [org, projectToUse, domainToUse, taskName],
  )

  // Memoize the query key to prevent unnecessary refetches
  const queryKey = useMemo(
    () => [
      'task-versions',
      domainToUse,
      projectToUse,
      org,
      taskName,
      sortForQueryKey,
    ],
    [domainToUse, projectToUse, org, taskName, sortForQueryKey],
  )

  const getVersionsPage = useCallback(
    async ({ pageParam: token = '' }: { pageParam?: string }) => {
      const request = create(ListVersionsRequestSchema, {
        request: {
          limit,
          token,
          sortByFields: [sortBy],
          filters: [],
          rawFilters: [],
        },
        taskName: taskNameObj!,
      })
      const versions = await client.listVersions(request)
      return versions
    },
    [limit, sortBy, taskNameObj, client],
  )

  return useInfiniteQuery({
    queryKey,
    queryFn: getVersionsPage,
    initialPageParam: '',
    getNextPageParam: (lastPage: ListVersionsResponse) => {
      // Only return the next token if the current page has data and a valid token
      if (
        lastPage.versions &&
        lastPage.versions.length > 0 &&
        lastPage.token &&
        lastPage.token.trim() !== ''
      ) {
        return lastPage.token
      }
      return undefined
    },
    enabled: Boolean(taskNameObj),
    // ensure we don't refetch too often
    staleTime: 5 * 60 * 1000, // 5 minutes
    gcTime: 10 * 60 * 1000, // 10 minutes (formerly cacheTime)
  })
}
