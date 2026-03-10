import { ActionPhase } from '@/gen/flyteidl2/common/phase_pb'
import { AccentColor } from '@/types/colors'

export const getColorsByPhase = (phase: ActionPhase | undefined): AccentColor => {
  switch (phase) {
    case ActionPhase.FAILED: {
      return 'red'
    }
    case ActionPhase.WAITING_FOR_RESOURCES:
    case ActionPhase.QUEUED:
    case ActionPhase.INITIALIZING: {
      return 'purple'
    }
    case ActionPhase.RUNNING: {
      return 'blue'
    }
    case ActionPhase.SUCCEEDED: {
      return 'green'
    }
    case ActionPhase.ABORTED: {
      return 'orange'
    }
    case ActionPhase.TIMED_OUT: {
      return 'yellow'
    }
    case ActionPhase.UNSPECIFIED:
    default: {
      return 'gray'
    }
  }
}