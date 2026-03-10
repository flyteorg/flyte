import NextLink from 'next/link'
import { ChevronRightIcon } from './icons/ChevronRightIcon'
import { ArrowTopRightIcon } from './icons/ArrowTopRightIcon'
/**
 * TODO: Update this component to use your client-side framework's link
 * component. We've provided examples of how to do this for Next.js, Remix, and
 * Inertia.js in the Catalyst documentation:
 *
 * https://catalyst.tailwindui.com/docs#client-side-router-integration
 */

import * as Headless from '@headlessui/react'
import React from 'react'

export const Link = function Link({
  ref,
  ...props
}: {
  href: string
  ref?: React.Ref<HTMLAnchorElement>
} & React.ComponentPropsWithoutRef<'a'>) {
  return (
    <Headless.DataInteractive>
      <a {...props} ref={ref} />
    </Headless.DataInteractive>
  )
}

export const LinkPill = ({
  displayText,
  href,
}: {
  displayText: string | undefined
  href: string
}) => (
  <NextLink
    onClick={(e: React.MouseEvent) => {
      // prevent parent click handlers
      e.stopPropagation()
    }}
    className="flex max-w-[270px] items-center gap-1 rounded-lg bg-(--system-gray-2) px-2"
    href={href}
  >
    <span className="truncate text-sm font-medium text-zinc-950 dark:text-white">
      {displayText || href}
    </span>
    <ChevronRightIcon className="ml-1 text-gray-500" />
  </NextLink>
)

export const LinkArrow = ({
  displayText,
  href,
}: {
  displayText: string | undefined
  href: string
}) => (
  <NextLink
    onClick={(e: React.MouseEvent) => {
      // prevent parent click handlers
      e.stopPropagation()
    }}
    className="group-arrow flex items-center gap-1 rounded-lg px-2"
    href={href}
  >
    <span className="truncate text-2xs font-medium text-(--system-gray-5) group-hover:text-(--system-white)">
      {displayText || href}
    </span>
    <ArrowTopRightIcon
      className="text-(--system-gray-5) group-hover:text-(--system-white)"
      width={16}
    />
  </NextLink>
)
