import { ActionDetails } from '@/gen/flyteidl2/workflow/run_definition_pb'
import { ActionIdentifier } from '@/gen/flyteidl2/common/identifier_pb'
import { ActionWithChildren } from './types'

export type ParentNode =
  | { type: 'group'; value: string }
  | { type: 'action'; value: ActionIdentifier }

export const getActionParents = ({
  allActions,
  selectedActionId,
}: {
  allActions: Record<string, ActionWithChildren>
  selectedActionId: string | null
}) => {
  const parents: ParentNode[] = []
  let thisNode: ActionWithChildren | ActionDetails | null = null
  if (selectedActionId && allActions[selectedActionId]) {
    thisNode = allActions[selectedActionId]
  }
  while (thisNode?.action?.metadata?.parent && thisNode.action?.id) {
    parents.push({ type: 'action', value: thisNode.action?.id })
    if (thisNode?.action?.metadata?.group) {
      parents.push({ type: 'group', value: thisNode.action?.metadata.group })
    }

    if (
      thisNode?.action?.metadata?.parent &&
      allActions[thisNode?.action?.metadata?.parent]
    ) {
      thisNode = allActions[thisNode?.action?.metadata?.parent]
    }
  }
  return parents.reverse()
}
