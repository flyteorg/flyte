import { CreateProjectButton } from '@/components/pages/Projects/components/CreateProjectButton'
import { CreateProjectDialog } from '@/components/pages/Projects/components/CreateProjectDialog'
import { NewProject } from '@/components/pages/Projects/components/types.ts'
import { getProjectId } from '@/components/pages/Projects/components/util.ts'
import {
  ProjectService,
  ProjectState,
} from '@/gen/flyteidl2/project/project_service_pb.ts'
import { useConnectRpcClient } from '@/hooks/useConnectRpc.ts'
import { useCreateProject } from '@/hooks/useProjects.ts'
import { useNotifications } from '@/providers/notifications'
import { Code, ConnectError } from '@connectrpc/connect'
import { useCallback, useState } from 'react'
import { useDebounce } from 'react-use'

export function CreateProjectWorkflow() {
  const { showNotification } = useNotifications()

  const [isDialogOpen, setIsDialogOpen] = useState(false)
  const [expectedProjectId, setExpectedProjectId] = useState<string>('')
  const [isValid, setIsValid] = useState<boolean>(true)
  const [waitingForValidation, setWaitingForValidation] =
    useState<boolean>(false)

  const client = useConnectRpcClient(ProjectService)
  const createProject = useCreateProject()

  const validateProjectId = useCallback(
    async (id: string) => {
      if (!id) {
        setWaitingForValidation(false)
        return
      }
      setWaitingForValidation(true)

      try {
        const result = await client.getProject({ id })
        setIsValid(!result?.project?.id)
        setWaitingForValidation(false)
      } catch (error) {
        setWaitingForValidation(false)
        if (error instanceof ConnectError && error.code === Code.NotFound) {
          setIsValid(true)
        } else {
          setIsValid(false)
          console.error(error)
          throw error
        }
      }
    },
    [client],
  )

  const onCreateProject = useCallback(
    (newProject: NewProject) => {
      createProject.mutate(
        {
          ...newProject,
          state: ProjectState.ACTIVE,
          domains: [],
          org: '', // Provide a default or get from context if needed
          $typeName: 'flyteidl2.project.Project',
        },
        {
          onSuccess: () => {
            showNotification(`Created ${newProject.name}`, 'success')
          },
          onError: () => {
            showNotification(`Error creating ${newProject.name}`, 'error')
          },
          onSettled: async () => {
            setIsDialogOpen(false)
          },
        },
      )
    },
    [createProject, showNotification],
  )

  const onProjectIdChange = useCallback(
    (id: string) => {
      setWaitingForValidation(true)
      setExpectedProjectId(id)
    },
    [setWaitingForValidation, setExpectedProjectId],
  )

  useDebounce(() => validateProjectId(expectedProjectId), 500, [
    expectedProjectId,
  ])

  const openDialog = useCallback(() => setIsDialogOpen(true), [setIsDialogOpen])
  const closeDialog = useCallback(
    () => setIsDialogOpen(false),
    [setIsDialogOpen],
  )

  return (
    <>
      <CreateProjectButton onClick={openDialog} />
      <CreateProjectDialog
        isOpen={isDialogOpen}
        onClose={closeDialog}
        getProjectId={getProjectId}
        onCreateProject={onCreateProject}
        isProjectIdValid={isValid}
        isWaitingForValidation={waitingForValidation}
        onProjectIdChange={onProjectIdChange}
        descriptionSizeLimit={300}
      />
    </>
  )
}
