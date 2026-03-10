import {
  SecretIdentifier,
  SecretIdentifierSchema,
  SecretType,
} from '@/gen/flyteidl2/secret/definition_pb'
import {
  ListSecretsRequestSchema,
  ListSecretsResponse,
} from '@/gen/flyteidl2/secret/payload_pb'
import { SecretService } from '@/gen/flyteidl2/secret/secret_pb'
import { create } from '@bufbuild/protobuf'
import {
  useInfiniteQuery,
  useMutation,
  useQuery,
  useQueryClient,
} from '@tanstack/react-query'
import { useCallback, useMemo } from 'react'
import { useConnectRpcClient } from './useConnectRpc'
import { useOrg } from './useOrg'

const getSecretDetailsQueryKey = ({
  org,
  domain,
  project,
  name,
}: SecretDetailsProps) => {
  const domainValue = domain?.length === 0 || !domain ? null : domain
  const projectValue = project?.length === 0 || !project ? null : project
  return ['secrets', org, domainValue, projectValue, name]
}

export const useListSecrets = ({
  limit,
  searchTerm,
}: {
  limit?: number
  searchTerm: string
}) => {
  const client = useConnectRpcClient(SecretService)
  const org = useOrg()
  const getSecretsPage = useCallback(
    async ({ pageParam = '' }: { pageParam?: string }) => {
      const request = create(ListSecretsRequestSchema, {
        organization: org,
        limit,
        token: pageParam,
      })
      const secrets = await client.listSecrets(request)
      return secrets
    },
    [client, limit, org],
  )

  return useInfiniteQuery({
    queryKey: ['secrets', org, searchTerm],
    queryFn: getSecretsPage,
    initialPageParam: '',
    getNextPageParam: (lastPage: ListSecretsResponse) => {
      // Only return the next token if the current page has data and a valid token
      if (
        lastPage.secrets &&
        lastPage.secrets.length > 0 &&
        lastPage.token &&
        lastPage.token.trim() !== ''
      ) {
        return lastPage.token
      }
      return undefined
    },
    // ensure we don't refetch too often
    staleTime: 5 * 60 * 1000, // 5 minutes
    gcTime: 10 * 60 * 1000, // 10 minutes (formerly cacheTime)
  })
}

export function useDeleteSecret() {
  const client = useConnectRpcClient(SecretService)
  const queryClient = useQueryClient()
  const org = useOrg()
  return useMutation({
    mutationFn: async (secretId: SecretIdentifier) => {
      return await client.deleteSecret({
        id: secretId,
      })
    },
    onSuccess: async () => {
      await new Promise<void>((resolve) => setTimeout(() => resolve(), 2000))
      queryClient.invalidateQueries({ queryKey: ['secrets', org] })
    },
  })
}

export function useCreateSecret() {
  const client = useConnectRpcClient(SecretService)
  const queryClient = useQueryClient()
  const org = useOrg()

  return useMutation({
    mutationFn: async ({
      name,
      domain,
      project,
      value,
    }: {
      name: string
      domain?: string
      project?: string
      value: string
    }) => {
      return client.createSecret({
        id: {
          name,
          organization: org,
          domain,
          project,
        },
        secretSpec: {
          value: { value, case: 'stringValue' },
          type: SecretType.GENERIC,
        },
      })
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['secrets', org] })
    },
  })
}

export function useUpdateSecret({ secretId }: { secretId?: SecretIdentifier }) {
  const client = useConnectRpcClient(SecretService)
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({ value }: { value: string }) => {
      return await client.updateSecret({
        id: secretId,
        secretSpec: {
          value: { value, case: 'stringValue' },
          type: SecretType.GENERIC,
        },
      })
    },
    onSuccess: async () => {
      if (secretId) {
        const { domain, project, name } = secretId
        await queryClient.invalidateQueries({
          queryKey: getSecretDetailsQueryKey({ domain, project, name }),
        })
      }
    },
  })
}

type SecretDetailsProps = {
  org?: string
  domain: string | undefined
  project: string | undefined
  name: string
}

const getSecretIdentifier = ({
  org,
  domain,
  name,
  project,
}: SecretDetailsProps) =>
  create(SecretIdentifierSchema, {
    name,
    domain,
    organization: org,
    project,
  })

export function useSecretDetails({
  domain,
  project,
  name,
}: SecretDetailsProps) {
  const client = useConnectRpcClient(SecretService)
  const secretIdentifier = useMemo(
    () => getSecretIdentifier({ domain, project, name }),
    [domain, project, name],
  )

  return useQuery({
    queryKey: getSecretDetailsQueryKey({ domain, project, name }),
    queryFn: async () => {
      if (!secretIdentifier) return null
      const response = await client.getSecret({ id: secretIdentifier })
      return response
    },
    enabled: !!name,
  })
}
