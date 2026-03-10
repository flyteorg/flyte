import { useDomainStore } from '@/lib/DomainStore'
import clsx from 'clsx'
import { AnimatePresence, motion } from 'motion/react'
import Link from 'next/link'
import { useParams, usePathname } from 'next/navigation'
import { Tooltip } from '../Tooltip'
import {
  NavItemType,
  NavPanelWidth,
  type NavLink as NavLinkType,
  type NavSectionHeading,
  type NavWidget as NavWidgetType,
} from './types'

export const getNavItem = (navItem: NavItemType, size: NavPanelWidth) => {
  switch (navItem.type) {
    case 'heading': {
      return <NavHeading {...navItem} size={size} />
    }
    case 'link': {
      return <NavLink {...navItem} icon={navItem.icon} size={size} />
    }
    case 'widget': {
      return <NavWidget {...navItem} size={size} />
    }
  }
}

type ProjectDomainParams = {
  domain: string
  project: string
}

export const NavLink = ({
  className,
  displayText,
  displayComponent,
  makeHref,
  icon,
  onClick,
  target,
  prefetch = false,
  size,
}: NavLinkType & { size: NavPanelWidth }) => {
  const params = useParams<ProjectDomainParams>()
  const { selectedDomain } = useDomainStore()
  const pathName = usePathname()
  // When in settings, use makeHref even without selectedDomain so sidebar links go to the correct settings page instead of /projects
  const isSettings = pathName.startsWith('/settings')
  const href =
    selectedDomain || isSettings
      ? makeHref({
          project: params.project ?? '',
          domain: selectedDomain?.id ?? '',
          pathname: pathName,
        })
      : '/projects'
  // Extract pathname from href (remove query parameters) for comparison
  const hrefPathname = href.split('?')[0]
  const isSelected =
    pathName === hrefPathname || pathName.startsWith(hrefPathname + '/')
  const isThin = size === 'thin'

  return (
    <Tooltip content={displayText} disabled={!isThin} placement="right">
      <Link
        href={href}
        prefetch={prefetch}
        target={target}
        onClick={() => onClick?.()}
        className={clsx(
          'flex min-h-8 w-full min-w-8 items-center overflow-hidden rounded-lg',
          'transition-colors duration-200 hover:text-(--system-white)',
          isSelected &&
            'bg-(--system-gray-2) text-(--system-white) dark:bg-(--system-gray-3)',
          !isSelected &&
            'hover:bg-(--system-gray-3) dark:hover:bg-(--system-gray-4)',
          className,
        )}
      >
        {icon && (
          <motion.div
            layout
            className={clsx(
              'flex w-8 flex-shrink-0 items-center justify-center p-2',
              isSelected && 'text-(--union)',
            )}
            transition={{ duration: 0.25, ease: 'easeInOut' }}
          >
            {icon}
          </motion.div>
        )}

        <AnimatePresence initial={false}>
          {!isThin && (
            <motion.span
              key="label"
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0 }}
              transition={{ duration: 0.2, ease: 'easeInOut' }}
              className={clsx(
                'w-full pr-2 text-sm font-medium whitespace-nowrap',
                !icon && 'ml-2',
              )}
            >
              {displayComponent ? displayComponent : displayText}
            </motion.span>
          )}
        </AnimatePresence>
      </Link>
    </Tooltip>
  )
}

export const NavHeading = ({
  className,
  displayText,
  icon,
  size,
}: NavSectionHeading & { size: NavPanelWidth }) => {
  const isThin = size === 'thin'

  return (
    <div
      className={clsx(
        'gap-.5 flex items-center rounded-lg',
        'transition-colors duration-200',
        className,
      )}
      title={isThin ? displayText : undefined}
    >
      {icon && (
        <motion.div className="flex w-8 items-center justify-center p-2">
          {icon}
        </motion.div>
      )}

      <AnimatePresence initial={false}>
        {!isThin && (
          <motion.span
            key="label"
            initial={{ opacity: 0, width: 0 }}
            animate={{ opacity: 1, width: 'auto' }}
            exit={{ opacity: 0, width: 0 }}
            transition={{ duration: 0.25, ease: 'easeInOut' }}
            className="overflow-hidden text-sm whitespace-nowrap text-(--system-gray-6)"
          >
            {displayText}
          </motion.span>
        )}
      </AnimatePresence>
    </div>
  )
}

export const NavWidget = ({
  className,
  size,
  widget,
}: NavWidgetType & { size: NavPanelWidth }) => (
  <div className={`${className}`}>{widget(size)}</div>
)
