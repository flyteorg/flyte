import { useQueryState } from 'nuqs'
import { useCallback, useMemo } from 'react'

type TimePeriod =
  | 'last_hour'
  | 'past_24_hours'
  | 'past_week'
  | 'past_month'
  | 'past_year'

interface TriggerTimeFilter {
  selectedPeriod: TimePeriod | null
  setSelectedPeriod: (period: TimePeriod | null) => void
  togglePeriod: (period: TimePeriod) => void
  clearFilter: () => void
}

export const useTriggerTimeFilter = (queryKey: string): TriggerTimeFilter => {
  const [selectedPeriod, setSelectedPeriod] = useQueryState<TimePeriod | null>(
    queryKey,
    {
      parse: (value) => {
        if (
          value === 'last_hour' ||
          value === 'past_24_hours' ||
          value === 'past_week' ||
          value === 'past_month' ||
          value === 'past_year'
        ) {
          return value
        }
        return null
      },
      serialize: (value) => value || '',
      defaultValue: null,
    },
  )

  const togglePeriod = useCallback(
    (period: TimePeriod) => {
      setSelectedPeriod((current) => {
        if (current === period) {
          return null
        }
        return period
      })
    },
    [setSelectedPeriod],
  )

  const clearFilter = useCallback(() => {
    setSelectedPeriod(null)
  }, [setSelectedPeriod])

  return useMemo(
    () => ({
      selectedPeriod,
      setSelectedPeriod,
      togglePeriod,
      clearFilter,
    }),
    [selectedPeriod, setSelectedPeriod, togglePeriod, clearFilter],
  )
}
