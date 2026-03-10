import React, { useMemo } from 'react'
import { FolderGroup } from '@/components/FolderGroup'
import { ActionPhase } from '@/gen/flyteidl2/common/phase_pb'
import { type ChildPhaseCounts } from '@/components/pages/RunDetails/types'
import { ActionWithChildren } from '../state/types'
import { usePhaseCounts } from './usePhaseCounts'
import { DisplayText } from './DisplayText'

export function useGroupPhase(counts: ChildPhaseCounts): ActionPhase | null {
  return useMemo(() => {
    if (!counts || Object.keys(counts).length === 0) return null

    const getCount = (phase: ActionPhase) => counts[phase] ?? 0

    const total = Object.values(counts).reduce((sum, val) => sum + val, 0)

    const terminalPhases = [
      ActionPhase.SUCCEEDED,
      ActionPhase.FAILED,
      ActionPhase.ABORTED,
      ActionPhase.TIMED_OUT,
    ]
    const hasTerminal = terminalPhases.some((phase) => getCount(phase) > 0)

    if (getCount(ActionPhase.UNSPECIFIED) === total) {
      return ActionPhase.UNSPECIFIED // "Not Started"
    }

    if (
      (getCount(ActionPhase.QUEUED) > 0 ||
        getCount(ActionPhase.WAITING_FOR_RESOURCES) > 0) &&
      getCount(ActionPhase.INITIALIZING) === 0 &&
      !hasTerminal
    ) {
      return ActionPhase.QUEUED
    }

    if (
      getCount(ActionPhase.INITIALIZING) > 0 &&
      getCount(ActionPhase.RUNNING) === 0 &&
      !hasTerminal
    ) {
      return ActionPhase.INITIALIZING
    }

    if (getCount(ActionPhase.RUNNING) > 0 && !hasTerminal) {
      return ActionPhase.RUNNING
    }

    if (getCount(ActionPhase.FAILED) > 0) return ActionPhase.FAILED
    if (getCount(ActionPhase.TIMED_OUT) > 0) return ActionPhase.TIMED_OUT
    if (getCount(ActionPhase.ABORTED) > 0) return ActionPhase.ABORTED
    if (getCount(ActionPhase.SUCCEEDED) === total) return ActionPhase.SUCCEEDED

    return null
  }, [counts])
}

export const GroupFolderItem = ({
  displayText,
  isSelected,
  maxWidth,
  node,
}: {
  displayText: string | undefined
  isSelected: boolean
  maxWidth: number
  node: ActionWithChildren
}) => {
  const phaseCounts = usePhaseCounts({ node })
  const phase = useGroupPhase(phaseCounts)
  return (
    <div className="flex min-w-0 flex-1 items-center gap-1">
      <FolderGroup phase={phase} />
      <DisplayText
        displayText={displayText}
        isSelected={isSelected}
        maxWidth={maxWidth}
        phase={phase}
      />
    </div>
  )
}