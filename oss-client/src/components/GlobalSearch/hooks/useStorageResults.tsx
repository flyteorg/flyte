import {
  useLatestApps,
  useLatestRuns,
  useLatestTasks,
  type StoredApp,
  type StoredRun,
  type StoredTask,
} from '@/hooks/useLatestResources'

export type StorageReturn = {
  apps: StoredApp[]
  runs: StoredRun[]
  tasks: StoredTask[]
}

export const useStorageResults = (searchQuery: string): StorageReturn => {
  const { latestApps } = useLatestApps()
  const { latestRuns } = useLatestRuns()
  const { latestTasks } = useLatestTasks()
  const normalizedQuery = searchQuery.trim().toLowerCase()
  return {
    apps: latestApps.filter((app) =>
      (app.metadata?.id?.name ?? '')
        .toLocaleLowerCase()
        .includes(normalizedQuery),
    ),
    runs: latestRuns,
    tasks: latestTasks.filter((task) =>
      (task.taskId?.name ?? '').toLocaleLowerCase().includes(normalizedQuery),
    ),
  }
}
