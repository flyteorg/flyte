import { ActionPhase } from '@/gen/flyteidl2/common/phase_pb'

export const mapPhaseToDisplayString: Record<ActionPhase, string> = {
  [ActionPhase.ABORTED]: 'Aborted',
  [ActionPhase.FAILED]: 'Failed',
  [ActionPhase.INITIALIZING]: 'Initializing',
  [ActionPhase.QUEUED]: 'Queued',
  [ActionPhase.RUNNING]: 'Running',
  [ActionPhase.SUCCEEDED]: 'Completed',
  [ActionPhase.TIMED_OUT]: 'Timed out',
  [ActionPhase.UNSPECIFIED]: 'Unspecified',
  [ActionPhase.WAITING_FOR_RESOURCES]: 'Waiting for resources',
}