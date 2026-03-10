import { PopoverMenu, type MenuItem } from '@/components/Popovers'
import { StatusDot, StatusDotBadge } from '@/components/StatusDotBadge'
import { useTriggerStatusFilter } from '@/hooks/filters/useTriggerStatusFilter'
import { useMemo } from 'react'

type FilterConfig = {
  label: string
  value: 'active' | 'inactive'
  color: 'green' | 'gray'
  filled: boolean
}

export const filterConfigs: FilterConfig[] = [
  {
    label: 'Active',
    value: 'active',
    color: 'green',
    filled: true,
  },
  {
    label: 'Inactive',
    value: 'inactive',
    color: 'gray',
    filled: false,
  },
]

export const useTriggerStatusFilterMenuItems = () => {
  const { activeStatus, toggleStatus } = useTriggerStatusFilter()
  const menuItems: MenuItem[] = useMemo(() => {
    return filterConfigs.map((config) => ({
      id: config.label,
      label: config.label,
      onClick: () => toggleStatus(config.value),
      type: 'item',
      icon: (
        <div className="flex h-full w-full items-center justify-center">
          <StatusDot color={config.color} filled={config.filled} />
        </div>
      ),
    }))
  }, [activeStatus, toggleStatus])
  return menuItems
}

export const TriggerStatusFilter = () => {
  const { activeStatus, clearFilter } = useTriggerStatusFilter()
  const menuItems = useTriggerStatusFilterMenuItems()

  const displayedValue = useMemo(() => {
    if (!activeStatus) return null
    const config = filterConfigs.find((c) => c.value === activeStatus)
    if (!config) return null
    return (
      <StatusDotBadge
        color={config.color}
        displayText={config.label}
        filled={config.filled}
        className="bg-transparent dark:bg-transparent"
      />
    )
  }, [activeStatus])

  return (
    <PopoverMenu
      label="Status"
      items={menuItems}
      variant="filter"
      menuClassName="min-w-56"
      filterProps={{
        displayedValues: displayedValue || <></>,
        valuesCount: activeStatus ? 1 : 0,
        onClearClick: () => clearFilter(),
      }}
      closeOnItemClick={false}
    />
  )
}
