import { PopoverMenu, type MenuItem } from '@/components/Popovers'
import { StatusIcon } from '@/components/StatusIcons'
import { ActionPhase } from '@/gen/flyteidl2/common/phase_pb'
import { useQueryFilters } from '@/hooks/useQueryFilters'
import { useMemo } from 'react'

type FilterConfig = {
  label: string
  phase: ActionPhase
  value: keyof typeof ActionPhase
}

export const filterConfigs: FilterConfig[] = [
  {
    label: 'Completed',
    phase: ActionPhase.SUCCEEDED,
    value: 'SUCCEEDED',
  },
  {
    label: 'Queued',
    phase: ActionPhase.QUEUED,
    value: 'QUEUED',
  },
  {
    label: 'Waiting for resources',
    phase: ActionPhase.WAITING_FOR_RESOURCES,
    value: 'WAITING_FOR_RESOURCES',
  },
  {
    label: 'Initializing',
    phase: ActionPhase.INITIALIZING,
    value: 'INITIALIZING',
  },
  {
    label: 'Running',
    phase: ActionPhase.RUNNING,
    value: 'RUNNING',
  },
  {
    label: 'Timed out',
    phase: ActionPhase.TIMED_OUT,
    value: 'TIMED_OUT',
  },
  {
    label: 'Aborted',
    phase: ActionPhase.ABORTED,
    value: 'ABORTED',
  },
  {
    label: 'Failed',
    phase: ActionPhase.FAILED,
    value: 'FAILED',
  },
]

export const useStatusFilterMenuItems = () => {
  const { filters, toggleFilter } = useQueryFilters()
  const menuItems: MenuItem[] = useMemo(() => {
    return filterConfigs.map((config) => ({
      id: config.label,
      label: config.label,
      onClick: () => toggleFilter({ type: 'status', status: config.value }),
      selected: !!filters.status?.includes(config.value),
      type: 'item',
      icon: <StatusIcon phase={config.phase} isStatic={true} />,
    }))
  }, [filters.status, toggleFilter])
  return menuItems
}

export const StatusFilter = () => {
  const { filters, clearFilter } = useQueryFilters()
  const menuItems = useStatusFilterMenuItems()
  return (
    <PopoverMenu
      label="Status"
      items={menuItems}
      variant="filter"
      menuClassName="min-w-56"
      filterProps={{
        displayedValues: (
          <div className="flex items-center">
            {filters.status?.slice(0, 3).map((s) => (
              <StatusIcon
                key={s}
                phase={ActionPhase[s as keyof typeof ActionPhase]}
                isStatic={true}
              />
            ))}
          </div>
        ),
        valuesCount: filters.status?.length || 0,
        onClearClick: () => clearFilter(),
      }}
      closeOnItemClick={false}
    />
  )
}
