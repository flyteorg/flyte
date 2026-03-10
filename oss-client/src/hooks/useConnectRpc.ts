import { useMemo } from 'react'
import { createClient } from '@connectrpc/connect'
import { DescService } from '@bufbuild/protobuf'
import { useTransport } from '@connectrpc/connect-query'

export function useConnectRpcClient<T extends DescService>(service: T) {
  const transport = useTransport()
  // We memoize the client, so that we only create one instance per service.
  return useMemo(() => createClient(service, transport), [service, transport])
}
