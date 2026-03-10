import { describe, it, expect, vi } from 'vitest'
import {
  transformRun,
  transformApp,
  transformTask,
  transformStorageResults,
} from './util'
import type { Action } from '@/gen/flyteidl2/workflow/run_definition_pb'
import type { App } from '@/gen/flyteidl2/app/app_definition_pb'
import type { TaskDetails } from '@/gen/flyteidl2/task/task_definition_pb'
import type { StorageReturn } from './hooks/useStorageResults'

vi.mock('./SearchResultComponents', () => ({
  RunResult: () => null,
  AppResult: () => null,
  TaskResult: () => null,
}))

vi.mock('@/lib/appUtils', () => ({
  getLastDeployed: (conditions: unknown[]) => conditions?.[0],
  getStatus: () => undefined,
}))

vi.mock('@/lib/dateUtils', () => ({
  timestampToMillis: (ts: { seconds?: bigint } | undefined) =>
    ts?.seconds ? Number(ts.seconds) * 1000 : null,
}))

const makeRun = (overrides: Partial<Action> = {}): Action =>
  ({
    id: {
      name: 'fn-name',
      run: { name: 'run-1', domain: 'development', project: 'my-project' },
    },
    metadata: { funtionName: 'my_function', environmentName: 'dev' },
    status: { startTime: { seconds: BigInt(1700000000) }, phase: undefined },
    ...overrides,
  }) as unknown as Action

const makeApp = (overrides: Partial<App> = {}): App =>
  ({
    metadata: {
      id: { name: 'my-app', domain: 'development', project: 'my-project' },
    },
    status: {
      conditions: [{ lastTransitionTime: { seconds: BigInt(1700000000) } }],
    },
    ...overrides,
  }) as unknown as App

const makeTask = (overrides: Partial<TaskDetails> = {}): TaskDetails =>
  ({
    taskId: {
      name: 'my-task',
      domain: 'development',
      project: 'my-project',
      version: 'v1',
    },
    metadata: {
      shortName: 'My Task',
      environmentName: 'dev',
      deployedAt: { seconds: BigInt(1700000000) },
    },
    spec: { shortName: 'My Task' },
    ...overrides,
  }) as unknown as TaskDetails

describe('transformRun', () => {
  it('returns null for undefined input', () => {
    expect(transformRun(undefined)).toBeNull()
  })

  it('returns null when run name is missing', () => {
    expect(
      transformRun(makeRun({ id: { name: 'fn', run: undefined } } as never)),
    ).toBeNull()
  })

  it('returns null when funtionName is missing', () => {
    expect(
      transformRun(
        makeRun({
          metadata: { funtionName: '', environmentName: 'dev' },
        } as never),
      ),
    ).toBeNull()
  })

  it('returns null when no sortable date exists', () => {
    expect(transformRun(makeRun({ status: undefined }))).toBeNull()
  })

  it('returns a valid SearchResult for a complete run', () => {
    const result = transformRun(makeRun())
    expect(result).not.toBeNull()
    expect(result?.href).toBe(
      '/domain/development/project/my-project/runs/run-1',
    )
    expect(result?.id).toBe('fn-name-run-1')
    expect(result?.displayText).toBe('fn-name')
    expect(result?.sortByDate).toBe(1700000000000)
  })

  it('uses endTime over startTime for sortByDate', () => {
    const run = makeRun({
      status: {
        startTime: { seconds: BigInt(1000) },
        endTime: { seconds: BigInt(2000) },
      } as never,
    })
    expect(transformRun(run)?.sortByDate).toBe(2000000)
  })

  it('falls back to empty strings for undefined domain/project in href', () => {
    const run = makeRun({
      id: {
        name: 'fn',
        run: { name: 'run-1', domain: undefined, project: undefined },
      },
    } as never)
    expect(transformRun(run)?.href).toBe('/domain//project//runs/run-1')
  })
})

describe('transformApp', () => {
  it('returns null for undefined input', () => {
    expect(transformApp(undefined)).toBeNull()
  })

  it('returns null when app name is missing', () => {
    expect(
      transformApp(
        makeApp({
          metadata: { id: { name: '', domain: 'development', project: 'p' } },
        } as never),
      ),
    ).toBeNull()
  })

  it('returns null when no sortable date exists', () => {
    expect(
      transformApp(makeApp({ status: { conditions: [] } } as never)),
    ).toBeNull()
  })

  it('returns a valid SearchResult for a complete app', () => {
    const result = transformApp(makeApp())
    expect(result).not.toBeNull()
    expect(result?.href).toContain('/apps/my-app')
    expect(result?.id).toBe('my-app')
    expect(result?.displayText).toBe('my-app')
    expect(result?.sortByDate).toBe(1700000000000)
  })
})

describe('transformTask', () => {
  it('returns null for undefined input', () => {
    expect(transformTask(undefined)).toBeNull()
  })

  it('returns null when task name is missing', () => {
    expect(
      transformTask(
        makeTask({
          taskId: {
            name: '',
            domain: 'development',
            project: 'p',
            version: 'v1',
          },
        } as never),
      ),
    ).toBeNull()
  })

  it('returns null when deployedAt is missing', () => {
    expect(
      transformTask(
        makeTask({
          metadata: {
            shortName: 'T',
            environmentName: 'dev',
            deployedAt: undefined,
          },
        } as never),
      ),
    ).toBeNull()
  })

  it('returns a valid SearchResult for a complete task', () => {
    const result = transformTask(makeTask())
    expect(result).not.toBeNull()
    expect(result?.href).toBe(
      '/domain/development/project/my-project/tasks/my-task',
    )
    expect(result?.id).toBe('my-task-v1')
    expect(result?.displayText).toBe('my-task')
    expect(result?.sortByDate).toBe(1700000000000)
  })
})

describe('transformStorageResults', () => {
  const opts = { domain: 'development', project: 'my-project', searchTerm: '' }

  const storageResults: StorageReturn = {
    runs: [makeRun()],
    apps: [makeApp()],
    tasks: [makeTask()],
  }

  it('returns results sorted by date descending', () => {
    const results = transformStorageResults(storageResults, opts)
    expect(results.length).toBeGreaterThan(0)
    for (let i = 1; i < results.length; i++) {
      expect(results[i - 1].sortByDate).toBeGreaterThanOrEqual(
        results[i].sortByDate,
      )
    }
  })

  it('marks all results as isLocalStorage', () => {
    const results = transformStorageResults(storageResults, opts)
    expect(results.every((r) => r.isLocalStorage)).toBe(true)
  })

  it('filters out runs from a different project', () => {
    const run = makeRun({
      id: {
        name: 'fn',
        run: { name: 'r', domain: 'development', project: 'other-project' },
      },
    } as never)
    const results = transformStorageResults(
      { runs: [run], apps: [], tasks: [] },
      opts,
    )
    expect(results).toHaveLength(0)
  })

  it('filters out apps from a different domain', () => {
    const app = makeApp({
      metadata: {
        id: { name: 'app', domain: 'staging', project: 'my-project' },
      },
    } as never)
    const results = transformStorageResults(
      { runs: [], apps: [app], tasks: [] },
      opts,
    )
    expect(results).toHaveLength(0)
  })

  it('filters by searchTerm against run function name', () => {
    const match = transformStorageResults(storageResults, {
      ...opts,
      searchTerm: 'my_func',
    })
    const noMatch = transformStorageResults(storageResults, {
      ...opts,
      searchTerm: 'zzznomatch',
    })
    expect(match.some((r) => r.id.includes('run-1'))).toBe(true)
    expect(noMatch.some((r) => r.id.includes('run-1'))).toBe(false)
  })

  it('fuzzy-matches runs when separators differ (space vs underscore)', () => {
    // "my function" should match "my_function"
    const result = transformStorageResults(storageResults, {
      ...opts,
      searchTerm: 'my function',
    })
    expect(result.some((r) => r.id.includes('run-1'))).toBe(true)
  })

  it('fuzzy-matches apps when separators differ (underscore vs hyphen)', () => {
    // "my_app" should match "my-app"
    const result = transformStorageResults(storageResults, {
      ...opts,
      searchTerm: 'my_app',
    })
    expect(result.some((r) => r.id === 'my-app')).toBe(true)
  })

  it('fuzzy-matches tasks when separators differ (space vs hyphen)', () => {
    // "my task" already matched via shortName; also confirm "my-task" matches taskId.name
    const result = transformStorageResults(storageResults, {
      ...opts,
      searchTerm: 'my-task',
    })
    expect(result.some((r) => r.id.includes('my-task'))).toBe(true)
  })

  it('filters by searchTerm against app name', () => {
    const match = transformStorageResults(storageResults, {
      ...opts,
      searchTerm: 'my-app',
    })
    const noMatch = transformStorageResults(storageResults, {
      ...opts,
      searchTerm: 'zzznomatch',
    })
    expect(match.some((r) => r.id === 'my-app')).toBe(true)
    expect(noMatch.some((r) => r.id === 'my-app')).toBe(false)
  })

  it('filters by searchTerm against task short name', () => {
    const match = transformStorageResults(storageResults, {
      ...opts,
      searchTerm: 'my task',
    })
    const noMatch = transformStorageResults(storageResults, {
      ...opts,
      searchTerm: 'zzznomatch',
    })
    expect(match.some((r) => r.id.includes('my-task'))).toBe(true)
    expect(noMatch.some((r) => r.id.includes('my-task'))).toBe(false)
  })

  it('does not throw when optional string fields are undefined', () => {
    const runWithUndefinedFields = makeRun({
      metadata: { funtionName: 'fn', environmentName: undefined } as never,
    })
    expect(() =>
      transformStorageResults(
        { runs: [runWithUndefinedFields], apps: [], tasks: [] },
        opts,
      ),
    ).not.toThrow()
  })

  it('caps results at 15', () => {
    const manyRuns = Array.from({ length: 20 }, (_, i) =>
      makeRun({
        id: {
          name: `fn-${i}`,
          run: {
            name: `run-${i}`,
            domain: 'development',
            project: 'my-project',
          },
        },
        metadata: { funtionName: `fn_${i}`, environmentName: 'dev' },
        status: { startTime: { seconds: BigInt(1700000000 + i) } },
      } as never),
    )
    const results = transformStorageResults(
      { runs: manyRuns, apps: [], tasks: [] },
      opts,
    )
    expect(results.length).toBeLessThanOrEqual(15)
  })
})
