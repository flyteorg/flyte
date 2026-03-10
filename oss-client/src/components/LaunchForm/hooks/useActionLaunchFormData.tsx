import { useRunStore } from '@/components/pages/RunDetails/state/RunStore'
import { getFormDataFromSchemaDefaults } from '@/components/SchemaForm/utils'
import { useActionData } from '@/hooks/useActionData'
import { useLiteralToJson } from '@/hooks/useLiteralToJsonSchema'
import { useRunDetails } from '@/hooks/useRunDetails'
import { useTaskSpecLaunchForm } from '@/hooks/useTaskSpecLaunchForm'
import { useWatchActionDetails } from '@/hooks/useWatchActionDetails'
import { getMergedRunAndTaskEnvs } from '@/lib/actionUtils'
import { isTaskSpec } from '@/lib/actionSpecUtils'
import { keys, merge } from 'lodash'
import { useEffect, useMemo, useRef } from 'react'
import { useForm } from 'react-hook-form'
import { type KVPair, LaunchFormState } from '../Tabs/types'

export type UseActionLaunchFormDataProps = {
  id: string | undefined
  actionDetailsQuery: ReturnType<typeof useWatchActionDetails>
}

export const useActionLaunchFormData = ({
  id,
  actionDetailsQuery,
}: UseActionLaunchFormDataProps) => {
  const initialRef = useRef<LaunchFormState | null>(null)
  const run = useRunStore((s) => s.run?.action)
  const runId = run?.id?.run?.name || ''
  const runData = useRunDetails(runId)

  const selectedActionSpec = isTaskSpec(actionDetailsQuery.data?.spec)
    ? actionDetailsQuery.data?.spec.value
    : undefined

  const { data: actionData, isFetched: isActionDataFetched } = useActionData({
    actionDetails: actionDetailsQuery.data,
    enabled: !!actionDetailsQuery.data?.id,
  })

  const formMethods = useForm<LaunchFormState>({
    defaultValues: {
      envs: [],
      labels: [],
      interruptible: undefined,
      overwriteCache: false,
      inputs: {},
    },
  })

  const taskQuery = useTaskSpecLaunchForm({
    taskSpec: selectedActionSpec,
    enabled: !!selectedActionSpec,
  })

  const literalsQuery = useLiteralToJson({
    literals: actionData?.inputs?.literals,
    variables: selectedActionSpec?.taskTemplate?.interface?.inputs,
  })

  const hydratedDefaults: LaunchFormState | undefined = useMemo(() => {
    const hasTaskData = taskQuery.isFetched && !!taskQuery.data?.json

    if (!hasTaskData && !literalsQuery.isFetched) {
      return undefined
    }

    const jsonInputs = merge(taskQuery.data?.json, literalsQuery.data?.json, {})
    const runSpec = runData.data?.details?.runSpec
    const processedLabels: KVPair[] = runSpec
      ? keys(runSpec?.labels?.values || {}).reduce((prev, key) => {
          const value = runSpec?.labels?.values?.[key] || ''
          prev.push({ key, value })
          return prev
        }, [] as KVPair[])
      : []

    const overwriteCache = runSpec
      ? (runSpec.cacheConfig?.overwriteCache ?? runSpec.overwriteCache ?? false)
      : false
    const interruptible = runSpec?.interruptible ?? undefined
    // Merge run envs + task envs (task wins); filter internal _F_E_VS for display
    const mergedEnvs =
      runSpec && selectedActionSpec
        ? getMergedRunAndTaskEnvs(
            runSpec.envs?.values ?? [],
            selectedActionSpec,
          )
        : selectedActionSpec
          ? getMergedRunAndTaskEnvs([], selectedActionSpec)
          : runSpec
            ? (runSpec.envs?.values ?? [])
            : []
    const envs = mergedEnvs
      .filter((e) => e.key !== '_F_E_VS')
      .map((e) => ({ key: e.key, value: e.value }))

    return {
      envs,
      labels: processedLabels,
      interruptible,
      overwriteCache,
      inputs: jsonInputs,
      formData: getFormDataFromSchemaDefaults(jsonInputs),
      runName: '',
      context: actionData?.inputs?.context ?? [],
    } as LaunchFormState
  }, [
    taskQuery.isFetched,
    taskQuery.data,
    selectedActionSpec,
    literalsQuery.isFetched,
    literalsQuery.data,
    actionData?.inputs?.context,
    runData.data,
  ])

  useEffect(() => {
    if (!hydratedDefaults) return
    const prev = initialRef.current
    const changed =
      !prev || JSON.stringify(prev) !== JSON.stringify(hydratedDefaults)
    if (changed) {
      initialRef.current = hydratedDefaults
      formMethods.reset(hydratedDefaults, {
        keepDirty: false,
        keepTouched: false,
        keepErrors: false,
      })
    }
  }, [hydratedDefaults, formMethods])

  const drawerMeta = useMemo(
    () => ({
      title: 'Run:',
      breadcrumbs: [
        { label: 'Task', value: selectedActionSpec?.shortName },
        {
          label: 'Version',
          value: id,
        },
      ],
    }),
    [selectedActionSpec, id],
  )

  return {
    drawerMeta,
    isDataFetched:
      taskQuery.isFetched &&
      isActionDataFetched &&
      actionDetailsQuery.isFetched,
    formMethods,
    initialSnapshot: initialRef.current,
  }
}
