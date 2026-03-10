import React from 'react'
import { NotificationInfo } from '@/providers/notifications'
import { Dialog, DialogPanel } from '@headlessui/react'
import {
  CheckCircleIcon,
  ExclamationCircleIcon,
  ExclamationTriangleIcon,
} from '@heroicons/react/20/solid'
import { XMarkIcon } from '@heroicons/react/24/outline'
import { CircularProgressIcon } from '@/components/icons/CircularProgressIcon'
import { Button } from './Button'

const icons = {
  success: (
    <CheckCircleIcon
      className="h-6 w-5 min-w-5 text-lime-400"
      data-slot="icon"
    />
  ),
  error: (
    <ExclamationCircleIcon
      className="h-6 w-5 min-w-5 text-red-400"
      data-slot="icon"
    />
  ),
  warning: (
    <ExclamationTriangleIcon
      className="mt-[3px] size-5 min-w-5 text-yellow-400"
      data-slot="icon"
    />
  ),
  running: (
    <CircularProgressIcon
      isStatic={true}
      className="h-6 w-5 min-w-5 text-(--accent-blue)"
      data-slot="icon"
    />
  ),
}

export interface NotificationProps {
  notificationInfo: NotificationInfo | null
  hideNotification: VoidFunction
}

export const Notification = ({
  notificationInfo,
  hideNotification,
}: NotificationProps) => {
  return (
    <Dialog open={!!notificationInfo} onClose={hideNotification}>
      <DialogPanel transition className={dialogPanelClassName}>
        <NotificationContent {...{ notificationInfo, hideNotification }} />
      </DialogPanel>
    </Dialog>
  )
}

export const dialogPanelClassName =
  'absolute top-0 right-0 z-10 m-4 justify-between flex max-w-md transform items-start gap-4 rounded-lg bg-white p-4.5 shadow-xl transition-all data-closed:translate-y-4 data-closed:opacity-0 data-enter:duration-300 data-enter:ease-out data-leave:duration-200 data-leave:ease-in sm:w-full data-closed:sm:translate-y-0 data-closed:sm:scale-95 dark:bg-zinc-700'

type NotificationContentProps = NotificationProps
export const NotificationContent = ({
  notificationInfo,
  hideNotification,
}: NotificationContentProps) => (
  <>
    <div className="flex flex-col">
      <div className="flex items-start gap-4">
        {notificationInfo && icons[notificationInfo?.variant]}
        <div className="text-sm font-semibold text-zinc-950 dark:text-white">
          {notificationInfo?.title}
        </div>
      </div>

      {notificationInfo?.text ? (
        <div className="ml-9 text-sm font-medium text-zinc-950 dark:text-zinc-400">
          {notificationInfo?.text}
        </div>
      ) : null}
    </div>

    <div className="flex items-center gap-3">
      {notificationInfo?.button ? (
        <Button
          plain
          className="whitespace-nowrap"
          onClick={notificationInfo.button.onClick}
          size="sm"
        >
          {notificationInfo.button.label}
        </Button>
      ) : null}
      {notificationInfo?.hasCloseBtn ? (
        <Button plain onClick={hideNotification} size="xxs" aria-label="Close">
          <XMarkIcon data-slot="icon" aria-hidden="true" />
        </Button>
      ) : null}
    </div>
  </>
)
