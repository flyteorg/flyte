import { PopoverMenu, type MenuItem } from '@/components/Popovers'
import { Switch } from '@/components/Switch'
import { useQueryState } from 'nuqs'
import React, { useCallback, useMemo } from 'react'

export const ProjectsHeaderDropdown: React.FC = () => {
  const [showArchived, setShowArchived] = useQueryState('showArchived')

  const isEnabled = useMemo(() => showArchived === 'yes', [showArchived])

  const setEnabled = useCallback(
    (state: boolean) => {
      setShowArchived(state ? 'yes' : null)
    },
    [setShowArchived],
  )

  const items: MenuItem[] = [
    {
      id: 'projects-header-dropdown',
      type: 'custom',
      component: (
        <div className="flex flex-row items-center gap-2 px-3 py-1 text-sm/5">
          <Switch checked={isEnabled} onChange={setEnabled} />
          Show archived
        </div>
      ),
    },
  ]

  return (
    <PopoverMenu closeOnItemClick={false} items={items} variant="overflow" />
  )
}
