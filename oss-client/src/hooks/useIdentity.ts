import { useQuery } from '@tanstack/react-query'
import { IdentityService } from '@/gen/flyteidl/service/identity_pb'
import { useConnectRpcClient } from './useConnectRpc'

export function useIdentity() {
  const client = useConnectRpcClient(IdentityService)

  const getIdentity = async () => {
    try {
      const response = await client.userInfo({})
      return response
    } catch (error) {
      console.error(error)
      throw error
    }
  }

  return useQuery({
    queryKey: ['identity'],
    queryFn: getIdentity,
    refetchInterval: 1000 * 60,
  })
}
