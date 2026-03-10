import { useState } from 'react'
import { Button } from '@/components/Button'
import { useFormContext } from 'react-hook-form'
import { CreateTriggerState } from './types'
import { useCreateTrigger } from '@/hooks/useCreateTrigger'
import { useParams } from 'next/navigation'
import { ProjectDomainPageParams } from '@/types/pageParams'
import { useOrg } from '@/hooks/useOrg'

export const CreateTriggerFooter = ({
  onCancel,
  taskName,
  taskVersion,
}: {
  onCancel: () => void
  taskName: string | undefined
  taskVersion: string | undefined
}) => {
  const [errorMessage, setErrorMessage] = useState('')
  const { domain, project } = useParams<ProjectDomainPageParams>()
  const org = useOrg()

  const { handleSubmit, setError } = useFormContext<CreateTriggerState>()
  const { mutateAsync } = useCreateTrigger({ domain, org, project })

  const handleClick = async (data: CreateTriggerState) => {
    if (!data.inputs) {
      setErrorMessage('Check inputs before creating trigger')
      setError('inputs', {
        type: 'validation',
        message: 'Check inputs before creating trigger',
      })
      return
    }

    try {
      await mutateAsync({
        createTriggerState: data,
        taskName,
        taskVersion,
      })
      onCancel()
    } catch (e) {
      if (e instanceof Error) {
        setErrorMessage(`There was an error creating trigger: ${e.message}`)
      }
    }
  }

  return (
    <div>
      {errorMessage && (
        <div className="p-4 text-(--accent-red)">{errorMessage}</div>
      )}
      <div className="flex justify-end gap-2">
        <Button outline onClick={onCancel}>
          Cancel
        </Button>
        <Button color="union" onClick={handleSubmit(handleClick)}>
          Create trigger
        </Button>
      </div>
    </div>
  )
}
