export type BulkAction<T> = {
  makeLabel: (selectedItems: T[]) => string
  onClick: (selectedItems: T[], actions: CallbackActions) => void
}

type CallbackActions = {
  clearNotification: () => void
  clearSelection: () => void
  showNotification: (props: NotificationProps) => void
}

export type NotificationProps = {
  message: string
  durationMs: number
  undoCallback?: () => void
  variant?: 'okay' | 'warn'
}
