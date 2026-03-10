import { useMemo } from 'react'
import { create } from '@bufbuild/protobuf'
import { useQueries } from '@tanstack/react-query'
import { SortSchema, Sort_Direction } from '@/gen/flyteidl2/common/list_pb'
import { AppService } from '@/gen/app/app_service_pb'
import { getAppIdentifier } from '@/hooks/useApps'
import { useListApps } from '@/hooks/useApps'
import { useListTasks } from '@/hooks/useListTasks'
import { useOrg } from '@/hooks/useOrg'
import { useConnectRpcClient } from '@/hooks/useConnectRpc'
import { RunService } from '@/gen/flyteidl2/workflow/run_service_pb'
import { RunIdentifierSchema } from '@/gen/flyteidl2/common/identifier_pb'
import { GetRequestSchema } from '@/gen/flyteidl2/app/app_payload_pb'
import {
  getSortParamForQueryKey,
  type QuerySort,
} from '@/hooks/useQueryParamSort'
import {
  transformApp,
  transformRun,
  transformTask,
  transformStorageResults,
} from '../util'
import { SearchFilter, type SearchResult } from '../types'
import { useStorageResults } from './useStorageResults'
import { useListRuns } from '@/hooks/useListRuns'
import { useGetRun } from '@/hooks/useGetRun'

type Options = {
  domain: string | undefined
  filter?: SearchFilter
  project: string | undefined
  searchTerm: string
}

const deployedAtSort = create(SortSchema, {
  key: 'created_at',
  direction: Sort_Direction.DESCENDING,
})
const TASK_SEARCH_SORT: QuerySort = {
  sortBy: deployedAtSort,
  sortForQueryKey: getSortParamForQueryKey(deployedAtSort),
}

export const useSearchResults = ({
  domain,
  filter = SearchFilter.none,
  project,
  searchTerm,
}: Options): SearchResult[] => {
  const org = useOrg()
  const storageResults = useStorageResults(searchTerm)

  const isNoFilter = filter === SearchFilter.none
  const canSearch = searchTerm.length > 2

  // Narrow localStorage results to the selected filter type before transforming
  const filteredStorageResults = useMemo(
    () => ({
      runs:
        isNoFilter || filter === SearchFilter.run ? storageResults.runs : [],
      apps:
        isNoFilter || filter === SearchFilter.app ? storageResults.apps : [],
      tasks:
        isNoFilter || filter === SearchFilter.task ? storageResults.tasks : [],
    }),
    [storageResults, filter, isNoFilter],
  )

  // Immediately derive results from localStorage (stale baseline, shown right away)
  const storageSearchResults = useMemo(
    () =>
      transformStorageResults(filteredStorageResults, {
        domain,
        project,
        searchTerm,
      }),
    [filteredStorageResults, domain, project, searchTerm],
  )

  const appClient = useConnectRpcClient(AppService)
  const runClient = useConnectRpcClient(RunService)

  // Run search by task name
  const { data: runsData } = useListRuns({
    domain,
    enabled:
      !!org &&
      !!domain &&
      !!project &&
      canSearch &&
      (isNoFilter || filter === SearchFilter.run),
    name: searchTerm,
    project,
  })

  // Run search by id
  const { data: runById } = useGetRun({
    domain,
    enabled:
      !!org &&
      !!domain &&
      !!project &&
      canSearch &&
      (isNoFilter || filter === SearchFilter.run),
    name: searchTerm,
    project,
  })

  const { data: appsData } = useListApps({
    enabled:
      !!org &&
      !!domain &&
      !!project &&
      canSearch &&
      (isNoFilter || filter === SearchFilter.app),
    org,
    domain,
    projectId: project,
    search: searchTerm || undefined,
  })

  const { data: tasksData } = useListTasks({
    enabled: canSearch && (isNoFilter || filter === SearchFilter.task),
    name: searchTerm || undefined,
    limit: 15,
    sort: TASK_SEARCH_SORT,
  })

  // When there's no search term the list queries above are disabled, so localStorage
  // results won't be overwritten. Revalidate each one individually to get fresh status.
  const appsToRevalidate =
    searchTerm || (!isNoFilter && filter !== SearchFilter.app)
      ? []
      : storageResults.apps.slice(0, 20)
  const runsToRevalidate =
    searchTerm || (!isNoFilter && filter !== SearchFilter.run)
      ? []
      : storageResults.runs
          .filter((r) => r.status?.phase && r.status?.phase < 5)
          .slice(0, 20)

  const appDetailQueries = useQueries({
    queries: appsToRevalidate.map((app) => {
      const id = app.metadata?.id
      return {
        queryKey: ['appDetails', org, id?.project, id?.domain, id?.name],
        queryFn: () =>
          appClient.get(
            create(GetRequestSchema, {
              identifier: {
                case: 'appId',
                value: getAppIdentifier({
                  domain: id?.domain ?? '',
                  name: id?.name ?? '',
                  org: org ?? '',
                  project: id?.project ?? '',
                }),
              },
            }),
          ),
        enabled: Boolean(org && id?.name),
      }
    }),
  })

  const runDetailQueries = useQueries({
    queries: runsToRevalidate.map((run) => {
      const runId = run.id?.run
      return {
        queryKey: ['runData', org, runId?.project, runId?.domain, runId?.name],
        queryFn: () =>
          runClient.getRunDetails({
            runId: create(RunIdentifierSchema, {
              name: runId?.name ?? '',
              domain: runId?.domain ?? domain ?? '',
              org: org ?? '',
              project: runId?.project ?? project ?? '',
            }),
          }),
        enabled: Boolean(runId?.name && org),
      }
    }),
  })

  return useMemo(() => {
    // Seed the map with stale localStorage results
    const resultsById = new Map<string, SearchResult>(
      storageSearchResults.map((r) => [r.id, r]),
    )

    if (isNoFilter || filter === SearchFilter.run) {
      for (const run of runsData?.runs ?? []) {
        const result = transformRun(run.action)
        if (result) resultsById.set(result.id, result)
      }
      if (runById?.details) {
        const result = transformRun(runById.details.action)
        if (result) resultsById.set(result.id, result)
      }
    }

    if (isNoFilter || filter === SearchFilter.app) {
      for (const app of appsData?.apps ?? []) {
        const result = transformApp(app)
        if (result) resultsById.set(result.id, result)
      }
    }

    if (isNoFilter || filter === SearchFilter.task) {
      const apiTasks = tasksData?.pages.flatMap((p) => p.tasks) ?? []
      for (const task of apiTasks) {
        const result = transformTask(task)
        if (result) resultsById.set(result.id, result)
      }
    }

    // Revalidate: overwrite stale localStorage entries with fresh app data.
    // Only entries still flagged isLocalStorage weren't updated by the list queries above.
    for (const { data: freshApp } of appDetailQueries) {
      const result = transformApp(freshApp?.app)
      if (result && resultsById.get(result.id)?.isLocalStorage) {
        resultsById.set(result.id, result)
      }
    }

    // Revalidate: overwrite stale localStorage entries with fresh run data.
    // Only entries still flagged isLocalStorage weren't updated by the list queries above.
    for (const { data: freshRun } of runDetailQueries) {
      const result = transformRun(freshRun?.details?.action)
      if (result && resultsById.get(result.id)?.isLocalStorage) {
        resultsById.set(result.id, result)
      }
    }

    return [...resultsById.values()]
      .sort((a, b) => b.sortByDate - a.sortByDate)
      .slice(0, 20)
  }, [
    storageSearchResults,
    filter,
    isNoFilter,
    appsData,
    runsData,
    runById,
    tasksData,
    appDetailQueries,
    runDetailQueries,
  ])
}
