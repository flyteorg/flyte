import { App, Condition } from '@/gen/flyteidl2/app/app_definition_pb'
import { Task, TaskDetails } from '@/gen/flyteidl2/task/task_definition_pb'
import {
  Action,
  ActionDetails,
} from '@/gen/flyteidl2/workflow/run_definition_pb'
import { SearchResult } from './types'
import { type StorageReturn } from './hooks/useStorageResults'
import {
  type StoredRun,
  type StoredApp,
  type StoredTask,
} from '@/hooks/useLatestResources'
import { timestampToMillis } from '@/lib/dateUtils'
import { getLastDeployed, getStatus } from '@/lib/appUtils'
import { AppResult, RunResult, TaskResult } from './SearchResultComponents'
import { SearchFilter } from './types'

// Strip spaces, hyphens, and underscores so "hello world" matches "hello_world", etc.
const normalize = (s: string) => s.replace(/[\s\-_]/g, '')

const matchesSearchTerm = (
  value: string | undefined,
  term: string,
): boolean => {
  if (!value) return false
  const v = value.toLocaleLowerCase()
  return v.includes(term) || normalize(v).includes(normalize(term))
}

export const transformRun = (
  action: Action | ActionDetails | StoredRun | undefined,
): SearchResult | null => {
  const runName = action?.id?.run?.name
  const fnName = action?.metadata?.funtionName
  if (!action || !runName || !fnName) return null
  const latestDate = action.status?.endTime ?? action.status?.startTime
  const sortByDate = timestampToMillis(latestDate)
  if (!sortByDate) return null
  const domain = action.id?.run?.domain ?? ''
  const project = action.id?.run?.project ?? ''
  const actionName = action.id?.name ?? ''
  return {
    href: `/domain/${domain}/project/${project}/runs/${runName}`,
    id: `${actionName}-${runName}`,
    displayText: actionName,
    content: (
      <RunResult
        env={action.metadata?.environmentName}
        phase={action.status?.phase}
        runFnName={fnName}
        runId={runName}
        sortByDate={sortByDate}
      />
    ),
    sortByDate,
  }
}

export const transformApp = (
  app: App | StoredApp | undefined,
): SearchResult | null => {
  if (!app || !app.metadata?.id?.name) return null
  // Cast needed because StoredCondition omits protobuf-specific fields (revision, $typeName)
  // but getStatus/getLastDeployed only access deploymentStatus and lastTransitionTime
  const conditions = app.status?.conditions as Condition[] | undefined
  const latestDate = getLastDeployed(conditions)?.lastTransitionTime
  const sortByDate = timestampToMillis(latestDate)
  if (!sortByDate) return null
  return {
    href: `/domain/${app.metadata.id.domain}/project/${app.metadata.id.project}//apps/${app.metadata.id.name}`,
    id: app.metadata.id.name,
    displayText: app.metadata.id.name,
    content: (
      <AppResult
        name={app.metadata.id.name}
        status={getStatus(conditions)}
        sortByDate={sortByDate}
      />
    ),
    sortByDate,
  }
}

export const transformTask = (
  task: TaskDetails | Task | StoredTask | undefined,
): SearchResult | null => {
  if (!task?.taskId?.name) return null
  const sortByDate = timestampToMillis(task.metadata?.deployedAt)
  if (!sortByDate) return null
  return {
    href: `/domain/${task.taskId.domain}/project/${task.taskId.project}/tasks/${task.taskId.name}`,
    id: `${task.taskId.name}-${task.taskId.version}`,
    displayText: task.taskId.name,
    content: (
      <TaskResult
        environmentName={task.metadata?.environmentName}
        name={task.taskId.name}
        shortName={task.metadata?.shortName}
        sortByDate={sortByDate}
        version={task.taskId.version}
      />
    ),
    sortByDate,
  }
}

type FilterOptions = {
  domain: string | undefined
  project: string | undefined
  searchTerm: string
}

const shouldIncludeRun = (
  action: StoredRun,
  { project, domain, searchTerm }: FilterOptions,
) => {
  if (
    action.id?.run?.project !== project ||
    action.id?.run?.domain !== domain
  ) {
    return false
  }
  const term = searchTerm.trim().toLocaleLowerCase()
  return (
    matchesSearchTerm(action.metadata?.funtionName, term) ||
    matchesSearchTerm(action.metadata?.environmentName, term) ||
    matchesSearchTerm(action.id?.run?.name, term)
  )
}

const shouldIncludeApp = (
  app: StoredApp,
  { domain, project, searchTerm }: FilterOptions,
) => {
  if (
    app.metadata?.id?.domain !== domain ||
    app.metadata?.id?.project !== project
  ) {
    return false
  }
  const term = searchTerm.trim().toLocaleLowerCase()
  return matchesSearchTerm(app.metadata?.id?.name, term)
}

const shouldIncludeTask = (
  task: StoredTask,
  { domain, project, searchTerm }: FilterOptions,
) => {
  if (task.taskId?.domain !== domain || task.taskId?.project !== project) {
    return false
  }
  const term = searchTerm.trim().toLocaleLowerCase()
  return (
    matchesSearchTerm(task.metadata?.shortName, term) ||
    matchesSearchTerm(task.taskId?.name, term) ||
    matchesSearchTerm(task.spec?.shortName, term)
  )
}

const markAsLocalStorage = (searchResult: SearchResult): SearchResult => ({
  ...searchResult,
  isLocalStorage: true,
})

export const transformStorageResults = (
  storageResults: StorageReturn,
  options: FilterOptions,
): SearchResult[] => {
  const runs = storageResults.runs
    .filter((run) => shouldIncludeRun(run, options))
    .map(transformRun)
    .filter((r): r is SearchResult => r !== null)
    .map(markAsLocalStorage)
  const apps = storageResults.apps
    .filter((app) => shouldIncludeApp(app, options))
    .map(transformApp)
    .filter((r): r is SearchResult => r !== null)
    .map(markAsLocalStorage)
  const tasks = storageResults.tasks
    .filter((task) => shouldIncludeTask(task, options))
    .map(transformTask)
    .filter((r): r is SearchResult => r !== null)
    .map(markAsLocalStorage)
  return [...runs, ...apps, ...tasks]
    .sort((a, b) => b.sortByDate - a.sortByDate)
    .slice(0, 15)
}

export const getPlaceholderText = (
  searchFilter: SearchFilter,
  project: string | undefined,
) => {
  if (!project) return 'Search'
  switch (searchFilter) {
    case SearchFilter.app: {
      return `Search for apps in ${project}`
    }
    case SearchFilter.run: {
      return `Search for runs in ${project}`
    }
    case SearchFilter.task: {
      return `Search for tasks in ${project}`
    }
    case SearchFilter.none:
    default: {
      return `Search ${project}`
    }
  }
}
