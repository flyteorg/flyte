import { PhaseTransition } from '@/gen/flyteidl2/workflow/run_definition_pb'
import { ActionPhase } from '@/gen/flyteidl2/common/phase_pb'
import { Timestamp } from '@bufbuild/protobuf/wkt'

const timestampToMillis = (ts?: Timestamp): number =>
  ts ? Number(ts.seconds) * 1000 + ts.nanos / 1_000_000 : 0

export const durationBetween = (start?: number, end?: number): number => {
  if (!start) return 0
  return Math.max((end ?? Date.now()) - start, 1) // avoid 0
}

export const getEarliestLatest = (phaseTransitions: PhaseTransition[]) => {
  if (phaseTransitions.length === 0) {
    return { startTime: undefined, endTime: undefined }
  }

  let earliestStart: number | undefined
  let latestStart: number | undefined
  let latestPhaseWithStart: PhaseTransition | undefined

  for (const phase of phaseTransitions) {
    const start = timestampToMillis(phase.startTime)

    if (earliestStart === undefined || start < earliestStart) {
      earliestStart = start
    }
    // handle case where dropped nanos on backend causes bug calculating duration
    const isLatestPhase =
      latestPhaseWithStart && phase.phase > latestPhaseWithStart?.phase
    if (latestStart === undefined || start >= latestStart || isLatestPhase) {
      latestStart = start
      latestPhaseWithStart = phase
    }
  }

  const endTime = latestPhaseWithStart?.endTime
    ? timestampToMillis(latestPhaseWithStart.endTime)
    : undefined

  return {
    startTime: earliestStart,
    endTime,
  }
}

export const getIsRunning = (phaseTransitions: PhaseTransition[]) => {
  if (phaseTransitions.some((p) => p.phase > ActionPhase.RUNNING)) {
    return false
  }
  return phaseTransitions.some((p) => p.phase === ActionPhase.RUNNING)
}