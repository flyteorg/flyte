import { Notification } from '@/components/Notification'
import React, { createContext, useCallback, useContext, useState } from 'react'

type NotificationVariant = 'success' | 'error' | 'warning' | 'running'
interface NotificationButton {
  label: string
  onClick: () => void
}
export interface NotificationInfo {
  title: string
  variant: NotificationVariant
  text?: string
  button?: NotificationButton
  hasCloseBtn?: boolean
}

type NotificationsContextType = {
  showNotification: (
    title: string,
    variant: NotificationVariant,
    text?: string,
    button?: NotificationButton,
    hasCloseBtn?: boolean,
  ) => void
  hideNotification: () => void
}

const NotificationsContext = createContext<NotificationsContextType | null>(
  null,
)

export const useNotifications = () => {
  const ctx = useContext(NotificationsContext)
  if (!ctx)
    throw new Error(
      'useNotifications must be used inside NotificationsProvider',
    )
  return ctx
}

export const NotificationsProvider: React.FC<{ children: React.ReactNode }> = ({
  children,
}) => {
  const [notificationInfo, setNotificationInfo] =
    useState<NotificationInfo | null>(null)

  const showNotification = (
    title: string,
    variant: NotificationVariant,
    text?: string,
    button?: NotificationButton,
    hasCloseBtn?: boolean,
  ) => {
    setNotificationInfo({
      title,
      text,
      variant,
      button,
      hasCloseBtn: hasCloseBtn ?? true,
    })
    setTimeout(() => setNotificationInfo(null), 3000)
  }

  const hideNotification = useCallback(() => {
    setNotificationInfo(null)
  }, [])

  return (
    <NotificationsContext.Provider
      value={{
        showNotification,
        hideNotification,
      }}
    >
      <Notification
        notificationInfo={notificationInfo}
        hideNotification={hideNotification}
      />
      {children}
    </NotificationsContext.Provider>
  )
}
