import { isNumber } from 'lodash'
import { toSeconds, parse as parseIso8601Duration } from 'iso8601-duration'
import parseDuration from 'parse-duration'
import dayjs from 'dayjs'
import duration from 'dayjs/plugin/duration'

dayjs.extend(duration)

export function getMillisecondsFromDurationString(
  input: string | number | null | undefined,
): number | undefined {
  if (input === null || input === undefined || !isValidDuration(input))
    return undefined
  if (isNumber(input)) {
    return input
  }
  if (Number(input).toString() === input) {
    return Number(input)
  }
  if (isValidIso8601Duration(input)) {
    return 1000 * toSeconds(parseIso8601Duration(input))
  }
  return parseDuration(input) ?? undefined
}

export function formatDuration(duration: plugin.Duration, separator?: string) {
  const parts = [
    duration.years() ? `${duration.years()}Y` : undefined,
    duration.months() ? `${duration.months}M` : undefined,
    duration.weeks() ? `${duration.weeks()}w` : undefined,
    duration.days() ? `${duration.days()}d` : undefined,
    duration.hours() ? `${duration.hours()}h` : undefined,
    duration.minutes() ? `${duration.minutes()}m` : undefined,
    duration.seconds() || duration.milliseconds()
      ? `${duration.seconds()}${duration.milliseconds() ? `.${duration.milliseconds()}` : ''}s`
      : undefined,
  ]
  return parts.filter((v) => !!v).join(separator ?? '')
}

export function isValidDuration(
  rawDuration: string | number | null | undefined,
) {
  if (rawDuration === null || rawDuration === undefined) {
    return false
  }
  if (isNumber(rawDuration) || Number(rawDuration).toString() === rawDuration) {
    return true
  }
  return (
    isValidReadableDuration(rawDuration) || isValidIso8601Duration(rawDuration)
  )
}

export function isValidIso8601Duration(
  rawDuration: string | number | null | undefined,
) {
  if (rawDuration === null || rawDuration === undefined) {
    return false
  }
  return !!String(rawDuration)
    .trim()
    .match(
      /^P(([0-9]+Y)?([0-9]+M)?([0-9]+W)?([0-9]+D)?(T([0-9]+H)?([0-9]+M)?([0-9]+(\.?[0-9]+)?S)?))?$/,
    )
}

export function isValidReadableDuration(
  rawDuration: string | number | null | undefined,
) {
  if (rawDuration === null || rawDuration === undefined) {
    return false
  }
  return !!String(rawDuration)
    .trim()
    .toLowerCase()
    .match(
      /^([0-9]+y)? *([0-9]+m)? *([0-9]+w)? *([0-9]+d)? *([0-9]+h)? *([0-9]+m)? *([0-9]+(\.[0-9]+)?s)? *([0-9]+ms)?$/,
    )
}

type RawDurationValue = string | number | null | undefined
export function formatIso8601Duration(durationInput: RawDurationValue):RawDurationValue {
  const rawMilliseconds = getMillisecondsFromDurationString(durationInput)

  return rawMilliseconds === undefined
    ? durationInput
    : dayjs.duration(rawMilliseconds, 'milliseconds').toISOString()
}
