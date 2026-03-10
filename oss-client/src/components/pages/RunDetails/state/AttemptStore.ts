import { create } from 'zustand'
import { ActionAttempt } from '@/gen/flyteidl2/workflow/run_definition_pb'

type SelectedAttempt = { attempt: ActionAttempt; isManual?: boolean }

export type SelectedAttemptState = {
  selectedAttemptNumber: number | null
  selectedAttempt: ActionAttempt | null
  setSelectedAttempt: (selectedAttempt: SelectedAttempt) => void
  clearSelectedAttempt: () => void
  manuallySelectedAttemptNumber: number | 'latest'
  clearManuallySelectedAttempt: () => void
}

export const useSelectedAttemptStore = create<SelectedAttemptState>((set) => ({
  selectedAttemptNumber: null,
  selectedAttempt: null,
  setSelectedAttempt: ({ attempt, isManual }) =>
    set((prev) => ({
      selectedAttemptNumber: attempt.attempt,
      selectedAttempt: attempt,
      manuallySelectedAttemptNumber: isManual
        ? attempt.attempt
        : prev.manuallySelectedAttemptNumber,
    })),
  clearSelectedAttempt: () =>
    set({ selectedAttempt: null, selectedAttemptNumber: null }),
  manuallySelectedAttemptNumber: 'latest',
  clearManuallySelectedAttempt: () =>
    set({ manuallySelectedAttemptNumber: 'latest' }),
}))
