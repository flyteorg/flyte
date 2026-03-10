import { EnrichedIdentity } from '@/gen/flyteidl2/common/identity_pb'
import { Run } from '@/gen/flyteidl2/workflow/run_definition_pb'
import { ActionPhase } from '@/gen/flyteidl2/common/phase_pb'

export type RunsTableRow = {
  runId: {
    status: ActionPhase | undefined
    id: string | undefined
  }
  name: {
    shortName: string | undefined
    fullName: string | undefined
  }
  startTime: string | undefined
  endTime: string | undefined
  runTime: string | undefined
  user: EnrichedIdentity | undefined
  actions: {
    url: string
    runId: string | undefined
    latestVersion: string | undefined
    actionId: string | undefined
  }
  original: Run
  trigger: {
    name: string | undefined
    type: string | undefined
  }
  environment: string | undefined
}
