import { TaskSpec } from '@/gen/flyteidl2/task/task_definition_pb'
import {
  TaskSpecToLaunchFormJsonRequestSchema,
  TaskSpecToLaunchFormJsonResponse,
  TranslatorService,
} from '@/gen/flyteidl2/workflow/translator_service_pb'
import { create } from '@bufbuild/protobuf'
import { useQuery } from '@tanstack/react-query'
import { useConnectRpcClient } from './useConnectRpc'

interface useTaskSpecLaunchFormParams {
  enabled: boolean
  taskSpec: TaskSpec | undefined | null
}

export function useTaskSpecLaunchForm({
  enabled,
  taskSpec,
}: useTaskSpecLaunchFormParams) {
  const client = useConnectRpcClient(TranslatorService)

  const query = useQuery({
    enabled,
    queryKey: ['taskSpecToLaunchFormJson', taskSpec?.taskTemplate?.id?.name],
    queryFn: async (): Promise<TaskSpecToLaunchFormJsonResponse> => {
      if (!taskSpec) {
        throw new Error(
          'No taskSpec provided for taskSpecToLaunchFormJson call',
        )
      }

      const request = create(TaskSpecToLaunchFormJsonRequestSchema, {
        taskSpec,
      })

      return await client.taskSpecToLaunchFormJson(request)
    },
    experimental_prefetchInRender: true,
  })

  return query
}
