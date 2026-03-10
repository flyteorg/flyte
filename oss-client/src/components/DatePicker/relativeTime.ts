import {
  addMinutes,
  subMinutes,
  subHours,
  subDays,
  subMonths,
  startOfToday,
  endOfToday,
  startOfYesterday,
  parseISO,
  isValid,
} from 'date-fns'

/**
 * Relative time format:
 * - `now` - current time
 * - `now-5m` - 5 minutes ago
 * - `now-1h` - 1 hour ago
 * - `now-7d` - 7 days ago
 * - `now-12M` - 12 months ago
 * - `now-30d+15m` - 30 days ago plus 15 minutes (compound format)
 * - `sod` - start of day (today)
 * - `eod` - end of day (today)
 * - `sod-1d` - start of yesterday
 * - `eod-1d` - end of yesterday
 * - ISO timestamp as fallback
 */

// Simple format: now-5m, now-1h, now-7d, now-12M
const RELATIVE_TIME_REGEX = /^now(-(\d+)([mhdM]))?$/
// Compound format: now-30d+15m (subtract days, then add minutes)
const COMPOUND_TIME_REGEX = /^now-(\d+)d\+(\d+)m$/
const SOD_REGEX = /^sod(-(\d+)d)?$/
const EOD_REGEX = /^eod(-(\d+)d)?$/

// takes relative time strings and parses to date ranges for query.
// note that this means that strings like "now-5m" will return different date ranges
// based on when the function is called
export function parseRelativeTime(str: string): Date | undefined {
  if (!str) return undefined

  // Check for "now" format
  if (str === 'now') {
    return new Date()
  }

  // Check for compound format first: now-30d+15m
  const compoundMatch = str.match(COMPOUND_TIME_REGEX)
  if (compoundMatch) {
    const daysBack = parseInt(compoundMatch[1], 10)
    const minutesForward = parseInt(compoundMatch[2], 10)
    const now = new Date()
    return addMinutes(subDays(now, daysBack), minutesForward)
  }

  const nowMatch = str.match(RELATIVE_TIME_REGEX)
  if (nowMatch) {
    const amount = nowMatch[2] ? parseInt(nowMatch[2], 10) : 0
    const unit = nowMatch[3]
    const now = new Date()

    if (amount === 0) return now

    switch (unit) {
      case 'm':
        return subMinutes(now, amount)
      case 'h':
        return subHours(now, amount)
      case 'd':
        return subDays(now, amount)
      case 'M':
        return subMonths(now, amount)
      default:
        return undefined
    }
  }

  // Check for "sod" (start of day) format
  const sodMatch = str.match(SOD_REGEX)
  if (sodMatch) {
    const daysBack = sodMatch[2] ? parseInt(sodMatch[2], 10) : 0
    if (daysBack === 0) return startOfToday()
    if (daysBack === 1) return startOfYesterday()
    return subDays(startOfToday(), daysBack)
  }

  // Check for "eod" (end of day) format
  const eodMatch = str.match(EOD_REGEX)
  if (eodMatch) {
    const daysBack = eodMatch[2] ? parseInt(eodMatch[2], 10) : 0
    const eod = endOfToday()
    if (daysBack === 0) return eod
    return subDays(eod, daysBack)
  }

  // Fallback to ISO parsing
  const parsed = parseISO(str)
  return isValid(parsed) ? parsed : undefined
}

export function isRelativeTimeString(str: string): boolean {
  return (
    str === 'now' ||
    RELATIVE_TIME_REGEX.test(str) ||
    COMPOUND_TIME_REGEX.test(str) ||
    SOD_REGEX.test(str) ||
    EOD_REGEX.test(str)
  )
}

/**
 * Mapping of quick range labels to their relative time representation.
 * The key is the label from quickRanges, value is [from, to] strings.
 *
 * Note: "Last 30 days" uses `now-30d+15m` (30 days ago plus 15 minutes) to
 * provide a small buffer. This matches the constrained quickRanges30Days
 * implementation which needs this buffer to avoid server errors when querying
 * exactly 30 days of data.
 */
export const quickRangeToRelative: Record<string, [string, string]> = {
  Today: ['sod', 'eod'],
  Yesterday: ['sod-1d', 'eod-1d'],
  'Last 5 min': ['now-5m', 'now'],
  'Last 30 min': ['now-30m', 'now'],
  'Last 1 hour': ['now-1h', 'now'],
  'Last 7 days': ['now-7d', 'now'],
  'Last 30 days': ['now-30d+15m', 'now'],
  'Last 90 days': ['now-90d', 'now'],
  'Last 12 months': ['now-12M', 'now'],
}

// Reverse mapping for looking up labels from relative strings
export const relativeToQuickRange: Record<string, string> = Object.fromEntries(
  Object.entries(quickRangeToRelative).map(([label, [from, to]]) => [
    `${from},${to}`,
    label,
  ]),
)
/**
 * Gets the quick range label for a pair of relative time strings, if it exists.
 */
export function getQuickRangeLabelFromRelative(
  from: string,
  to: string,
): string | undefined {
  return relativeToQuickRange[`${from},${to}`]
}
