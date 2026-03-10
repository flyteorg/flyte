import clsx from 'clsx'
import React, { useState } from 'react'
import { NavPanel } from './NavPanel'
import { NavPanelSettings } from './NavPanelSettings'
import { NavItemType, NavPanelMode, NavPanelType, NavPanelWidth } from './types'

export type NavPanelLayoutProps = {
  children: React.ReactNode
  navItems?: NavItemType[]
  type?: NavPanelType
  mode?: NavPanelMode
  initialSize?: NavPanelWidth
}

export const NavPanelLayout = ({
  children,
  navItems,
  type = 'default',
  mode = 'overlay',
  initialSize = 'thin',
}: NavPanelLayoutProps) => {
  const [navSize, setNavSize] = useState<NavPanelWidth>(initialSize)
  const isOverlayMode = mode === 'overlay'

  return (
    <div className="relative flex h-full w-full dark:bg-(--system-black)">
      {/* Reserve layout space */}
      {isOverlayMode && (
        <div className="w-[var(--sidebar-w-thin)] min-w-[var(--sidebar-w-thin)] flex-shrink-0" />
      )}

      {/* Absolutely positioned panel for smooth transitions */}
      <div
        className={clsx(
          'z-30 h-full transition-[width] duration-300 ease-in-out',
          isOverlayMode ? 'absolute top-0 left-0' : 'relative',
          isOverlayMode &&
            navSize === 'wide' &&
            'shadow-[4px 0 4px 0 rgba(0, 0, 0, 0.5)]',
        )}
        style={{
          width:
            navSize === 'thin'
              ? 'var(--sidebar-w-thin)'
              : 'var(--sidebar-w-wide)',
        }}
      >
        <div className="flex h-full flex-col bg-(--system-gray-1) text-(--system-gray-5)">
          {type === 'settings' ? (
            <NavPanelSettings navItems={navItems} />
          ) : (
            <NavPanel
              navSize={navSize}
              setNavSize={setNavSize}
              navItems={navItems}
            />
          )}
        </div>
      </div>

      {/* Page content (never gets overlapped) */}
      <div className="flex min-h-0 w-full flex-1 flex-col overflow-hidden">
        {children}
      </div>
    </div>
  )
}
