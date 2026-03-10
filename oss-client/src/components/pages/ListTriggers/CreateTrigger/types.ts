import { EnrichedIdentity } from '@/gen/flyteidl2/common/identity_pb'
import { Task } from '@/gen/flyteidl2/task/task_definition_pb'

export type TriggerMode = 'select-task' | 'define-trigger'
export type TaskDetails = {
  taskId: string
  taskVersion: string
}

export type TableTask = {
  name: {
    shortName: string | undefined
    fullName: string | undefined
    shortDescription: string | undefined
  }
  lastDeployed: {
    date: number | undefined
    version: string | undefined
  }
  createdUser: EnrichedIdentity | undefined
  actions: Task
}

export type CreateTriggerTab =
  | 'definition'
  | 'inputs'
  | 'labels'
  | 'env-vars'
  | 'annotations'
  | 'advanced'
