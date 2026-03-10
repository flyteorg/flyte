import {
  DialogActions,
  DialogBody,
  DialogDescription,
  DialogTitle,
  Dialog,
} from '@/components/Dialog'
import { Button } from '@/components/Button'
import { Input } from '@/components/Input'
import { NewProject } from '@/components/pages/Projects/components/types.ts'
import { useCallback, useEffect, useMemo, useState } from 'react'
import { ErrorMessage, Field } from '@/components/Fieldset'
import { Textarea } from '@/components/Textarea'

export const CreateProjectDialog: React.FC<{
  isOpen: boolean
  onClose: () => void
  isProjectIdValid: boolean
  isWaitingForValidation: boolean
  onProjectIdChange: (id: string) => void
  getProjectId: (projectName: string) => string
  onCreateProject: (newProject: NewProject) => void
  descriptionSizeLimit: number
}> = ({
  isOpen,
  isProjectIdValid,
  isWaitingForValidation,
  onProjectIdChange,
  onClose,
  onCreateProject,
  getProjectId,
  descriptionSizeLimit,
}) => {
  const [projectName, setProjectName] = useState<string>('')
  const [projectId, setProjectId] = useState<string>('')
  const [projectDescription, setProjectDescription] = useState<string>('')

  const [enableEditProjectId, setEnableEditProjectId] = useState<boolean>(false)

  useEffect(() => {
    if (!isOpen) {
      setProjectId('')
      setProjectName('')
      setProjectDescription('')
      setEnableEditProjectId(false)
    }
  }, [isOpen])

  useEffect(() => {
    setProjectId(getProjectId(projectName))
  }, [getProjectId, projectName])

  useEffect(() => {
    onProjectIdChange(projectId)
  }, [onProjectIdChange, projectId])

  useEffect(() => {
    if (!isProjectIdValid) {
      setEnableEditProjectId(true)
    }
  }, [isProjectIdValid])

  const isDescriptionValid = useMemo(
    () => projectDescription.length <= descriptionSizeLimit,
    [projectDescription, descriptionSizeLimit],
  )

  const isFormDisabled = useMemo(
    () =>
      !(
        isProjectIdValid &&
        isDescriptionValid &&
        !isWaitingForValidation &&
        projectId.length > 0 &&
        projectName.length > 0
      ),
    [
      isProjectIdValid,
      isDescriptionValid,
      isWaitingForValidation,
      projectName,
      projectId,
    ],
  )

  const handleCreateProject = useCallback(() => {
    if (isProjectIdValid) {
      onCreateProject({
        name: projectName,
        description: projectDescription,
        id: projectId,
      })
      onClose()
    }
  }, [
    projectName,
    projectDescription,
    isProjectIdValid,
    projectId,
    onCreateProject,
    onClose,
  ])

  return (
    <Dialog onClose={onClose} open={isOpen}>
      <DialogTitle className="text-14/20 font-semibold">
        Create new project
      </DialogTitle>
      <DialogDescription className="text-14/20 font-medium text-zinc-500">
        Projects are created across all domains
      </DialogDescription>
      <DialogBody className="flex flex-col gap-6">
        <Field>
          <div className="text m-0 text-sm/5 font-semibold">Name</div>
          <Input
            className="mt-2"
            onChange={(e) => setProjectName(e.target.value)}
            value={projectName}
          />
          {projectName &&
            (enableEditProjectId ? (
              <>
                <div className="text mt-6 text-sm/5 font-semibold">
                  Project ID
                </div>
                <Input
                  invalid={!isProjectIdValid}
                  className="mt-2"
                  onChange={(e) => setProjectId(e.target.value)}
                  value={projectId}
                />
                {!isProjectIdValid && (
                  <ErrorMessage className="!mt-2 !text-sm/5">
                    {projectId} already exists
                  </ErrorMessage>
                )}
              </>
            ) : (
              <div className="mt-2 text-sm/5">
                Your project ID will be created as {projectId}{' '}
                <span
                  className="cursor-pointer font-semibold"
                  onClick={() => setEnableEditProjectId(true)}
                >
                  Edit project ID
                </span>
              </div>
            ))}
        </Field>
        <Field>
          <div className="text m-0 text-sm/5 font-semibold">
            Description <span className="text-zinc-500">(Optional)</span>
          </div>
          <Textarea
            invalid={!isDescriptionValid}
            className="mt-2"
            onChange={(e) => setProjectDescription(e.target.value)}
          />
          <div
            className={`mt-2 text-sm/5 ${isDescriptionValid ? '' : 'text-red-400'}`}
          >
            {projectDescription.length}/{descriptionSizeLimit}
          </div>
        </Field>
      </DialogBody>
      <DialogActions>
        <Button outline onClick={onClose}>
          Cancel
        </Button>
        <Button
          onClick={handleCreateProject}
          disabled={isFormDisabled}
          color="union"
        >
          Create project
        </Button>
      </DialogActions>
    </Dialog>
  )
}
