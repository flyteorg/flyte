import { ActionPhase } from '@/gen/flyteidl2/common/phase_pb'
import { ActionWithChildren } from '../state/types'

export type GroupTableItem = {
  name: {
    actionId: string
    taskName: string
    phase: ActionPhase | undefined
  }
  duration: string
  startTime: string
  original: ActionWithChildren
}