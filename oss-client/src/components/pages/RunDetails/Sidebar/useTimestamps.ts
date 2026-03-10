import { useMemo } from 'react'
import { useRunStore } from '../state/RunStore'
import { Timestamp } from '@bufbuild/protobuf/wkt'
import { ActionWithChildren } from '../state/types'

function compareTimestamps(a?: Timestamp, b?: Timestamp): number {
  if (!a) return 1
  if (!b) return -1
  return a.seconds === b.seconds
    ? a.nanos - b.nanos
    : Number(a.seconds) - Number(b.seconds)
}

function isAfter(a?: Timestamp, b?: Timestamp): boolean {
  return compareTimestamps(a, b) > 0
}

function isBefore(a?: Timestamp, b?: Timestamp): boolean {
  return compareTimestamps(a, b) < 0
}

// for groups that contain multiple children, return earliest startTime and latest endTime.
// for endTime in groups, we bias to returning undefined for the ActionBar visualization
export const useGroupTimestamp = ({
  isGroup,
  node,
}: {
  isGroup: boolean
  node: ActionWithChildren
}) => {
  const allActions = useRunStore((s) => s.actions)
  return useMemo(() => {
    if (!isGroup) {
      return {
        startTime: node.action?.status?.startTime,
        endTime: node.action?.status?.endTime,
      }
    }

    let earliestStart: Timestamp | undefined
    let latestEnd: Timestamp | undefined
    let hasIncompleteChild = false

    for (const actionId of node.children) {
      const child = allActions[actionId]
      if (!child) continue
      const childStart = child.action?.status?.startTime
      const childEnd = child.action?.status?.endTime

      if (
        childStart &&
        (!earliestStart || isBefore(childStart, earliestStart))
      ) {
        earliestStart = childStart
      }

      if (!childEnd) {
        hasIncompleteChild = true
      } else if (
        !hasIncompleteChild &&
        (!latestEnd || isAfter(childEnd, latestEnd))
      ) {
        latestEnd = childEnd
      }
    }

    return {
      startTime: earliestStart,
      endTime: hasIncompleteChild ? undefined : latestEnd,
    }
  }, [
    allActions,
    isGroup,
    node.action?.status?.endTime,
    node.action?.status?.startTime,
    node.children,
  ])
}
