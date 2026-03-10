import { ActionAttempt } from '@/gen/flyteidl2/workflow/run_definition_pb'
import { ActionPhase } from '@/gen/flyteidl2/common/phase_pb'

export const isAttemptSucceeded = (attempt?: ActionAttempt | null) => {
  return (attempt?.phase || ActionPhase.UNSPECIFIED) == ActionPhase.SUCCEEDED
}

export const isAttemptTerminal = (attempt?: ActionAttempt | null) => {
  return [
    ActionPhase.SUCCEEDED,
    ActionPhase.ABORTED,
    ActionPhase.FAILED,
    ActionPhase.TIMED_OUT,
  ].includes(attempt?.phase || ActionPhase.UNSPECIFIED)
}