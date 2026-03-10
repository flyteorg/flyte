import { ChevronDownIcon } from '@heroicons/react/20/solid'
import React, { useEffect, useMemo, useRef, useState } from 'react'
import { Button, ButtonProps } from './Button'

interface ComboButtonOption {
  name: string | React.ReactElement
  onClick: () => void
}

export interface ComboButtonProps {
  color?: ButtonProps['color']
  size?: ButtonProps['size']
  options: ComboButtonOption[]
  outline?: boolean
  firstButtonFilled?: boolean
}

export default function ComboButton({
  color,
  size,
  options,
  outline,
  firstButtonFilled,
}: ComboButtonProps) {
  const [open, setOpen] = useState(false)
  // for outside click detection
  const menuRef = useRef<HTMLDivElement>(null)
  // for focus management & outside click detection
  const buttonRef = useRef<HTMLButtonElement>(null)

  // close dropdown on outside click
  useEffect(() => {
    if (!open) return
    function handleClick(e: MouseEvent) {
      if (
        menuRef.current &&
        !menuRef.current.contains(e.target as Node) &&
        buttonRef.current &&
        !buttonRef.current.contains(e.target as Node)
      ) {
        setOpen(false)
      }
    }
    document.addEventListener('mousedown', handleClick)
    return () => document.removeEventListener('mousedown', handleClick)
  }, [open])

  // keyboard accessibility
  function handleButtonKeyDown(e: React.KeyboardEvent) {
    if (e.key === 'ArrowDown' || e.key === 'Enter' || e.key === ' ') {
      e.preventDefault()
      setOpen((o) => !o)
    }
  }
  function handleMenuKeyDown(e: React.KeyboardEvent) {
    if (e.key === 'Escape') {
      setOpen(false)
      buttonRef.current?.focus()
    }
  }

  const chevronClassNames = useMemo(
    () =>
      `h-4 w-4 ${firstButtonFilled || !outline ? '!text-(--system-black)' : '!text-(--system-gray-5)'}`,
    [outline, firstButtonFilled],
  )

  const btnClassNames = useMemo(() => {
    if (firstButtonFilled && outline) {
      // When first button is filled but we want separator, keep the right border
      return '!border-y-0 border-l-0 border-r-1 border-(--system-gray-4) rounded-r-none !px-6 [&::after]:rounded-none [&::after]:shadow-none [&::before]:shadow-none'
    }
    return `${outline ? '!border-y-0 border-l-0  border-r-1 border-(--system-gray-4)' : 'border-none'} rounded-r-none !px-6 [&::after]:rounded-none [&::after]:shadow-none [&::before]:shadow-none`
  }, [outline, firstButtonFilled])

  return (
    <div className="relative inline-flex">
      <div className="inline-flex overflow-hidden rounded-md bg-transparent">
        <Button
          type="button"
          className={btnClassNames}
          size={size}
          color={color}
          outline={firstButtonFilled ? false : outline}
          onClick={options[0].onClick}
        >
          {options[0].name}
        </Button>
        {options.length > 1 && (
          <Button
            outline={firstButtonFilled ? false : outline}
            ref={buttonRef}
            type="button"
            aria-haspopup="menu"
            aria-expanded={open}
            onClick={() => setOpen((o) => !o)}
            onKeyDown={handleButtonKeyDown}
            className="rounded-none border-none [&::after]:rounded-none [&::after]:border-l-1 [&::after]:border-zinc-100 [&::after]:shadow-none dark:[&::after]:border-(--system-black) [&::before]:shadow-none"
            size={size}
            color={color}
            tabIndex={0}
            disabled={options.length <= 1}
          >
            <span className="sr-only">Open options</span>
            <ChevronDownIcon className={chevronClassNames} aria-hidden="true" />
          </Button>
        )}
      </div>
      {open && options.length > 1 && (
        <div
          ref={menuRef}
          className="bg-secondary absolute top-full right-0 z-50 mt-2 w-40 origin-top-right rounded-md shadow-lg ring-1 ring-black/5 focus:outline-none"
          role="menu"
          tabIndex={-1}
          onKeyDown={handleMenuKeyDown}
        >
          <div className="py-1">
            {options.slice(1).map((item, idx) => (
              <button
                key={typeof item.name === 'string' ? item.name : idx}
                onClick={() => {
                  setOpen(false)
                  item.onClick()
                }}
                className="bg-secondary block w-full cursor-pointer rounded-none px-4 py-2 text-left text-sm font-semibold text-(--system-white) hover:bg-(--system-gray-2)"
                type="button"
                role="menuitem"
              >
                {item.name}
              </button>
            ))}
          </div>
        </div>
      )}
    </div>
  )
}
