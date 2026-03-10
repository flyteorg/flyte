import { type ButtonConfig, SimpleDialog } from '@/components/SimpleDialog'
import { create } from '@bufbuild/protobuf'
import { EnvsSchema, RunSpecSchema } from '@/gen/flyteidl2/task/run_pb'
import { RunSource } from '@/gen/flyteidl2/workflow/run_definition_pb'
import { RunService } from '@/gen/flyteidl2/workflow/run_service_pb'
import { useActionData } from '@/hooks/useActionData'
import { useConnectRpcClient } from '@/hooks/useConnectRpc'
import { useOrg } from '@/hooks/useOrg'
import { useRunDetails } from '@/hooks/useRunDetails'
import { useWatchActionDetails } from '@/hooks/useWatchActionDetails'
import { getMergedRunAndTaskEnvs } from '@/lib/actionUtils'
import { useParams, useRouter } from 'next/navigation'
import { useCallback, useEffect, useMemo, useState } from 'react'
import { RunDetailsPageParams } from './types'

// You can optionally provide a reason to help track why it was stopped.

const ConfirmRerunDialogContent = ({
  message: message,
}: {
  message: string
}) => {
  return (
    <>
      <div className="mb-2 text-lg font-semibold">Rerun</div>
      {message && <div className="mb-2 text-green-500">{message}</div>}
      <div className="text-zinc-500">
        This will create a new Run using the same inputs as the current run.
      </div>
    </>
  )
}

export interface RerunModalProps {
  isOpen: boolean
  setIsOpen: (isOpen: boolean) => void
}

export const RerunModal = ({ isOpen, setIsOpen }: RerunModalProps) => {
  const router = useRouter()

  const client = useConnectRpcClient(RunService)
  const params = useParams<RunDetailsPageParams>()
  const runDetails = useRunDetails(params.runId)
  const mainActionDetails = useWatchActionDetails(
    runDetails.data?.details?.action?.id?.name,
  )

  const actionData = useActionData({
    actionDetails: mainActionDetails.data,
    enabled: !!mainActionDetails.data,
  })

  const [errorMessage, setErrorMessage] = useState('')
  const organization = useOrg()

  const handleRerunClick = useCallback(async () => {
    try {
      const actionSpec = runDetails.data?.details?.action?.spec
      const taskSpec =
        actionSpec?.case === 'task' ? actionSpec.value : undefined
      if (!taskSpec) {
        setErrorMessage('Missing task spec for rerun.')
        return
      }
      const existingRunSpec = runDetails.data?.details?.runSpec
      const mergedEnvs = getMergedRunAndTaskEnvs(
        existingRunSpec?.envs?.values ?? [],
        taskSpec,
      )
      const runSpec = create(RunSpecSchema, {
        labels: existingRunSpec?.labels,
        annotations: existingRunSpec?.annotations,
        envs: create(EnvsSchema, { values: mergedEnvs }),
        interruptible: existingRunSpec?.interruptible,
        overwriteCache: existingRunSpec?.overwriteCache ?? false,
        cluster: existingRunSpec?.cluster ?? '',
        rawDataStorage: existingRunSpec?.rawDataStorage,
        securityContext: existingRunSpec?.securityContext,
        cacheConfig: existingRunSpec?.cacheConfig,
      })
      const response = await client.createRun({
        id: {
          case: 'projectId',
          value: {
            organization: organization,
            name: params.project,
            domain: params.domain,
          },
        },
        task: {
          value: taskSpec,
          case: 'taskSpec',
        },
        inputs: actionData.data!.inputs,
        runSpec,
        source: RunSource.WEB,
      })
      const action = response.run!.action
      if (!action) {
        return setErrorMessage(
          'There was an error creating the rerun. Try again later.',
        )
      }
      const path = `/domain/${action?.id?.run?.domain}/project/${action?.id?.run?.project}/runs/${action?.id?.run?.name}?i=${action?.id?.name}`
      console.log('Rerun modal pushing to path', path)
      router.push(path)
      setIsOpen(false)
    } catch (e) {
      console.error(e)
      setErrorMessage('There was an error. Try again later.')
    }
  }, [
    client,
    runDetails.data,
    params.domain,
    params.project,
    actionData.data,
    setIsOpen,
    router,
    organization,
  ])

  useEffect(() => {
    // reset error message on close
    if (!isOpen) {
      setErrorMessage('')
    }
  }, [isOpen])

  const dialogButtons: ButtonConfig[] = useMemo(() => {
    return [
      {
        color: 'union',
        displayText: 'Rerun',
        onClick: handleRerunClick,
      },
      {
        color: 'dark/zinc',
        displayText: 'Cancel',
        onClick: () => setIsOpen(false),
        outline: true,
      },
    ]
  }, [handleRerunClick, setIsOpen])

  return (
    <>
      <SimpleDialog
        buttons={dialogButtons}
        content={<ConfirmRerunDialogContent message={errorMessage} />}
        isOpen={isOpen}
        setIsOpen={setIsOpen}
      />
    </>
  )
}
