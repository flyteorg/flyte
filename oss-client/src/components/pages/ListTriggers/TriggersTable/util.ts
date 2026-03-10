import { TriggerName } from '@/gen/flyteidl2/common/identifier_pb'
import { Trigger } from '@/gen/flyteidl2/trigger/trigger_definition_pb'
import { toDateFormat } from '@/lib/dateUtils'
import { highlightMatches } from '@/lib/highlightMatches'
import {
  getNextExecutionTime,
  getTriggerScheduleString,
} from '@/lib/triggerUtils'
import type { TriggerTableRow, TriggerTableRowWithHighlights } from './types'

export const formatForTable = (trigger: Trigger): TriggerTableRow => {
  const updatedAtFormattedDate = toDateFormat({
    timestamp: trigger?.status?.updatedAt,
    relativeThreshold: 'week',
  })

  const triggeredAtFormattedDate = toDateFormat({
    timestamp: trigger?.status?.triggeredAt,
    relativeThreshold: 'week',
  })

  const triggerScheduleString = getTriggerScheduleString(
    trigger?.automationSpec?.automation,
  )

  return {
    active: {
      id: trigger.id?.name?.name || '',
      status: trigger.active,
    },
    nextRun: trigger?.active
      ? getNextExecutionTime(
          trigger?.automationSpec?.automation,
          trigger?.status?.triggeredAt,
          trigger?.status?.updatedAt,
        )
      : undefined,
    name: {
      name: trigger.id?.name?.name,
      schedule: triggerScheduleString,
      nextRun: trigger?.active
        ? getNextExecutionTime(
            trigger?.automationSpec?.automation,
            trigger?.status?.triggeredAt,
            trigger?.status?.updatedAt,
          )
        : undefined,
    },
    task: trigger.id?.name,
    triggered: {
      date: triggeredAtFormattedDate,

      // TODO: add triggered run status
    },
    updated: {
      date: updatedAtFormattedDate,
      updatedBy: trigger?.metadata?.updatedBy,
    },
    createdDate: updatedAtFormattedDate,
    createdUser: trigger.metadata?.deployedBy,
    actions: trigger,
  }
}

export const formatHighlights = (
  rows: TriggerTableRow[],
  searchTerm: string,
): TriggerTableRowWithHighlights[] => {
  return rows.map((row) => ({
    ...row,
    name: {
      name: searchTerm
        ? highlightMatches(row.name?.name ?? '', searchTerm)
        : (row.name?.name ?? ''),
      schedule: row.name?.schedule ?? '',
      nextRun: row.name?.nextRun,
    },
  }))
}

export const getTriggerName = (trigger: Trigger): TriggerName =>
  ({
    domain: trigger.id?.name?.domain || '',
    name: trigger.id?.name?.name || '',
    org: trigger.id?.name?.org || '',
    project: trigger.id?.name?.project || '',
    taskName: trigger.id?.name?.taskName || '',
  }) as TriggerName
