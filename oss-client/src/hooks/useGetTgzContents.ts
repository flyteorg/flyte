import { CloudDataProxyService } from '@/gen/clouddataproxy/clouddataproxy_pb'
import { ArtifactType } from '@/gen/clouddataproxy/payload_pb'
import { ActionAttemptIdentifier } from '@/gen/flyteidl2/common/identifier_pb'
import { TaskIdentifier } from '@/gen/flyteidl2/task/task_definition_pb'
import { Identifier as AppIdentifier } from '@/gen/flyteidl2/app/app_definition_pb'

import {
  extractTgzFromUrl,
  FileSizeExceededError,
  AccessDeniedError,
} from '@/lib/tgzUtils'
import {
  is403OrCorsError,
  is404Error,
  isFailedPreconditionError,
} from '@/lib/errorUtils'
import { useQuery } from '@tanstack/react-query'
import { useConnectRpcClient } from './useConnectRpc'

type UseGetTgzContentsProps = {
  actionAttemptId?: ActionAttemptIdentifier
  taskId?: TaskIdentifier
  appId?: AppIdentifier
  enabled?: boolean
  includeFileContents?: boolean
}

export function useGetTgzContents({
  actionAttemptId,
  taskId,
  appId,
  enabled = true,
  includeFileContents = true,
}: UseGetTgzContentsProps) {
  const client = useConnectRpcClient(CloudDataProxyService)

  // Determine which identifier is provided
  const identifier = actionAttemptId || taskId || appId
  let identifierType: 'actionAttemptId' | 'taskId' | 'appId' | undefined
  if (actionAttemptId) {
    identifierType = 'actionAttemptId'
  } else if (taskId) {
    identifierType = 'taskId'
  } else if (appId) {
    identifierType = 'appId'
  }

  const query = useQuery({
    queryKey: ['getTgzContents', identifier, includeFileContents],
    queryFn: async () => {
      if (!identifier || !identifierType) {
        throw new Error('One of actionAttemptId, taskId, or appId is required')
      }

      // Construct the source based on the identifier type
      let source:
        | { case: 'actionAttemptId'; value: ActionAttemptIdentifier }
        | { case: 'taskId'; value: TaskIdentifier }
        | { case: 'appId'; value: AppIdentifier }

      if (identifierType === 'actionAttemptId' && actionAttemptId) {
        source = {
          case: 'actionAttemptId',
          value: actionAttemptId,
        }
      } else if (identifierType === 'taskId' && taskId) {
        // Validate that taskId has all required fields
        if (!taskId.project) {
          throw new Error('TaskIdentifier is missing required field: project')
        }
        if (!taskId.name || !taskId.domain || !taskId.org || !taskId.version) {
          throw new Error(
            'TaskIdentifier is missing required fields. Required: name, domain, org, project, version',
          )
        }
        source = {
          case: 'taskId',
          value: taskId,
        }
      } else if (identifierType === 'appId' && appId) {
        // Validate that appId has all required fields
        if (!appId.project) {
          throw new Error('AppIdentifier is missing required field: project')
        }
        if (!appId.name || !appId.domain || !appId.org) {
          throw new Error(
            'AppIdentifier is missing required fields. Required: name, domain, org, project',
          )
        }
        source = {
          case: 'appId',
          value: appId,
        }
      } else {
        throw new Error('Invalid identifier type')
      }

      // Use createDownloadLinkV2 to get presigned URL for code bundle
      const response = await client.createDownloadLinkV2({
        artifactType: ArtifactType.CODE_BUNDLE,
        source,
      })

      // Get the presigned URL from the response
      const presignedUrls = response.preSignedUrls?.signedUrl ?? []
      const presignedUrl = presignedUrls[0]

      if (!presignedUrl) {
        throw new Error('Presigned URL not found in response')
      }

      // Download and extract the tgz file client-side
      try {
        const root = await extractTgzFromUrl(
          presignedUrl,
          includeFileContents,
          false, // Size check disabled by default (optional)
        )

        return {
          root,
          presignedUrl,
        }
      } catch (error) {
        // If it's a 403/CORS error, wrap it in AccessDeniedError to preserve presignedUrl
        if (is403OrCorsError(error)) {
          throw new AccessDeniedError(
            error instanceof Error ? error.message : 'Access denied',
            presignedUrl,
            error,
          )
        }
        throw error
      }
    },
    retry: (failureCount, error) => {
      // Don't retry errors that won't resolve on retry
      if (
        error instanceof FileSizeExceededError ||
        error instanceof AccessDeniedError ||
        is403OrCorsError(error) ||
        is404Error(error) ||
        isFailedPreconditionError(error)
      ) {
        return false
      }

      // For other errors, allow retry up to 3 times
      return failureCount < 3
    },
    enabled: Boolean(identifier) && enabled,
    refetchOnWindowFocus: false,
    // Cache the result to avoid re-requesting on every render
    staleTime: 10 * 60 * 1000, // 10 minutes - data is considered fresh for 10 minutes
    gcTime: 30 * 60 * 1000, // 30 minutes - keep in cache for 30 minutes (formerly cacheTime)
  })

  return query
}
