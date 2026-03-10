import { create } from '@bufbuild/protobuf'
import { useConnectRpcClient } from './useConnectRpc'
import { UsageService } from '@/gen/usage/service_pb'
import { GetActionAttemptMetricsRequestSchema } from '@/gen/usage/payload_pb'
import { ExecutionMetric } from '@/gen/usage/payload_pb'
import { useQuery } from '@tanstack/react-query'

export type UseActionMetricsProps = {
  actionName: string
  attempt: number
  domain: string
  enabled: boolean
  isTerminal: boolean
  metrics: ExecutionMetric[]
  org: string
  project: string
  runName: string
}

const getMetricsIdentifier = ({
  actionId,
  attempt,
  domain,
  metrics,
  org,
  project,
  runName,
}: {
  actionId: string
  attempt: number
  domain: string
  metrics: ExecutionMetric[]
  org: string
  project: string
  runName: string | undefined
}) => {
  return create(GetActionAttemptMetricsRequestSchema, {
    id: {
      actionId: {
        name: actionId,
        run: {
          domain,
          org,
          project,
          name: runName,
        },
      },
      attempt,
    },
    metrics,
  })
}

export const useActionMetrics = ({
  actionName,
  attempt,
  domain,
  enabled,
  isTerminal,
  metrics,
  org,
  project,
  runName,
}: UseActionMetricsProps) => {
  const client = useConnectRpcClient(UsageService)
  const requests = metrics.map((m) => ({
    ...getMetricsIdentifier({
      actionId: actionName,
      attempt,
      domain,
      metrics: [m],
      org,
      project,
      runName,
    }),
  }))
  const getData = async () => {
    const promises = requests.map((r) => client.getActionAttemptMetrics(r))
    const response = await Promise.all(promises)
    return response
  }
  const queryResult = useQuery({
    enabled,
    queryKey: ['metrics', actionName, attempt, metrics],
    queryFn: getData,
    refetchInterval: isTerminal ? false : 30000,
    retry: false,
  })
  return queryResult
}
