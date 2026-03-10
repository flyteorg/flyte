import { useConnectRpcClient } from './useConnectRpc'
import { AuthorizerService } from '@/gen/authorizer/authorizer_pb'
import { AuthorizeRequestSchema } from '@/gen/authorizer/payload_pb'
import { IdentitySchema, type Identity } from '@/gen/common/identity_pb'
import { UserIdentifierSchema } from '@/gen/common/identifier_pb'
import {
  type Resource,
  OrganizationSchema,
  ResourceSchema,
  Action,
} from '@/gen/common/authorization_pb'
import { create } from '@bufbuild/protobuf'
import { useOrg } from './useOrg'
import { useIdentity } from './useIdentity'
import { useQuery } from '@tanstack/react-query'

interface AuthorizeParams {
  action: Action
  resource?: Resource
}

/** Builds a Resource scoped to the current org (used when no resource is provided). */
function useOrgResource(): Resource {
  const org = useOrg()
  return create(ResourceSchema, {
    resource: {
      case: 'organization',
      value: create(OrganizationSchema, { name: org }),
    },
  })
}

/** Builds an Identity from the current user's subject claim. */
function useCurrentIdentity(): Identity | undefined {
  const { data } = useIdentity()
  if (!data?.subject) return undefined
  return create(IdentitySchema, {
    principal: {
      case: 'userId',
      value: create(UserIdentifierSchema, { subject: data.subject }),
    },
  })
}

// ---------------------------------------------------------------------------
// useAuthorizationQuery
// ---------------------------------------------------------------------------

/**
 * Low-level hook that returns the full react-query result for an authorization
 * check. Use this when you need loading/error states (e.g. skeleton UI while
 * the check is in flight).
 *
 * The query is automatically disabled until the user's identity is available.
 *
 * @example
 * ```tsx
 * const { data, isLoading, isError } = useAuthorizationQuery({
 *   action: Action.CREATE,
 * })
 * if (isLoading) return <Skeleton />
 * if (data?.allowed) return <CreateButton />
 * ```
 */
export function useAuthorizationQuery({ action, resource }: AuthorizeParams) {
  const client = useConnectRpcClient(AuthorizerService)
  const org = useOrg()
  const orgResource = useOrgResource()
  const identity = useCurrentIdentity()

  const resolvedResource = resource ?? orgResource

  return useQuery({
    queryKey: [
      'authorize',
      org,
      action,
      resolvedResource.resource,
      identity?.principal,
    ],
    queryFn: () =>
      client.authorize(
        create(AuthorizeRequestSchema, {
          action,
          organization: org,
          identity,
          resource: resolvedResource,
        }),
      ),
    // Don't fire until we know who the user is.
    enabled: identity !== undefined && !!org,
    // Cache authorization results for better performance.
    staleTime: 5 * 60 * 1000, // 5 minutes
    gcTime: 10 * 60 * 1000, // 10 minutes
  })
}

// ---------------------------------------------------------------------------
// useIsAuthorized
// ---------------------------------------------------------------------------

/**
 * Convenience hook that resolves to a boolean. Returns `false` while the
 * authorization check is in flight or if the user is not authorized.
 *
 * Use this for simple show/hide decisions where you don't need loading states.
 *
 * @example
 * ```tsx
 * const canCreate = useIsAuthorized({ action: Action.CREATE })
 * return canCreate ? <CreateButton /> : null
 * ```
 */
export function useIsAuthorized({
  action,
  resource,
}: AuthorizeParams): boolean {
  return useAuthorizationQuery({ action, resource }).data?.allowed === true
}
