import { useMemo } from 'react'
import { useRunStore } from '../state/RunStore'
import { ActionWithChildren } from '../state/types'

export const usePhaseCounts = ({ node }: { node?: ActionWithChildren }) => {
  const allActions = useRunStore((s) => s.actions)
  return useMemo(() => {
    if (!node) return {}
    if (!node.isGroup) return node.childrenPhaseCounts
    // aggregate phase count summaries for groups
    const groupChildren = node.children ?? []

    const aggregate: { [key: number]: number } = {}

    for (const childId of groupChildren) {
      const childNode = allActions[childId]
      if (!childNode) continue

      const childCounts: { [key: number]: number } =
        childNode.childrenPhaseCounts ?? {}
      for (const [phaseStr, count] of Object.entries(childCounts)) {
        const phase = Number(phaseStr)
        aggregate[phase] = (aggregate[phase] ?? 0) + count
      }

      // Optionally include each child's own status
      const ownPhase = childNode.action?.status?.phase
      if (ownPhase != null) {
        aggregate[ownPhase] = (aggregate[ownPhase] ?? 0) + 1
      }
    }

    return aggregate
  }, [allActions, node])
}
