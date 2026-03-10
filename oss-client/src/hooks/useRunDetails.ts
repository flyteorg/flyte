import { RunDetailsPageParams } from '@/components/pages/RunDetails/types'
import { RunIdentifierSchema } from '@/gen/flyteidl2/common/identifier_pb'
import {
  GetRunDetailsResponse,
  RunService,
} from '@/gen/flyteidl2/workflow/run_service_pb'
import { create } from '@bufbuild/protobuf'
import { useQuery } from '@tanstack/react-query'
import { useParams } from 'next/navigation'
import { useMemo } from 'react'
import { useConnectRpcClient } from './useConnectRpc'
import { useOrg } from './useOrg'

export const getRunIdentifier = ({
  domain,
  name,
  org,
  project,
}: {
  domain: string
  name: string
  org: string
  project: string
}) =>
  create(RunIdentifierSchema, {
    name,
    domain,
    org,
    project,
  })

const useRunIdentifier = (runId?: string | null) => {
  const org = useOrg()
  const params = useParams<RunDetailsPageParams>()
  return useMemo(() => {
    return getRunIdentifier({
      domain: params.domain,
      name: runId || params.runId,
      org,
      project: params.project,
    })
  }, [runId, org, params.domain, params.project, params.runId])
}

export function useRunDetails(runId?: string | null) {
  const client = useConnectRpcClient(RunService)
  const runIdentifier = useRunIdentifier(runId)
  return useQuery({
    queryKey: ['runData', runId],
    queryFn: async () => {
      if (!runIdentifier) return null
      const response = await client.getRunDetails({ runId: runIdentifier })
      return response as unknown as GetRunDetailsResponse
    },
    enabled: !!runId,
  })
}
