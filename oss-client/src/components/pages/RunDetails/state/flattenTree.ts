import { ActionId, ActionWithChildren, FlatRunNode } from '../state/types'

export const GROUP_SEPARATOR = '::group::'

export function flattenTree(
  id: ActionId,
  nodes: Record<ActionId, ActionWithChildren>,
  collapsed: Set<ActionId>,
  depth: number = 0,
): FlatRunNode[] {
  const result: FlatRunNode[] = []
  const node = nodes[id]
  if (!node) return result

  result.push({ id, node, depth, isGroup: false })

  if (collapsed.has(id)) return result

  // Helper function to get start time for sorting
  const getStartTime = (nodeId: ActionId): number => {
    const nodeToCheck = nodes[nodeId]
    const startTime = nodeToCheck?.action?.status?.startTime

    if (!startTime) return 0

    // Handle protobuf timestamp: { seconds: bigint, nanos: number }
    const seconds = Number(startTime.seconds) || 0
    const nanos = Number(startTime.nanos) || 0

    // Convert to milliseconds for comparison
    return seconds * 1000 + nanos / 1000000
  }

  // Sort group entries by the earliest start time in each group
  const sortedGroupEntries = Object.entries(node.groupChildren).sort(
    ([, groupIdsA], [, groupIdsB]) => {
      const earliestA = Math.min(...groupIdsA.map(getStartTime))
      const earliestB = Math.min(...groupIdsB.map(getStartTime))
      return earliestA - earliestB
    },
  )

  // Sort regular children by start time
  const sortedChildren = [...node.children].sort(
    (a, b) => getStartTime(a) - getStartTime(b),
  )

  // Create a combined sorted list of groups and individual children
  const combinedItems: Array<
    | {
        type: 'group'
        groupName: string
        groupIds: ActionId[]
        sortTime: number
      }
    | { type: 'child'; id: ActionId; sortTime: number }
  > = [
    ...sortedGroupEntries.map(([groupName, groupIds]) => ({
      type: 'group' as const,
      groupName,
      groupIds,
      sortTime: Math.min(...groupIds.map(getStartTime)),
    })),
    ...sortedChildren.map((childId) => ({
      type: 'child' as const,
      id: childId,
      sortTime: getStartTime(childId),
    })),
  ]

  // Sort the combined list by start time
  combinedItems.sort((a, b) => a.sortTime - b.sortTime)

  // Process the combined sorted list
  for (const item of combinedItems) {
    if (item.type === 'group') {
      const { groupName, groupIds } = item
      const folderId = `${id}${GROUP_SEPARATOR}${groupName}`

      result.push({
        id: folderId,
        node: {
          ...node,
          isGroup: true,
          groupChildren: {},
          children: groupIds,
          ...(node.action?.metadata && {
            action: {
              ...node.action,
              metadata: {
                ...node.action?.metadata,
                group: groupName,
              },
            },
          }),
        },
        depth: depth + 1,
        isGroup: true,
      })

      if (!collapsed.has(folderId)) {
        // Sort group children by start time before processing them recursively
        // This is necessary because the children within a group also need to be in chronological order
        const sortedGroupChildren = [...groupIds].sort(
          (a, b) => getStartTime(a) - getStartTime(b),
        )

        for (const childId of sortedGroupChildren) {
          result.push(...flattenTree(childId, nodes, collapsed, depth + 2))
        }
      }
    } else {
      // Regular child
      result.push(...flattenTree(item.id, nodes, collapsed, depth + 1))
    }
  }

  return result
}
