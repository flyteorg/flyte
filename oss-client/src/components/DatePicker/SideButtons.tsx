import React from 'react'
import clsx from 'clsx'
import { DateRange } from 'react-day-picker'
import { Button } from '../Button'

interface QuickRangeSidebarProps {
  lastClickedQuickKey: string | undefined
  sideButtons: SideButtonItem[]
  onSideButtonClick: (dateRange: DateRange, quickLabel: string) => void
}

export type SideButtonItem = SideButtonDivider | SideButtonType

export const isSideButton = (
  sideButtonItem: SideButtonItem,
): sideButtonItem is SideButtonType => {
  return (sideButtonItem as SideButtonType).displayText !== undefined
}

type SideButtonDivider = { key: string; type: 'divider' }

type SideButtonType = {
  displayText: string
  key: string
  onClick: () => DateRange
}

export const QuickRangeSidebar: React.FC<QuickRangeSidebarProps> = ({
  lastClickedQuickKey,
  sideButtons,
  onSideButtonClick,
}) => {
  return (
    <div className="flex h-full w-32 flex-col border-r-1 border-(--system-gray-3) px-3 py-2">
      {sideButtons.map((props: SideButtonItem) => {
        return isSideButton(props) ? (
          <SideButton
            key={props.displayText}
            isActive={lastClickedQuickKey === props.displayText}
            onClick={() => {
              const dateRange = props.onClick()
              onSideButtonClick(dateRange, props.displayText)
            }}
          >
            {props.displayText}
          </SideButton>
        ) : (
          <hr className="text-(--system-gray-5) opacity-25" key={props.key} />
        )
      })}
    </div>
  )
}

const SideButton = ({
  children,
  className,
  isActive,
  onClick,
}: {
  children: React.ReactNode
  className?: string
  isActive: boolean
  onClick: () => void
}) => {
  return (
    <Button
      className={clsx(
        `font-xs justify-start border-none !p-1 !font-medium ${className}`,
        isActive && 'dark:bg-(--system-gray-4)',
      )}
      color={isActive ? 'white' : 'zinc'}
      onClick={onClick}
      plain
      size="xs"
    >
      {children}
    </Button>
  )
}
