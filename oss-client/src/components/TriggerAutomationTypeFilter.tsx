import { PopoverMenu, type MenuItem } from '@/components/Popovers'
import { TriggerAutomationSpecType } from '@/gen/flyteidl2/task/common_pb'
import { useTriggerAutomationTypeFilter } from '@/hooks/filters/useTriggerAutomationTypeFilter'
import { useMemo } from 'react'
import { CalendarIcon } from './icons/CalendarIcon'

type AutomationType = Exclude<
  keyof typeof TriggerAutomationSpecType,
  'TYPE_UNSPECIFIED'
>

type FilterConfig = {
  label: string
  value: AutomationType
  enumValue: TriggerAutomationSpecType
  icon?: React.ReactNode
}

export const filterConfigs: FilterConfig[] = [
  {
    label: 'Schedule',
    value: TriggerAutomationSpecType[
      TriggerAutomationSpecType.TYPE_SCHEDULE
    ] as AutomationType,
    enumValue: TriggerAutomationSpecType.TYPE_SCHEDULE,
    icon: <CalendarIcon />,
  },
  {
    label: 'No automation',
    value: TriggerAutomationSpecType[
      TriggerAutomationSpecType.TYPE_NONE
    ] as AutomationType,
    enumValue: TriggerAutomationSpecType.TYPE_NONE,
  },
]

export const useTriggerAutomationTypeFilterMenuItems = () => {
  const { selectedTypes, toggleType } = useTriggerAutomationTypeFilter()
  const menuItems: MenuItem[] = useMemo(() => {
    return filterConfigs.map((config) => ({
      id: config.label,
      label: config.label,
      onClick: () => toggleType(config.value),
      selected: selectedTypes.includes(config.value),
      type: 'item',
      icon: config.icon ? (
        <div className="flex h-full w-full items-center justify-center">
          {config.icon}
        </div>
      ) : undefined,
    }))
  }, [selectedTypes, toggleType])
  return menuItems
}

export const TriggerAutomationTypeFilter = () => {
  const { selectedTypes, clearFilter } = useTriggerAutomationTypeFilter()
  const menuItems = useTriggerAutomationTypeFilterMenuItems()

  const displayedValues = useMemo(() => {
    if (selectedTypes.length === 0) return null

    return (
      <div className="flex items-center gap-1">
        {selectedTypes.slice(0, 2).map((type, index) => {
          const config = filterConfigs.find((c) => c.value === type)
          if (!config) return null

          return (
            <span key={type} className="flex items-center gap-1 text-xs">
              {config.icon ? (
                <span className="flex h-3 w-3 items-center justify-center">
                  {config.icon}
                </span>
              ) : (
                <span>{config.label}</span>
              )}
              {index < selectedTypes.slice(0, 2).length - 1 && (
                <span className="mr-0.5">;</span>
              )}
            </span>
          )
        })}
      </div>
    )
  }, [selectedTypes])

  return (
    <PopoverMenu
      label="Automation type"
      items={menuItems}
      variant="filter"
      menuClassName="min-w-56"
      filterProps={{
        displayedValues: displayedValues || <></>,
        valuesCount: selectedTypes.length,
        onClearClick: () => clearFilter(),
      }}
      closeOnItemClick={false}
    />
  )
}
