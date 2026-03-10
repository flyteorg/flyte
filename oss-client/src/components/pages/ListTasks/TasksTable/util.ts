import { TriggerAutomationSpecType } from '@/gen/flyteidl2/task/common_pb'
import { Task } from '@/gen/flyteidl2/task/task_definition_pb'
import { toDateFormat } from '@/lib/dateUtils'
import { highlightMatches } from '@/lib/highlightMatches'
import {
  getActiveTaskTriggersCount,
  getTriggerScheduleString,
} from '@/lib/triggerUtils'
import type { TaskTableRow, TaskTableRowWithHighlights } from './types'

export const formatForTable = (task: Task): TaskTableRow => {
  const formatDate = toDateFormat({ timestamp: task.metadata?.deployedAt })

  const activeCount =
    getActiveTaskTriggersCount(task.metadata?.triggersSummary) || 0

  let triggerColumn
  switch (task.metadata?.triggersSummary?.summary.case) {
    case 'details':
      const triggerDetails = task.metadata?.triggersSummary?.summary
      const isSchedule =
        triggerDetails?.value.automationSpec?.type ===
        TriggerAutomationSpecType.TYPE_SCHEDULE
      const triggerScheduleString = getTriggerScheduleString(
        triggerDetails?.value?.automationSpec?.automation,
      )

      triggerColumn = {
        active: !!activeCount,
        title: isSchedule ? 'Schedule' : '',
        subtitle: isSchedule ? triggerScheduleString : '',
      }
      break
    case 'stats':
      const triggersStats = task.metadata?.triggersSummary?.summary
      const totalCount = triggersStats?.value?.total || 0

      triggerColumn = {
        active: !!activeCount,
        title:
          totalCount > 0
            ? `${totalCount} ${triggersStats?.value?.total === 1 ? 'Trigger' : 'Triggers'}`
            : '',
        subtitle: activeCount > 0 ? `${activeCount} active` : '',
      }
      break
    default:
      triggerColumn = undefined
  }

  // Extract last run information from taskSummary
  const latestRun = task?.taskSummary?.latestRun
  const lastRunTime = latestRun?.runTime
    ? toDateFormat({
        timestamp: latestRun.runTime,
        relativeThreshold: 'week',
      })
    : undefined

  const lastRun = latestRun
    ? {
        phase: latestRun.phase,
        time: lastRunTime ?? '-',
        runId: latestRun.runId?.name,
      }
    : undefined

  return {
    name: {
      shortName: task.metadata?.shortName,
      fullName: task.taskId?.name,
      shortDescription: task.metadata?.shortDescription,
    },
    environment: task.metadata?.environmentName,
    createdAt: {
      version: task.taskId?.version,
      date: formatDate,
    },
    createdUser: task.metadata?.deployedBy,
    trigger: triggerColumn,
    lastRun,
    copyAction: task,
    original: task,
  }
}

export const formatHighlights = (
  rows: TaskTableRow[],
  searchTerm?: string,
): TaskTableRowWithHighlights[] => {
  if (!searchTerm) {
    //nothing to highlight
    return rows
  }
  return rows.map((row) => ({
    ...row,
    name: {
      shortName: highlightMatches(row.name.shortName ?? '', searchTerm),
      fullName: highlightMatches(row.name.fullName ?? '', searchTerm),
      shortDescription: highlightMatches(
        row.name.shortDescription ?? '',
        searchTerm,
      ),
    },
  }))
}
