import { Filter, Filter_Function } from '@/gen/flyteidl2/common/list_pb'
import { getFilter } from '@/lib/filterUtils'
import dayjs from 'dayjs'
import utc from 'dayjs/plugin/utc'
import { useCallback, useMemo } from 'react'
import { useOwnerFilter } from './useOwnerFilter'
import { useTriggerAutomationTypeFilter } from './useTriggerAutomationTypeFilter'
import { useTriggerStatusFilter } from './useTriggerStatusFilter'
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
 * Hook that encapsulates all trigger filter logic and returns the combined filters array,
 * whether filters are active, and a function to clear all filters
 */
export const useTriggerAPIFilters = () => {
  const { activeStatus, clearFilter: clearStatusFilter } =
    useTriggerStatusFilter()
  const { selectedTypes, clearFilter: clearTypeFilter } =
    useTriggerAutomationTypeFilter()
  const { selectedOwners, clearFilter: clearOwnerFilter } =
    useOwnerFilter('owner')
  const {
    selectedOwners: selectedUpdatedBy,
    clearFilter: clearUpdatedByFilter,
  } = useOwnerFilter('updatedBy')
  const {
    selectedPeriod: selectedTriggeredAt,
    clearFilter: clearTriggeredAtFilter,
  } = useTriggerTimeFilter('triggered_at')
  const {
    selectedPeriod: selectedUpdatedAt,
    clearFilter: clearUpdatedAtFilter,
  } = useTriggerTimeFilter('updated_at')

  const filters = useMemo(() => {
    const filterList: (Filter | undefined)[] = []

    if (activeStatus) {
      filterList.push(
        getFilter({
          function: Filter_Function.EQUAL,
          field: 'active',
          values: [activeStatus === 'active' ? 'true' : 'false'],
        }),
      )
    }

    if (selectedTypes.length > 0) {
      filterList.push(
        getFilter({
          function: Filter_Function.VALUE_IN,
          field: 'automation_type',
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

    if (selectedUpdatedBy.length > 0) {
      filterList.push(
        getFilter({
          function: Filter_Function.VALUE_IN,
          field: 'updated_by',
          values: selectedUpdatedBy,
        }),
      )
    }

    if (selectedTriggeredAt) {
      const timestamp = getTimePeriodTimestamp(selectedTriggeredAt)
      filterList.push(
        getFilter({
          function: Filter_Function.GREATER_THAN_OR_EQUAL,
          field: 'triggered_at',
          values: [timestamp],
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
  }, [
    activeStatus,
    selectedTypes,
    selectedOwners,
    selectedUpdatedBy,
    selectedTriggeredAt,
    selectedUpdatedAt,
  ])

  const hasActiveFilters = useMemo(() => {
    return (
      !!activeStatus ||
      selectedTypes.length > 0 ||
      selectedOwners.length > 0 ||
      selectedUpdatedBy.length > 0 ||
      !!selectedTriggeredAt ||
      !!selectedUpdatedAt
    )
  }, [
    activeStatus,
    selectedTypes.length,
    selectedOwners.length,
    selectedUpdatedBy.length,
    selectedTriggeredAt,
    selectedUpdatedAt,
  ])

  const clearAllFilters = useCallback(() => {
    clearStatusFilter()
    clearTypeFilter()
    clearOwnerFilter()
    clearUpdatedByFilter()
    clearTriggeredAtFilter()
    clearUpdatedAtFilter()
  }, [
    clearStatusFilter,
    clearTypeFilter,
    clearOwnerFilter,
    clearUpdatedByFilter,
    clearTriggeredAtFilter,
    clearUpdatedAtFilter,
  ])

  return {
    filters,
    hasActiveFilters,
    clearAllFilters,
  }
}
