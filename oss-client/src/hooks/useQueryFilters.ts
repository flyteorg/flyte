import { useCallback, useMemo } from 'react'
import { useQueryState } from 'nuqs'
import { ActionPhase } from '@/gen/flyteidl2/common/phase_pb'
import { PhaseKey } from '@/lib/phaseUtils'

type StatusFilter = {
  type: 'status'
  status: PhaseKey | PhaseKey[]
}

interface QueryFilters {
  addStatusFilters: (statuses: PhaseKey | PhaseKey[]) => void
  filters: {
    status?: PhaseKey[] | null
  }
  removeStatusFilters: (statuses: PhaseKey | PhaseKey[]) => void
  toggleFilter: (p: StatusFilter) => void
  clearFilter: () => void
}

const excludeFalsyValues = (s: PhaseKey | string) => !!s

const DELIMITER = ';'

export const useQueryFilters = (): QueryFilters => {
  const [statusValue, setStatus] = useQueryState('status')

  const handleStatusUpdate = useCallback(
    (updatedStatus: PhaseKey | PhaseKey[]) => {
      setStatus((currentStatus) => {
        const current = currentStatus
          ? currentStatus.split(DELIMITER).filter(excludeFalsyValues)
          : []

        const updatedArray = Array.isArray(updatedStatus)
          ? updatedStatus
          : [updatedStatus]

        let newStatuses = [...current]

        for (const status of updatedArray) {
          if (!status) continue
          if (newStatuses.includes(status)) {
            newStatuses = newStatuses.filter((s) => s !== status)
          } else {
            newStatuses.push(status)
          }
        }

        return newStatuses.length > 0
          ? newStatuses.join(DELIMITER) + DELIMITER
          : null
      })
    },
    [setStatus],
  )

  const handleStatusAdd = useCallback(
    (toAdd: PhaseKey | PhaseKey[]) => {
      setStatus((currentStatus) => {
        const current = currentStatus
          ? currentStatus.split(DELIMITER).filter(excludeFalsyValues)
          : []
        const additions = normalize(toAdd)
        const next = Array.from(new Set([...current, ...additions]))
        return next.length > 0 ? next.join(DELIMITER) + DELIMITER : null
      })
    },
    [setStatus],
  )
  const normalize = (s: PhaseKey | PhaseKey[]) =>
    (Array.isArray(s) ? s : [s]).filter(Boolean) as (keyof typeof ActionPhase)[]

  const handleStatusRemove = useCallback(
    (toRemove: PhaseKey | PhaseKey[]) => {
      setStatus((currentStatus) => {
        const current = currentStatus
          ? currentStatus.split(DELIMITER).filter(excludeFalsyValues)
          : []
        const removals = new Set(normalize(toRemove))
        const next = current.filter(
          (s) => !removals.has(s as keyof typeof ActionPhase),
        )
        return next.length > 0 ? next.join(DELIMITER) + DELIMITER : null
      })
    },
    [setStatus],
  )

  const addStatusFilters = useCallback(
    (statuses: PhaseKey | PhaseKey[]) => {
      handleStatusAdd(statuses)
    },
    [handleStatusAdd],
  )

  const removeStatusFilters = useCallback(
    (statuses: PhaseKey | PhaseKey[]) => {
      handleStatusRemove(statuses)
    },
    [handleStatusRemove],
  )

  const toggleFilter = useCallback(
    (filter: StatusFilter) => {
      if (filter.type === 'status') {
        handleStatusUpdate(filter.status)
      }
    },
    [handleStatusUpdate],
  )

  const clearFilter = useCallback(() => {
    setStatus(null)
  }, [setStatus])

  const filters: { status: PhaseKey[] } = useMemo(
    () => ({
      status: statusValue
        ? (statusValue
            .split(DELIMITER)
            .filter(excludeFalsyValues) as PhaseKey[])
        : [],
    }),
    [statusValue],
  )

  return {
    addStatusFilters,
    removeStatusFilters,
    filters,
    clearFilter,
    toggleFilter,
  }
}