import {
  FixedRateUnit,
  TriggerAutomationSpec,
  TriggerAutomationSpecType,
} from '@/gen/flyteidl2/task/common_pb'
import { TaskTriggersSummary } from '@/gen/flyteidl2/task/task_definition_pb'
import { TriggerRevisionAction } from '@/gen/flyteidl2/trigger/trigger_definition_pb'
import { Timestamp } from '@/gen/google/protobuf/timestamp_pb'
import { CronExpressionParser } from 'cron-parser'
import cronstrue from 'cronstrue'
import dayjs from 'dayjs'
import { getDateObject, getFormatDate } from './dateUtils'

/**
 * Gets the abbreviated timezone name (e.g., "PST", "PDT", "CST", "CDT")
 * @param timezone IANA timezone string (e.g., "America/Los_Angeles")
 * @param date Optional date to determine DST (defaults to current date)
 * @returns Timezone abbreviation or empty string
 */
const getTimezoneAbbr = (timezone: string, date: Date = new Date()): string => {
  return (
    new Intl.DateTimeFormat('en-US', {
      timeZone: timezone,
      timeZoneName: 'short',
    })
      .formatToParts(date)
      .find((part) => part.type === 'timeZoneName')?.value || ''
  )
}

export const getTriggerScheduleString = (
  trigger?: TriggerAutomationSpec['automation'],
) => {
  if (!trigger?.case) {
    return '-'
  }

  switch (trigger.value?.expression.case) {
    case 'cron': {
      const cronExpr = trigger.value?.expression.value?.expression ?? '-'
      if (cronExpr === '-') return cronExpr
      const tz = trigger.value?.expression.value?.timezone || 'UTC'

      try {
        // Validate cron expression (will throw if invalid)
        CronExpressionParser.parse(cronExpr, {
          currentDate: dayjs().toDate(),
          tz: tz,
        })

        // Get abbreviated timezone of the trigger's timezone (e.g., "PST", "PDT" for America/Los_Angeles)
        const triggerTzAbbr = getTimezoneAbbr(tz)

        // Get base description from cronstrue
        const baseDescription = cronstrue.toString(cronExpr)

        // Append trigger's timezone abbreviation at the end
        return triggerTzAbbr
          ? `${baseDescription} (${triggerTzAbbr})`
          : baseDescription
      } catch {
        if (tz != 'UTC') {
          return `CRON_TZ=${tz} ${cronExpr}`
        }
        return cronExpr // Return raw expression if invalid
      }
    }
    case 'rate': {
      const rate = trigger.value?.expression.value
      if (!rate) return '-'

      const unit = rate.unit
      const value = rate.value || 1

      const unitText = getRateUnitText(unit, value)
      return `Every ${value} ${unitText}`
    }
    default:
      return '-'
  }
}

export const getNextExecutionTime = (
  trigger?: TriggerAutomationSpec['automation'],
  lastTriggeredAt?: Timestamp,
  lastUpdatedAt?: Timestamp,
): string | null => {
  if (!trigger?.case) return null

  switch (trigger.value?.expression.case) {
    case 'cron': {
      const cronExpr = trigger.value?.expression.value.expression
      if (!cronExpr) return null
      const tz = trigger.value?.expression.value?.timezone || 'UTC'

      try {
        const interval = CronExpressionParser.parse(cronExpr, {
          // set the reference time for the cron parser to calculate the next run time
          currentDate: dayjs().toDate(),
          tz: tz,
        })
        const nextRun = interval.next().toDate()
        return getFormatDate(nextRun.getTime())
      } catch {
        return null
      }
    }
    case 'rate': {
      const rate = trigger.value?.expression.value
      if (!rate) return null

      const unit = rate.unit
      const value = rate.value || 1

      // Convert raw timestamps to dayjs objects and calculate next execution time
      const lastTriggeredDayjs = lastTriggeredAt
        ? getDateObject(lastTriggeredAt)
        : undefined
      const lastUpdatedDayjs = lastUpdatedAt
        ? getDateObject(lastUpdatedAt)
        : undefined
      const baseTime = getMaxDate(lastTriggeredDayjs, lastUpdatedDayjs)

      const nextRun = addRateToDate(baseTime, value, unit)
      return getFormatDate(nextRun.getTime())
    }
    default:
      return null
  }
}

const getRateUnitText = (unit: FixedRateUnit, value: number): string => {
  const unitMap: Record<FixedRateUnit, string> = {
    [FixedRateUnit.MINUTE]: 'minute',
    [FixedRateUnit.HOUR]: 'hour',
    [FixedRateUnit.DAY]: 'day',
    [FixedRateUnit.UNSPECIFIED]: 'minute',
  }

  const base = unitMap[unit]
  return value === 1 ? base : `${base}s`
}

const addRateToDate = (
  date: Date,
  value: number,
  unit: FixedRateUnit,
): Date => {
  const result = new Date(date)

  switch (unit) {
    case FixedRateUnit.MINUTE:
      result.setMinutes(result.getMinutes() + value)
      break
    case FixedRateUnit.HOUR:
      result.setHours(result.getHours() + value)
      break
    case FixedRateUnit.DAY:
      result.setDate(result.getDate() + value)
      break
    case FixedRateUnit.UNSPECIFIED:
    default:
      result.setMinutes(result.getMinutes() + value)
      break
  }

  return result
}

const getMaxDate = (
  lastTriggeredAt?: dayjs.Dayjs,
  lastUpdatedAt?: dayjs.Dayjs,
): Date => {
  if (!lastTriggeredAt && !lastUpdatedAt) {
    return new Date()
  }

  if (!lastTriggeredAt) {
    return lastUpdatedAt!.toDate()
  }

  if (!lastUpdatedAt) {
    return lastTriggeredAt.toDate()
  }

  // Return the more recent date
  return lastTriggeredAt.isAfter(lastUpdatedAt)
    ? lastTriggeredAt.toDate()
    : lastUpdatedAt.toDate()
}

export const getActionDisplayName = (action: TriggerRevisionAction): string => {
  switch (action) {
    case TriggerRevisionAction.DEPLOY:
      return 'deployed trigger'
    case TriggerRevisionAction.ACTIVATE:
      return 'activated trigger'
    case TriggerRevisionAction.DEACTIVATE:
      return 'deactivated trigger'
    case TriggerRevisionAction.DELETE:
      return 'deleted trigger'
    default:
      return 'modified trigger'
  }
}

export const getActiveTaskTriggersCount = (
  triggersSummary?: TaskTriggersSummary | undefined,
): number | undefined => {
  switch (triggersSummary?.summary.case) {
    case 'stats':
      return triggersSummary?.summary.value?.active || 0
    case 'details':
      return triggersSummary?.summary.value?.active ? 1 : 0

    default:
      return undefined
  }
}

export function getTriggerTypeString(
  triggerType: TriggerAutomationSpecType | undefined,
): string | undefined {
  if (!triggerType) return undefined
  switch (triggerType) {
    case TriggerAutomationSpecType.TYPE_SCHEDULE:
      return 'Scheduled'
    case TriggerAutomationSpecType.TYPE_NONE:
      return 'None'
    default:
      return undefined
  }
}
