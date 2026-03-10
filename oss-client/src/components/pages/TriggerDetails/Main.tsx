'use client'

import { DetailsLayout } from '@/components/DetailsLayout'
import { DetailsMetadata } from '@/components/DetailsLayout/DetailsLayout'
import { LogsIcon } from '@/components/icons/LogsIcon'
import { TriggerRunButton } from '@/components/RunButton'
import { Switch } from '@/components/Switch'
import { Tabs, TabType } from '@/components/Tabs'
import { LaunchFormStateProvider } from '@/hooks/useLaunchFormState'
import { useSelectedTab } from '@/hooks/useQueryParamState'
import {
  UpdateTriggersProps,
  useGetTriggerDetails,
  useUpdateTriggers,
} from '@/hooks/useTriggers'
import { toDateFormat } from '@/lib/dateUtils'
import {
  getNextExecutionTime,
  getTriggerScheduleString,
} from '@/lib/triggerUtils'

import { LinkPill } from '@/components/Link'
import { useOrg } from '@/hooks/useOrg'
import { useParams } from 'next/navigation'
import { useCallback, useMemo } from 'react'
import { TriggerActivityTab } from './TriggerActivityTab'
import { TriggerRunsTab } from './TriggerRunsTab'
import { TriggerSpecTab } from './TriggerSpecTab'
import { TriggerDetailsPageParams, TriggerDetailsTab } from './types'

const tabs: TabType<TriggerDetailsTab>[] = [
  {
    label: 'Runs',
    content: <TriggerRunsTab />,
    icon: <LogsIcon width={16} />,
    path: TriggerDetailsTab.RUNS,
  },
  {
    label: 'Trigger',
    content: <TriggerSpecTab />,
    icon: <LogsIcon width={16} />,
    path: TriggerDetailsTab.SPEC,
  },
  {
    label: 'Activity Log',
    content: <TriggerActivityTab />,
    icon: <LogsIcon width={16} />,
    path: TriggerDetailsTab.ACTIVITY,
  },
]

export const TriggerDetailsPage = () => {
  const { selectedTab, setSelectedTab } = useSelectedTab(
    TriggerDetailsTab.RUNS,
    TriggerDetailsTab,
  )
  const params = useParams<TriggerDetailsPageParams>()
  const org = useOrg()

  const decodedName = useMemo(
    () => (params.name ? decodeURIComponent(params.name) : ''),
    [params.name],
  )

  const triggerDetailsQuery = useGetTriggerDetails({
    org,
    name: decodedName,
    projectId: params.project,
    domain: params.domain,
    taskName: params.taskName,
  })

  const triggerDetails = triggerDetailsQuery.data?.trigger

  const { mutate } = useUpdateTriggers({
    org,
    domain: params.domain,
    projectId: params.project,
    search: decodedName,
    taskNames: [params.taskName],
    invalidationTimeoutMs: 0,
  })

  const onChangeCallback = useCallback(
    (props: UpdateTriggersProps) => {
      mutate(props)
    },
    [mutate],
  )

  const metadata: DetailsMetadata = useMemo(() => {
    const lastTriggeredAt = triggerDetails?.status?.triggeredAt
    const lastUpdatedAt = triggerDetails?.status?.updatedAt

    return {
      leftControls: (
        <div className="text-sm text-zinc-950 dark:text-white">
          <Switch
            size="sm"
            color="green"
            checked={triggerDetails?.spec?.active ?? false}
            onChange={() => {
              if (!triggerDetails || !triggerDetails?.id?.name) return

              onChangeCallback({
                active: !triggerDetails.spec?.active,
                triggerNames: [triggerDetails.id?.name],
              })
            }}
          />
        </div>
      ),
      title: { value: triggerDetails?.id?.name?.name ?? '' },
      subtitle: [
        {
          disableCopy: true,
          label: 'Schedule:',
          value: getTriggerScheduleString(
            triggerDetails?.automationSpec?.automation,
          ),
        },
      ],
      dataList: [
        {
          label: 'Next run',
          value: triggerDetails?.spec?.active
            ? getNextExecutionTime(
                triggerDetails?.automationSpec?.automation,
                lastTriggeredAt,
                lastUpdatedAt,
              )
            : // If not enabled, don't compute next run
              '-',
          className: 'min-w-38',
        },
        {
          label: 'Associated task',
          value: (
            <LinkPill
              displayText={triggerDetails?.id?.name?.taskName}
              href={`/domain/${triggerDetails?.id?.name?.domain}/project/${triggerDetails?.id?.name?.project}/tasks/${triggerDetails?.id?.name?.taskName}`}
            />
          ),
        },
        {
          label: 'Updated',
          value: toDateFormat({
            timestamp: triggerDetails?.status?.updatedAt,
          }),
        },
      ],
    }
  }, [triggerDetails, onChangeCallback])

  return (
    <LaunchFormStateProvider buttonText="Run">
      <DetailsLayout
        metadata={metadata}
        actionBtn={
          triggerDetails?.id?.name ? (
            <TriggerRunButton triggerName={triggerDetails.id.name} />
          ) : undefined
        }
      >
        <div className="flex min-h-0 w-full min-w-0 justify-between bg-(--system-gray-2)">
          <Tabs<TriggerDetailsTab>
            tabs={tabs}
            currentTab={selectedTab}
            onClickTab={setSelectedTab}
            scrollShadow
          />
        </div>
      </DetailsLayout>
    </LaunchFormStateProvider>
  )
}
