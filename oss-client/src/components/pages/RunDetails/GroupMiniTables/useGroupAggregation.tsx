import { useMemo } from 'react'
import { useSelectedGroup } from '../hooks/useSelectedItem'
import { ActionPhase } from '@/gen/flyteidl2/common/phase_pb'
import { useRunStore } from '../state/RunStore'
import { transformForTable } from './util'

const LONGEST_DURATION_ITEMS = 3

function toMillis(ts?: { seconds?: bigint; nanos?: number }): number {
  if (!ts || ts.seconds == null) return 0
  const nanos = ts.nanos ?? 0
  return Number(ts.seconds) * 1000 + Math.floor(nanos / 1e6)
}

// get failed and longestDuration for selected group
export const useGroupAggregation = () => {
  const selectedGroup = useSelectedGroup()
  const flatItems = useRunStore((s) => s.flatItems)

  return useMemo(() => {
    if (!selectedGroup?.name) {
      return { failed: [], longest: [] }
    }

    const groupItems = flatItems.filter(
      (item) =>
        !item.isGroup &&
        item.node.action?.metadata?.group === selectedGroup.name,
    )

    const failed = groupItems
      .filter((item) => item.node.action?.status?.phase === ActionPhase.FAILED)
      .map((item) => transformForTable(item.node))

    const longest = groupItems
      .map((item) => {
        const start = toMillis(item.node.action?.status?.startTime)
        const end = toMillis(item.node.action?.status?.endTime)
        const duration = start && end ? end - start : 0
        return { item, duration }
      })
      .filter(({ duration }) => duration > 0)
      .sort((a, b) => b.duration - a.duration)
      .slice(0, LONGEST_DURATION_ITEMS)
      .map(({ item }) => transformForTable(item.node))

    return { failed, longest }
  }, [flatItems, selectedGroup?.name])
}