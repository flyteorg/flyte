import { Identifier, App } from '@/gen/flyteidl2/app/app_definition_pb'
import { EnrichedIdentity } from '@/gen/flyteidl2/common/identity_pb'
import { Status_DeploymentStatus } from '@/gen/flyteidl2/app/app_definition_pb'

export type AppTableItem = {
  id: Identifier | undefined
  status: Status_DeploymentStatus
  replicas: {
    min: number
    max: number
  }
  name: {
    displayText: string
    endpoint: string
  }
  type: string
  lastDeployed: {
    relativeTime: string
    version: string
  }
  deployedBy: EnrichedIdentity | undefined
  actions: App
  original: App
}
