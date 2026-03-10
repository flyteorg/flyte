import { useSelectedActionId } from '@/components/pages/RunDetails/hooks/useSelectedItem'
import {
  ActionAttempt,
  ActionDetails,
} from '@/gen/flyteidl2/workflow/run_definition_pb'
import { useEffect } from 'react'
import { useSelectedAttemptStore } from '../state/AttemptStore'

export const useSyncSelectedAttempt = (actionDetails?: ActionDetails) => {
  const selectedActionId = useSelectedActionId()
  const {
    selectedAttemptNumber,
    setSelectedAttempt,
    clearSelectedAttempt,
    manuallySelectedAttemptNumber,
    clearManuallySelectedAttempt,
  } = useSelectedAttemptStore()

  useEffect(() => {
    clearSelectedAttempt()
    clearManuallySelectedAttempt()
  }, [selectedActionId, clearManuallySelectedAttempt, clearSelectedAttempt])

  useEffect(() => {
    if (actionDetails?.metadata?.spec.case === 'trace') {
      setSelectedAttempt({
        // we have to manually create the attempt data for trace ActionTimeline because it has no phases
        attempt: {
          phaseTransitions: [
            {
              startTime: actionDetails?.status?.startTime,
              endTime: actionDetails?.status?.endTime,
              phase: actionDetails?.status?.phase,
            },
          ],
        } as ActionAttempt,
        isManual: true,
      })
      return
    }

    if (!actionDetails?.attempts?.length) {
      return
    }

    const attempts = actionDetails.attempts
    const latest = attempts[attempts.length - 1]

    // check if we have manually selected attempt number and that attempt exists
    const manuallySelected =
      manuallySelectedAttemptNumber === 'latest'
        ? undefined
        : attempts?.[manuallySelectedAttemptNumber - 1]

    // select manually-selected attempt if it's available
    if (manuallySelectedAttemptNumber !== 'latest' && !!manuallySelected) {
      setSelectedAttempt({ attempt: manuallySelected })
      return
    }

    //if set to "latest" and new attempt arrived, change to latest
    if (
      manuallySelectedAttemptNumber === 'latest' &&
      selectedAttemptNumber !== null &&
      latest.attempt > selectedAttemptNumber
    ) {
      setSelectedAttempt({ attempt: latest })
      return
    }

    // If no attempt is selected, select the latest
    if (selectedAttemptNumber == null) {
      setSelectedAttempt({ attempt: manuallySelected ?? latest })
      return
    }

    // Find the current attempt by number
    const currentAttempt = attempts.find(
      (a) => a.attempt === selectedAttemptNumber,
    )

    // If the selected attempt doesn't exist anymore, fall back to latest
    if (!currentAttempt) {
      setSelectedAttempt({ attempt: latest })
      return
    }

    // Always update with fresh data for the selected attempt
    setSelectedAttempt({ attempt: currentAttempt })
  }, [
    actionDetails, // This will trigger when ANY nested data changes
    selectedAttemptNumber,
    setSelectedAttempt,
    clearSelectedAttempt,
    manuallySelectedAttemptNumber,
  ])
}
