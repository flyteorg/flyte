import { Button } from '@/components/Button'
import {
  Dialog,
  DialogActions,
  DialogBody,
  DialogDescription,
} from '@/components/Dialog'
import { Input } from '@/components/Input'
import { ProjectArchiveRestoreDialogProps } from '@/components/pages/Projects/components/types.ts'
import { ProjectState } from '@/gen/flyteidl2/project/project_service_pb.ts'
import { useEffect, useMemo, useState } from 'react'

export const ProjectArchiveRestoreDialog: React.FC<
  ProjectArchiveRestoreDialogProps
> = ({ project, onClose, onArchive, onRestore }) => {
  const [projectName, setProjectName] = useState<string>('')
  const isArchived = useMemo(
    () => project?.state === ProjectState.ARCHIVED,
    [project?.state],
  )

  const headerText = useMemo(
    () => `${isArchived ? 'Restore' : 'Archive'} ${project?.name}`,
    [project?.name, isArchived],
  )

  const isActionEnabled = useMemo(
    () => (isArchived ? true : !!project?.name && projectName === project.name),
    [isArchived, projectName, project?.name],
  )

  useEffect(() => {
    setProjectName('')
  }, [project?.name])

  return project?.name ? (
    <Dialog onClose={onClose} open={!!project}>
      <div className="text-sm/5 font-semibold">{headerText}</div>

      <DialogDescription>
        {isArchived ? (
          <div className="text-sm/5 font-medium text-zinc-500">
            Project <strong>{project?.name}</strong> will become read-write for
            all domains. All previously active triggers will reactivate.
          </div>
        ) : (
          <div className="text-sm/5 font-medium text-zinc-500">
            Project <strong>{project.name}</strong> will become read-only for
            all domains. You will still be able to view and unarchive it at any
            time. All active triggers will stop running.
          </div>
        )}
      </DialogDescription>
      {!isArchived && (
        <DialogBody className="text-sm/5 font-semibold text-zinc-500">
          Please type{' '}
          <strong className="leading-5 font-semibold text-[#131313] dark:text-white">
            {project.name}
          </strong>{' '}
          to confirm
          <Input
            className="mt-2"
            onChange={(e) => setProjectName(e.target.value)}
            value={projectName}
          />
        </DialogBody>
      )}
      <DialogActions>
        <Button plain onClick={onClose} size="sm">
          Cancel
        </Button>
        <Button
          onClick={isArchived ? onRestore : onArchive}
          disabled={!isActionEnabled}
          color="union"
          size="sm"
        >
          {isArchived ? 'Restore' : 'Archive'}
        </Button>
      </DialogActions>
    </Dialog>
  ) : null
}
