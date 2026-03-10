import { useQueryState, parseAsString } from 'nuqs'
import { useCallback, useMemo } from 'react'
import { DateRange } from 'react-day-picker'
import { formatISO } from 'date-fns'
import {
  parseRelativeTime,
  isRelativeTimeString,
  quickRangeToRelative,
} from './relativeTime'

export const useDateRangeParams = () => {
  const [fromParam, setFromParam] = useQueryState(
    'from',
    parseAsString.withDefault(''),
  )
  const [toParam, setToParam] = useQueryState(
    'to',
    parseAsString.withDefault(''),
  )

  const dateRange: DateRange | undefined = useMemo(() => {
    if (!fromParam || !toParam) return undefined

    try {
      const from = parseRelativeTime(fromParam)
      const to = parseRelativeTime(toParam)

      if (!from || !to) return undefined

      return { from, to }
    } catch {
      return undefined
    }
  }, [fromParam, toParam])

  const setDateRange = useCallback(
    (range: DateRange | undefined, quickRangeLabel?: string) => {
      if (!range?.from || !range?.to) {
        setFromParam(null)
        setToParam(null)
        return
      }

      // If a quick range label is provided, use the relative format
      if (quickRangeLabel && quickRangeToRelative[quickRangeLabel]) {
        const [fromStr, toStr] = quickRangeToRelative[quickRangeLabel]
        setFromParam(fromStr)
        setToParam(toStr)
        return
      }

      // Otherwise fall back to ISO format
      const fromStr = formatISO(range.from)
      const toStr = formatISO(range.to)
      setFromParam(fromStr)
      setToParam(toStr)
    },
    [setFromParam, setToParam],
  )

  const clearDateRange = useCallback(() => {
    setFromParam(null)
    setToParam(null)
  }, [setFromParam, setToParam])

  // Check if current params are relative (for preserving on refresh)
  const isRelativeRange = useMemo(() => {
    return isRelativeTimeString(fromParam) && isRelativeTimeString(toParam)
  }, [fromParam, toParam])

  return {
    dateRange,
    setDateRange,
    clearDateRange,
    // Raw params for debugging or custom use
    fromParam,
    toParam,
    isRelativeRange,
  }
}
