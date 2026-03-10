import { useLatestProjectDomainPairs } from '@/hooks/useLatestProjects'
import { useUpdateProject } from '@/hooks/useProjects.ts'
import { useNotifications } from '@/providers/notifications'
import { useCallback, useState } from 'react'
import {
  ArchiveRestoreProjectItem,
  ProjectArchiveRestoreDialogProps,
} from './types.ts'
import { ProjectState } from '@/gen/flyteidl2/project/project_service_pb.ts'

export function useProjectArchiveRestoreDialog(): ProjectArchiveRestoreDialogProps & {
  setActiveProject: (project: ArchiveRestoreProjectItem | undefined) => void
} {
  const { showNotification } = useNotifications()
  const { removeProjectById } = useLatestProjectDomainPairs()

  const [activeProject, setActiveProject] = useState<
    ArchiveRestoreProjectItem | undefined
  >(undefined)
  const updateProject = useUpdateProject()

  const onClose = useCallback(
    () => setActiveProject(undefined),
    [setActiveProject],
  )

  const onArchive = useCallback(() => {
    if (activeProject) {
      updateProject.mutate(
        {
          ...activeProject,
          domains: [],
          state: ProjectState.ARCHIVED,
        },
        {
          onSuccess: () => {
            removeProjectById(activeProject.id)
            showNotification(`Archived ${activeProject.name}`, 'success')
          },
          onError: () => {
            showNotification(`Error archiving ${activeProject.name}`, 'error')
          },
          onSettled: () => setActiveProject(undefined),
        },
      )
    }
  }, [activeProject, removeProjectById, showNotification, updateProject])

  const onRestore = useCallback(() => {
    if (activeProject) {
      updateProject.mutate(
        {
          ...activeProject,
          domains: [],
          state: ProjectState.ACTIVE,
        },
        {
          onSuccess: () => {
            showNotification(`Restored ${activeProject.name}`, 'success')
          },
          onError: () => {
            showNotification(`Error restoring ${activeProject.name}`, 'error')
          },
          onSettled: () => setActiveProject(undefined),
        },
      )
    }
  }, [activeProject, showNotification, updateProject])

  return {
    project: activeProject,
    setActiveProject,
    onArchive,
    onRestore,
    onClose,
  }
}
