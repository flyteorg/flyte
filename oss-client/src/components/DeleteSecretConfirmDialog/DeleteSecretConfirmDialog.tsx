'use client'

import { type ButtonConfig, SimpleDialog } from '@/components/SimpleDialog'
import { Secret } from '@/gen/flyteidl2/secret/definition_pb'
import { useDeleteSecret } from '@/hooks/useSecrets'
import { useNotifications } from '@/providers/notifications'
import { useCallback, useMemo, useState } from 'react'

type DeleteSecretConfirmDialog = {
  secret: Secret | undefined
  isOpen: boolean
  setIsOpen: (isOpen: boolean) => void
  onSuccess?: () => void
  onDeleteStart?: () => void
  onDeleteError?: () => void
}

export const DeleteSecretConfirmDialog = ({
  secret,
  isOpen,
  setIsOpen,
  onSuccess,
  onDeleteStart,
  onDeleteError,
}: DeleteSecretConfirmDialog) => {
  const { showNotification } = useNotifications()
  const [errorMessage, setErrorMessage] = useState('')

  const deleteSecretMutation = useDeleteSecret()

  const handleDeleteClick = useCallback(async () => {
    if (secret?.id) {
      onDeleteStart?.()

      await deleteSecretMutation.mutate(secret.id, {
        onSuccess: () => {
          showNotification('Secret deleted', 'success')
          if (onSuccess) {
            onSuccess()
          } else {
            setIsOpen(false)
          }
        },
        onError: () => {
          setErrorMessage('There was an error. Try again later.')
          onDeleteError?.()
        },
      })
    }
  }, [
    secret?.id,
    onDeleteStart,
    deleteSecretMutation,
    showNotification,
    onSuccess,
    setIsOpen,
    onDeleteError,
  ])

  const dialogButtons: ButtonConfig[] = useMemo(() => {
    return [
      {
        color: 'rose',
        displayText: 'Delete',
        disabled: deleteSecretMutation.isPending,
        onClick: () => {
          setErrorMessage('')
          handleDeleteClick()
        },
      },
      {
        color: 'dark/zinc',
        displayText: 'Cancel',
        onClick: () => setIsOpen(false),
        plain: true,
      },
    ]
  }, [deleteSecretMutation.isPending, handleDeleteClick, setIsOpen])

  return (
    <SimpleDialog
      buttons={dialogButtons}
      headerText={`Delete ${secret?.id?.name ?? ''}?`}
      content={
        <>
          {errorMessage && (
            <div className="mb-2 text-red-600 dark:text-red-400">
              {errorMessage}
            </div>
          )}
          <div className="text-sm text-zinc-500 dark:text-(--system-gray-5)">
            Are you sure you want to delete this secret? This action cannot be
            undone.
          </div>
        </>
      }
      isOpen={isOpen}
      setIsOpen={(open) => {
        // Prevent closing the modal while deletion is in progress
        if (!deleteSecretMutation.isPending) {
          setIsOpen(open)
        }
      }}
      shouldShowCloseButton={!deleteSecretMutation.isPending}
    />
  )
}
