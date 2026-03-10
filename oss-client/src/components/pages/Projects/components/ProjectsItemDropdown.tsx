import { PopoverMenu, type MenuItem } from '@/components/Popovers'

import { ProjectState } from '@/gen/flyteidl2/project/project_service_pb'
import React, { useCallback, useMemo } from 'react'
import { ArchiveRestoreProjectItem } from './types'

export const ProjectsItemDropdown: React.FC<{
  project: ArchiveRestoreProjectItem
  setActiveProject: (project: ArchiveRestoreProjectItem | undefined) => void
}> = ({ setActiveProject, project }) => {
  const isArchived = useMemo(
    () => project.state === ProjectState.ARCHIVED,
    [project],
  )

  const onClick = useCallback(
    (e: React.MouseEvent) => {
      e.preventDefault()
      setActiveProject(project)
    },
    [setActiveProject, project],
  )

  const label = useMemo(
    () => `${isArchived ? 'Restore' : 'Archive'} project`,
    [isArchived],
  )

  const items: MenuItem[] = [
    {
      id: 'projects-items',
      type: 'item',
      onClick,
      label,
    },
  ]

  return <PopoverMenu items={items} variant="overflow" />
}
