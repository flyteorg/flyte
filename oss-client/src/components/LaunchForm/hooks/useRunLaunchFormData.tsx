import { useRunStore } from '@/components/pages/RunDetails/state/RunStore'
import { getFormDataFromSchemaDefaults } from '@/components/SchemaForm/utils'
import { ActionPhase } from '@/gen/flyteidl2/common/phase_pb'
import { useActionData } from '@/hooks/useActionData'
import { useLaunchFormState } from '@/hooks/useLaunchFormState'
import { useLiteralToJson } from '@/hooks/useLiteralToJsonSchema'
import { useRunDetails } from '@/hooks/useRunDetails'
import { useTaskSpecLaunchForm } from '@/hooks/useTaskSpecLaunchForm'
import { keys, merge } from 'lodash'
import { useEffect, useMemo, useRef } from 'react'
import { useForm } from 'react-hook-form'
import { KVPair, LaunchFormState } from '../Tabs/types'

export const useRunLaunchFormData = () => {
  const run = useRunStore((s) => s.run?.action)
  const { taskSpec } = useLaunchFormState()
  const runId = run?.id?.run?.name || ''
  const runData = useRunDetails(runId)
  const isTerminalPhase =
    run?.status?.phase && run?.status?.phase >= ActionPhase.SUCCEEDED
  const initialRef = useRef<LaunchFormState | null>(null)

  const actionDetails = runData.data?.details?.action

  const { data: actionData, isError: isActionDataError } = useActionData({
    actionDetails,
    enabled: !!actionDetails,
  })
  const { inputs } = taskSpec?.taskTemplate?.interface || {}

  const taskQuery = useTaskSpecLaunchForm({
    taskSpec: taskSpec ?? undefined,
    enabled: !!taskSpec,
  })

  const literalsQuery = useLiteralToJson({
    literals: actionData?.inputs?.literals ?? [],
    variables: inputs,
  })

  const methods = useForm<LaunchFormState>({
    defaultValues: {
      envs: [],
      labels: [],
      interruptible: undefined,
      overwriteCache: false,
      inputs: {},
    },
  })

  const hydratedDefaults: LaunchFormState | undefined = useMemo(() => {
    const runSpec = runData.data?.details?.runSpec

    const processedLabels: KVPair[] = keys(
      runSpec?.labels?.values || {},
    ).reduce((prev, key) => {
      const value = runSpec?.labels?.values?.[key] || ''
      prev.push({ key, value })
      return prev
    }, [] as KVPair[])

    // If we have taskSpec, we can show the inputs schema even if literalsQuery fails
    // Merge taskQuery (schema from taskSpec) with literalsQuery (actual values from run)
    // Prioritize literalsQuery data if available, fallback to taskQuery schema
    const taskSchema = taskQuery.data?.json ?? {}
    const literalsData = literalsQuery.data?.json ?? {}
    const jsonInputs = merge({}, taskSchema, literalsData)

    // Allow showing form if:
    // 1. We have taskSpec schema (taskQuery is fetched) OR
    // 2. We have runData OR
    // 3. literalsQuery is fetched (even if it failed, we still want to show the schema)
    const hasTaskSchema = taskQuery.isFetched && !!taskQuery.data?.json
    const hasRunData = !!runData.data
    const hasLiteralsFetched = literalsQuery.isFetched || isActionDataError

    if (!hasTaskSchema && !hasRunData && !hasLiteralsFetched) {
      return undefined
    }

    // Prefer cacheConfig.overwriteCache when present (RunSpec.overwrite_cache is deprecated)
    const overwriteCache =
      runSpec?.cacheConfig?.overwriteCache ?? runSpec?.overwriteCache ?? false

    // Filter out internal env var that should not be shown when hydrating from a run
    const envs = (runSpec?.envs?.values || []).filter(
      (e) => e.key !== '_F_E_VS',
    )

    return {
      envs,
      labels: processedLabels,
      interruptible: runSpec?.interruptible ?? undefined,
      overwriteCache,
      inputs: jsonInputs,
      formData: getFormDataFromSchemaDefaults(jsonInputs),
      runName: '',
      context: actionData?.inputs?.context ?? [],
    } as LaunchFormState
  }, [
    taskQuery.isFetched,
    taskQuery.data,
    literalsQuery.isFetched,
    literalsQuery.data,
    isActionDataError,
    runData.data,
    actionData?.inputs?.context,
  ])

  // reset form once when the hydrated defaults first appear and if they change
  useEffect(() => {
    if (!hydratedDefaults) return
    const prev = initialRef.current
    const changed =
      !prev || JSON.stringify(prev) !== JSON.stringify(hydratedDefaults)
    if (changed) {
      initialRef.current = hydratedDefaults
      methods.reset(hydratedDefaults, {
        keepDirty: false,
        keepTouched: false,
        keepErrors: false,
      })
    }
  }, [hydratedDefaults, methods])

  const drawerMeta = useMemo(
    () => ({
      title: 'Rerun:',
      breadcrumbs: [
        { value: run?.id?.name },
        {
          label: 'Version:',
          value: taskSpec?.taskTemplate?.id?.version,
        },
        { label: 'Run ID:', value: run?.id?.run?.name },
      ],
    }),
    [run, taskSpec],
  )

  return {
    action: run,
    drawerMeta,
    // Consider data fetched if we have taskSpec schema OR literals query is fetched/errored
    isDataFetched:
      (taskQuery.isFetched && !!taskQuery.data?.json) ||
      literalsQuery.isFetched ||
      isActionDataError,
    isTerminalPhase,
    formMethods: methods,
    spec: actionDetails?.spec?.value,
    initialSnapshot: initialRef.current,
  }
}
