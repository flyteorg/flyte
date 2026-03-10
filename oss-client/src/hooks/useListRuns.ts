import { Filter_Function } from '@/gen/flyteidl2/common/list_pb'
import {
  ListRunsRequestSchema,
  RunService,
} from '@/gen/flyteidl2/workflow/run_service_pb'
import { create } from '@bufbuild/protobuf'
import { useQuery } from '@tanstack/react-query'
import { useMemo } from 'react'
import { useConnectRpcClient } from './useConnectRpc'
import { useOrg } from './useOrg'
import { ProjectIdentifierSchema } from '@/gen/flyteidl2/common/identifier_pb'

export const useListRuns = ({
  domain,
  enabled,
  name = '',
  project,
}: {
  domain: string | undefined
  enabled?: boolean
  name?: string
  project: string | undefined
}) => {
  const client = useConnectRpcClient(RunService)
  const org = useOrg()

  const filters = useMemo(() => {
    const nameFilter = name
      ? [
          {
            function: Filter_Function.CONTAINS_CASE_INSENSITIVE,
            field: 'task_name',
            values: [name],
          },
        ]
      : []
    return [...nameFilter]
  }, [name])

  const projectIdentifier = useMemo(() => {
    return create(ProjectIdentifierSchema, {
      domain,
      name: project,
      organization: org,
    })
  }, [domain, org, project])

  const request = useMemo(() => {
    return create(ListRunsRequestSchema, {
      request: {
        limit: 15,
        filters,
      },
      scopeBy: {
        case: 'projectId',
        value: projectIdentifier,
      },
    })
  }, [filters, projectIdentifier])

  const queryFn = () => {
    return client.listRuns(request)
  }

  const isEnabled = (enabled ?? true) && !!org && !!domain && !!project

  return useQuery({
    enabled: isEnabled,
    queryKey: ['listRuns', project, domain, name],
    queryFn,
  })
}
