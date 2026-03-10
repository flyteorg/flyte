import { TaskDetailsPageParams } from '@/components/pages/TaskDetails/types'
import { useDefaultInputsJson } from '@/hooks/useDefaultInputsJson'
import { useOrg } from '@/hooks/useOrg'
import { useTaskDetails } from '@/hooks/useTaskDetails'
import { useTaskSpecLaunchForm } from '@/hooks/useTaskSpecLaunchForm'
import { merge } from 'lodash'
import { useParams } from 'next/navigation'
import { useEffect, useMemo, useRef } from 'react'
import { useForm } from 'react-hook-form'
import { LaunchFormState } from '../Tabs/types'
import { getFormDataFromSchemaDefaults } from '@/components/SchemaForm/utils'

type UseTaskLaunchFormDataProps = {
  version: string
  taskName?: string
  project?: string
  domain?: string
}

export const useTaskLaunchFormData = ({
  version,
  taskName,
  project,
  domain,
}: UseTaskLaunchFormDataProps) => {
  const params = useParams<TaskDetailsPageParams>()
  const org = useOrg()
  // Use provided parameters if available, otherwise fall back to params
  const taskNameToUse = taskName ?? params.name
  const projectToUse = project ?? params.project
  const domainToUse = domain ?? params.domain

  const taskDetails = useTaskDetails({
    name: taskNameToUse,
    version: version,
    project: projectToUse,
    domain: domainToUse,
    org: org,
    enabled: !!version && !!taskNameToUse && !!projectToUse && !!domainToUse,
  })

  const taskQuery = useTaskSpecLaunchForm({
    taskSpec: taskDetails.data?.details?.spec,
    enabled: !!taskDetails.data?.details?.spec,
  })

  const literalsQuery = useDefaultInputsJson(
    version,
    taskNameToUse,
    projectToUse,
    domainToUse,
  )

  const initialRef = useRef<LaunchFormState | null>(null)

  const formMethods = useForm<LaunchFormState>({
    defaultValues: {
      envs: [],
      labels: [],
      interruptible: undefined,
      overwriteCache: false,
      inputs: {},
    },
  })

  const hydratedDefaults: LaunchFormState | undefined = useMemo(() => {
    // Wait for all queries to be fetched
    if (!taskDetails.isFetched) return undefined

    // We need at least taskDetails.data to proceed (taskQuery.data?.json can be empty/undefined)
    if (!taskDetails.data) return undefined

    const literalsFetched = literalsQuery.isFetched && !literalsQuery.isError
    if (!taskQuery.isFetched && !literalsFetched) return undefined

    const jsonInputs = merge(taskQuery.data?.json, literalsQuery.data?.json, {})
    const { target, metadata } =
      taskDetails?.data?.details?.spec?.taskTemplate ?? {}
    const isContainer = target?.case === 'container'
    const envs = isContainer ? target?.value?.env : undefined
    // Prepopulate interruptible from task spec when task defines it (TaskMetadata.interruptible_value)
    const interruptible =
      metadata?.interruptibleValue?.case === 'interruptible'
        ? metadata.interruptibleValue.value
        : undefined
    if (!literalsQuery.isFetched && !literalsQuery.isError) return undefined
    if (!taskQuery.isFetched) return undefined
    if (!taskQuery.data?.json && !taskDetails.data) return undefined

    return {
      envs: envs ?? [],
      labels: [],
      interruptible,
      overwriteCache: false,
      inputs: jsonInputs,
      formData: getFormDataFromSchemaDefaults(jsonInputs),
      runName: '',
    } as LaunchFormState
  }, [
    taskQuery.data,
    taskQuery.isFetched,
    literalsQuery.data,
    literalsQuery.isFetched,
    literalsQuery.isError,
    taskDetails.data,
    taskDetails.isFetched,
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
        { label: 'Task', value: taskDetails?.data?.details?.taskId?.name },
        {
          label: 'Version',
          value: taskDetails?.data?.details?.taskId?.version,
        },
      ],
    }),
    [taskDetails?.data?.details],
  )

  // Data is fully fetched when all queries are complete and we have hydrated defaults
  // hydratedDefaults will be undefined while queries are loading, but should be set once all are fetched
  const isDataFetched = useMemo(() => {
    // All queries must be fetched (or errored for literalsQuery)
    const allQueriesFetched =
      taskDetails.isFetched &&
      taskQuery.isFetched &&
      (literalsQuery.isFetched || literalsQuery.isError)

    // We need hydratedDefaults to be ready, which means the form can be populated
    // If all queries are fetched but hydratedDefaults is still undefined, it means
    // there's an issue with the data (e.g., no task data returned)
    return allQueriesFetched && !!hydratedDefaults
  }, [
    taskDetails.isFetched,
    taskQuery.isFetched,
    literalsQuery.isFetched,
    literalsQuery.isError,
    hydratedDefaults,
  ])

  return {
    drawerMeta,
    isDataFetched,
    formMethods,
    initialSnapshot: initialRef.current,
    taskDetails: taskDetails.data?.details,
  }
}
