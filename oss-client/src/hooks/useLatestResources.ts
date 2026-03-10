import { useCallback } from 'react'
import stringify from 'safe-stable-stringify'
import useLocalStorage from 'use-local-storage'
import { Action } from '@/gen/flyteidl2/workflow/run_definition_pb'
import {
  App,
  Status_DeploymentStatus,
} from '@/gen/flyteidl2/app/app_definition_pb'
import { TaskDetails } from '@/gen/flyteidl2/task/task_definition_pb'
import { ActionPhase } from '@/gen/flyteidl2/common/phase_pb'
import { Timestamp } from '@/gen/google/protobuf/timestamp_pb'

// Slim types — only the fields consumed by GlobalSearch

export type StoredRun = {
  id?: {
    run?: { name?: string; domain?: string; project?: string }
    name?: string
  }
  metadata?: {
    funtionName?: string
    environmentName?: string
  }
  status?: {
    phase?: ActionPhase
    startTime?: Timestamp
    endTime?: Timestamp
  }
}

export type StoredCondition = {
  deploymentStatus?: Status_DeploymentStatus
  lastTransitionTime?: Timestamp
}

export type StoredApp = {
  metadata?: {
    id?: { name?: string; domain?: string; project?: string }
  }
  status?: {
    conditions?: StoredCondition[]
  }
}

export type StoredTask = {
  taskId?: {
    name?: string
    domain?: string
    project?: string
    version?: string
  }
  metadata?: {
    deployedAt?: Timestamp
    environmentName?: string
    shortName?: string
  }
  spec?: { shortName?: string }
}

// Projection helpers

const toStoredRun = (run: Action): StoredRun => ({
  id: run.id
    ? {
        run: run.id.run
          ? {
              name: run.id.run.name,
              domain: run.id.run.domain,
              project: run.id.run.project,
            }
          : undefined,
        name: run.id.name,
      }
    : undefined,
  metadata: run.metadata
    ? {
        funtionName: run.metadata.funtionName,
        environmentName: run.metadata.environmentName,
      }
    : undefined,
  status: run.status
    ? {
        phase: run.status.phase,
        startTime: run.status.startTime,
        endTime: run.status.endTime,
      }
    : undefined,
})

const toStoredApp = (app: App): StoredApp => ({
  metadata: app.metadata
    ? {
        id: app.metadata.id
          ? {
              name: app.metadata.id.name,
              domain: app.metadata.id.domain,
              project: app.metadata.id.project,
            }
          : undefined,
      }
    : undefined,
  status: app.status
    ? {
        conditions: app.status.conditions?.map((c) => ({
          deploymentStatus: c.deploymentStatus,
          lastTransitionTime: c.lastTransitionTime,
        })),
      }
    : undefined,
})

const toStoredTask = (task: TaskDetails): StoredTask => ({
  taskId: task.taskId
    ? {
        name: task.taskId.name,
        domain: task.taskId.domain,
        project: task.taskId.project,
        version: task.taskId.version,
      }
    : undefined,
  metadata: task.metadata
    ? {
        deployedAt: task.metadata.deployedAt,
        environmentName: task.metadata.environmentName,
        shortName: task.metadata.shortName,
      }
    : undefined,
  spec: task.spec ? { shortName: task.spec.shortName } : undefined,
})

const MAX_LATEST_RUNS = 100

const serializer = <T>(value: T | undefined): string => stringify(value) ?? '[]'

export const useLatestRuns = () => {
  const [latestRuns, setLatestRuns] = useLocalStorage<StoredRun[]>(
    'latestRuns',
    [],
    { serializer },
  )

  const setLatestRun = useCallback(
    (run: Action) => {
      const stored = toStoredRun(run)
      setLatestRuns((prev: StoredRun[] | undefined) => {
        const previous = prev || []
        const runName = stored.id?.run?.name
        const withoutDuplicate = previous.filter(
          (r) => r.id?.run?.name !== runName,
        )
        return [stored, ...withoutDuplicate].slice(0, MAX_LATEST_RUNS)
      })
    },
    [setLatestRuns],
  )

  return { latestRuns, setLatestRun }
}

const MAX_LATEST_APPS = 20

export const useLatestApps = () => {
  const [latestApps, setLatestApps] = useLocalStorage<StoredApp[]>(
    'latestApps',
    [],
    { serializer },
  )

  const setLatestApp = useCallback(
    (app: App) => {
      const stored = toStoredApp(app)
      setLatestApps((prev: StoredApp[] | undefined) => {
        const previous = prev || []
        const appId = stored.metadata?.id
        const withoutDuplicate = previous.filter(
          (a) =>
            !(
              a.metadata?.id?.name === appId?.name &&
              a.metadata?.id?.project === appId?.project &&
              a.metadata?.id?.domain === appId?.domain
            ),
        )
        return [stored, ...withoutDuplicate].slice(0, MAX_LATEST_APPS)
      })
    },
    [setLatestApps],
  )

  return { latestApps, setLatestApp }
}

const MAX_LATEST_TASKS = 50

export const useLatestTasks = () => {
  const [latestTasks, setLatestTasks] = useLocalStorage<StoredTask[]>(
    'latestTasks',
    [],
    { serializer },
  )

  const setLatestTask = useCallback(
    (task: TaskDetails) => {
      const stored = toStoredTask(task)
      setLatestTasks((prev: StoredTask[] | undefined) => {
        const previous = prev || []
        const taskId = stored.taskId
        const withoutDuplicate = previous.filter(
          (t) =>
            !(
              t.taskId?.name === taskId?.name &&
              t.taskId?.project === taskId?.project &&
              t.taskId?.domain === taskId?.domain
            ),
        )
        return [stored, ...withoutDuplicate].slice(0, MAX_LATEST_TASKS)
      })
    },
    [setLatestTasks],
  )

  return { latestTasks, setLatestTask }
}
