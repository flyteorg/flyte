import { TriggerName } from '@/gen/flyteidl2/common/identifier_pb'
import { EnrichedIdentity } from '@/gen/flyteidl2/common/identity_pb'
import { Trigger } from '@/gen/flyteidl2/trigger/trigger_definition_pb'
import { ReactNode } from 'react'

export type TriggerTableRow = {
  // same name as on BE to simplify sorting
  active: {
    id: string
    status: boolean
  }
  nextRun: string | null | undefined
  name: {
    name: string | undefined
    schedule: string | null | undefined
    nextRun: string | null | undefined
  }
  task: TriggerName | undefined
  triggered: {
    date: string | undefined
  }
  updated: {
    date: string | undefined
    updatedBy: EnrichedIdentity | undefined
  }
  createdDate: string
  createdUser: EnrichedIdentity | undefined
  actions: Trigger
}

export type TriggerTableRowWithHighlights = Omit<TriggerTableRow, 'name'> & {
  name: Omit<TriggerTableRow['name'], 'name'> & {
    name: ReactNode
  }
}
