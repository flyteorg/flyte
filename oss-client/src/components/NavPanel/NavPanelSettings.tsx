import { ChevronLeftIcon } from '@heroicons/react/20/solid'
import { UserIcon, UsersIcon } from '@heroicons/react/24/solid'
import { AnimatePresence, motion } from 'motion/react'
import Link from 'next/link'
import { useSearchParams } from 'next/navigation'
import { KeyIcon } from '../icons/KeyIcon'
import { getNavItem } from './NavItems'
import { type NavItemType } from './types'

type NavPanelSettingsProps = {
  navItems?: NavItemType[]
}

export const NavPanelSettings = ({ navItems }: NavPanelSettingsProps) => {
  const searchParams = useSearchParams()

  // Get returnTo from query params and construct the href
  const returnTo = searchParams.get('returnTo')
  let returnToHref = '/projects' // Default fallback

  if (returnTo) {
    try {
      const decodedPath = decodeURIComponent(returnTo)
      // Only use returnTo if it's not a settings page
      if (!decodedPath.startsWith('/settings')) {
        returnToHref = decodedPath
      }
    } catch {
      // If decoding fails, fall through to default
    }
  }

  // Preserve returnTo query param when navigating between settings pages
  // searchParams.get() returns decoded value, so we need to encode it again for the URL
  const returnToQuery = returnTo
    ? `?returnTo=${encodeURIComponent(returnTo)}`
    : ''

  const settingsNavItems: NavItemType[] = navItems || [
    {
      displayText: 'Account',
      type: 'heading',
      className: 'text-xs font-semibold',
    },
    {
      displayText: 'Profile',
      makeHref: () => `/settings/profile${returnToQuery}`,
      icon: <UserIcon className="size-4" />,
      type: 'link',
      className: 'font-medium  hover:text-(--system-white)',
    },
    {
      displayText: 'Assets & Configuration',
      type: 'heading',
      className: 'text-xs font-semibold',
    },
    {
      displayText: 'Secrets',
      makeHref: () => `/settings/secrets${returnToQuery}`,
      icon: <KeyIcon />,
      type: 'link',
    },
    {
      displayText: 'User Management',
      makeHref: () => `/settings/user-management${returnToQuery}`,
      icon: <UsersIcon className="size-4" />,
      type: 'link',
      className: 'font-medium hover:text-(--system-white)',
    },
  ]

  return (
    <>
      <div className="flex flex-1 flex-col items-start gap-10 bg-(--system-gray-2) p-4 py-4">
        <Link
          href={returnToHref}
          className="flex cursor-pointer items-center gap-2 text-(--system-gray-5) transition-colors hover:text-(--system-white)"
        >
          <div className="flex h-5 w-5 items-center justify-center rounded-sm bg-(--system-gray-2) backdrop-blur-sm">
            <ChevronLeftIcon
              className="h-6 w-4 font-bold"
              aria-hidden="true"
              strokeWidth={2.5}
            />
          </div>
          <span className="text-xs font-semibold tracking-normal">
            Return to app
          </span>
        </Link>

        <motion.div
          className="flex w-full flex-col gap-1"
          initial="hidden"
          animate="visible"
        >
          <AnimatePresence mode="popLayout">
            {settingsNavItems.map((item) => (
              <motion.div
                key={item.displayText}
                initial="hidden"
                animate="visible"
                exit="exit"
                layout
                className="flex w-full min-w-5.5"
              >
                {getNavItem(item, 'wide')}
              </motion.div>
            ))}
          </AnimatePresence>
        </motion.div>
      </div>
    </>
  )
}
