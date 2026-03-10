import { LaunchFormTab } from '@/components/LaunchForm/Tabs/types'
import { RjsfSchemaFormState } from '@/components/SchemaForm/SchemaForm'
import { TaskSpec } from '@/gen/flyteidl2/task/task_definition_pb'
import { useSearchParams } from 'next/navigation'
import { useQueryState } from 'nuqs'
import React, {
  createContext,
  useContext,
  useMemo,
  useState,
  useCallback,
} from 'react'
import { TriggerName } from '@/gen/flyteidl2/common/identifier_pb'

const initialRjsfFormState = { errors: [], touchedFields: {} }

type Ctx = {
  buttonText: string
  isOpen: boolean
  launchFormTab: LaunchFormTab
  setLaunchFormTab: (tab?: LaunchFormTab | null) => void
  setIsOpen: (newState: boolean) => void
  taskSpec: TaskSpec | null | undefined
  setTaskSpec: React.Dispatch<React.SetStateAction<TaskSpec | null | undefined>>
  /** Errors kept in custom state so they aren't cleared when onBlur validation runs or when form unmounts while switching tabs or pertty/raw view */
  /** Touched fields kept to show errors only for touched inputs */
  rjsfFormState: RjsfSchemaFormState
  setRjsfFormState: React.Dispatch<React.SetStateAction<RjsfSchemaFormState>>
  triggerName: TriggerName | null | undefined
  setTriggerName: React.Dispatch<
    React.SetStateAction<TriggerName | null | undefined>
  >
  /** True while a blob field is uploading; used to disable submit button */
  isBlobUploading: boolean
  setIsBlobUploading: React.Dispatch<React.SetStateAction<boolean>>
}

const LaunchFormStateContext = createContext<Ctx | null>(null)

export const LaunchFormStateProvider: React.FC<{
  buttonText: string
  children: React.ReactNode
}> = ({ buttonText, children }) => {
  const searchParams = useSearchParams()
  const [launchFormTabParam, setLaunchFormTabParam] = useQueryState(
    'launchTab',
    {
      defaultValue: 'inputs',
    },
  )
  const [rjsfFormState, setRjsfFormState] =
    useState<RjsfSchemaFormState>(initialRjsfFormState)

  const [taskSpec, setTaskSpec] = useState<TaskSpec | null | undefined>()
  const [triggerName, setTriggerName] = useState<
    TriggerName | null | undefined
  >()
  const [isOpen, setIsOpen] = useState(!!searchParams.get('launchTab'))
  const [isBlobUploading, setIsBlobUploading] = useState(false)

  const setLaunchFormTab = useCallback(
    (newValue: LaunchFormTab | null = 'inputs') =>
      setLaunchFormTabParam(newValue),
    [setLaunchFormTabParam],
  )

  const value = useMemo<Ctx>(
    () => ({
      buttonText,
      isOpen,
      launchFormTab: launchFormTabParam as LaunchFormTab,
      setIsOpen,
      setLaunchFormTab,
      taskSpec,
      setTaskSpec,
      rjsfFormState,
      setRjsfFormState,
      triggerName,
      setTriggerName,
      isBlobUploading,
      setIsBlobUploading,
    }),
    [
      buttonText,
      isOpen,
      launchFormTabParam,
      setLaunchFormTab,
      taskSpec,
      rjsfFormState,
      setRjsfFormState,
      triggerName,
      setTriggerName,
      isBlobUploading,
      setIsBlobUploading,
    ],
  )

  return (
    <LaunchFormStateContext.Provider value={value}>
      {children}
    </LaunchFormStateContext.Provider>
  )
}

export const useLaunchFormState = () => {
  const ctx = useContext(LaunchFormStateContext)
  if (!ctx)
    throw new Error(
      'useLaunchFormState must be used within LaunchFormStateProvider',
    )
  return ctx
}
