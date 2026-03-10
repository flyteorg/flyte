import {
  GetRunDetailsRequestSchema,
  RunService,
} from '@/gen/flyteidl2/workflow/run_service_pb'
import { create } from '@bufbuild/protobuf'
import { useQuery } from '@tanstack/react-query'
import { useMemo } from 'react'
import { useConnectRpcClient } from './useConnectRpc'
import { useOrg } from './useOrg'

export const useGetRun = ({
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

  const request = useMemo(() => {
    return create(GetRunDetailsRequestSchema, {
      runId: {
        domain,
        name,
        org,
        project,
      },
    })
  }, [domain, name, org, project])

  const queryFn = () => {
    return client.getRunDetails(request)
  }

  const isEnabled =
    (enabled ?? true) && !!name && !!org && !!domain && !!project

  return useQuery({
    enabled: isEnabled,
    queryKey: ['getRun', org, project, domain, name],
    queryFn,
  })
}
