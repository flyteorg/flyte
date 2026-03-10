import { Dispatch, useCallback, useEffect, useState } from 'react'
import Drawer from '@/components/Drawer'
import { ChevronLeftIcon } from '@heroicons/react/20/solid'
import { SelectTask } from './SelectTask'
import { TriggerMode, TaskDetails } from './types'
import { Button } from '@/components/Button'
import { Badge } from '@/components/Badge'
import { CreateTriggerTabs } from './CreateTriggerTabs'
import { CreateTriggerFormProvider } from './Form/FormProvider'

const CreateTriggerTitle = ({
  onBackClick,
  taskDetails,
}: {
  onBackClick: () => void
  taskDetails: TaskDetails | undefined
}) => {
  if (!taskDetails) return 'Create a new trigger'
  return (
    <div className="flex items-center gap-1">
      <Button onClick={onBackClick} size="xs" plain>
        <ChevronLeftIcon width={12} />
      </Button>
      Create a new trigger <Badge>{taskDetails.taskId}</Badge>
    </div>
  )
}

export const CreateTriggerDrawer = ({
  isOpen,
  setIsOpen,
}: {
  isOpen: boolean
  setIsOpen: Dispatch<React.SetStateAction<boolean>>
}) => {
  const [mode, updateMode] = useState<TriggerMode>('select-task')
  const [taskDetails, setTaskDetails] = useState<TaskDetails | undefined>()

  const switchMode = useCallback(
    (newMode: TriggerMode, options: TaskDetails | undefined) => {
      updateMode(newMode)
      setTaskDetails(options)
    },
    [],
  )

  // Cleanup state when drawer closes
  useEffect(() => {
    if (!isOpen) {
      updateMode('select-task')
      setTaskDetails(undefined)
    }
  }, [isOpen])

  return (
    <Drawer
      isOpen={isOpen}
      setIsOpen={setIsOpen}
      size={802}
      tabs={
        mode === 'select-task' ? (
          <SelectTask
            onSelectTask={(taskDetails: TaskDetails) => {
              switchMode('define-trigger', taskDetails)
            }}
          />
        ) : (
          <CreateTriggerFormProvider>
            <CreateTriggerTabs
              taskDetails={taskDetails}
              onCancel={() => {
                switchMode('select-task', undefined)
                setIsOpen(false)
              }}
            />
          </CreateTriggerFormProvider>
        )
      }
      title={
        <CreateTriggerTitle
          onBackClick={() => {
            switchMode('select-task', undefined)
          }}
          taskDetails={taskDetails}
        />
      }
    />
  )
}
