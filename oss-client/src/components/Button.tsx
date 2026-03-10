import * as Headless from '@headlessui/react'
import clsx from 'clsx'
import React from 'react'
import { Link } from './Link'

/**
 * Style configurations for the Button component.
 * @const
 */
export const styles = {
  size: {
    xxs: 'px-[calc(--spacing(1)+2px)] py-0 text-[10px] leading-3',
    xs: 'px-[calc(--spacing(2.5)-1px)] py-[calc(--spacing(1)-1px)] text-[13px] leading-5',
    sm: 'px-[calc(--spacing(3)-1px)] py-[calc(--spacing(1.5)-1px)] text-sm/5',
    md: 'px-[calc(--spacing(3.5)-1px)] py-[calc(--spacing(2.5)-1px)] sm:px-[calc(--spacing(3)-1px)] sm:py-[calc(--spacing(1.5)-1px)] sm:text-sm/6',
    lg: 'px-[calc(--spacing(4.5)-1px)] py-[calc(--spacing(2.5)-1px)] text-base/6',
    xl: 'px-[calc(--spacing(5.5)-1px)] py-[calc(--spacing(3)-1px)] text-lg/6',
  },
  base: [
    // Base
    'relative isolate inline-flex items-baseline justify-center gap-x-2 rounded-lg border text-base/6 font-medium cursor-pointer',

    // Focus
    'focus:outline-hidden data-focus:outline data-focus:outline-2 data-focus:outline-offset-2 data-focus:outline-blue-500',
    // Disabled
    'data-disabled:opacity-50',
    // Icon
    '*:data-[slot=icon]:-mx-0.5 *:data-[slot=icon]:my-0.5 *:data-[slot=icon]:size-5 *:data-[slot=icon]:shrink-0 *:data-[slot=icon]:self-center *:data-[slot=icon]:text-(--btn-icon) sm:*:data-[slot=icon]:my-1 sm:*:data-[slot=icon]:size-4 forced-colors:[--btn-icon:ButtonText] forced-colors:data-hover:[--btn-icon:ButtonText]',
  ],
  solid: [
    // Optical border, implemented as the button background to avoid corner artifacts
    'border-transparent bg-(--btn-border)',
    // Dark mode: border is rendered on `after` so background is set to button background
    'dark:bg-(--btn-bg)',
    // Button background, implemented as foreground layer to stack on top of pseudo-border layer
    'before:absolute before:inset-0 before:-z-10 before:rounded-[calc(var(--radius-lg)-1px)] before:bg-(--btn-bg)',
    // Drop shadow, applied to the inset `before` layer so it blends with the border
    'before:shadow-sm',
    // Background color is moved to control and shadow is removed in dark mode so hide `before` pseudo
    'dark:before:hidden',
    // Dark mode: Subtle white outline is applied using a border
    'dark:border-white/5',
    // Shim/overlay, inset to match button foreground and used for hover state + highlight shadow
    'after:absolute after:inset-0 after:-z-10 after:rounded-[calc(var(--radius-lg)-1px)]',
    // Inner highlight shadow
    'after:shadow-[shadow:inset_0_1px_--theme(--color-white/15%)]',
    // White overlay on hover
    'data-active:after:bg-(--btn-hover-overlay) data-hover:after:bg-(--btn-hover-overlay)',
    // Dark mode: `after` layer expands to cover entire button
    'dark:after:-inset-px dark:after:rounded-lg',
    // Disabled
    'data-disabled:before:shadow-none data-disabled:after:shadow-none',
  ],
  outline: [
    // Base
    'border data-active:bg-current/5 data-hover:bg-current/5',
    // Dark mode
    'dark:[--btn-bg:transparent]',
  ],
  plain: [
    // Base
    'border-transparent data-active:bg-current/5 data-hover:bg-current/5',
  ],
  colors: {
    'dark/zinc': {
      solid: [
        'text-white [--btn-bg:var(--color-zinc-900)] [--btn-border:var(--color-zinc-950)]/90 [--btn-hover-overlay:var(--color-white)]/10',
        'dark:text-white dark:[--btn-bg:var(--color-zinc-600)] dark:[--btn-hover-overlay:var(--color-white)]/5',
        '[--btn-icon:var(--color-zinc-400)] data-active:[--btn-icon:var(--color-zinc-300)] data-hover:[--btn-icon:var(--color-zinc-300)]',
      ],
      outline: [
        'text-zinc-900 dark:text-white border-zinc-900/20 dark:border-white/20',
        '[--btn-icon:var(--color-zinc-400)] data-active:[--btn-icon:var(--color-zinc-500)] data-hover:[--btn-icon:var(--color-zinc-500)]',
      ],
      plain: [
        'text-zinc-900 dark:text-white',
        '[--btn-icon:var(--color-zinc-400)] data-active:[--btn-icon:var(--color-zinc-500)] data-hover:[--btn-icon:var(--color-zinc-500)]',
      ],
    },
    light: {
      solid: [
        'text-zinc-950 [--btn-bg:white] [--btn-border:var(--color-zinc-950)]/10 [--btn-hover-overlay:var(--color-zinc-950)]/[2.5%] data-active:[--btn-border:var(--color-zinc-950)]/15 data-hover:[--btn-border:var(--color-zinc-950)]/15',
        'dark:text-white dark:[--btn-hover-overlay:var(--color-white)]/5 dark:[--btn-bg:var(--color-zinc-800)]',
        '[--btn-icon:var(--color-zinc-500)] data-active:[--btn-icon:var(--color-zinc-700)] data-hover:[--btn-icon:var(--color-zinc-700)] dark:[--btn-icon:var(--color-zinc-500)] dark:data-active:[--btn-icon:var(--color-zinc-400)] dark:data-hover:[--btn-icon:var(--color-zinc-400)]',
      ],
      outline: [
        'text-zinc-900 dark:text-white border-zinc-900/10 dark:border-white/20',
        '[--btn-icon:var(--color-zinc-400)] data-active:[--btn-icon:var(--color-zinc-500)] data-hover:[--btn-icon:var(--color-zinc-500)]',
      ],
      plain: [
        'text-zinc-900 dark:text-white',
        '[--btn-icon:var(--color-zinc-400)] data-active:[--btn-icon:var(--color-zinc-500)] data-hover:[--btn-icon:var(--color-zinc-500)]',
      ],
    },
    'dark/white': {
      solid: [
        'text-white [--btn-bg:var(--color-zinc-900)] [--btn-border:var(--color-zinc-950)]/90 [--btn-hover-overlay:var(--color-white)]/10',
        'dark:text-zinc-950 dark:[--btn-bg:white] dark:[--btn-hover-overlay:var(--color-zinc-950)]/5',
        '[--btn-icon:var(--color-zinc-400)] data-active:[--btn-icon:var(--color-zinc-300)] data-hover:[--btn-icon:var(--color-zinc-300)] dark:[--btn-icon:var(--color-zinc-500)] dark:data-active:[--btn-icon:var(--color-zinc-400)] dark:data-hover:[--btn-icon:var(--color-zinc-400)]',
      ],
      outline: [
        'text-zinc-900 dark:text-white border-zinc-900/20 dark:border-white/20',
        '[--btn-icon:var(--color-zinc-400)] data-active:[--btn-icon:var(--color-zinc-500)] data-hover:[--btn-icon:var(--color-zinc-500)]',
      ],
      plain: [
        'text-zinc-900 dark:text-white',
        '[--btn-icon:var(--color-zinc-400)] data-active:[--btn-icon:var(--color-zinc-500)] data-hover:[--btn-icon:var(--color-zinc-500)]',
      ],
    },
    dark: {
      solid: [
        'text-white [--btn-bg:var(--color-zinc-900)] [--btn-border:var(--color-zinc-950)]/90 [--btn-hover-overlay:var(--color-white)]/10',
        'dark:[--btn-hover-overlay:var(--color-white)]/5 dark:[--btn-bg:var(--color-zinc-800)]',
        '[--btn-icon:var(--color-zinc-400)] data-active:[--btn-icon:var(--color-zinc-300)] data-hover:[--btn-icon:var(--color-zinc-300)]',
      ],
      outline: [
        'text-zinc-900 dark:text-white border-zinc-900/20 dark:border-white/20',
        '[--btn-icon:var(--color-zinc-400)] data-active:[--btn-icon:var(--color-zinc-500)] data-hover:[--btn-icon:var(--color-zinc-500)]',
      ],
      plain: [
        'text-zinc-900 dark:text-white',
        '[--btn-icon:var(--color-zinc-400)] data-active:[--btn-icon:var(--color-zinc-500)] data-hover:[--btn-icon:var(--color-zinc-500)]',
      ],
    },
    white: {
      solid: [
        'text-zinc-950 [--btn-bg:white] [--btn-border:var(--color-zinc-950)]/10 [--btn-hover-overlay:var(--color-zinc-950)]/[2.5%] data-active:[--btn-border:var(--color-zinc-950)]/15 data-hover:[--btn-border:var(--color-zinc-950)]/15',
        'dark:[--btn-hover-overlay:var(--color-zinc-950)]/5',
        '[--btn-icon:var(--color-zinc-400)] data-active:[--btn-icon:var(--color-zinc-500)] data-hover:[--btn-icon:var(--color-zinc-500)]',
      ],
      outline: [
        'text-zinc-900 dark:text-white border-zinc-900/10 dark:border-white/20',
        '[--btn-icon:var(--color-zinc-400)] data-active:[--btn-icon:var(--color-zinc-500)] data-hover:[--btn-icon:var(--color-zinc-500)]',
      ],
      plain: [
        'text-zinc-900 dark:text-white',
        '[--btn-icon:var(--color-zinc-400)] data-active:[--btn-icon:var(--color-zinc-500)] data-hover:[--btn-icon:var(--color-zinc-500)]',
      ],
    },
    zinc: {
      solid: [
        'text-white [--btn-hover-overlay:var(--color-white)]/10 [--btn-bg:var(--color-zinc-600)] [--btn-border:var(--color-zinc-700)]/90',
        'dark:[--btn-hover-overlay:var(--color-white)]/5',
        '[--btn-icon:var(--color-zinc-400)] data-active:[--btn-icon:var(--color-zinc-300)] data-hover:[--btn-icon:var(--color-zinc-300)]',
      ],
      outline: [
        'text-zinc-600 dark:text-zinc-400 border-zinc-500/50 dark:border-zinc-400/50',
        '[--btn-icon:var(--color-zinc-500)] data-active:[--btn-icon:var(--color-zinc-600)] data-hover:[--btn-icon:var(--color-zinc-600)]',
      ],
      plain: [
        'text-zinc-600 dark:text-(--system-gray-5)',
        '[--btn-icon:var(--color-zinc-500)] data-active:[--btn-icon:var(--color-zinc-600)] data-hover:[--btn-icon:var(--color-zinc-600)]',
      ],
    },
    indigo: {
      solid: [
        'text-white [--btn-hover-overlay:var(--color-white)]/10 [--btn-bg:var(--color-indigo-500)] [--btn-border:var(--color-indigo-600)]/90',
        '[--btn-icon:var(--color-indigo-300)] data-active:[--btn-icon:var(--color-indigo-200)] data-hover:[--btn-icon:var(--color-indigo-200)]',
      ],
      outline: [
        'text-indigo-600 dark:text-indigo-400 border-indigo-500/50 dark:border-indigo-400/50',
        '[--btn-icon:var(--color-indigo-500)] data-active:[--btn-icon:var(--color-indigo-600)] data-hover:[--btn-icon:var(--color-indigo-600)]',
      ],
      plain: [
        'text-indigo-600 dark:text-indigo-400',
        '[--btn-icon:var(--color-indigo-500)] data-active:[--btn-icon:var(--color-indigo-600)] data-hover:[--btn-icon:var(--color-indigo-600)]',
      ],
    },
    orange: {
      solid: [
        'text-white [--btn-hover-overlay:var(--color-white)]/10 [--btn-bg:var(--color-orange-500)] [--btn-border:var(--color-orange-600)]/90',
        '[--btn-icon:var(--color-orange-300)] data-active:[--btn-icon:var(--color-orange-200)] data-hover:[--btn-icon:var(--color-orange-200)]',
      ],
      outline: [
        'text-orange-600 dark:text-orange-400 border-orange-500/50 dark:border-orange-400/50',
        '[--btn-icon:var(--color-orange-500)] data-active:[--btn-icon:var(--color-orange-600)] data-hover:[--btn-icon:var(--color-orange-600)]',
      ],
      plain: [
        'text-orange-600 dark:text-orange-400',
        '[--btn-icon:var(--color-orange-500)] data-active:[--btn-icon:var(--color-orange-600)] data-hover:[--btn-icon:var(--color-orange-600)]',
      ],
    },
    amber: {
      solid: [
        'text-amber-950 [--btn-hover-overlay:var(--color-white)]/25 [--btn-bg:var(--color-amber-400)] [--btn-border:var(--color-amber-500)]/80',
        '[--btn-icon:var(--color-amber-600)]',
      ],
      outline: [
        'text-amber-600 dark:text-amber-400 border-amber-500/50 dark:border-amber-400/50',
        '[--btn-icon:var(--color-amber-500)] data-active:[--btn-icon:var(--color-amber-600)] data-hover:[--btn-icon:var(--color-amber-600)]',
      ],
      plain: [
        'text-amber-600 dark:text-amber-400',
        '[--btn-icon:var(--color-amber-500)] data-active:[--btn-icon:var(--color-amber-600)] data-hover:[--btn-icon:var(--color-amber-600)]',
      ],
    },
    yellow: {
      solid: [
        'text-yellow-950 [--btn-hover-overlay:var(--color-white)]/25 [--btn-bg:var(--color-yellow-300)] [--btn-border:var(--color-yellow-400)]/80',
        '[--btn-icon:var(--color-yellow-600)] data-active:[--btn-icon:var(--color-yellow-700)] data-hover:[--btn-icon:var(--color-yellow-700)]',
      ],
      outline: [
        'text-yellow-600 dark:text-yellow-400 border-yellow-400/50 dark:border-yellow-400/50',
        '[--btn-icon:var(--color-yellow-500)] data-active:[--btn-icon:var(--color-yellow-600)] data-hover:[--btn-icon:var(--color-yellow-600)]',
      ],
      plain: [
        'text-yellow-600 dark:text-yellow-400',
        '[--btn-icon:var(--color-yellow-500)] data-active:[--btn-icon:var(--color-yellow-600)] data-hover:[--btn-icon:var(--color-yellow-600)]',
      ],
    },
    lime: {
      solid: [
        'text-lime-950 [--btn-hover-overlay:var(--color-white)]/25 [--btn-bg:var(--color-lime-300)] [--btn-border:var(--color-lime-400)]/80',
        '[--btn-icon:var(--color-lime-600)] data-active:[--btn-icon:var(--color-lime-700)] data-hover:[--btn-icon:var(--color-lime-700)]',
      ],
      outline: [
        'text-lime-600 dark:text-lime-400 border-lime-400/50 dark:border-lime-400/50',
        '[--btn-icon:var(--color-lime-500)] data-active:[--btn-icon:var(--color-lime-600)] data-hover:[--btn-icon:var(--color-lime-600)]',
      ],
      plain: [
        'text-lime-600 dark:text-lime-400',
        '[--btn-icon:var(--color-lime-500)] data-active:[--btn-icon:var(--color-lime-600)] data-hover:[--btn-icon:var(--color-lime-600)]',
      ],
    },
    green: {
      solid: [
        'text-white [--btn-hover-overlay:var(--color-white)]/10 [--btn-bg:var(--color-green-600)] [--btn-border:var(--color-green-700)]/90',
        '[--btn-icon:var(--color-white)]/60 data-active:[--btn-icon:var(--color-white)]/80 data-hover:[--btn-icon:var(--color-white)]/80',
      ],
      outline: [
        'text-green-600 dark:text-green-400 border-green-500/50 dark:border-green-400/50',
        '[--btn-icon:var(--color-green-500)] data-active:[--btn-icon:var(--color-green-600)] data-hover:[--btn-icon:var(--color-green-600)]',
      ],
      plain: [
        'text-green-600 dark:text-green-400',
        '[--btn-icon:var(--color-green-500)] data-active:[--btn-icon:var(--color-green-600)] data-hover:[--btn-icon:var(--color-green-600)]',
      ],
    },
    emerald: {
      solid: [
        'text-white [--btn-hover-overlay:var(--color-white)]/10 [--btn-bg:var(--color-emerald-600)] [--btn-border:var(--color-emerald-700)]/90',
        '[--btn-icon:var(--color-white)]/60 data-active:[--btn-icon:var(--color-white)]/80 data-hover:[--btn-icon:var(--color-white)]/80',
      ],
      outline: [
        'text-emerald-600 dark:text-emerald-400 border-emerald-500/50 dark:border-emerald-400/50',
        '[--btn-icon:var(--color-emerald-500)] data-active:[--btn-icon:var(--color-emerald-600)] data-hover:[--btn-icon:var(--color-emerald-600)]',
      ],
      plain: [
        'text-emerald-600 dark:text-emerald-400',
        '[--btn-icon:var(--color-emerald-500)] data-active:[--btn-icon:var(--color-emerald-600)] data-hover:[--btn-icon:var(--color-emerald-600)]',
      ],
    },
    teal: {
      solid: [
        'text-white [--btn-hover-overlay:var(--color-white)]/10 [--btn-bg:var(--color-teal-600)] [--btn-border:var(--color-teal-700)]/90',
        '[--btn-icon:var(--color-white)]/60 data-active:[--btn-icon:var(--color-white)]/80 data-hover:[--btn-icon:var(--color-white)]/80',
      ],
      outline: [
        'text-teal-600 dark:text-teal-400 border-teal-500/50 dark:border-teal-400/50',
        '[--btn-icon:var(--color-teal-500)] data-active:[--btn-icon:var(--color-teal-600)] data-hover:[--btn-icon:var(--color-teal-600)]',
      ],
      plain: [
        'text-teal-600 dark:text-teal-400',
        '[--btn-icon:var(--color-teal-500)] data-active:[--btn-icon:var(--color-teal-600)] data-hover:[--btn-icon:var(--color-teal-600)]',
      ],
    },
    sky: {
      solid: [
        'text-white [--btn-hover-overlay:var(--color-white)]/10 [--btn-bg:var(--color-sky-500)] [--btn-border:var(--color-sky-600)]/80',
        '[--btn-icon:var(--color-white)]/60 data-active:[--btn-icon:var(--color-white)]/80 data-hover:[--btn-icon:var(--color-white)]/80',
      ],
      outline: [
        'text-sky-600 dark:text-sky-400 border-sky-500/50 dark:border-sky-400/50',
        '[--btn-icon:var(--color-sky-500)] data-active:[--btn-icon:var(--color-sky-600)] data-hover:[--btn-icon:var(--color-sky-600)]',
      ],
      plain: [
        'text-sky-600 dark:text-sky-400',
        '[--btn-icon:var(--color-sky-500)] data-active:[--btn-icon:var(--color-sky-600)] data-hover:[--btn-icon:var(--color-sky-600)]',
      ],
    },
    blue: {
      solid: [
        'text-white [--btn-hover-overlay:var(--color-white)]/10 [--btn-bg:var(--color-blue-600)] [--btn-border:var(--color-blue-700)]/90',
        '[--btn-icon:var(--color-blue-400)] data-active:[--btn-icon:var(--color-blue-300)] data-hover:[--btn-icon:var(--color-blue-300)]',
      ],
      outline: [
        'text-blue-600 dark:text-blue-400 border-blue-500/50 dark:border-blue-400/50',
        '[--btn-icon:var(--color-blue-500)] data-active:[--btn-icon:var(--color-blue-600)] data-hover:[--btn-icon:var(--color-blue-600)]',
      ],
      plain: [
        'text-blue-600 dark:text-blue-400',
        '[--btn-icon:var(--color-blue-500)] data-active:[--btn-icon:var(--color-blue-600)] data-hover:[--btn-icon:var(--color-blue-600)]',
      ],
    },
    violet: {
      solid: [
        'text-white [--btn-hover-overlay:var(--color-white)]/10 [--btn-bg:var(--color-violet-500)] [--btn-border:var(--color-violet-600)]/90',
        '[--btn-icon:var(--color-violet-300)] data-active:[--btn-icon:var(--color-violet-200)] data-hover:[--btn-icon:var(--color-violet-200)]',
      ],
      outline: [
        'text-violet-600 dark:text-violet-400 border-violet-500/50 dark:border-violet-400/50',
        '[--btn-icon:var(--color-violet-500)] data-active:[--btn-icon:var(--color-violet-600)] data-hover:[--btn-icon:var(--color-violet-600)]',
      ],
      plain: [
        'text-violet-600 dark:text-violet-400',
        '[--btn-icon:var(--color-violet-500)] data-active:[--btn-icon:var(--color-violet-600)] data-hover:[--btn-icon:var(--color-violet-600)]',
      ],
    },
    purple: {
      solid: [
        'text-white [--btn-hover-overlay:var(--color-white)]/10 [--btn-bg:var(--color-purple-500)] [--btn-border:var(--color-purple-600)]/90',
        '[--btn-icon:var(--color-purple-300)] data-active:[--btn-icon:var(--color-purple-200)] data-hover:[--btn-icon:var(--color-purple-200)]',
      ],
      outline: [
        'text-purple-600 dark:text-purple-400 border-purple-500/50 dark:border-purple-400/50',
        '[--btn-icon:var(--color-purple-500)] data-active:[--btn-icon:var(--color-purple-600)] data-hover:[--btn-icon:var(--color-purple-600)]',
      ],
      plain: [
        'text-purple-600 dark:text-purple-400',
        '[--btn-icon:var(--color-purple-500)] data-active:[--btn-icon:var(--color-purple-600)] data-hover:[--btn-icon:var(--color-purple-600)]',
      ],
    },
    fuchsia: {
      solid: [
        'text-white [--btn-hover-overlay:var(--color-white)]/10 [--btn-bg:var(--color-fuchsia-500)] [--btn-border:var(--color-fuchsia-600)]/90',
        '[--btn-icon:var(--color-fuchsia-300)] data-active:[--btn-icon:var(--color-fuchsia-200)] data-hover:[--btn-icon:var(--color-fuchsia-200)]',
      ],
      outline: [
        'text-fuchsia-600 dark:text-fuchsia-400 border-fuchsia-500/50 dark:border-fuchsia-400/50',
        '[--btn-icon:var(--color-fuchsia-500)] data-active:[--btn-icon:var(--color-fuchsia-600)] data-hover:[--btn-icon:var(--color-fuchsia-600)]',
      ],
      plain: [
        'text-fuchsia-600 dark:text-fuchsia-400',
        '[--btn-icon:var(--color-fuchsia-500)] data-active:[--btn-icon:var(--color-fuchsia-600)] data-hover:[--btn-icon:var(--color-fuchsia-600)]',
      ],
    },
    pink: {
      solid: [
        'text-white [--btn-hover-overlay:var(--color-white)]/10 [--btn-bg:var(--color-pink-500)] [--btn-border:var(--color-pink-600)]/90',
        '[--btn-icon:var(--color-pink-300)] data-active:[--btn-icon:var(--color-pink-200)] data-hover:[--btn-icon:var(--color-pink-200)]',
      ],
      outline: [
        'text-pink-600 dark:text-pink-400 border-pink-500/50 dark:border-pink-400/50',
        '[--btn-icon:var(--color-pink-500)] data-active:[--btn-icon:var(--color-pink-600)] data-hover:[--btn-icon:var(--color-pink-600)]',
      ],
      plain: [
        'text-pink-600 dark:text-pink-400',
        '[--btn-icon:var(--color-pink-500)] data-active:[--btn-icon:var(--color-pink-600)] data-hover:[--btn-icon:var(--color-pink-600)]',
      ],
    },
    rose: {
      solid: [
        'text-white dark:text-black [--btn-hover-overlay:var(--color-white)]/10 [--btn-bg:var(--color-rose-500)] [--btn-border:var(--color-rose-600)]/90',
        '[--btn-icon:var(--color-rose-300)] data-active:[--btn-icon:var(--color-rose-200)] data-hover:[--btn-icon:var(--color-rose-200)]',
      ],
      outline: [
        'text-rose-600 dark:text-rose-400 border-rose-500/50 dark:border-rose-400/50',
        '[--btn-icon:var(--color-rose-500)] data-active:[--btn-icon:var(--color-rose-600)] data-hover:[--btn-icon:var(--color-rose-600)]',
      ],
      plain: [
        'text-rose-600 dark:text-rose-400',
        '[--btn-icon:var(--color-rose-500)] data-active:[--btn-icon:var(--color-rose-600)] data-hover:[--btn-icon:var(--color-rose-600)]',
      ],
    },
    union: {
      solid: [
        // Text/icon color from theme (var(--union-on-union): dark in light/dark, light in flyte)
        '',
      ],
      outline: ['', ''],
      plain: ['', ''],
    },
  } as const satisfies Record<
    string,
    {
      solid: string[]
      outline: string[]
      plain: string[]
    }
  >,
} as const

/**
 * Props for the Button component.
 * @typedef {Object} ButtonProps
 * @property {keyof typeof styles.colors} [color='dark/zinc'] - The color scheme of the button
 * @property {boolean} [outline] - Whether to render the button in outline style
 * @property {boolean} [plain] - Whether to render the button in plain style
 * @property {keyof typeof styles.size} [size='md'] - The size of the button
 * @property {string} [className] - Additional CSS classes to apply
 * @property {React.ReactNode} children - The content to render inside the button
 * @property {string} [title] - The title attribute for the button or link
 * @property {React.Ref<HTMLElement>} [ref] - Ref to the button or link element
 */
export type ButtonProps = {
  color?: keyof typeof styles.colors
  outline?: boolean
  plain?: boolean
  size?: keyof typeof styles.size
  className?: string
  children: React.ReactNode
  title?: string
  ref?: React.Ref<HTMLElement>
} & (
  | Omit<Headless.ButtonProps, 'as' | 'className'>
  | Omit<React.ComponentPropsWithoutRef<typeof Link>, 'className'>
)

/**
 * A versatile button component that supports multiple variants and styles.
 *
 * @component
 * @example
 * // Solid button (default)
 * <Button color="blue">Click me</Button>
 *
 * @example
 * // Outline button
 * <Button outline color="blue">Click me</Button>
 *
 * @example
 * // Plain button
 * <Button plain color="blue">Click me</Button>
 *
 * @example
 * // Link button
 * <Button href="/some-path">Navigate</Button>
 *
 * @param {ButtonProps} props - The component props
 * @returns {JSX.Element} A button or link component with the specified styling
 */
export const Button = function Button({
  color = 'dark/zinc',
  outline,
  plain,
  className,
  size = 'md',
  children,
  title,
  ref,
  ...props
}: ButtonProps) {
  const variant = outline ? 'outline' : plain ? 'plain' : 'solid'
  const isUnion = color === 'union'
  const classes = clsx(
    styles.base,
    styles.colors[color][variant],
    styles.size[size],
    variant === 'outline' && styles.outline,
    variant === 'plain' && styles.plain,
    variant === 'solid' && styles.solid,
    className,
  )
  // Inline style for union color (uses --union from theme so Flyte theme gets purple)
  const unionStyle = isUnion
    ? outline
      ? {
          color: 'var(--union)',
          borderColor: 'var(--union)',
          backgroundColor: 'transparent',
        }
      : plain
        ? {
            color: 'var(--union)',
            backgroundColor: 'transparent',
            borderColor: 'transparent',
          }
        : {
            backgroundColor: 'var(--union)',
            borderColor: 'var(--union)',
            color: 'var(--union-on-union)',
          }
    : undefined

  return 'href' in props ? (
    <Link
      {...props}
      className={classes}
      ref={ref as React.Ref<HTMLAnchorElement>}
      title={title}
      style={unionStyle}
    >
      <TouchTarget>{children}</TouchTarget>
    </Link>
  ) : (
    <Headless.Button
      {...props}
      className={classes}
      ref={ref}
      title={title}
      style={unionStyle}
    >
      <TouchTarget>{children}</TouchTarget>
    </Headless.Button>
  )
}

/**
 * A wrapper component that expands the hit area of its children on touch devices.
 * This improves accessibility by ensuring touch targets are at least 44×44px.
 * The expanded area is only visible on touch devices and is hidden on devices with fine pointer control.
 *
 * @component
 * @param {Object} props - The component props
 * @param {React.ReactNode} props.children - The content to wrap with an expanded touch target
 * @returns {JSX.Element} A wrapper with an expanded touch target area
 */
export function TouchTarget({ children }: { children: React.ReactNode }) {
  return (
    <>
      <span
        className="absolute top-1/2 left-1/2 size-[max(100%,2.75rem)] -translate-x-1/2 -translate-y-1/2 [@media(pointer:fine)]:hidden"
        aria-hidden="true"
      />
      {children}
    </>
  )
}
