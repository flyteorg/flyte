import { PopoverMenu, type MenuItem } from '@/components/Popovers'
import { useTriggerTimeFilter } from '@/hooks/filters/useTriggerTimeFilter'
import { useMemo } from 'react'

type TimePeriod =
  | 'last_hour'
  | 'past_24_hours'
  | 'past_week'
  | 'past_month'
  | 'past_year'

type FilterConfig = {
  label: string
  value: TimePeriod
}

export const filterConfigs: FilterConfig[] = [
  {
    label: 'Last hour',
    value: 'last_hour',
  },
  {
    label: 'Past 24 hours',
    value: 'past_24_hours',
  },
  {
    label: 'Past week',
    value: 'past_week',
  },
  {
    label: 'Past month',
    value: 'past_month',
  },
  {
    label: 'Past year',
    value: 'past_year',
  },
]

export interface TriggerTimeFilterProps {
  label: string
  queryKey: string
}

export const RelativeTimeFilter = ({
  label,
  queryKey,
}: TriggerTimeFilterProps) => {
  const { selectedPeriod, togglePeriod, clearFilter } =
    useTriggerTimeFilter(queryKey)

  const menuItems: MenuItem[] = useMemo(() => {
    return filterConfigs.map((config) => ({
      id: config.value,
      label: config.label,
      onClick: () => togglePeriod(config.value),
      type: 'item',
    }))
  }, [togglePeriod])

  const displayedValue = useMemo(() => {
    if (!selectedPeriod) return null
    const config = filterConfigs.find((c) => c.value === selectedPeriod)
    if (!config) return null
    return <span className="text-xs">{config.label}</span>
  }, [selectedPeriod])

  return (
    <PopoverMenu
      label={label}
      items={menuItems}
      variant="filter"
      menuClassName="min-w-56"
      filterProps={{
        displayedValues: displayedValue || <></>,
        valuesCount: selectedPeriod ? 1 : 0,
        onClearClick: () => clearFilter(),
      }}
      closeOnItemClick={true}
    />
  )
}
