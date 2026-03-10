import {
  CreateTriggerState,
  CronTrigger,
  FixedRateTrigger,
  KVPair,
} from '@/components/pages/ListTriggers/CreateTrigger/Form/types'
import type { Cron, FixedRate } from '@/gen/flyteidl2/task/common_pb'
import {
  FixedRateUnit,
  TriggerAutomationSpecSchema,
  TriggerAutomationSpecType,
} from '@/gen/flyteidl2/task/common_pb'
import { useMemo } from 'react'

import { TriggerNameSchema } from '@/gen/flyteidl2/common/identifier_pb'
import { useQueryClient } from '@tanstack/react-query'

import {
  DeployTriggerRequestSchema,
  GetTriggerDetailsRequestSchema,
  TriggerService,
} from '@/gen/flyteidl2/trigger/trigger_service_pb'
import { create, JsonObject } from '@bufbuild/protobuf'
import { useMutation } from '@tanstack/react-query'
import { useConnectRpcClient } from './useConnectRpc'
import { TriggerSpecSchema } from '@/gen/flyteidl2/trigger/trigger_definition_pb'
import { TranslatorService } from '@/gen/flyteidl2/workflow/translator_service_pb'
import {
  VariableEntry,
  VariableMapSchema,
} from '@/gen/flyteidl2/core/interface_pb'
import { getTriggersQueryKey } from './useTriggers'
import { getTaskIdentifier } from '@/hooks/useTaskDetails'
import {
  convertRjsfInternalFormats,
  skipUndefinedFormValues,
} from '@/components/SchemaForm/utils'
import { TaskService } from '@/gen/workflow/task_service_pb'
import { Code, ConnectError } from '@connectrpc/connect'

const OPTIMISTIC_LOCK_MESSAGE = 'optimistic lock failure'

function isOptimisticLockError(error: unknown): boolean {
  if (
    !(error instanceof ConnectError) ||
    error.code !== Code.FailedPrecondition
  ) {
    return false
  }
  const rawMessage = (error as { rawMessage?: string }).rawMessage ?? ''
  return (
    error.message.includes(OPTIMISTIC_LOCK_MESSAGE) ||
    rawMessage.includes(OPTIMISTIC_LOCK_MESSAGE)
  )
}

const getAutomationType = (
  createTriggerState: CreateTriggerState,
):
  | { case: 'cron'; value: Cron }
  | { case: 'rate'; value: FixedRate }
  | { case: undefined; value?: undefined } => {
  switch (createTriggerState.automationType) {
    case 'cron-schedule': {
      const automation = createTriggerState as CronTrigger
      return {
        case: 'cron',
        value: {
          expression: automation.cronExpression,
          timezone: automation.timezone,
        } as Cron,
      }
    }
    case 'fixed-rate': {
      const rate = createTriggerState as FixedRateTrigger
      const now = Date.now()
      const defaultStartTime = new Date(now + 10_000).toISOString()

      let startTimeValue: number | string = defaultStartTime

      // Use the user's start time only if it's valid and in the future
      if (rate.startTime) {
        const inputDate = new Date(rate.startTime)
        if (Number.isFinite(inputDate.getTime()) && inputDate.getTime() > now) {
          startTimeValue = rate.startTime
        }
      }

      const startDate = new Date(startTimeValue)

      return {
        case: 'rate',
        value: {
          value: rate.interval > 0 ? rate.interval : 1,
          unit: FixedRateUnit.MINUTE,
          startTime: {
            seconds: Math.floor(startDate.getTime() / 1000),
            nanos: (startDate.getMilliseconds() || 0) * 1_000_000,
          },
        } as unknown as FixedRate,
      }
    }
    default:
      return {
        case: undefined,
        value: undefined,
      }
  }
}

const convertKvPairs = (kvPairs: KVPair[]) =>
  kvPairs
    .filter(({ key, value }) => key?.trim() && value?.trim())
    .reduce(
      (acc, { key, value }) => {
        const trimmedKey = key.trim()
        const trimmedValue = value.trim()
        return { ...acc, [trimmedKey]: trimmedValue }
      },
      {} as { [key: string]: string },
    )

export const useCreateTrigger = ({
  domain,
  org,
  project,
}: {
  domain: string
  org: string
  project: string
}) => {
  const client = useConnectRpcClient(TriggerService)
  const queryClient = useQueryClient()
  const translatorClient = useConnectRpcClient(TranslatorService)
  const taskClient = useConnectRpcClient(TaskService)

  const queryKey = useMemo(() => {
    return getTriggersQueryKey({ org, domain, projectId: project, search: '' })
  }, [domain, org, project])

  const handleDeployTrigger = async ({
    createTriggerState,
    taskName,
    taskVersion,
  }: {
    createTriggerState: CreateTriggerState
    taskName: string | undefined
    taskVersion: string | undefined
  }) => {
    try {
      const triggerNameReq = create(TriggerNameSchema, {
        domain,
        org,
        project,
        name: createTriggerState.name,
        taskName,
      })

      const automationReq = create(TriggerAutomationSpecSchema, {
        type: TriggerAutomationSpecType.TYPE_SCHEDULE,
        automation: {
          case: 'schedule',
          value: {
            expression: getAutomationType(createTriggerState),
            kickoffTimeInputArg: createTriggerState.kickoffTimeInputArg,
          },
        },
      })

      const taskId = getTaskIdentifier({
        ...triggerNameReq,
        name: taskName ?? '',
        version: taskVersion ?? '',
      })

      const triggerTaskDetailsData = await taskClient.getTaskDetails({ taskId })

      let inputs =
        triggerTaskDetailsData?.details?.spec?.taskTemplate?.interface?.inputs

      // if an input is being used as kickoffTimeArg, remove it from the variable map
      if (createTriggerState.kickoffTimeInputArg) {
        const updatedVariables: VariableEntry[] = (
          triggerTaskDetailsData?.details?.spec?.taskTemplate?.interface?.inputs
            ?.variables ?? []
        ).filter((v) => v.key !== createTriggerState.kickoffTimeInputArg)
        // Properly construct VariableMap using the schema/factory
        inputs = create(VariableMapSchema, {
          ...inputs,
          variables: updatedVariables,
        })
      }

      const { literals } = await translatorClient.jsonValuesToLiterals({
        variables: inputs,
        values: skipUndefinedFormValues(
          convertRjsfInternalFormats(
            createTriggerState.inputs,
            createTriggerState.formData as JsonObject | undefined,
          ),
        ),
      })

      const triggerSpecReq = create(TriggerSpecSchema, {
        active: createTriggerState.activeOnCreation === false ? false : true,
        description: createTriggerState.description,
        inputs: {
          literals: literals,
        },
        runSpec: {
          ...(createTriggerState.labels &&
            createTriggerState.labels.length > 0 && {
              labels: {
                values: convertKvPairs(createTriggerState.labels),
              },
            }),
          ...(createTriggerState.annotations &&
            createTriggerState.annotations.length > 0 && {
              annotations: {
                values: convertKvPairs(createTriggerState.annotations),
              },
            }),
          ...(createTriggerState.envVars &&
            createTriggerState.envVars.length > 0 && {
              envs: {
                values: createTriggerState.envVars,
              },
            }),
          ...(createTriggerState.interruptible !== undefined && {
            interruptible: createTriggerState.interruptible,
          }),
          ...(createTriggerState.overwriteCache !== undefined && {
            overwriteCache: createTriggerState.overwriteCache,
          }),
          ...(createTriggerState.dataOutputLocation && {
            rawDataStorage: {
              rawDataPrefix: createTriggerState.dataOutputLocation,
            },
          }),
          ...(createTriggerState.serviceAccount && {
            securityContext: {
              runAs: {
                k8sServiceAccount: createTriggerState.serviceAccount,
              },
            },
          }),
        },
        taskVersion,
      })

      const req = create(DeployTriggerRequestSchema, {
        name: triggerNameReq,
        automationSpec: automationReq,
        spec: triggerSpecReq,
      })

      try {
        await client.deployTrigger(req)
      } catch (e) {
        if (isOptimisticLockError(e)) {
          let details
          try {
            details = await client.getTriggerDetails(
              create(GetTriggerDetailsRequestSchema, { name: triggerNameReq }),
            )
          } catch {
            throw e
          }
          const revision = details.trigger?.id?.revision
          if (revision != null) {
            await client.deployTrigger(
              create(DeployTriggerRequestSchema, {
                name: triggerNameReq,
                automationSpec: automationReq,
                spec: triggerSpecReq,
                revision,
              }),
            )
            return
          }
        }
        throw e
      }
    } catch (e) {
      console.error('Error creating trigger', e)
      throw e
    }
  }

  return useMutation({
    mutationFn: handleDeployTrigger,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey })
    },
  })
}
