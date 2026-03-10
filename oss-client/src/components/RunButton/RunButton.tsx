'use client'

import React, { useEffect, useState } from 'react'
import { PlayIcon } from '@heroicons/react/24/solid'
import { LaunchFormDrawer } from '@/components/LaunchForm'
import { Button } from '@/components/Button'
import ComboButton from '@/components/ComboButton'
import { RerunIcon } from '@/components/icons/RerunIcon'
import { AbortModal } from './components/AbortModal'
import { useRunLaunchFormData } from '../LaunchForm/hooks/useRunLaunchFormData'
import { useTaskLaunchFormData } from '../LaunchForm/hooks/useTaskLaunchFormData'
import { TaskSpec } from '@/gen/flyteidl2/task/task_definition_pb'
import { useLaunchFormState } from '@/hooks/useLaunchFormState'

export const RunButton = () => {
  const [abortOpen, setAbortOpen] = useState(false)
  const { setTaskSpec, setIsOpen } = useLaunchFormState()

  const { drawerMeta, isDataFetched, isTerminalPhase, formMethods, spec } =
    useRunLaunchFormData()

  useEffect(() => {
    if (spec && spec.$typeName === 'flyteidl2.task.TaskSpec') {
      setTaskSpec((prev) => (prev ? prev : spec))
    }
  }, [spec, setTaskSpec])

  return (
    <>
      <ComboButton
        color={isTerminalPhase ? 'union' : 'rose'}
        options={
          isTerminalPhase
            ? [
                // In terminal phase, show only rerun option
                {
                  name: (
                    <span className="flex items-center">
                      <RerunIcon className="mr-2 size-3" />
                      Rerun
                    </span>
                  ),
                  onClick: () => setIsOpen(true),
                },
              ]
            : [
                // Per design request, when not in terminal phase, show abort in the dropdown as well
                {
                  name: 'Abort run',
                  onClick: () => {
                    setAbortOpen(true)
                  },
                },
                {
                  name: 'Abort run',
                  onClick: () => {
                    setAbortOpen(true)
                  },
                },
                {
                  name: 'Rerun',
                  onClick: () => setIsOpen(true),
                },
              ]
        }
      />
      <AbortModal isOpen={abortOpen} setIsOpen={setAbortOpen} />
      <LaunchFormDrawer {...{ drawerMeta, formMethods, isDataFetched }} />
    </>
  )
}

type TaskRunButtonProps = {
  taskSpec: TaskSpec | undefined
  version: string
}

export const TaskRunButton = ({ taskSpec, version }: TaskRunButtonProps) => {
  const [abortOpen, setAbortOpen] = useState(false)
  const { buttonText, setIsOpen, setTaskSpec } = useLaunchFormState()

  const { drawerMeta, isDataFetched, formMethods } = useTaskLaunchFormData({
    version,
  })

  useEffect(() => {
    if (taskSpec) {
      setTaskSpec((prev) => (prev ? prev : taskSpec))
    }
  }, [taskSpec, setTaskSpec])

  return (
    <>
      <Button
        type="button"
        className="border-none !px-3 [&::after]:shadow-none [&::before]:shadow-none"
        color="union"
        onClick={() => setIsOpen(true)}
      >
        <span className="flex items-center">
          <PlayIcon className="mr-2 size-4 text-(--union-on-union)" />
          {buttonText}
        </span>
      </Button>
      <AbortModal isOpen={abortOpen} setIsOpen={setAbortOpen} />
      <LaunchFormDrawer
        drawerMeta={drawerMeta}
        formMethods={formMethods}
        isDataFetched={isDataFetched}
      />
    </>
  )
}
