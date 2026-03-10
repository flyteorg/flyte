import { ArrowsRightIcon } from '@/components/icons/ArrowsRightIcon'
import { ChevronRightIcon } from '@/components/icons/ChevronRightIcon'
import { PopoverMenu, type MenuItem } from '@/components/Popovers'
import { ActionDetails } from '@/gen/flyteidl2/workflow/run_definition_pb'
import { useMemo, useState } from 'react'
import { useSelectedAttemptStore } from './state/AttemptStore'

export const AttemptPicker = ({
  actionDetails,
}: {
  actionDetails: ActionDetails | undefined
}) => {
  const { selectedAttemptNumber, setSelectedAttempt } =
    useSelectedAttemptStore()
  const attemptsCount = actionDetails?.attempts?.length || 0
  const [isOpen, setIsOpen] = useState(false)

  const menuItems: MenuItem[] = useMemo(() => {
    const items: MenuItem[] = []

    // Add header with attempt count
    items.push({
      id: 'header',
      type: 'custom',
      component: (
        <div className="px-2 py-0.5 text-sm font-semibold text-(--system-gray-7)">
          Attempts ({attemptsCount})
        </div>
      ),
    })

    // Add divider
    items.push({
      id: 'divider',
      type: 'divider',
    })

    for (let i = 1; i <= attemptsCount; i++) {
      items.push({
        id: `Attempt-${i}`,
        type: 'item',
        label: `Attempt ${i}`,
        selected: i === selectedAttemptNumber,
        onClick: () => {
          if (!actionDetails) return
          const number = Number(i)
          const attempt = actionDetails.attempts.find(
            (a) => a.attempt === number,
          )
          if (attempt) {
            setSelectedAttempt({ attempt, isManual: true })
          }
        },
      })
    }
    return items
  }, [actionDetails, attemptsCount, selectedAttemptNumber, setSelectedAttempt])

  if (!actionDetails?.attempts?.length) return null

  return (
    <PopoverMenu
      items={menuItems}
      showChevron={false}
      showCheckboxes={false}
      outline={false}
      size="sm"
      variant="dropdown"
      open={isOpen}
      onOpenChange={setIsOpen}
      itemClassName="!py-1 -my-1"
      noSelectedBackground={true}
    >
      <div
        data-testid="attempt-picker-trigger"
        className="flex cursor-pointer items-center gap-1 text-sm text-(--system-gray-5)"
      >
        <ArrowsRightIcon />
        Attempt {selectedAttemptNumber}/{attemptsCount}
        <div className="ml-1 flex h-4 w-4 items-center justify-center">
          <ChevronRightIcon className="rotate-90 transition-all" width={8} />
        </div>
      </div>
    </PopoverMenu>
  )
}
