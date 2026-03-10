import { ActionDetails } from '@/gen/flyteidl2/workflow/run_definition_pb'
import { RunService } from '@/gen/flyteidl2/workflow/run_service_pb'
import { useQuery } from '@tanstack/react-query'
import { useConnectRpcClient } from './useConnectRpc'

export interface UseActionDataProps {
  actionDetails?: ActionDetails | null
  enabled?: boolean
}
export function useActionData({
  actionDetails,
  enabled = true,
}: UseActionDataProps) {
  const client = useConnectRpcClient(RunService)

  return useQuery({
    queryKey: ['actionData', actionDetails?.id],
    queryFn: async () => {
      if (!actionDetails) return null

      const response = await client.getActionData({
        actionId: actionDetails.id,
      })

      return response
    },
    enabled: !!actionDetails && enabled,
    refetchOnWindowFocus: false,
    staleTime: Infinity,
    experimental_prefetchInRender: true,
  })
}
