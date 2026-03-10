import { PopoverMenu } from '@/components/Popovers/PopoverMenu.tsx'
// import { BUTTON_MAX_HEIGHT } from './constants.ts'
import { useShowArchived } from '@/hooks/useQueryParamState.ts'
import { Switch } from '@/components/Switch'

const DropdownToggle = () => {
  const { shouldShowArchived, toggleShouldShowArchived } = useShowArchived()
  return (
    <div className="flex gap-2">
      Show archived
      <Switch
        checked={!!shouldShowArchived}
        onChange={toggleShouldShowArchived}
      />
    </div>
  )
}

export const ListRunsArchivedFilter = () => {
  const { shouldShowArchived } = useShowArchived()
  return (
    <PopoverMenu
      items={[
        {
          id: 'dropdown-toggle',
          type: 'custom',
          component: <DropdownToggle />,
        },
      ]}
      outline={!shouldShowArchived}
      variant="overflow"
    ></PopoverMenu>
  )
}
