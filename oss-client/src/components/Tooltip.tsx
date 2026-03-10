'use client'
import { usePopoverStore } from '@/components/Popovers'
import { useEffect } from 'react'
import {
  useFloating,
  offset,
  shift,
  flip,
  useHover,
  useFocus,
  useRole,
  useDismiss,
  useInteractions,
  FloatingPortal,
  Placement,
  useClientPoint,
  OffsetOptions,
} from '@floating-ui/react'
import { useState, cloneElement, isValidElement } from 'react'
import clsx from 'clsx'

export type TooltipProps = {
  content: string | React.ReactNode
  children: React.ReactNode
  closeDelay?: number
  offsetProp?: number | OffsetOptions
  openDelay?: number
  placement?: Placement | undefined
  shouldUseClientPoint?: boolean
  disabled?: boolean
  contentClassName?: string
}

export function Tooltip({
  content,
  children,
  closeDelay = 100,
  offsetProp = 8,
  openDelay = 100,
  placement = 'top',
  shouldUseClientPoint = false,
  disabled = false,
  contentClassName,
}: TooltipProps) {
  const [open, setOpen] = useState(false)
  const { openId } = usePopoverStore()

  const handleOpenChange = (next: boolean) => {
    if (disabled) return
    // Guard opening when a popover is open.
    if (next && openId) return
    setOpen(next)
  }

  const { refs, floatingStyles, context } = useFloating({
    open: disabled ? false : open,
    onOpenChange: handleOpenChange,
    middleware: [offset(offsetProp), flip(), shift()],
    placement,
  })

  // close tooltip if a popover opens
  useEffect(() => {
    if (openId) {
      setOpen(false)
    }
  }, [openId])

  const hover = useHover(context, {
    move: false,
    delay: { open: openDelay, close: closeDelay },
  })
  const focus = useFocus(context, {
    enabled: !disabled && !openId,
  })
  const dismiss = useDismiss(context, {
    enabled: !disabled,
  })
  const role = useRole(context, { role: 'tooltip' })
  const clientPoint = useClientPoint(context, {
    enabled: shouldUseClientPoint && !disabled,
  })

  const { getReferenceProps, getFloatingProps } = useInteractions([
    clientPoint,
    hover,
    focus,
    dismiss,
    role,
  ])

  const child = isValidElement(children) ? children : <span>{children}</span>

  if (disabled) {
    return <>{children}</>
  }

  return (
    <>
      {cloneElement(child, {
        ref: refs.setReference,
        ...getReferenceProps(),
      })}

      {open && (
        <FloatingPortal>
          <div
            ref={refs.setFloating}
            style={floatingStyles}
            className={`z-50 rounded-lg px-2 py-1 text-sm`}
            {...getFloatingProps()}
          >
            <div
              className={clsx(
                `bg-inverse-tertiary rounded-lg text-2xs/6 font-medium text-white dark:bg-zinc-700`,
                typeof content === 'string' && 'px-1.5',
                contentClassName,
              )}
            >
              {content}
            </div>
          </div>
        </FloatingPortal>
      )}
    </>
  )
}
