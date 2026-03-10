import {
  TaskIdentifier,
  TaskIdentifierSchema,
} from '@/gen/flyteidl2/task/task_definition_pb'
import { TaskService } from '@/gen/flyteidl2/task/task_service_pb'
import { create } from '@bufbuild/protobuf'
import { useQuery } from '@tanstack/react-query'
import { useConnectRpcClient } from './useConnectRpc'

export const getTaskIdentifier = ({
  domain,
  name,
  org,
  project,
  version,
}: Pick<TaskIdentifier, 'domain' | 'name' | 'org' | 'project' | 'version'>) =>
  create(TaskIdentifierSchema, {
    name,
    domain,
    org,
    project,
    version,
  })

export const getTaskDetailsQueryKey = (taskIdentifier: TaskIdentifier) =>
  [
    'taskDetails',
    taskIdentifier.org,
    taskIdentifier.project,
    taskIdentifier.domain,
    taskIdentifier.name,
    taskIdentifier.version,
  ].filter(Boolean)

export const useTaskDetails = ({
  name,
  version,
  project,
  domain,
  org,
  enabled,
}: {
  name: string
  version: string
  project: string
  domain: string
  org: string
  enabled?: boolean
}) => {
  const client = useConnectRpcClient(TaskService)
  const taskIdentifier = getTaskIdentifier({
    domain,
    name,
    org,
    project,
    version,
  })

  return useQuery({
    queryKey: getTaskDetailsQueryKey(taskIdentifier),
    queryFn: async () => {
      if (!taskIdentifier) return null
      const response = await client.getTaskDetails({ taskId: taskIdentifier })
      return response
    },
    enabled: enabled && !!org && !!taskIdentifier,
    refetchOnWindowFocus: false,
    // Cache the result to avoid re-requesting on every render
    staleTime: 10 * 60 * 1000, // 10 minutes - data is considered fresh for 10 minutes
    gcTime: 30 * 60 * 1000, // 30 minutes - keep in cache for 30 minutes (formerly cacheTime)
  })
}
