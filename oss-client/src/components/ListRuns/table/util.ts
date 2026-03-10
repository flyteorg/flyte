import {
  ActionMetadata,
  Run,
  TaskActionMetadata,
} from '@/gen/flyteidl2/workflow/run_definition_pb'
import { getRunningTime, toDateFormat } from '@/lib/dateUtils'
import { getTriggerTypeString } from '@/lib/triggerUtils'
import { type RunsTableRow } from './types'

function getTaskSpec(
  actionMetadata: ActionMetadata | undefined,
): TaskActionMetadata | null {
  if (actionMetadata?.spec.case === 'task') {
    return actionMetadata.spec.value
  }
  return null
}

// transform api Run data for easy table consumption

export const formatForTable = (r: Run): RunsTableRow => {
  const taskSpec = getTaskSpec(r.action?.metadata)

  // Format timing info
  const relativeStartTime = toDateFormat({
    timestamp: r.action?.status?.startTime,
    relativeThreshold: 'week',
  })
  const relativeEndTime = toDateFormat({
    timestamp: r.action?.status?.endTime,
    relativeThreshold: 'week',
  })

  const durationStr = getRunningTime({
    timestamp: r.action?.status?.startTime,
    endTimestamp: r.action?.status?.endTime,
  })

  return {
    runId: {
      status: r.action?.status?.phase,
      id: r.action?.id?.run?.name,
    },
    name: {
      shortName: taskSpec?.shortName ?? '',
      fullName: taskSpec?.id?.name ?? '',
    },
    startTime: relativeStartTime,
    endTime: relativeEndTime,
    runTime: durationStr,
    user: r.action?.metadata?.executedBy,
    actions: {
      url: `/domain/${r.action?.id?.run?.domain}/project/${r.action?.id?.run?.project}/runs/${r.action?.id?.run?.name}`,
      runId: r.action?.id?.run?.name,
      latestVersion: taskSpec?.id?.version ?? '',
      actionId: r.action?.id?.name,
    },
    original: r,
    trigger: {
      name:
        r.action?.metadata?.triggerName ||
        r.action?.metadata?.triggerId?.name?.name ||
        undefined,
      type: getTriggerTypeString(r.action?.metadata?.triggerType?.type),
    },
    environment: r.action?.metadata?.environmentName || undefined,
  }
}
