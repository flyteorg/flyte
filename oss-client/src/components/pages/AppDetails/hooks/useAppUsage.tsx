import { useMemo } from 'react'
import { useConnectRpcClient } from '@/hooks/useConnectRpc'
import { UsageService } from '@/gen/usage/service_pb'
import {
  ExecutionMetric,
  GetAppMetricsRequestSchema,
} from '@/gen/usage/payload_pb'
import { useQuery } from '@tanstack/react-query'
import { create } from '@bufbuild/protobuf'
import { Timestamp } from '@bufbuild/protobuf/wkt'

export const useAppUsage = ({
  appId,
  domain,
  endTime,
  org,
  project,
  startTime,
}: {
  appId: string | undefined
  domain: string
  metrics: ExecutionMetric[]
  org: string
  project: string
  endTime?: Timestamp | undefined
  startTime?: Timestamp | undefined
}) => {
  const client = useConnectRpcClient(UsageService)

  const req = useMemo(() => {
    return create(GetAppMetricsRequestSchema, {
      id: {
        name: appId,
        domain,
        org,
        project,
      },
      ...(startTime && { startTime }),
      ...(endTime && { endTime }),
    })
  }, [appId, domain, endTime, org, project, startTime])

  const getMetrics = async () => {
    return client.getAppMetrics(req)
  }

  return useQuery({
    queryKey: [
      appId,
      domain,
      org,
      project,
      `s-${startTime?.seconds}`,
      `e-${endTime?.seconds}`,
    ],
    queryFn: getMetrics,
    refetchInterval: 30000,
    retry: false,
  })
}
