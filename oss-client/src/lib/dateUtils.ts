import { Duration } from '@/gen/google/protobuf/duration_pb'
import { Timestamp } from '@/gen/google/protobuf/timestamp_pb'
import { format } from 'date-fns'
import dayjs from 'dayjs'
import duration from 'dayjs/plugin/duration'
import isToday from 'dayjs/plugin/isToday'
import relativeTime from 'dayjs/plugin/relativeTime'

dayjs.extend(relativeTime)
dayjs.extend(isToday)
dayjs.extend(duration)

const FORMAT = 'MMM D, YYYY h:mm A'
const FORMAT_WITH_SECONDS = 'MMM D, YYYY h:mm:ss A'
const FORMAT_NO_DATE = 'h:mm:ss A'

export const getLogDateString = (timestamp?: Timestamp) => {
  return timestamp
    ? format(
        new Date(
          Number(timestamp.seconds) * 1000 +
            Math.round(Number(timestamp.nanos) / 1e6),
        ),
        'MMM dd HH:mm:ss.SSS',
      )
    : ''
}

/**
 * Converts a protobuf Duration to a human-readable string format.
 * Returns a formatted string showing only non-zero time units (e.g., "1h 2m 30s", "5d 3h", "45s").
 * Returns '-' if no duration is provided, or '0s' if the duration is zero.
 */
export const getDurationStringFromProto = (duration?: Duration) => {
  if (!duration) return '-'
  const totalMs =
    Number(duration.seconds) * 1000 + Math.round(duration.nanos / 1e6)

  const preparedDuration = dayjs.duration(totalMs)
  const years = preparedDuration.years()
  const days = preparedDuration.days()
  const hours = preparedDuration.hours()
  const minutes = preparedDuration.minutes()
  const seconds = preparedDuration.seconds()

  const units = [
    { key: 'y', value: years },
    { key: 'd', value: days },
    { key: 'h', value: hours },
    { key: 'm', value: minutes },
    { key: 's', value: seconds },
  ]

  // Get all units that have a value > 0
  const unitsWithValue = units.filter(({ value }) => value > 0)

  // If no units have a value, return '0s'
  if (unitsWithValue.length === 0) {
    return '0s'
  }

  // Render all sections that have a value
  const result = unitsWithValue
    .map(({ key, value }) => `${value}${key}`)
    .join(' ')

  return result
}

// convert timestamp to dayjs object
export const getDateObject = (timestamp: Timestamp | string | number) => {
  if (typeof timestamp === 'string' || typeof timestamp === 'number')
    return dayjs(timestamp)

  const seconds = Number(timestamp.seconds || 0)
  const nanos = Number(timestamp.nanos || 0)

  const millis = seconds * 1000 + nanos / 1_000_000

  return dayjs(millis)
}

export const getFormatDate = (
  timestamp: Timestamp | string | number,
  includeSeconds?: boolean,
) => {
  const format = includeSeconds ? FORMAT_WITH_SECONDS : FORMAT
  return getDateObject(timestamp).format(format)
}

function abbreviateRelative(str: string): string {
  return str.replace(/\bhours?\b/g, 'hr').replace(/\bminutes?\b/g, 'min')
}

// how long to show relative date before showing calendar date
// 'always' show relative date, never show calendar date
// 'today' show relative date if comparison date is today, etc
type RelativeThreshold = 'always' | 'today' | 'week' | 'month' | 'year'

export const getRelativeDate = (
  timestamp: Timestamp | string | number,
  relativeThreshold?: RelativeThreshold,
) => {
  const timestampObject = getDateObject(timestamp)
  const now = dayjs()

  if (relativeThreshold === 'always') {
    return abbreviateRelative(timestampObject.fromNow())
  }
  if (relativeThreshold === 'today') {
    return timestampObject.isToday()
      ? abbreviateRelative(timestampObject.fromNow())
      : getFormatDate(timestamp)
  }
  if (Math.abs(timestampObject.diff(now, relativeThreshold)) < 1) {
    return abbreviateRelative(timestampObject.fromNow())
  }

  // return standard date format
  return getFormatDate(timestamp)
}

/**
 * @param timestamp Timestamp | string;
 * @param relativeThreshold when this argument is passed, return value will be
 * relative if timestamp is within the threshold. if 'today' is passed, then return
 * value will be relative for a timestamp that is from today but not if older.
 * @returns formatted date
 */

export const toDateFormat = ({
  includeSeconds = true,
  relativeThreshold,
  timestamp,
}: {
  includeSeconds?: boolean
  relativeThreshold?: RelativeThreshold
  timestamp?: Timestamp | string | number
}): string => {
  if (!timestamp) return '-'
  if (!relativeThreshold) {
    return getFormatDate(timestamp, includeSeconds)
  }
  return getRelativeDate(timestamp, relativeThreshold)
}

type UnitKey = 'y' | 'd' | 'h' | 'm' | 's'

/**
 * @param timestamp Timestamp | Protobuf.ITimestamp | string;
 * @param maxUnits when this argument is passed, return value will include
 * up to the number of units passed. units will only be pass if more than 0
 * e.g maxUnits: 1 --> 1h, maxUnits: 4 --> 365d 23h 42m 10s
 * @returns running time in form such as '1h 2m', '10m 59s', '500d 23h', '59s'
 */

export const getRunningTime = ({
  endTimestamp,
  excludeUnits,
  maxUnits = 2,
  showSubSecondPrecision,
  timestamp,
}: {
  endTimestamp?: Timestamp | number | undefined
  excludeUnits?: UnitKey[] | UnitKey
  maxUnits?: number
  timestamp?: Timestamp | number | string
  showSubSecondPrecision?: boolean
}) => {
  if (!timestamp) return '-'
  const dayjsObj = getDateObject(timestamp)
  const end = endTimestamp ? getDateObject(endTimestamp) : dayjs()

  const totalMs = Math.abs(end.diff(dayjsObj))
  const totalSeconds = totalMs / 1000

  if (showSubSecondPrecision && totalSeconds < 1) {
    return `${totalSeconds.toFixed(2)}s`
  }

  const years = Math.abs(dayjsObj.diff(end, 'year'))
  const days = Math.abs(dayjsObj.diff(end, 'day')) % 365
  const hours = Math.abs(dayjsObj.diff(end, 'hours')) % 24
  const minutes = Math.abs(dayjsObj.diff(end, 'minute')) % 60
  const seconds = Math.abs(dayjsObj.diff(end, 'second')) % 60
  const accum = [
    { key: 'y', value: years },
    { key: 'd', value: days },
    { key: 'h', value: hours },
    { key: 'm', value: minutes },
    { key: 's', value: seconds },
  ]
    .filter(({ key }) => !excludeUnits?.includes(key as UnitKey))
    .map(({ key, value }) => {
      if (!value) return null
      if (key === 's' && showSubSecondPrecision && value % 1 !== 0) {
        return `${value.toFixed(2)}${key}`
      }
      return `${value}${key}`
    })
  const result = accum.filter(Boolean).slice(0, maxUnits).join(' ')
  return result || (showSubSecondPrecision ? '0.00s' : '0s')
}

export const toLocalTime = (ts: number) => {
  return dayjs(ts).format(FORMAT_NO_DATE)
}

/**
 * Formats a date with conditional year display.
 * Shows year only when it's not the current year.
 * Format: "MMM D, h:mm A" for current year, "MMM D, YYYY, h:mm A" for other years.
 * Returns '-' if no timestamp is provided.
 */
export const formatDateWithConditionalYear = (
  timestamp?: Timestamp | string | number,
): string => {
  if (!timestamp) return '-'
  const dateObj = getDateObject(timestamp)
  const currentYear = new Date().getFullYear()
  const dateYear = dateObj.year()
  const format =
    dateYear === currentYear ? 'MMM D, h:mm A' : 'MMM D, YYYY, h:mm A'
  return dateObj.format(format)
}

/**
 * Formats a timestamp for compact display in search results:
 * - Today → relative time ("2 hr ago", "5 min ago")
 * - This year → locale-aware month + day ("Mar 15" or "15 Mar" depending on locale)
 * - Older → full date without seconds ("Mar 15, 2023, 2:30 PM")
 */
export const formatSearchResultDate = (
  timestamp: number | undefined,
): string => {
  if (!timestamp) return '-'
  const dateObj = dayjs(timestamp)
  const now = dayjs()

  if (dateObj.isToday()) {
    return abbreviateRelative(dateObj.fromNow())
  }

  if (dateObj.year() === now.year()) {
    return new Intl.DateTimeFormat(undefined, {
      month: 'short',
      day: 'numeric',
    }).format(timestamp)
  }

  return toDateFormat({ timestamp, includeSeconds: false })
}

export function timestampToMillis(
  ts: Timestamp | undefined,
): number | undefined {
  if (!ts) return undefined
  return Number(ts.seconds) * 1000 + Math.floor(ts.nanos / 1_000_000)
}
