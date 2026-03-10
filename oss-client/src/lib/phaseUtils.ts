import { ActionPhase } from '@/gen/flyteidl2/common/phase_pb'

export type PhaseKey = keyof typeof ActionPhase | null

export function getPhaseString(phase: ActionPhase | undefined) {
  switch (phase) {
    case ActionPhase.QUEUED:
      return 'Queued'
    case ActionPhase.WAITING_FOR_RESOURCES:
      return 'Waiting'
    case ActionPhase.INITIALIZING:
      return 'Initializing'
    case ActionPhase.RUNNING:
      return 'Running'
    case ActionPhase.SUCCEEDED:
      return 'Succeeded'
    case ActionPhase.FAILED:
      return 'Failed'
    case ActionPhase.ABORTED:
      return 'Aborted'
    case ActionPhase.TIMED_OUT:
      return 'Timed out'
    case ActionPhase.UNSPECIFIED:
    default:
      return 'Unknown'
  }
}

export function getPhaseClass(phase: ActionPhase | undefined): string {
  switch (phase) {
    case ActionPhase.QUEUED:
      return 'phase-queued'
    case ActionPhase.WAITING_FOR_RESOURCES:
      return 'phase-waiting'
    case ActionPhase.INITIALIZING:
      return 'phase-initializing'
    case ActionPhase.RUNNING:
      return 'phase-running'
    case ActionPhase.SUCCEEDED:
      return 'phase-succeeded'
    case ActionPhase.FAILED:
      return 'phase-failed'
    case ActionPhase.ABORTED:
      return 'phase-aborted'
    case ActionPhase.TIMED_OUT:
      return 'phase-timed-out'
    case ActionPhase.UNSPECIFIED:
    default:
      return 'phase-unspecified'
  }
}

export const getPhaseEnumValue = (
  phaseStringValue: PhaseKey | undefined,
): ActionPhase | undefined => {
  switch (phaseStringValue) {
    case 'ABORTED':
      return ActionPhase.ABORTED
    case 'FAILED':
      return ActionPhase.FAILED
    case 'INITIALIZING':
      return ActionPhase.INITIALIZING
    case 'SUCCEEDED':
      return ActionPhase.SUCCEEDED
    case 'TIMED_OUT':
      return ActionPhase.TIMED_OUT
    case 'QUEUED':
      return ActionPhase.QUEUED
    case 'RUNNING':
      return ActionPhase.RUNNING
    case 'WAITING_FOR_RESOURCES':
      return ActionPhase.WAITING_FOR_RESOURCES
    case 'UNSPECIFIED':
      return ActionPhase.UNSPECIFIED
    default: {
      return undefined
    }
  }
}

export const getPhaseFroπmEnum = (
  p: ActionPhase | undefined,
): PhaseKey | undefined => {
  switch (p) {
    case ActionPhase.ABORTED:
      return 'ABORTED'
    case ActionPhase.FAILED:
      return 'FAILED'
    case ActionPhase.INITIALIZING:
      return 'INITIALIZING'
    case ActionPhase.QUEUED:
      return 'QUEUED'
    case ActionPhase.RUNNING:
      return 'RUNNING'
    case ActionPhase.SUCCEEDED:
      return 'SUCCEEDED'
    case ActionPhase.TIMED_OUT:
      return 'TIMED_OUT'
    case ActionPhase.UNSPECIFIED:
      return 'UNSPECIFIED'
    case ActionPhase.WAITING_FOR_RESOURCES:
      return 'WAITING_FOR_RESOURCES'
    default:
      return undefined
  }
}