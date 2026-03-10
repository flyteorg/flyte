import * as Headless from '@headlessui/react'
import clsx from 'clsx'
import React, { forwardRef } from 'react'

export function InputGroup({
  children,
  className,
  size = 'md',
}: React.ComponentPropsWithoutRef<'span'> & { size?: 'sm' | 'md' }) {
  return (
    <span
      data-slot="control"
      className={clsx(
        className,
        'group relative isolate block',
        'has-[[data-slot=icon]:first-child]:[&_input]:pl-10 has-[[data-slot=icon]:last-child]:[&_input]:pr-10 sm:has-[[data-slot=icon]:first-child]:[&_input]:pl-8 sm:has-[[data-slot=icon]:last-child]:[&_input]:pr-8',
        '*:data-[slot=icon]:pointer-events-none *:data-[slot=icon]:absolute *:data-[slot=icon]:z-10 *:data-[slot=icon]:size-5 sm:*:data-[slot=icon]:size-4',
        size === 'sm' && '*:data-[slot=icon]:top-1.5',
        size === 'md' &&
          '*:data-[slot=icon]:top-3 sm:*:data-[slot=icon]:top-2.5',
        '[&>[data-slot=icon]:first-child]:left-3 sm:[&>[data-slot=icon]:first-child]:left-2.5 [&>[data-slot=icon]:last-child]:right-3 sm:[&>[data-slot=icon]:last-child]:right-2.5',
        '*:data-[slot=icon]:text-(--system-gray-5)',
      )}
    >
      {children}
    </span>
  )
}

const dateTypes = ['date', 'datetime-local', 'month', 'time', 'week']
type DateType = (typeof dateTypes)[number]

export type InputProps = {
  size?: 'sm' | 'md'
  className?: string
  typographyClassName?: string
  hideBorder?: boolean
  noBackground?: boolean
  /** When true, removes drop shadow and focus ring (e.g. for search bars) */
  noOutline?: boolean
  type?:
    | 'email'
    | 'number'
    | 'password'
    | 'search'
    | 'tel'
    | 'text'
    | 'url'
    | DateType
} & Omit<Headless.InputProps, 'as' | 'className' | 'size'>

export const Input = forwardRef(function Input(
  {
    className,
    hideBorder,
    noBackground,
    noOutline,
    typographyClassName,
    size = 'md',
    ...props
  }: InputProps,
  ref: React.ForwardedRef<HTMLInputElement>,
) {
  return (
    <span
      data-slot="control"
      className={clsx([
        // Basic layout
        'relative block w-full',
        // Background color + shadow applied to inset pseudo element, so shadow blends with border in light mode
        !noOutline && [
          'before:absolute before:inset-px before:rounded-[calc(var(--radius-lg)-1px)] before:bg-white before:shadow-sm',
          'dark:before:hidden',
        ],
        // Focus ring (none when noOutline)
        noOutline
          ? 'after:pointer-events-none after:absolute after:inset-0 after:rounded-lg after:ring-0'
          : 'after:pointer-events-none after:absolute after:inset-0 after:rounded-lg after:ring-transparent after:ring-inset sm:focus-within:after:ring-2 sm:focus-within:after:ring-blue-500',
        // Disabled state
        'has-data-disabled:opacity-50 has-data-disabled:before:bg-zinc-950/5 has-data-disabled:before:shadow-none',
        // Invalid state
        'has-data-invalid:before:shadow-red-500/10',
      ])}
    >
      <Headless.Input
        ref={ref}
        {...props}
        className={clsx([
          // Date classes
          props.type &&
            dateTypes.includes(props.type) && [
              '[&::-webkit-datetime-edit-fields-wrapper]:p-0',
              '[&::-webkit-date-and-time-value]:min-h-[1.5em]',
              '[&::-webkit-datetime-edit]:inline-flex',
              '[&::-webkit-datetime-edit]:p-0',
              '[&::-webkit-datetime-edit-year-field]:p-0',
              '[&::-webkit-datetime-edit-month-field]:p-0',
              '[&::-webkit-datetime-edit-day-field]:p-0',
              '[&::-webkit-datetime-edit-hour-field]:p-0',
              '[&::-webkit-datetime-edit-minute-field]:p-0',
              '[&::-webkit-datetime-edit-second-field]:p-0',
              '[&::-webkit-datetime-edit-millisecond-field]:p-0',
              '[&::-webkit-datetime-edit-meridiem-field]:p-0',
            ],
          // Basic layout
          'relative block w-full appearance-none truncate rounded-lg',
          // Size sm (height 28px)
          size === 'sm' &&
            'px-[calc(--spacing(3)-1px)] py-[calc(--spacing(1.5)-1px)] sm:px-[calc(--spacing(2.5)-1px)] sm:py-0.5',
          // Size md (height 36px) - default
          size === 'md' &&
            'px-[calc(--spacing(3.5)-1px)] py-[calc(--spacing(2.5)-1px)] sm:px-[calc(--spacing(3)-1px)] sm:py-[calc(--spacing(1.5)-1px)]',
          // Typography
          typographyClassName ??
            'text-xs/6 font-medium placeholder:font-medium placeholder:text-(--system-gray-5)',
          // Border
          !hideBorder && [
            'border border-zinc-950/10 data-hover:border-zinc-950/20',
            'dark:border-zinc-700 dark:data-hover:border-white/20',
          ],
          // Background color
          noBackground ? null : 'bg-transparent',
          // Hide default focus styles
          'focus:outline-hidden',
          // Invalid state
          'data-invalid:border-red-500 data-invalid:data-hover:border-red-500 dark:data-invalid:border-red-500 dark:data-invalid:data-hover:border-red-500',
          // Disabled state
          'data-disabled:border-zinc-950/20 dark:data-disabled:border-white/15 dark:data-disabled:bg-white/[2.5%] dark:data-hover:data-disabled:border-white/15',
          // System icons
          'dark:[color-scheme:dark]',
          className,
        ])}
      />
    </span>
  )
})
