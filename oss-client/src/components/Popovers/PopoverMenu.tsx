import { EllipsisHorizontalIcon } from '@heroicons/react/16/solid'
import { EllipsisVerticalIcon } from '@heroicons/react/20/solid'
import { XMarkIcon } from '@heroicons/react/24/outline'
import clsx from 'clsx'
import React, { ReactNode, useLayoutEffect, useRef } from 'react'
import { CheckedBoxIcon, UncheckedBoxIcon } from '../icons/CheckboxIcons'
import { ChevronDownIcon } from '../icons/ChevronDownIcon'
import { Popover, PopoverProps } from './Popover'
import { usePopoverStore } from './PopoverStore'

export interface MenuItem {
  className?: string
  component?: ReactNode // used for "custom" type
  disabled?: boolean
  icon?: ReactNode
  id: string
  label?: ReactNode
  onClick?: (e: React.MouseEvent<HTMLButtonElement>) => void
  selected?: boolean
  type?: 'item' | 'divider' | 'custom'
}

export interface PopoverMenuProps extends Omit<
  PopoverProps,
  'content' | 'children'
> {
  items: MenuItem[]
  children?: React.ReactElement
  closeOnItemClick?: boolean | 'default-only'
  disabled?: boolean
  variant?: 'dropdown' | 'overflow' | 'filter'
  label?: string | ReactNode
  outline?: boolean
  size?: 'xs' | 'sm' | 'md' | 'lg'
  menuClassName?: string
  open?: boolean
  onOpenChange?: (open: boolean) => void
  itemClassName?: string
  itemCustomClassName?: string
  triggerClassName?: string
  showChevron?: boolean
  showCheckboxes?: boolean
  noSelectedBackground?: boolean
  width?: 'auto' | 'trigger'
  filterProps?: {
    maxDisplayedValues?: number // max number of values to display before "and X others"
    displayedValues: React.ReactNode // first few selected values to display (3 by default)
    valuesCount: number // total count of selected values
    onClearClick: () => void // clear filter
  }
  overflowProps?: {
    orientation?: 'horizontal' | 'vertical'
  }
}

export const PopoverMenu: React.FC<PopoverMenuProps> = ({
  items,
  children,
  closeOnItemClick = true,
  disabled,
  variant = 'dropdown',
  label,
  open,
  onOpenChange,
  outline = false,
  size = 'md',
  menuClassName = '',
  itemClassName = '',
  itemCustomClassName = '',
  triggerClassName = '',
  showChevron = true,
  showCheckboxes = true,
  noSelectedBackground = false,
  width = 'auto',
  placement = 'bottom-start',
  filterProps,
  overflowProps,
  ...popoverProps
}) => {
  // Use controlled state if open/onOpenChange are provided, otherwise use internal state
  const [internalOpen, setInternalOpen] = React.useState(false)
  const isControlled = open !== undefined && onOpenChange !== undefined
  const isOpen = isControlled ? open : internalOpen
  const setIsOpen = isControlled ? onOpenChange : setInternalOpen
  const { setOpenId } = usePopoverStore()

  const menuRef = useRef<HTMLDivElement>(null)
  const scrollPositionRef = useRef<number>(0)
  const sizeClasses = {
    xs: 'px-1 py-1 text-xs',
    sm: 'px-2 py-1 text-sm',
    md: 'px-3 py-2 text-sm',
    lg: 'px-4 py-2 text-md',
  }

  // Reset scroll position when menu closes
  React.useEffect(() => {
    if (!isOpen) {
      scrollPositionRef.current = 0
      setOpenId(null)
    }
  }, [isOpen, setOpenId])

  // Preserve scroll position when items change but menu stays open
  useLayoutEffect(() => {
    if (isOpen && menuRef.current) {
      // Restore scroll position after items update
      menuRef.current.scrollTop = scrollPositionRef.current
    }
  }, [items, isOpen])

  // Store scroll position before updates
  const handleScroll = React.useCallback(() => {
    if (menuRef.current) {
      scrollPositionRef.current = menuRef.current.scrollTop
    }
  }, [])

  const handleItemClick = React.useCallback(
    (item: MenuItem, event: React.MouseEvent<HTMLButtonElement>) => {
      // Determine if we should close the menu
      const shouldClose =
        closeOnItemClick === true || // Always close
        (closeOnItemClick === 'default-only' && !item.component) // Close only for default items

      // Call the original onClick if it exists
      if (item.onClick) {
        item.onClick(event)
      }

      // Close the menu if configured to do so
      if (shouldClose) {
        setIsOpen(false)
      }
    },
    [closeOnItemClick, setIsOpen],
  )

  // Update the filter variant clear button click handler
  const handleClearClick = React.useCallback(
    (e: React.MouseEvent) => {
      filterProps?.onClearClick?.()
      if (!isOpen) {
        e.stopPropagation() // to not open dropdown
      }
    },
    [filterProps, isOpen],
  )

  const renderTrigger = () => {
    const baseProps = {
      disabled,
      'aria-disabled': disabled,
      tabIndex: disabled ? -1 : 0,
      className: clsx(
        `inline-flex cursor-pointer items-center gap-2 rounded-lg transition-colors`,
        outline && 'border border-(--system-gray-1) bg-white ',
        sizeClasses[size],
        triggerClassName,
        disabled && 'opacity-50 pointer-events-none select-none',
      ),
    }

    if (children) {
      return children
    }

    if (variant === 'overflow') {
      const orientation = overflowProps?.orientation || 'vertical'
      const overflowSizeClasses = {
        xs: 'h-4 w-4',
        sm: 'h-4 w-4',
        md: 'h-6 w-6',
        lg: 'h-8 w-8',
      }
      return (
        <button
          className={clsx(
            `inline-flex cursor-pointer items-center justify-center rounded-lg`,
            overflowSizeClasses[size],
            triggerClassName,
          )}
        >
          {orientation === 'vertical' ? (
            <EllipsisVerticalIcon
              className="leading-5 text-(--system-gray-5)"
              width="16px"
            />
          ) : (
            <EllipsisHorizontalIcon
              className="leading-5 text-(--system-gray-5)"
              width="16px"
            />
          )}
        </button>
      )
    }

    if (variant === 'filter' && filterProps) {
      const {
        displayedValues,
        valuesCount,
        maxDisplayedValues = 3,
      } = filterProps

      const otherValuesCount =
        valuesCount - maxDisplayedValues > 0
          ? valuesCount - maxDisplayedValues
          : 0

      return (
        <button
          className={clsx(
            `inline-flex h-6 cursor-pointer items-center gap-1 rounded-lg px-2 py-0.5 transition-colors focus:outline-none`,
            'border-[1.5px] border-(--system-gray-3)',
            `text-xs font-medium ${valuesCount ? 'text-(--system-white)' : isOpen ? 'dark:text-(--system-gray-7)' : 'text-(--accent-gray) dark:text-(--system-gray-6)'}`,
            triggerClassName,
          )}
        >
          <span>
            {label}
            {valuesCount ? ': ' : ''}
          </span>

          {valuesCount ? (
            <>
              {displayedValues}
              {valuesCount > maxDisplayedValues ? (
                <span>
                  and {otherValuesCount} other{otherValuesCount > 1 ? 's' : ''}
                </span>
              ) : null}
              <span
                onClick={handleClearClick}
                className="cursor-pointer text-(--system-gray-5)"
                aria-label="Clear"
              >
                <XMarkIcon data-slot="icon" width={16} aria-hidden="true" />
              </span>
            </>
          ) : (
            showChevron && (
              <ChevronDownIcon
                className={`text-(--system-gray-5) transition-transform ${isOpen ? 'rotate-180' : ''}`}
              />
            )
          )}
        </button>
      )
    }

    // variant === "dropdown"
    return (
      <button {...baseProps}>
        <span>{label}</span>
        {showChevron && (
          <ChevronDownIcon
            className={`transition-transform ${isOpen ? 'rotate-180' : ''}`}
          />
        )}
      </button>
    )
  }

  const getMenuWidth = () => {
    if (width === 'auto') return 'min-w-40'
    if (width === 'trigger') return 'w-full'
    return ''
  }

  const menuContent = (
    <div
      data-testid="popover-menu"
      ref={menuRef}
      onScroll={handleScroll}
      className={clsx(
        `overflow-hidden rounded-xl border border-(--system-gray-3) bg-white py-1 shadow-lg dark:bg-(--system-gray-2) dark:text-white`,
        getMenuWidth(),
        menuClassName,
      )}
    >
      {items.map((item, index) => {
        if (item.type === 'divider') {
          return (
            <div
              key={`separator-${index}`}
              className="pointer-events-none mx-4.5 my-1 border-t border-(--system-gray-1) dark:border-(--system-gray-3)"
            />
          )
        }

        if (item.type === 'custom') {
          return (
            <div
              key={item.id}
              className={`px-1 py-1 dark:text-white ${itemCustomClassName}`}
            >
              {item.component}
            </div>
          )
        }

        // item.type === "item"
        return (
          <button
            data-testid={`popover-item-${item.id}`}
            data-checked={!!item.selected}
            key={item.id}
            onClick={(e) => handleItemClick(item, e)}
            disabled={item.disabled}
            className={clsx(
              `flex h-full w-full cursor-pointer items-center gap-1 px-4.5 text-left text-sm transition-colors`,
              item.disabled ? 'cursor-default' : 'hover:bg-(--system-gray-3)',
              variant === 'filter' ? 'py-0.5 text-(--system-gray-6)' : 'py-2',
              !showCheckboxes ? 'text-(--system-gray-5)' : '',
              variant === 'filter' && item.selected
                ? 'bg-(--system-gray-3)'
                : '',
              !noSelectedBackground && !showCheckboxes && item.selected
                ? 'bg-(--system-gray-3)'
                : '',
              itemClassName,
              item.className,
            )}
          >
            <div className="flex flex-1 items-center gap-2">
              {showCheckboxes && typeof item.selected === 'boolean' ? (
                <>
                  {item.selected ? (
                    <CheckedBoxIcon className="text-(--system-white)" />
                  ) : (
                    <UncheckedBoxIcon />
                  )}
                </>
              ) : null}
              {item.icon && (
                <span className="h-4 w-4 flex-shrink-0">{item.icon}</span>
              )}
              <span
                className={clsx(
                  'flex items-center gap-2',
                  !showCheckboxes && item.selected
                    ? 'text-(--system-white)'
                    : item.selected
                      ? 'text-(--system-white)'
                      : '',
                )}
              >
                {item.label}
              </span>
            </div>
          </button>
        )
      })}
    </div>
  )

  return (
    <Popover
      content={menuContent}
      disabled={disabled}
      placement={placement}
      open={isOpen}
      onOpenChange={setIsOpen}
      {...popoverProps}
    >
      {renderTrigger()}
    </Popover>
  )
}
