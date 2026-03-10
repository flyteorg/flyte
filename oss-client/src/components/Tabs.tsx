'use client'

import { Tab, TabGroup, TabList } from '@headlessui/react'
import clsx from 'clsx'
import { AnimatePresence } from 'motion/react'
import * as motion from 'motion/react-client'
import { useEffect, useRef, useState } from 'react'

export type TabWidget = {
  key: string
  content: React.ReactNode
}

export interface TabType<T> {
  label: string | React.ReactNode
  icon?: React.ReactNode
  /** Rendered after the label (e.g. error indicator). */
  trailingIcon?: React.ReactNode
  path: T
  content: React.ReactNode
  hidden?: boolean // If true, tab is accessible via direct link but not shown in tab bar
}

export interface TabsProps<T> {
  currentTab?: T | null | undefined
  footer?: React.ReactNode
  onClickTab: (path: T) => void
  scrollShadow?: boolean
  classes?: { tabListContainer?: string; tabContentContainer?: string }
  tabs: TabType<T>[]
  widgets?: TabWidget[]
}

export function Tabs<T extends string>({
  classes,
  currentTab,
  footer,
  onClickTab,
  scrollShadow = false,
  tabs,
  widgets,
}: TabsProps<T>) {
  const selectedIndex =
    currentTab && tabs.some((tab) => tab.path === currentTab)
      ? tabs.findIndex((tab) => tab.path === currentTab)
      : 0
  const activeTab = tabs[selectedIndex]

  const contentRef = useRef<HTMLDivElement>(null)
  const [scrolled, setScrolled] = useState(false)

  useEffect(() => {
    if (scrollShadow) {
      const el = contentRef.current
      const handleScroll = () => setScrolled((el?.scrollTop ?? 0) > 0)
      if (el) {
        setScrolled(el.scrollTop > 0) // check initial scroll
        el.addEventListener('scroll', handleScroll)
      }
      return () => el?.removeEventListener('scroll', handleScroll)
    }
  }, [scrollShadow])

  return (
    <TabGroup
      className="relative flex min-w-0 flex-1 flex-col"
      selectedIndex={selectedIndex}
    >
      <div
        className={clsx(
          'flex w-full justify-between px-10 py-5',
          classes?.tabListContainer,
        )}
      >
        <TabList className="flex min-h-7 w-full gap-1">
          {tabs
            .filter((tab) => !tab.hidden)
            .map((tab) => {
              const isSelected = tab.path === currentTab
              return (
                <Tab
                  key={tab.path}
                  className={`group relative cursor-pointer rounded-lg text-(--system-gray-5) transition outline-none hover:bg-(--system-gray-3) hover:text-(--system-white) ${
                    isSelected
                      ? 'bg-(--system-gray-4) text-(--system-white) hover:bg-(--system-gray-4)'
                      : ''
                  }`}
                  onClick={() => onClickTab?.(tab.path)}
                >
                  <div className="flex items-center gap-1.5 px-2 py-0.5 text-left align-middle text-xs font-medium">
                    {tab.icon ? (
                      <span className={isSelected ? 'text-(--union)' : ''}>
                        {tab.icon}
                      </span>
                    ) : null}
                    <span className="capitalize">{tab.label}</span>
                    {tab.trailingIcon ?? null}
                  </div>
                </Tab>
              )
            })}
        </TabList>
        <div className="flex shrink-0 items-center gap-3">
          {widgets?.map(({ content, key }) => (
            <div key={key}>{content}</div>
          ))}
        </div>
      </div>

      {scrollShadow && (
        <div
          className={clsx(
            'pointer-events-none absolute top-17 right-0 left-0 z-20 h-3 bg-gradient-to-b from-black/20 via-black/10 to-transparent transition-opacity duration-300 ease-in-out',
            scrolled ? 'opacity-100' : 'opacity-0',
          )}
        />
      )}

      {activeTab && (
        <div
          ref={contentRef}
          className="relative z-0 min-h-0 flex-1 overflow-auto [scrollbar-gutter:stable_both-edges] [&::-webkit-scrollbar]:w-2 [&::-webkit-scrollbar-corner]:bg-transparent [&::-webkit-scrollbar-thumb]:rounded-full [&::-webkit-scrollbar-thumb]:bg-zinc-400 dark:[&::-webkit-scrollbar-thumb]:bg-zinc-600 [&::-webkit-scrollbar-thumb:hover]:bg-zinc-500 dark:[&::-webkit-scrollbar-thumb:hover]:bg-zinc-500 [&::-webkit-scrollbar-track]:bg-transparent"
        >
          <AnimatePresence mode="wait">
            <motion.div
              key={activeTab.path}
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0 }}
              transition={{ duration: 0.2 }}
              className={clsx(
                'absolute inset-0 flex h-full min-h-0 flex-col',
                classes?.tabContentContainer,
              )}
            >
              <div className="flex-1 overflow-auto">{activeTab.content}</div>
              {footer && <div className="mt-auto pt-4">{footer}</div>}
            </motion.div>
          </AnimatePresence>
        </div>
      )}
    </TabGroup>
  )
}
