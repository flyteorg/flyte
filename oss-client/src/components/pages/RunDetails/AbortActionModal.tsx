'use client'

import { type ButtonConfig, SimpleDialog } from '@/components/SimpleDialog'
import { ActionDetails } from '@/gen/flyteidl2/workflow/run_definition_pb'
import {
  AbortActionRequestSchema,
  RunService,
} from '@/gen/flyteidl2/workflow/run_service_pb'
import { useConnectRpcClient } from '@/hooks/useConnectRpc'
import { create } from '@bufbuild/protobuf'
import { useCallback, useEffect, useMemo, useState } from 'react'

const DEFAULT_REASON = 'Terminated from UI'

const ConfirmAbortActionDialogContent = ({
  errorMessage,
  reason,
  setReason,
}: {
  errorMessage: string
  reason: string
  setReason: (reason: string) => void
}) => {
  return (
    <>
      <div className="mb-2 text-lg font-semibold">Abort action</div>
      {errorMessage && <div className="mb-2 text-rose-500">{errorMessage}</div>}
      <div className="mb-4 text-zinc-500">
        This action will stop immediately, preventing it from completing and
        marking it as Aborted.
      </div>
      <div className="mb-2">
        <label
          htmlFor="abort-reason"
          className="block text-sm font-medium text-zinc-700 dark:text-(--system-gray-5)"
        >
          Reason (optional)
        </label>
        <input
          id="abort-reason"
          type="text"
          value={reason}
          onChange={(e) => setReason(e.target.value)}
          placeholder={DEFAULT_REASON}
          className="mt-1 block w-full rounded-md border border-zinc-300 bg-white px-3 py-2 text-sm text-zinc-900 placeholder-zinc-400 focus:border-zinc-500 focus:ring-1 focus:ring-zinc-500 focus:outline-none dark:border-(--system-gray-3) dark:bg-(--system-gray-1) dark:text-(--system-white) dark:placeholder-(--system-gray-4)"
        />
      </div>
    </>
  )
}

export interface AbortActionModalProps {
  actionDetails: ActionDetails | undefined
  isOpen: boolean
  setIsOpen: (isOpen: boolean) => void
}

export const AbortActionModal = ({
  actionDetails,
  isOpen,
  setIsOpen,
}: AbortActionModalProps) => {
  const client = useConnectRpcClient(RunService)
  const [errorMessage, setErrorMessage] = useState('')
  const [reason, setReason] = useState('')
  const [isAborting, setIsAborting] = useState(false)

  const handleAbortClick = useCallback(async () => {
    if (!actionDetails?.id) {
      setErrorMessage('Action identifier is missing.')
      return
    }

    setIsAborting(true)
    setErrorMessage('')

    try {
      const trimmedReason = reason.trim()
      const request = create(AbortActionRequestSchema, {
        actionId: actionDetails.id,
        reason: trimmedReason || DEFAULT_REASON,
      })

      await client.abortAction(request)
      setIsOpen(false)
      setReason('')
    } catch (e) {
      console.error('Error aborting action:', e)
      setErrorMessage(
        'There was an error aborting the action. Try again later.',
      )
    } finally {
      setIsAborting(false)
    }
  }, [actionDetails?.id, reason, client, setIsOpen])

  useEffect(() => {
    // Reset error message and reason on close
    if (!isOpen) {
      setErrorMessage('')
      setReason('')
    }
  }, [isOpen])

  const dialogButtons: ButtonConfig[] = useMemo(() => {
    return [
      {
        color: 'rose',
        displayText: isAborting ? 'Aborting...' : 'Abort action',
        onClick: handleAbortClick,
        disabled: isAborting,
      },
      {
        color: 'dark/zinc',
        displayText: 'Cancel',
        onClick: () => setIsOpen(false),
        outline: true,
        disabled: isAborting,
      },
    ]
  }, [handleAbortClick, setIsOpen, isAborting])

  return (
    <SimpleDialog
      buttons={dialogButtons}
      content={
        <ConfirmAbortActionDialogContent
          errorMessage={errorMessage}
          reason={reason}
          setReason={setReason}
        />
      }
      isOpen={isOpen}
      setIsOpen={(open) => {
        // Prevent closing the modal while aborting is in progress
        if (!isAborting) {
          setIsOpen(open)
        }
      }}
      shouldShowCloseButton={!isAborting}
    />
  )
}
