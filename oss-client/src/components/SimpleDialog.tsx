'use client'

import {
  Dialog,
  DialogBackdrop,
  DialogPanel,
  DialogTitle,
} from '@headlessui/react'
import { XMarkIcon } from '@heroicons/react/24/outline'
import { Button, styles } from './Button'

export type ButtonConfig = {
  className?: string
  color?: keyof typeof styles.colors
  displayText: string
  onClick: () => void
  outline?: boolean
  plain?: boolean
  disabled?: boolean
}

interface SimpleDialogProps {
  buttons?: ButtonConfig[]
  content: string | React.ReactNode
  headerText?: string | React.ReactNode
  headerIcon?: React.ReactNode
  isOpen: boolean
  setIsOpen: (newState: boolean) => void
  shouldShowCloseButton?: boolean
  maxWidth?: 'sm' | 'md' | 'lg' | 'xl' | '2xl'
}

const maxWidthClasses = {
  sm: 'sm:max-w-sm',
  md: 'sm:max-w-md',
  lg: 'sm:max-w-lg',
  xl: 'sm:max-w-xl',
  '2xl': 'sm:max-w-2xl',
}

export function SimpleDialog({
  buttons,
  content,
  headerIcon,
  headerText,
  isOpen,
  setIsOpen,
  shouldShowCloseButton,
  maxWidth = 'lg',
}: SimpleDialogProps) {
  return (
    <Dialog open={isOpen} onClose={setIsOpen} className="relative z-10">
      <DialogBackdrop
        transition
        className="fixed inset-0 bg-black/60 transition-opacity data-closed:opacity-0 data-enter:duration-300 data-enter:ease-out data-leave:duration-200 data-leave:ease-in"
      />

      <div
        className="fixed inset-0 z-10 w-screen overflow-y-auto"
        data-testid="simple-dialog"
      >
        <div className="flex min-h-full items-end justify-center p-4 text-center sm:items-center sm:p-0">
          <DialogPanel
            transition
            className={`relative transform overflow-hidden rounded-xl bg-white text-left shadow-xl transition-all data-closed:translate-y-4 data-closed:opacity-0 data-enter:duration-300 data-enter:ease-out data-leave:duration-200 data-leave:ease-in sm:my-8 sm:w-full ${maxWidthClasses[maxWidth]} flex max-h-[90vh] flex-col data-closed:sm:translate-y-0 data-closed:sm:scale-95 dark:border dark:border-[#3F3F46] dark:bg-(--system-black)`}
          >
            <div className="absolute top-0 right-0 hidden pt-4 pr-4 sm:block">
              {shouldShowCloseButton && (
                <button
                  type="button"
                  onClick={() => setIsOpen(false)}
                  className="cursor-pointer rounded-md bg-white text-gray-400 hover:text-gray-500 focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 focus:outline-hidden dark:bg-black"
                >
                  <span className="sr-only">Close</span>
                  <XMarkIcon aria-hidden="true" className="size-6" />
                </button>
              )}
            </div>
            <div className="flex-shrink-0 px-4 pt-5 pb-4 sm:p-6">
              <div className={headerIcon ? 'sm:flex sm:items-start' : ''}>
                {headerIcon && (
                  <div className="mx-auto flex size-12 shrink-0 items-center justify-center rounded-full bg-red-100 sm:mx-0 sm:size-10">
                    {headerIcon}
                  </div>
                )}
                <div
                  className={`mt-3 text-center sm:mt-0 ${headerIcon ? 'sm:ml-4 sm:text-left' : 'sm:text-left'}`}
                >
                  {headerText && (
                    <DialogTitle
                      as="h3"
                      className="text-sm font-semibold text-zinc-800 dark:text-(--system-white)"
                    >
                      {headerText}
                    </DialogTitle>
                  )}
                </div>
              </div>
            </div>
            <div
              className="min-h-0 flex-1 overflow-y-auto px-4 [scrollbar-gutter:stable] sm:px-6 [&::-webkit-scrollbar]:w-2 [&::-webkit-scrollbar-corner]:bg-transparent [&::-webkit-scrollbar-thumb]:rounded-full [&::-webkit-scrollbar-thumb]:bg-zinc-400 dark:[&::-webkit-scrollbar-thumb]:bg-zinc-600 [&::-webkit-scrollbar-thumb:hover]:bg-zinc-500 dark:[&::-webkit-scrollbar-thumb:hover]:bg-zinc-500 [&::-webkit-scrollbar-track]:bg-transparent"
              style={{
                scrollbarWidth: 'thin',
                scrollbarColor: 'rgb(161 161 170) transparent', // zinc-400 - works reasonably in both modes
              }}
            >
              <div className={headerText ? 'mt-2' : ''}>{content}</div>
            </div>
            <div className="flex-shrink-0 gap-2 border-t border-zinc-200 px-4 pt-4 pb-4 sm:flex sm:flex-row-reverse sm:px-6 sm:pt-4 sm:pb-6 dark:border-zinc-700">
              {buttons?.map((btn) => {
                return (
                  <Button
                    key={btn.displayText}
                    className={btn.className}
                    color={btn.color}
                    onClick={btn.onClick}
                    disabled={btn.disabled}
                    {...(btn.outline && { outline: true })}
                    {...(btn.plain && { plain: true })}
                  >
                    {btn.displayText}
                  </Button>
                )
              })}
            </div>
          </DialogPanel>
        </div>
      </div>
    </Dialog>
  )
}
