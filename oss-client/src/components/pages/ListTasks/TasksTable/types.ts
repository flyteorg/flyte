import { EnrichedIdentity } from '@/gen/flyteidl2/common/identity_pb'
import { Task } from '@/gen/flyteidl2/task/task_definition_pb'
import { ActionPhase } from '@/gen/flyteidl2/common/phase_pb'
import type { ReactNode } from 'react'

export type TaskTableRow = {
  name: {
    shortName: string | undefined
    fullName: string | undefined
    shortDescription: string | undefined
  }
  environment: string | undefined
  createdAt: {
    version: string | undefined
    date: string
  }
  createdUser: EnrichedIdentity | undefined
  trigger:
    | {
        active: boolean
        title: string
        subtitle: string
      }
    | undefined
  lastRun:
    | {
        phase: ActionPhase | undefined
        time: string
        runId: string | undefined
      }
    | undefined
  copyAction: Task
  original: Task
}

export type TaskTableRowWithHighlights = Omit<TaskTableRow, 'name'> & {
  name: {
    shortName: ReactNode
    fullName: ReactNode
    shortDescription: ReactNode
  }
}
