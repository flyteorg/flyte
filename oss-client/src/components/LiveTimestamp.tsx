import { getDateObject } from '@/lib/dateUtils'
import { Timestamp } from '@bufbuild/protobuf/wkt'
import clsx from 'clsx'
import dayjs from 'dayjs'
import duration from 'dayjs/plugin/duration'
import { useMemo, useRef } from 'react'
import { useGlobalNow } from './pages/RunDetails/state/GlobalTimestamp'

dayjs.extend(duration)

type LiveTimestampProps = {
  decimalPlaces?: number
  timestamp?: Timestamp | number
  endTimestamp?: Timestamp | number
  className?: string
  minWidth?: number
}

export const LiveTimestamp = ({
  decimalPlaces = 2,
  timestamp,
  endTimestamp,
  className,
  minWidth = 60,
}: LiveTimestampProps) => {
  const hasEnded = Boolean(endTimestamp)
  const globalNow = useGlobalNow((s) => (hasEnded ? null : s.now))
  const nowRef = useRef(Date.now())

  if (!hasEnded && globalNow) {
    nowRef.current = globalNow
  }

  const effectiveEnd = endTimestamp
    ? getDateObject(endTimestamp).valueOf()
    : nowRef.current

  const displayTimestamp = useMemo(() => {
    return formatRuntime({
      decimalPlaces,
      startTime: timestamp,
      endTime: effectiveEnd,
      now: effectiveEnd,
    })
  }, [timestamp, effectiveEnd, decimalPlaces])

  return displayTimestamp ? (
    <div
      className={clsx(
        className ? className : 'text-2xs leading-5 font-normal tracking-[0px]',
      )}
      style={{
        minWidth: minWidth || 'fit-content',
      }}
    >
      {displayTimestamp}
    </div>
  ) : null
}
type FormatOptions = {
  decimalPlaces: number
  startTime?: Timestamp | number
  endTime?: Timestamp | number
  now: number
}

function formatRuntime({
  decimalPlaces,
  startTime,
  endTime,
  now,
}: FormatOptions): string {
  if (!startTime) return '-'

  const start = getDateObject(startTime)
  const end = endTime ? getDateObject(endTime) : dayjs(now)

  const formatSeconds = (s: number, decimalPlaces: number = 2) => {
    let decimalsToUse = 2
    if (decimalPlaces === 1) {
      if (s < 1) {
        decimalsToUse = 2
      } else {
        decimalsToUse = 1
      }
    }
    const fixed = s.toFixed(decimalsToUse)
    return fixed.endsWith('.00') ? `${Math.floor(s)}s` : `${fixed}s`
  }

  if (end.isBefore(start)) {
    if (start.diff(end, 'seconds') === 0) {
      return formatSeconds(0, decimalPlaces)
    }
    return ''
  }
  const ms = end.diff(start)
  const duration = dayjs.duration(end.diff(start))

  const totalSeconds = ms / 1000
  const totalMinutes = duration.asMinutes()
  const totalHours = duration.asHours()
  const totalDays = duration.asDays()

  if (totalSeconds < 60) {
    return formatSeconds(totalSeconds, decimalPlaces)
  } else if (totalMinutes < 60) {
    return `${Math.floor(totalMinutes)}m ${Math.round(duration.seconds())}s`
  } else if (totalHours < 24) {
    return `${Math.floor(totalHours)}h ${Math.round(duration.minutes())}m`
  } else {
    return `${Math.floor(totalDays)}d ${Math.round(duration.hours())}h`
  }
}
