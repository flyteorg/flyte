import { useMemo, useState } from 'react'
import { Tabs } from '@/components/Tabs'
import { CreateTriggerTab, TaskDetails } from './types'
import {
  TriggerAdvancedInputs,
  TriggerAnnotations,
  TriggerDefinition,
  TriggerLabels,
  TriggerEnvVars,
  TriggerInputs,
} from './Form'
import { CreateTriggerFooter } from './Form/CreateTriggerFooter'

type TabConfig = {
  label: string
  content: React.ReactNode
  path: CreateTriggerTab
}

export const CreateTriggerTabs = ({
  onCancel,
  taskDetails,
}: {
  onCancel: () => void
  taskDetails: TaskDetails | undefined
}) => {
  const [currentTab, updateCurrentTab] =
    useState<CreateTriggerTab>('definition')
  const launchFormTabConfig: TabConfig[] = useMemo(
    () => [
      {
        label: 'Definition',
        path: 'definition',
        content: <TriggerDefinition />,
      },
      {
        label: 'Inputs',
        path: 'inputs',
        content: taskDetails ? (
          <TriggerInputs taskDetails={taskDetails} />
        ) : null,
      },
      {
        label: 'Labels',
        path: 'labels',
        content: <TriggerLabels />,
      },
      {
        label: 'Env vars',
        path: 'env-vars',
        content: <TriggerEnvVars />,
      },
      {
        label: 'Annotations',
        path: 'annotations',
        content: <TriggerAnnotations />,
      },
      {
        label: 'Advanced settings',
        path: 'advanced',
        content: <TriggerAdvancedInputs />,
      },
    ],
    [taskDetails],
  )
  return (
    <Tabs<CreateTriggerTab>
      currentTab={currentTab}
      onClickTab={updateCurrentTab}
      tabs={launchFormTabConfig}
      classes={{
        tabListContainer: 'px-4! py-3! border-b-1 border-(--bg-gray)',
        tabContentContainer: 'p-2',
      }}
      footer={
        <CreateTriggerFooter
          onCancel={onCancel}
          taskName={taskDetails?.taskId}
          taskVersion={taskDetails?.taskVersion}
        />
      }
    />
  )
}
