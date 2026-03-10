import { EnrichedAction } from '@/gen/flyteidl2/workflow/run_definition_pb'

export type ActionId = string

export type ActionWithChildren = EnrichedAction & {
  children: ActionId[]
  groupChildren: Record<string, ActionId[]>
  isGroup: boolean
}

export type FlatRunNode = {
  id: ActionId
  depth: number
  node: ActionWithChildren
  isGroup: boolean
}
