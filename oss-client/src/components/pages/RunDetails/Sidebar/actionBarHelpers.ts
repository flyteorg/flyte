import { useMemo, useRef } from 'react'
import { Timestamp } from '@/gen/google/protobuf/timestamp_pb'
import { useRunStore } from '../state/RunStore'
import { useGlobalNow } from '../state/GlobalTimestamp'
import { timestampToMillis } from '@/lib/dateUtils'

export function getDurationMs(
  start: Timestamp | undefined,
  end: Timestamp | undefined,
  now: number,
): number {
  const startMs = timestampToMillis(start)
  const endMs = end ? timestampToMillis(end) : now
  if (startMs == null || endMs == null) return 0
  return Math.max(endMs - startMs, 0)
}

type UseActionBarMetricsProps = {
  startTime: Timestamp | undefined
  endTime: Timestamp | undefined
}

export function useActionBarMetrics({
  startTime,
  endTime,
}: UseActionBarMetricsProps) {
  const runStart = useRunStore((s) => s.run?.action?.status?.startTime)
  const runEnd = useRunStore((s) => s.run?.action?.status?.endTime)
  const globalNow = useGlobalNow((s) => (runEnd ? null : s.now))
  const nowRef = useRef(Date.now())

  // If not running, keep last known time frozen
  if (globalNow) {
    nowRef.current = globalNow
  }

  return useMemo(() => {
    const runStartMs = timestampToMillis(runStart)
    const runEndMs = timestampToMillis(runEnd) ?? nowRef.current
    const actionStartMs = timestampToMillis(startTime)
    const actionEndMs = endTime ? timestampToMillis(endTime) : nowRef.current

    if (runStartMs == null || actionStartMs == null) {
      return { offsetPercent: 0, widthPercent: 0 }
    }

    const runDuration = Math.max(runEndMs - runStartMs, 1)
    const actionOffset = Math.max(actionStartMs - runStartMs, 0)
    const actionDuration = Math.max(
      (actionEndMs ?? nowRef.current) - actionStartMs,
      0,
    )

    return {
      offsetPercent: (actionOffset / runDuration) * 100,
      widthPercent: (actionDuration / runDuration) * 100,
    }
  }, [runStart, runEnd, startTime, endTime])
}
