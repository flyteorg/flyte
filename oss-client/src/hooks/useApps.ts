import { useEffect, useMemo } from 'react'
import cloneDeep from 'lodash/cloneDeep'
import {
  App,
  Identifier,
  IdentifierSchema,
  Spec_DesiredState,
} from '@/gen/flyteidl2/app/app_definition_pb'
import { useQueryClient, useMutation, useQuery } from '@tanstack/react-query'

import {
  GetRequestSchema,
  ListRequestSchema,
  UpdateRequestSchema,
} from '@/gen/flyteidl2/app/app_payload_pb'
import { FilterSchema } from '@/gen/flyteidl2/common/list_pb'
import { ProjectIdentifierSchema } from '@/gen/flyteidl2/common/identifier_pb'
import { Filter_Function } from '@/gen/flyteidl2/common/list_pb'
import { AppService } from '@/gen/flyteidl2/app/app_service_pb'
import { create } from '@bufbuild/protobuf'
import { useConnectRpcClient } from './useConnectRpc'

type ListAppsProps = {
  enabled?: boolean
  org: string
  domain: string | undefined
  projectId: string | undefined
  search?: string
  limit?: number
}

const getAppsQueryKey = ({ org, projectId, domain, search }: ListAppsProps) => {
  const key = ['apps', org, projectId, domain]
  return key.concat(search ? [search] : [])
}

// todo: this is intended as a placeholder for development. eventually we should swap this out for
// watchApps, with pagination and infinite scroll
export const useListApps = ({
  enabled = true,
  domain,
  org,
  projectId,
  search,
  limit = 100, // todo - replace with smaller limit when implementing watch api
}: ListAppsProps) => {
  const client = useConnectRpcClient(AppService)
  const queryKey = useMemo(
    () => getAppsQueryKey({ org, projectId, domain, search, limit }),
    [org, projectId, domain, search, limit],
  )

  const listRequest = create(ListRequestSchema, {
    request: {
      filters: search
        ? [
            create(FilterSchema, {
              function: Filter_Function.CONTAINS_CASE_INSENSITIVE,
              field: 'name',
              values: [search],
            }),
          ]
        : [],
      limit,
    },
    filterBy: {
      case: 'project',
      value: create(ProjectIdentifierSchema, {
        organization: org,
        domain,
        name: projectId,
      }),
    },
  })

  const fetchApps = async () => {
    return client.list(listRequest)
  }

  const isEnabled = enabled && !!org && !!domain && !!projectId

  return useQuery({
    enabled: isEnabled,
    queryKey,
    queryFn: fetchApps,
    refetchInterval: 10000,
  })
}

export const getAppIdentifier = ({
  domain,
  name,
  org,
  project,
}: Pick<Identifier, 'domain' | 'name' | 'org' | 'project'>) =>
  create(IdentifierSchema, {
    domain,
    name,
    org,
    project,
  })

export const useAppDetails = ({
  domain,
  name,
  org,
  projectId,
}: {
  domain: string
  org: string
  name: string
  projectId: string
}) => {
  const client = useConnectRpcClient(AppService)

  const queryKey = useMemo(
    () => getAppsQueryKey({ org, projectId, domain }),
    [org, projectId, domain],
  )

  const appIdentifier = getAppIdentifier({
    domain,
    name,
    org,
    project: projectId,
  })

  const getRequest = create(GetRequestSchema, {
    identifier: {
      case: 'appId',
      value: appIdentifier,
    },
  })
  return useQuery({
    queryFn: async () => {
      return client.get(getRequest)
    },
    queryKey: [...queryKey, name],
    refetchInterval: 5000,
  })
}

type UpdateAppStatus = {
  app: App
  desiredState: Spec_DesiredState
}

export const useUpdateAppStatus = ({ app, desiredState }: UpdateAppStatus) => {
  const client = useConnectRpcClient(AppService)
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async () => {
      if (!app.metadata?.id) return null
      const getRequest = create(GetRequestSchema, {
        identifier: {
          case: 'appId',
          value: app.metadata?.id,
        },
      })
      const appDetails = await client.get(getRequest)
      const appClone = cloneDeep(appDetails.app)
      if (!appClone?.spec)
        throw new Error('Could not update app without app spec')
      const withUpdatedStatus: App = {
        ...appClone,
        spec: {
          ...appClone.spec,
          desiredState,
        },
      }
      const request = create(UpdateRequestSchema, {
        app: withUpdatedStatus,
      })
      return client.update(request)
    },
    onSuccess: () => {
      const domain = app.metadata?.id?.domain || ''
      const org = app.metadata?.id?.org || ''
      const projectId = app.metadata?.id?.project || ''
      queryClient.invalidateQueries({
        queryKey: getAppsQueryKey({ domain, projectId, org }),
      })
    },
  })
}

export const useStartApp = (props: { app: App }) =>
  useUpdateAppStatus({ ...props, desiredState: Spec_DesiredState.ACTIVE })

export const useStopApp = (props: { app: App }) =>
  useUpdateAppStatus({
    ...props,
    desiredState: Spec_DesiredState.STOPPED,
  })
