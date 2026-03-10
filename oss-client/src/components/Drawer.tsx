'use client'

import React, { useState } from 'react'
import {
  Dialog,
  DialogBackdrop,
  DialogPanel,
  DialogTitle,
} from '@headlessui/react'
import {
  ArrowsPointingInIcon,
  ArrowsPointingOutIcon,
  XMarkIcon,
} from '@heroicons/react/24/solid'
import { Button } from './Button'
import { Tooltip } from './Tooltip'

type DrawerSize = 'sm' | 'md' | 'lg' | 'xl' | number

export const drawerSizes: Record<string, string> = {
  sm: 'w-sm max-w-sm',
  md: 'w-md max-w-md',
  lg: 'w-lg max-w-lg',
  xl: 'w-xl max-w-xl',
}

const getDrawerSize = (size: DrawerSize, isFullscreen: boolean) => {
  if (isFullscreen) {
    return { classes: 'w-screen', style: {} }
  }
  if (typeof size === 'number') {
    return { classes: '', style: { width: size, maxWidth: size } }
  }
  return { classes: drawerSizes[size], style: {} }
}

type BaseDrawerProps = {
  isOpen: boolean
  setIsOpen: (newState: boolean) => void
  title: React.ReactNode
  size?: DrawerSize
  titleSection?: React.ReactNode
  hasFullscreen?: boolean
}

type DrawerWithChildren = BaseDrawerProps & {
  children: React.ReactNode
  tabs?: never
}

type DrawerWithTabs = BaseDrawerProps & {
  tabs: React.ReactNode
  children?: never
}

export type DrawerProps = DrawerWithChildren | DrawerWithTabs

export default function Drawer({
  children,
  isOpen,
  setIsOpen,
  tabs,
  title,
  size = 'lg',
  titleSection,
  hasFullscreen = false,
}: DrawerProps) {
  const [isFullscreen, setIsFullscreen] = useState(false)
  const { classes, style } = getDrawerSize(size, isFullscreen)

  return (
    <Dialog open={isOpen} onClose={setIsOpen} className="relative z-10">
      <DialogBackdrop className="fixed inset-0 bg-black/50 shadow-2xl" />
      <div className="fixed inset-0 overflow-hidden">
        <div className="absolute inset-0 overflow-hidden">
          <div className="pointer-events-none fixed inset-y-0 right-0 flex max-w-full">
            <DialogPanel
              transition
              style={style}
              className={`pointer-events-auto transform transition duration-250 ease-in-out data-closed:translate-x-full ${classes}`}
            >
              <div
                data-testid="drawer-dialog"
                className="flex h-full w-[100vw] flex-col bg-(--system-gray-1) py-1 shadow-xl md:w-auto"
              >
                <div className="flex w-full justify-between gap-2 px-4 py-3">
                  <div className="flex min-w-0 flex-1 items-center gap-2">
                    <DialogTitle className="shrink-0 text-base font-semibold">
                      {title}
                    </DialogTitle>
                    {titleSection}
                  </div>

                  <div className="flex shrink-0 items-center gap-3">
                    {hasFullscreen ? (
                      <>
                        {isFullscreen ? (
                          <Tooltip content="Exit fullscreen" placement="bottom">
                            <Button
                              outline
                              type="button"
                              size="sm"
                              onClick={() => setIsFullscreen(false)}
                            >
                              <span className="absolute -inset-2.5" />
                              <span className="sr-only">Exit fullscreen</span>
                              <ArrowsPointingInIcon
                                aria-hidden="true"
                                className="size-6 dark:fill-(--system-white)"
                              />
                            </Button>
                          </Tooltip>
                        ) : (
                          <Tooltip content="Fullscreen" placement="bottom">
                            <Button
                              outline
                              type="button"
                              size="sm"
                              onClick={() => setIsFullscreen(true)}
                            >
                              <span className="absolute -inset-2.5" />
                              <span className="sr-only">Fullscreen</span>
                              <ArrowsPointingOutIcon
                                aria-hidden="true"
                                className="size-6 dark:fill-(--system-white)"
                              />
                            </Button>
                          </Tooltip>
                        )}
                      </>
                    ) : null}

                    <Button
                      outline
                      type="button"
                      size="sm"
                      onClick={() => setIsOpen(false)}
                    >
                      <span className="absolute -inset-2.5" />
                      <span className="sr-only">Close panel</span>
                      <XMarkIcon
                        aria-hidden="true"
                        className="dark:!text-(--system-white)"
                      />
                    </Button>
                  </div>
                </div>
                <div className="flex flex-1 flex-col overflow-hidden">
                  {tabs || children}
                </div>
              </div>
            </DialogPanel>
          </div>
        </div>
      </div>
    </Dialog>
  )
}
