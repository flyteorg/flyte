import { Button } from '@/components/Button'
import { RerunIcon } from '@/components/icons/RerunIcon'
import { PlayIcon } from '@/components/icons/PlayIcon'
import { RunDetailsPageParams } from '@/components/pages/RunDetails/types'
import { useTriggerRun } from '@/components/RunButton/useTriggerRun'
import {
  convertRjsfInternalFormats,
  skipUndefinedFormValues,
} from '@/components/SchemaForm/utils'
import {
  RunIdentifierSchema,
  TriggerName,
  TriggerNameSchema,
} from '@/gen/flyteidl2/common/identifier_pb'
import { KeyValuePair } from '@/gen/flyteidl2/core/literals_pb'
import { NamedLiteral } from '@/gen/flyteidl2/task/common_pb'
import { TaskSpec } from '@/gen/flyteidl2/task/task_definition_pb'
import { RunSource } from '@/gen/flyteidl2/workflow/run_definition_pb'
import {
  CreateRunRequest,
  CreateRunRequestSchema,
  RunService,
} from '@/gen/flyteidl2/workflow/run_service_pb'
import { TranslatorService } from '@/gen/flyteidl2/workflow/translator_service_pb'
import { useConnectRpcClient } from '@/hooks/useConnectRpc'
import { useLaunchFormState } from '@/hooks/useLaunchFormState'
import { useOrg } from '@/hooks/useOrg'
import { useTaskDetails } from '@/hooks/useTaskDetails'
import { create, JsonObject } from '@bufbuild/protobuf'
import { useParams, useRouter } from 'next/navigation'
import { useCallback } from 'react'
import { useFormContext } from 'react-hook-form'
import {
  ErrorWithRawMessage,
  isRunNameValid,
  KVPair,
  LaunchFormState,
} from './types'

const filterOutEmptyKVPairs = (items: KVPair[] | undefined) => {
  if (
    !items ||
    items?.length === 0 ||
    items.every((i) => i.key === '' && i.value === '')
  ) {
    return undefined
  }
  return items.filter((i) => i.key !== '' && i.value !== '')
}

/** For env vars only: keep pairs with a non-empty key; empty value is allowed. */
const filterEnvsForSubmit = (items: KVPair[] | undefined) => {
  if (
    !items ||
    items.length === 0 ||
    items.every((i) => !(i.key?.trim() ?? ''))
  ) {
    return undefined
  }
  const filtered = items.filter((i) => (i.key?.trim() ?? '') !== '')
  return filtered.length > 0 ? filtered : undefined
}

const formatEnvs = (envs: LaunchFormState['envs']) => {
  const filteredEnvs = filterEnvsForSubmit(envs)
  if (!filteredEnvs) return undefined
  return {
    values: filteredEnvs.map((kv) => ({
      key: kv.key,
      value: kv.value,
    })),
  }
}

const formatLabels = (labels: LaunchFormState['labels']) => {
  const labelsChecked = filterOutEmptyKVPairs(labels)
  if (!labelsChecked) return undefined
  return {
    values: labelsChecked.reduce(
      (acc, { key, value }) => {
        acc[key] = value
        return acc
      },
      {} as Record<string, string>,
    ),
  }
}

const filterOutEmptyKeyValuePairs = (
  items: KeyValuePair[] | undefined,
): KeyValuePair[] | undefined => {
  if (
    !items ||
    items.length === 0 ||
    items.every(
      (i) =>
        (!i.key || i.key.trim() === '') && (!i.value || i.value.trim() === ''),
    )
  ) {
    return undefined
  }
  const filtered = items.filter(
    (i) => i.key && i.key.trim() !== '' && i.value && i.value.trim() !== '',
  )
  return filtered.length > 0 ? filtered : undefined
}

type MergeLaunchFormDataProps = {
  org: string
  domain: string
  formValues: Omit<LaunchFormState, 'inputs'>
  literals: NamedLiteral[]
  name: string
  taskSpec?: TaskSpec
  triggerName?: TriggerName
}

const mergeLaunchFormData = ({
  org,
  literals,
  formValues,
  domain,
  name,
  taskSpec,
  triggerName,
}: MergeLaunchFormDataProps): CreateRunRequest | null => {
  // If triggerName is provided, use it; otherwise use taskSpec
  const taskField = triggerName
    ? {
        case: 'triggerName' as const,
        value: create(TriggerNameSchema, {
          org: triggerName.org,
          project: triggerName.project,
          domain: triggerName.domain,
          name: triggerName.name,
          taskName: triggerName.taskName,
        }),
      }
    : taskSpec
      ? {
          case: 'taskSpec' as const,
          value: taskSpec,
        }
      : null

  if (!taskField) {
    return null
  }

  const runName = formValues.runName?.trim()
  const id =
    runName && isRunNameValid(runName)
      ? {
          case: 'runId' as const,
          value: create(RunIdentifierSchema, {
            org,
            project: name,
            domain,
            name: runName,
          }),
        }
      : {
          case: 'projectId' as const,
          value: {
            domain,
            name,
            organization: org,
          },
        }

  return create(CreateRunRequestSchema, {
    id,
    task: taskField,
    inputs: {
      literals: literals,
      context: filterOutEmptyKeyValuePairs(formValues.context) ?? [],
    },
    runSpec: {
      envs: formatEnvs(formValues.envs),
      labels: formatLabels(formValues.labels),
      interruptible: formValues.interruptible,
      overwriteCache: formValues.overwriteCache,
    },
    source: RunSource.WEB,
  })
}

export const LaunchFormButtons = () => {
  const router = useRouter()
  const runClient = useConnectRpcClient(RunService)
  const translatorClient = useConnectRpcClient(TranslatorService)
  const params = useParams<RunDetailsPageParams>()
  const org = useOrg()
  const {
    buttonText,
    setIsOpen,
    taskSpec,
    triggerName,
    isBlobUploading,
  } = useLaunchFormState()

  const { handleSubmit, getValues, setError, formState } =
    useFormContext<LaunchFormState>()

  const formatErrorMessage = useCallback(
    (genericMessage: string, e: Error | unknown): string => {
      const rawMessage = (e as ErrorWithRawMessage)?.rawMessage
      if (!rawMessage) return genericMessage
      return `Error: ${rawMessage}`
    },
    [],
  )

  const { latestVersion } = useTriggerRun(triggerName ?? undefined)
  const { data: triggerTaskDetailsData } = useTaskDetails({
    version: latestVersion || '',
    name: triggerName?.taskName ?? '',
    project: triggerName?.project ?? '',
    domain: triggerName?.domain ?? '',
    org: triggerName?.org ?? '',
    enabled: !!triggerName,
  })

  const onSubmit = useCallback(async () => {
    try {
      const formValues = getValues()

      // Check for JSON errors in inputs
      if (formState.errors.inputs) {
        setError('root', {
          message: 'Please fix JSON errors in inputs before submitting',
          type: 'validation',
        })
        return
      }

      if (
        formState.errors.context ||
        formState.errors.envs ||
        formState.errors.labels
      ) {
        setError('root', {
          message: 'Please fix validation errors in the form',
          type: 'validation',
        })
        return
      }

      if (!taskSpec && !triggerName) {
        setError('root', {
          message: 'Error: missing task spec or trigger name',
          type: 'data',
        })
        return
      }

      const values = skipUndefinedFormValues(
        convertRjsfInternalFormats(
          (formValues.inputs ?? {}) as import('json-schema').JSONSchema7,
          (formValues.formData ?? {}) as JsonObject,
        ),
      )
      const literals = await translatorClient.jsonValuesToLiterals({
        variables:
          taskSpec?.taskTemplate?.interface?.inputs ??
          triggerTaskDetailsData?.details?.spec?.taskTemplate?.interface
            ?.inputs,
        values,
      })

      // Use trigger's project/domain if triggerName is present, otherwise use params
      const projectDomain = triggerName
        ? { domain: triggerName.domain, project: triggerName.project }
        : { domain: params.domain || '', project: params.project || '' }

      const requestData = mergeLaunchFormData({
        org,
        domain: projectDomain.domain,
        formValues,
        literals: literals.literals,
        name: projectDomain.project,
        taskSpec: taskSpec ?? undefined,
        triggerName: triggerName ?? undefined,
      })
      if (!requestData) {
        return
      }
      const newRun = await runClient.createRun(requestData)
      const newAction = newRun.run?.action
      const path = `/domain/${newAction?.id?.run?.domain}/project/${newAction?.id?.run?.project}/runs/${newAction?.id?.run?.name}?i=${newAction?.id?.name}`
      router.push(path)
    } catch (e) {
      console.error('Error cloning run', e)
      setError('root', {
        type: 'api',
        message: formatErrorMessage(
          'An error occurred while submitting your run',
          e,
        ),
      })
    }
  }, [
    getValues,
    formState.errors.inputs,
    formState.errors.context,
    formState.errors.envs,
    formState.errors.labels,
    taskSpec,
    translatorClient,
    org,
    params.domain,
    params.project,
    runClient,
    router,
    setError,
    formatErrorMessage,
    triggerName,
    triggerTaskDetailsData,
  ])

  return (
    <>
      <Button
        outline
        color="zinc"
        className="ml-auto dark:!text-(--system-white)"
        onClick={() => setIsOpen(false)}
      >
        Cancel
      </Button>
      <Button
        color="union"
        onClick={handleSubmit(onSubmit)}
        disabled={
          !!formState.errors.inputs ||
          !!formState.errors.runName ||
          !!formState.errors.context ||
          !!formState.errors.envs ||
          !!formState.errors.labels ||
          formState.isSubmitting ||
          isBlobUploading
        }
        title={
          isBlobUploading
            ? 'Please wait for file upload to complete'
            : formState.errors.inputs
              ? 'Please fix JSON errors in inputs before submitting'
              : formState.errors.runName
                ? 'Please fix run name validation errors'
                : formState.errors.context
                  ? 'Please fix context validation errors'
                  : formState.errors.envs
                    ? 'Please fix environment variable validation errors'
                    : formState.errors.labels
                      ? 'Please fix label validation errors'
                      : 'Submit the form'
        }
      >
        {buttonText === 'Run' ? (
          <PlayIcon className="size-3" fill="currentColor" />
        ) : (
          <RerunIcon className="size-3" />
        )}
        {buttonText}
      </Button>
    </>
  )
}
