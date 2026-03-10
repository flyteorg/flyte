import clsx from 'clsx'
import { AnimatePresence, motion } from 'motion/react'
import { Dispatch, SetStateAction, useMemo } from 'react'
import Link from 'next/link'
import { useParams } from 'next/navigation'
import { HamburgerIcon } from '../icons/HamburgerIcon'
import { type NavItemType, NavPanelWidth } from './types'
import { NavItems } from '.'
import { Button } from '../Button'
import { Logo } from '../Logo'
import { RunDetailsPageParams } from '../pages/RunDetails/types'
import { getNavItem } from './NavItems'
import { DomainPicker } from '../DomainPicker'
import { useDomainStore } from '@/lib/DomainStore'
import { ProjectsPicker } from '../ProjectsPicker'

type NavPanelProps = {
  navItems?: NavItemType[] // custom main nav items
  orgItems?: NavItemType[] // custom organization nav items
  navSize: NavPanelWidth
  setNavSize: Dispatch<SetStateAction<NavPanelWidth>>
}

export const NavPanel = ({
  navItems,
  orgItems,
  navSize,
  setNavSize,
}: NavPanelProps) => {
  const isThin = navSize === 'thin'
  const params = useParams<RunDetailsPageParams>()
  const { selectedDomain } = useDomainStore()
  const project = params.project
  // show the project picker when project is in url params (excludes /projects)
  const shouldShowProjectPicker = !!project
  const defaultNavItems = NavItems.useDefaultItems()
  const defaultOrgNavItems = NavItems.useDefaultOrgItems()
  const itemsToShow = navItems || defaultNavItems
  const orgItemsToShow = orgItems || defaultOrgNavItems

  const logoClickLink = useMemo(() => {
    if (project && selectedDomain) {
      return `/domain/${selectedDomain.id}/project/${project}/runs`
    }
    // if no selected project/domain, send to project selection page
    return `/projects`
  }, [project, selectedDomain])

  return (
    <>
      <div className="flex flex-1 flex-col items-start gap-8 px-2.5 py-3">
        <div>
          <div className="flex items-center gap-1">
            <Button
              size="xs"
              className="!p-2 hover:bg-(--system-gray-3) dark:hover:bg-(--system-gray-4)"
              plain
              onClick={() => {
                setNavSize((prev) => (prev === 'thin' ? 'wide' : 'thin'))
              }}
            >
              <HamburgerIcon className="size-4 text-(--system-gray-5)" />
            </Button>
            <Link href={logoClickLink} aria-label="Home">
              <Logo height={20} width={42} />
            </Link>
            <DomainPicker />
            {shouldShowProjectPicker && (
              <div className="flex items-center gap-1">
                <span className="mx-2 text-xl text-(--system-gray-3)">|</span>
                <ProjectsPicker />
              </div>
            )}
          </div>
        </div>

        {/* Top nav items */}
        <motion.div
          className="flex w-full flex-col gap-1"
          initial="hidden"
          animate="visible"
        >
          <AnimatePresence mode="popLayout">
            {itemsToShow.map((item) =>
              item.shouldHideIconOnCollapse && isThin ? null : (
                <motion.div
                  key={item.displayText}
                  initial="hidden"
                  animate="visible"
                  exit="exit"
                  layout
                  className={clsx(
                    `flex w-full min-w-5.5`,
                    isThin && 'justify-center',
                  )}
                >
                  {getNavItem(item, navSize)}
                </motion.div>
              ),
            )}
          </AnimatePresence>
        </motion.div>
      </div>

      {/* Bottom org items */}
      <div
        className={`flex shrink-0 flex-col ${isThin ? 'justify-center' : ''} gap-y-0.5 px-2.5`}
      >
        <motion.div
          className="flex w-full flex-col gap-1 pb-5"
          initial="hidden"
          animate="visible"
        >
          <AnimatePresence mode="popLayout">
            {orgItemsToShow.map((item) => (
              <motion.div
                key={item.displayText}
                initial="hidden"
                animate="visible"
                exit="exit"
                layout
                className={clsx(
                  `flex w-full min-w-5.5`,
                  isThin && 'justify-center',
                )}
              >
                {getNavItem(item, navSize)}
              </motion.div>
            ))}
          </AnimatePresence>
        </motion.div>
      </div>
    </>
  )
}
