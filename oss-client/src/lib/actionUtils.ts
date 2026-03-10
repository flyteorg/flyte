import { ActionPhase } from '@/gen/flyteidl2/common/phase_pb'
import { KeyValuePair } from '@/gen/flyteidl2/core/literals_pb'
import { TaskSpec } from '@/gen/flyteidl2/task/task_definition_pb'
import {
  Action,
  ActionDetails,
} from '@/gen/flyteidl2/workflow/run_definition_pb'
import { isTaskSpec } from './actionSpecUtils'

export const isActionInitializing = (action?: Action | ActionDetails) => {
  return [
    ActionPhase.UNSPECIFIED,
    ActionPhase.QUEUED,
    ActionPhase.INITIALIZING,
    ActionPhase.WAITING_FOR_RESOURCES,
  ].includes(action?.status?.phase || ActionPhase.UNSPECIFIED)
}

export const isActionSucceeded = (action?: Action | ActionDetails) => {
  return (
    (action?.status?.phase || ActionPhase.UNSPECIFIED) == ActionPhase.SUCCEEDED
  )
}

export const isActionTerminal = (action?: Action | ActionDetails) => {
  return [
    ActionPhase.SUCCEEDED,
    ActionPhase.ABORTED,
    ActionPhase.FAILED,
    ActionPhase.TIMED_OUT,
  ].includes(action?.status?.phase || ActionPhase.UNSPECIFIED)
}

export const isActionRunning = (action?: Action | ActionDetails) => {
  return [ActionPhase.RUNNING].includes(
    action?.status?.phase || ActionPhase.UNSPECIFIED,
  )
}

export const isActionTrace = (action?: Action | ActionDetails) => {
  return action?.metadata?.spec?.case === 'trace'
}

export function getTaskType(metadata?: ActionDetails['metadata']) {
  switch (metadata?.spec?.case) {
    case 'task':
      return metadata?.spec?.value?.taskType || ''
    case 'trace':
      return 'trace'
    default:
      return ''
  }
}

export function getCachingStatus(action?: Action | ActionDetails) {
  return action?.status?.cacheStatus
}

export function getTaskRuntimeVersion(action?: ActionDetails) {
  const spec = action?.spec
  if (isTaskSpec(spec)) {
    switch (spec.value?.taskTemplate?.target?.case) {
      case 'container':
        return spec.value.taskTemplate?.metadata?.runtime?.version || ''
      case 'k8sPod':
        return ''
      case 'sql':
      default:
        return ''
    }
  } else {
    return ''
  }
}

export const getActionDisplayString = (
  action?: Action | ActionDetails | null,
) => {
  switch (action?.metadata?.spec?.case) {
    case 'task':
      return action?.metadata?.spec?.value?.shortName || action?.id?.name || ''
    case 'trace':
      return action?.metadata?.spec?.value?.name || action?.id?.name || ''
    case 'condition':
    default:
      return action?.id?.name || ''
  }
}

/**
 * Collect env vars from an action's task spec: container target (key/value) or
 * k8sPod target (podSpec.containers[].env with name/value). Returns KeyValuePair[].
 */
function getActionEnvsFromTaskSpec(resolvedTaskSpec: TaskSpec): KeyValuePair[] {
  const target = resolvedTaskSpec?.taskTemplate?.target
  if (!target?.case) return []
  if (target.case === 'container') {
    return target.value.env ?? []
  }
  if (target.case === 'k8sPod') {
    const podSpec = target.value?.podSpec
    if (!podSpec || typeof podSpec !== 'object') return []
    const containers = podSpec['containers']
    if (!Array.isArray(containers)) return []
    const envMap = new Map<string, string>()
    for (const c of containers) {
      const env =
        c && typeof c === 'object'
          ? (c as { env?: Array<{ name?: string; value?: string }> }).env
          : undefined
      if (!Array.isArray(env)) continue
      for (const e of env) {
        const name = e?.name
        const value = e?.value
        if (typeof name === 'string') {
          envMap.set(name, typeof value === 'string' ? value : '')
        }
      }
    }
    return Array.from(envMap.entries()).map(
      ([key, value]) =>
        ({
          key,
          value,
        }) as KeyValuePair,
    )
  }
  return []
}

/**
 * Merge run envs with task envs (task wins). Optionally apply overrides last.
 * Order: runEnvs → task envs from taskSpec → overrides. Later layers override earlier for same key.
 */
export function getMergedRunAndTaskEnvs(
  runEnvs: KeyValuePair[],
  taskSpec: TaskSpec,
  overrides?: KeyValuePair[],
): KeyValuePair[] {
  const taskEnvs = getActionEnvsFromTaskSpec(taskSpec)
  const envMap = new Map<string, string>()
  for (const e of runEnvs) {
    envMap.set(e.key, e.value)
  }
  for (const e of taskEnvs) {
    envMap.set(e.key, e.value)
  }
  if (overrides?.length) {
    for (const e of overrides) {
      envMap.set(e.key, e.value)
    }
  }
  return Array.from(envMap.entries()).map(
    ([key, value]) =>
      ({
        key,
        value,
      }) as KeyValuePair,
  )
}
