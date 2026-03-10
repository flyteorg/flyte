import { VariableMap } from '@/gen/flyteidl2/core/interface_pb'
import { NamedLiteral } from '@/gen/flyteidl2/task/common_pb'
import {
  LiteralsToLaunchFormJsonRequestSchema,
  LiteralsToLaunchFormJsonResponse,
  TranslatorService,
} from '@/gen/flyteidl2/workflow/translator_service_pb'
import { create } from '@bufbuild/protobuf'
import { useQuery } from '@tanstack/react-query'
import stringify from 'safe-stable-stringify'
import { useConnectRpcClient } from './useConnectRpc'

interface UseLiteralToJsonParams {
  literals?: NamedLiteral[]
  variables?: VariableMap
}

export function useLiteralToJson(params: UseLiteralToJsonParams | null) {
  const client = useConnectRpcClient(TranslatorService)

  const queryKey = stringify(params)
  const query = useQuery({
    queryKey: ['literalsToLaunchForJson', queryKey],
    queryFn: async (): Promise<LiteralsToLaunchFormJsonResponse> => {
      if (!params) {
        throw new Error(
          'No parameters provided for literalsToLaunchFormJson call',
        )
      }

      const request = create(LiteralsToLaunchFormJsonRequestSchema, {
        literals: params.literals,
        variables: params.variables,
      })

      return await client.literalsToLaunchFormJson(request)
    },
    enabled: !!(params?.literals && params?.variables),
    experimental_prefetchInRender: true,
  })

  return query
}
