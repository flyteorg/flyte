import { Filter, Filter_Function } from '@/gen/flyteidl2/common/list_pb'
import { getFilter } from '@/lib/filterUtils'
import dayjs from 'dayjs'
import utc from 'dayjs/plugin/utc'
import { useCallback, useMemo } from 'react'
import { useOwnerFilter } from './useOwnerFilter'
import { useTriggerAutomationTypeFilter } from './useTriggerAutomationTypeFilter'
import { useTriggerTimeFilter } from './useTriggerTimeFilter'

dayjs.extend(utc)

/**
 * Converts a time period string to an ISO timestamp string in UTC
 */
const getTimePeriodTimestamp = (
  period:
    | 'last_hour'
    | 'past_24_hours'
    | 'past_week'
    | 'past_month'
    | 'past_year',
): string => {
  const now = dayjs.utc()
  let timestamp: dayjs.Dayjs

  switch (period) {
    case 'last_hour':
      timestamp = now.subtract(1, 'hour')
      break
    case 'past_24_hours':
      timestamp = now.subtract(24, 'hour')
      break
    case 'past_week':
      timestamp = now.subtract(7, 'day')
      break
    case 'past_month':
      timestamp = now.subtract(1, 'month')
      break
    case 'past_year':
      timestamp = now.subtract(1, 'year')
      break
    default:
      timestamp = now
  }

  return timestamp.toISOString()
}

/**
 * Hook that encapsulates all task filter logic and returns the combined filters array,
 * whether filters are active, and a function to clear all filters
 */
export const useTaskAPIFilters = () => {
  const { selectedTypes, clearFilter: clearTypeFilter } =
    useTriggerAutomationTypeFilter()
  const { selectedOwners, clearFilter: clearDeployedByFilter } =
    useOwnerFilter('deployed_by')
  const {
    selectedPeriod: selectedUpdatedAt,
    clearFilter: clearUpdatedAtFilter,
  } = useTriggerTimeFilter('updated_at')

  const filters = useMemo(() => {
    const filterList: (Filter | undefined)[] = []

    if (selectedTypes.length > 0) {
      filterList.push(
        getFilter({
          function: Filter_Function.VALUE_IN,
          field: 'trigger_types',
          values: selectedTypes,
        }),
      )
    }

    if (selectedOwners.length > 0) {
      filterList.push(
        getFilter({
          function: Filter_Function.VALUE_IN,
          field: 'deployed_by',
          values: selectedOwners,
        }),
      )
    }

    if (selectedUpdatedAt) {
      const timestamp = getTimePeriodTimestamp(selectedUpdatedAt)
      filterList.push(
        getFilter({
          function: Filter_Function.GREATER_THAN_OR_EQUAL,
          field: 'updated_at',
          values: [timestamp],
        }),
      )
    }

    return filterList.filter(Boolean) as Filter[]
  }, [selectedTypes, selectedOwners, selectedUpdatedAt])

  const hasActiveFilters = useMemo(() => {
    return (
      selectedTypes.length > 0 ||
      selectedOwners.length > 0 ||
      !!selectedUpdatedAt
    )
  }, [selectedTypes.length, selectedOwners.length, selectedUpdatedAt])

  const clearAllFilters = useCallback(() => {
    clearTypeFilter()
    clearDeployedByFilter()
    clearUpdatedAtFilter()
  }, [clearTypeFilter, clearDeployedByFilter, clearUpdatedAtFilter])

  return {
    filters,
    hasActiveFilters,
    clearAllFilters,
  }
}
