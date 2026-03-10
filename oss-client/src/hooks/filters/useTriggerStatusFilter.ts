import { useCallback, useMemo } from 'react'
import { useQueryState } from 'nuqs'

type TriggerStatus = 'active' | 'inactive'

interface TriggerStatusFilter {
  activeStatus: TriggerStatus | null
  setActiveStatus: (status: TriggerStatus | null) => void
  toggleStatus: (status: TriggerStatus) => void
  clearFilter: () => void
}

export const useTriggerStatusFilter = (): TriggerStatusFilter => {
  const [activeStatus, setActiveStatus] = useQueryState<TriggerStatus | null>(
    'status',
    {
      parse: (value) => {
        if (value === 'active' || value === 'inactive') {
          return value
        }
        return null
      },
      serialize: (value) => value || '',
      defaultValue: null,
    },
  )

  const toggleStatus = useCallback(
    (status: TriggerStatus) => {
      setActiveStatus((current) => {
        if (current === status) {
          return null
        }
        return status
      })
    },
    [setActiveStatus],
  )

  const clearFilter = useCallback(() => {
    setActiveStatus(null)
  }, [setActiveStatus])

  return useMemo(
    () => ({
      activeStatus,
      setActiveStatus,
      toggleStatus,
      clearFilter,
    }),
    [activeStatus, setActiveStatus, toggleStatus, clearFilter],
  )
}
