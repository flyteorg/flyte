import { Project } from '@/gen/flyteidl2/project/project_service_pb'
export type ArchiveRestoreProjectItem = Project
export type NewProject = Pick<Project, 'id' | 'name' | 'description'>

export type ProjectArchiveRestoreDialogProps = {
  project?: ArchiveRestoreProjectItem
  onClose: VoidFunction
  onArchive: () => void
  onRestore: () => void
}
