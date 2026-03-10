import {
  Project,
  ProjectService,
} from '@/gen/flyteidl2/project/project_service_pb'
import { Message } from '@bufbuild/protobuf'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useConnectRpcClient } from './useConnectRpc'

export function useProjects() {
  const client = useConnectRpcClient(ProjectService)

  const getProjects = async () => {
    try {
      const response = await client.listProjects({
        limit: 200,
        filters: 'value_in(state,0;1;2)',
      })
      return response.projects
    } catch (error) {
      console.error(error)
      throw error
    }
  }

  return useQuery({
    queryKey: ['projects'],
    queryFn: getProjects,
  })
}

export function useProjectData({ id }: { id: string }) {
  const client = useConnectRpcClient(ProjectService)
  const getProjectData = async () => {
    try {
      const response = await client.getProject({ id })
      return response.project
    } catch (error) {
      console.error(error)
      throw error
    }
  }
  return useQuery({
    enabled: !!id,
    queryKey: ['projectData', id],
    queryFn: getProjectData,
  })
}

export function useUpdateProject() {
  const client = useConnectRpcClient(ProjectService)
  const queryClient = useQueryClient()

  const updateProjectCb = async (project: Project) => {
    try {
      const response = await client.updateProject({ project })
      return response
    } catch (error) {
      console.error(error)
      throw error
    }
  }

  return useMutation({
    mutationFn: updateProjectCb,
    onSettled: async (result, error, project) => {
      await queryClient.invalidateQueries({
        queryKey: ['projectData', project.id],
      })
      await queryClient.invalidateQueries({ queryKey: ['projects'] })
    },
  })
}

export function useCreateProject() {
  const client = useConnectRpcClient(ProjectService)
  const queryClient = useQueryClient()

  const createProjectCb = async (
    project: Pick<Project, Exclude<keyof Project, Message>>,
  ) => {
    try {
      const response = await client.createProject({ project })

      return response
    } catch (error) {
      console.error(error)
      throw error
    }
  }

  return useMutation({
    mutationFn: createProjectCb,
    onSettled: async () => {
      await queryClient.invalidateQueries({ queryKey: ['projects'] })
    },
  })
}
