import { RunDetailsPageParams } from '@/components/pages/RunDetails/types'
import { type ButtonConfig, SimpleDialog } from '@/components/SimpleDialog'
import {
  AbortRunRequestSchema,
  RunService,
} from '@/gen/flyteidl2/workflow/run_service_pb'
import { useConnectRpcClient } from '@/hooks/useConnectRpc'
import { useOrg } from '@/hooks/useOrg'
import { getRunIdentifier } from '@/hooks/useRunDetails'
import { create } from '@bufbuild/protobuf'
import { useParams } from 'next/navigation'
import { useCallback, useEffect, useMemo, useState } from 'react'

const DEFAULT_REASON = 'Terminated from UI'

const ConfirmAbortDialogContent = ({
  errorMessage,
  reason,
  setReason,
  isAborting,
}: {
  errorMessage: string
  reason: string
  setReason: (reason: string) => void
  isAborting: boolean
}) => {
  return (
    <>
      <div className="mb-2 text-lg font-semibold">Abort run</div>
      {errorMessage && <div className="mb-2 text-rose-500">{errorMessage}</div>}
      <div className="mb-4 text-zinc-500">
        This run will stop immediately, preventing any remaining actions from
        running and marking the run as Stopped.
      </div>
      <div className="mb-2" aria-busy={isAborting}>
        <label
          htmlFor="abort-run-reason"
          className="block text-sm font-medium text-zinc-700 dark:text-(--system-gray-5)"
        >
          Reason (optional)
        </label>
        <input
          id="abort-run-reason"
          type="text"
          value={reason}
          onChange={(e) => setReason(e.target.value)}
          placeholder={isAborting ? 'Aborting...' : DEFAULT_REASON}
          disabled={isAborting}
          aria-busy={isAborting}
          className="mt-1 block w-full rounded-md border border-zinc-300 bg-white px-3 py-2 text-sm text-zinc-900 placeholder-zinc-400 focus:border-zinc-500 focus:ring-1 focus:ring-zinc-500 focus:outline-none disabled:cursor-not-allowed disabled:opacity-60 dark:border-(--system-gray-3) dark:bg-(--system-gray-1) dark:text-(--system-white) dark:placeholder-(--system-gray-4)"
        />
      </div>
    </>
  )
}

export interface AbortModalProps {
  isOpen?: boolean
  setIsOpen?: (isOpen: boolean) => void
}

export const AbortModal = ({
  isOpen = false,
  setIsOpen = () => {},
}: AbortModalProps) => {
  const client = useConnectRpcClient(RunService)
  const params = useParams<RunDetailsPageParams>()
  const [errorMessage, setErrorMessage] = useState('')
  const [reason, setReason] = useState('')
  const [isAborting, setIsAborting] = useState(false)
  const org = useOrg()

  const handleAbortClick = useCallback(async () => {
    setIsAborting(true)
    setErrorMessage('')
    try {
      const runIdentifier = getRunIdentifier({
        domain: params.domain,
        name: params.runId,
        org,
        project: params.project,
      })
      const trimmedReason = reason.trim()
      const request = create(AbortRunRequestSchema, {
        runId: runIdentifier,
        reason: trimmedReason || DEFAULT_REASON,
      })
      await client.abortRun(request)
      setIsOpen(false)
      setReason('')
    } catch (e) {
      console.error(e)
      setErrorMessage('There was an error. Try again later.')
    } finally {
      setIsAborting(false)
    }
  }, [
    client,
    params.domain,
    params.project,
    params.runId,
    reason,
    setIsOpen,
    org,
  ])

  useEffect(() => {
    if (!isOpen) {
      setErrorMessage('')
      setReason('')
    }
  }, [isOpen])

  const dialogButtons: ButtonConfig[] = useMemo(() => {
    return [
      {
        color: 'rose',
        displayText: isAborting ? 'Aborting...' : 'Abort run',
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
        <ConfirmAbortDialogContent
          errorMessage={errorMessage}
          reason={reason}
          setReason={setReason}
          isAborting={isAborting}
        />
      }
      isOpen={isOpen}
      setIsOpen={(open) => {
        if (!isAborting) {
          setIsOpen(open)
        }
      }}
      shouldShowCloseButton={!isAborting}
    />
  )
}
