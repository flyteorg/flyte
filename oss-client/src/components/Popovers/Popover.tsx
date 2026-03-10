import React, { useState, ReactNode, ReactElement, useId } from 'react'
import {
  useFloating,
  autoUpdate,
  offset,
  flip,
  shift,
  useHover,
  useClick,
  useDismiss,
  useRole,
  useInteractions,
  FloatingPortal,
  Placement,
  OffsetOptions,
} from '@floating-ui/react'
import { usePopoverStore } from './PopoverStore'

export interface PopoverProps {
  children: ReactElement
  content: ReactNode
  placement?: Placement
  offset?: OffsetOptions
  trigger?: 'click' | 'hover'
  disabled?: boolean
  open?: boolean
  onOpenChange?: (open: boolean) => void
  portal?: boolean
  className?: string
  contentClassName?: string
  noFlip?: boolean
  noShift?: boolean
}

/**
 * unstyled popover component. Most consumers will want to use PopoverMenu but
 * this component is intended as a reusable building block for non-menu popovers
 * (e.g an info popover that appears on click)
 *  */
export const Popover: React.FC<PopoverProps> = ({
  children,
  content,
  placement = 'bottom',
  offset: offsetValue = 5,
  trigger = 'click',
  disabled = false,
  open: controlledOpen,
  onOpenChange,
  portal = true,
  className = '',
  contentClassName = '',
  noFlip = false,
  noShift = false,
}) => {
  const { openId, setOpenId } = usePopoverStore()
  const [uncontrolledOpen, setUncontrolledOpen] = useState(false)
  const id = useId()

  // Use controlled state if provided, otherwise use internal state
  const isStoreOpen = openId === id
  const isOpen =
    controlledOpen !== undefined
      ? controlledOpen
      : isStoreOpen || uncontrolledOpen

  const setIsOpen = (next: boolean) => {
    if (onOpenChange) {
      onOpenChange(next)
    } else {
      setUncontrolledOpen(next)
    }

    // Update global store so only one stays open
    setOpenId(next ? id : null)
  }

  const { refs, floatingStyles, context } = useFloating({
    open: isOpen,
    onOpenChange: setIsOpen,
    placement,
    middleware: [
      offset(offsetValue),
      ...(noFlip ? [] : [flip({ fallbackAxisSideDirection: 'start' })]),
      ...(noShift ? [] : [shift()]),
    ],
    whileElementsMounted: autoUpdate,
  })

  const hover = useHover(context, {
    enabled: trigger === 'hover' && !disabled,
    delay: { open: 100, close: 300 },
  })

  const click = useClick(context, {
    enabled: trigger === 'click' && !disabled,
  })

  const dismiss = useDismiss(context)
  const role = useRole(context)

  const { getReferenceProps, getFloatingProps } = useInteractions([
    hover,
    click,
    dismiss,
    role,
  ])

  const triggerElement = React.cloneElement(children, {
    ref: refs.setReference,
    ...getReferenceProps(),
    className:
      `${typeof children.props === 'object' && children.props !== null && 'className' in children.props ? (children.props.className as string) || '' : ''} ${className}`.trim(),
  } as React.HTMLAttributes<HTMLElement>)

  const popoverContent = (
    <div
      ref={refs.setFloating}
      style={floatingStyles}
      {...getFloatingProps()}
      className={`z-50 ${contentClassName}`}
    >
      {content}
    </div>
  )

  return (
    <>
      {triggerElement}
      {isOpen &&
        !disabled &&
        (portal ? (
          <FloatingPortal>{popoverContent}</FloatingPortal>
        ) : (
          popoverContent
        ))}
    </>
  )
}
