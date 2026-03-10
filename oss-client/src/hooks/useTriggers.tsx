import {
  ProjectIdentifierSchema,
  TriggerName,
  TriggerNameSchema,
} from '@/gen/flyteidl2/common/identifier_pb'
import { Filter, Filter_Function } from '@/gen/flyteidl2/common/list_pb'
import type {
  Trigger,
  TriggerDetails,
} from '@/gen/flyteidl2/trigger/trigger_definition_pb'
import type {
  GetTriggerDetailsResponse,
  ListTriggersResponse,
} from '@/gen/flyteidl2/trigger/trigger_service_pb'
import { TriggerService } from '@/gen/flyteidl2/trigger/trigger_service_pb'
import { getFilter } from '@/lib/filterUtils'
import { create } from '@bufbuild/protobuf'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useMemo, useRef } from 'react'
import { useConnectRpcClient } from './useConnectRpc'
import { QuerySort } from './useQueryParamSort'

type ListTriggersProps = {
  org: string
  domain: string
  projectId: string
  search: string
  sort?: QuerySort
  taskName?: string
  taskVersion?: string
  filters?: Filter[]
  enabled?: boolean
}

const getProjectIdentifier = ({
  org,
  domain,
  projectId,
}: Pick<ListTriggersProps, 'domain' | 'projectId' | 'org'>) =>
  create(ProjectIdentifierSchema, {
    domain,
    name: projectId,
    organization: org,
  })

export const getTriggersQueryKey = ({
  org,
  projectId,
  domain,
  search,
  sort,
  taskName,
  taskVersion,
  filters,
}: ListTriggersProps) => {
  const key = ['triggers', org, projectId, domain]
  return key
    .concat(search ? [search] : [])
    .concat(taskName ? [taskName] : [])
    .concat(taskVersion ? [taskVersion] : [])
    .concat(filters ? [JSON.stringify(filters)] : [])
    .concat(sort ? sort.sortForQueryKey : [])
}
export const useListTriggers = ({
  org,
  domain,
  projectId,
  search,
  sort,
  taskName,
  taskVersion,
  filters,
  enabled = true,
}: ListTriggersProps) => {
  const client = useConnectRpcClient(TriggerService)

  const finalFilters = useMemo(() => {
    const finalFilters = search
      ? [
          {
            function: Filter_Function.CONTAINS_CASE_INSENSITIVE,
            field: 'name',
            values: [search],
          },
        ]
      : []

    if (taskName) {
      finalFilters.push(
        getFilter({
          function: Filter_Function.EQUAL,
          field: 'task_name',
          values: [taskName],
        }),
      )
    }

    if (taskVersion) {
      finalFilters.push(
        getFilter({
          function: Filter_Function.EQUAL,
          field: 'task_version',
          values: [taskVersion],
        }),
      )
    }

    return finalFilters.concat(filters || [])
  }, [search, taskName, taskVersion, filters])

  const projectIdentifier = useMemo(() => {
    return getProjectIdentifier({ domain, projectId, org })
  }, [domain, projectId, org])

  const queryKey = useMemo(
    () =>
      getTriggersQueryKey({
        org,
        projectId,
        domain,
        search,
        sort,
        taskName,
        taskVersion,
        filters,
      }),
    [projectId, domain, search, sort, org, taskName, taskVersion, filters],
  )

  const getTriggers = async () => {
    const result = await client.listTriggers({
      request: {
        filters: finalFilters,
        ...(sort && { sortByFields: [sort.sortBy] }),
      },
      scopeBy: {
        case: 'projectId',
        value: projectIdentifier,
      },
    })
    return result
  }

  return useQuery({
    queryKey,
    queryFn: getTriggers,
    enabled: enabled && !!domain && !!projectId && !!org,
    refetchInterval: 30000,
  })
}

export type UpdateTriggersProps = {
  active: boolean
  triggerNames: TriggerName[]
}

export const useUpdateTriggers = ({
  org,
  domain,
  projectId,
  search,
  taskNames,
  invalidationTimeoutMs = 3000,
}: ListTriggersProps & {
  taskNames: string[]
  invalidationTimeoutMs?: number
}) => {
  const client = useConnectRpcClient(TriggerService)
  const queryClient = useQueryClient()
  const invalidationTimeoutRef = useRef<number | null>(null)

  const queryKey = useMemo(
    () => getTriggersQueryKey({ projectId, domain, search, org }),
    [projectId, domain, search, org],
  )

  const updateTriggers = async ({
    active,
    triggerNames,
  }: UpdateTriggersProps) => {
    return await client.updateTriggers({ active, names: triggerNames })
  }

  return useMutation({
    mutationFn: updateTriggers,
    onMutate: async ({ active, triggerNames }) => {
      // Cancel any pending invalidation timeout to restart the timer
      if (invalidationTimeoutRef.current) {
        clearTimeout(invalidationTimeoutRef.current)
        invalidationTimeoutRef.current = null
      }

      // Cancel any outgoing refetches to avoid overwriting optimistic update
      await queryClient.cancelQueries({ queryKey })

      // Optimistically update ALL queries that match the base key (including those with filters)
      queryClient.setQueriesData<ListTriggersResponse>({ queryKey }, (old) => {
        if (!old?.triggers) return old
        return {
          ...old,
          triggers: old.triggers.map((trigger: Trigger) => {
            const triggerName = trigger.id?.name
            const matches = triggerNames.some(
              (tn) =>
                tn.name === triggerName?.name &&
                tn.taskName === triggerName?.taskName &&
                tn.project === triggerName?.project &&
                tn.domain === triggerName?.domain &&
                tn.org === triggerName?.org,
            )
            if (matches) {
              return { ...trigger, active }
            }
            return trigger
          }),
        }
      })
    },
    onError: () => {
      // On error, invalidate all matching queries to refetch with correct data
      queryClient.invalidateQueries({ queryKey })
    },
    onSettled: async () => {
      // invalidate all task details queries (partial match on the root key)
      await Promise.allSettled(
        taskNames.map(async (taskName) =>
          queryClient.invalidateQueries({
            queryKey: ['taskDetails', org, projectId, domain, taskName],
          }),
        ),
      )

      // Clear any existing timeout and start a new timer
      if (invalidationTimeoutRef.current) {
        clearTimeout(invalidationTimeoutRef.current)
      }
      // eslint-disable-next-line no-restricted-globals
      invalidationTimeoutRef.current = window.setTimeout(() => {
        // Invalidate all queries that match the base key (including those with filters)
        queryClient.invalidateQueries({
          queryKey: queryKey,
        })
        invalidationTimeoutRef.current = null
      }, invalidationTimeoutMs)
    },
  })
}

export interface GetTriggerDetailsProps {
  org: string
  domain: string
  projectId: string
  name: string
  taskName: string
}
export const useGetTriggerDetails = ({
  org,
  domain,
  projectId,
  taskName,
  name,
}: GetTriggerDetailsProps) => {
  const client = useConnectRpcClient(TriggerService)
  const queryClient = useQueryClient()

  const queryKey = useMemo(
    () => ['triggers', org, projectId, domain, name, taskName] as const,
    [projectId, domain, name, taskName, org],
  )

  const getTrigger = async () => {
    const result = await client.getTriggerDetails({
      name: create(TriggerNameSchema, {
        project: projectId,
        domain,
        name,
        org,
        taskName,
      }),
    })
    return result
  }

  // Compute initial data from cache if org is available
  const initialData = useMemo(() => {
    // Skip if org is empty - can't match queries without org
    if (!org || !domain || !projectId) {
      return undefined
    }

    // Try to find the trigger in cached list queries
    // Match queries that start with ['triggers', org, projectId, domain]
    const queries = queryClient.getQueriesData<ListTriggersResponse>({
      predicate: (query) => {
        const key = query.queryKey
        return (
          Array.isArray(key) &&
          key.length >= 4 &&
          key[0] === 'triggers' &&
          key[1] === org &&
          key[2] === projectId &&
          key[3] === domain
        )
      },
    })

    for (const [, data] of queries) {
      if (!data?.triggers) continue

      const trigger = data.triggers.find(
        (t) => t.id?.name?.name === name && t.id?.name?.taskName === taskName,
      )
      if (trigger) {
        return {
          trigger: {
            ...trigger,
            spec: { active: trigger.active },
          } as unknown as TriggerDetails,
        } as GetTriggerDetailsResponse
      }
    }
    return undefined
  }, [queryClient, org, projectId, domain, name, taskName])

  return useQuery({
    queryKey,
    queryFn: getTrigger,
    initialData,
    enabled: !!domain && !!projectId && !!name && !!taskName && !!org,
  })
}

export interface GetTriggerActivityProps {
  org: string
  domain: string
  projectId: string
  name: string
  taskName: string
  limit?: number
}
