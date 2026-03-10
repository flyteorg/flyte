'use client'

import { DetailsLayout } from '@/components/DetailsLayout'
import { DetailsMetadata } from '@/components/DetailsLayout/DetailsLayout'
import { CodeIcon } from '@/components/icons/CodeIcon'
import { GithubIcon } from '@/components/icons/GithubIcon'
import { LogsIcon } from '@/components/icons/LogsIcon'
import { TaskIcon } from '@/components/icons/TaskIcon'
import { TaskRunButton } from '@/components/RunButton'
import { StatusDotBadge } from '@/components/StatusDotBadge'
import { Tabs, TabType } from '@/components/Tabs'
import { Tooltip } from '@/components/Tooltip'
import { Sort, Sort_Direction } from '@/gen/flyteidl2/common/list_pb'
import { useLatestTasks } from '@/hooks/useLatestResources'
import { LaunchFormStateProvider } from '@/hooks/useLaunchFormState'
import { useListTaskVersions } from '@/hooks/useListTaskVersions'
import { useOrg } from '@/hooks/useOrg'
import { getSortParamForQueryKey } from '@/hooks/useQueryParamSort'
import { useSelectedTab } from '@/hooks/useQueryParamState'
import { useTaskDetails } from '@/hooks/useTaskDetails'
import { toDateFormat } from '@/lib/dateUtils'
import { getActiveTaskTriggersCount } from '@/lib/triggerUtils'
import { ArrowUpRightIcon } from '@heroicons/react/16/solid'
import { isNumber } from 'lodash'
import Link from 'next/link'
import { useParams, useSearchParams } from 'next/navigation'
import { useEffect, useMemo } from 'react'
import { TaskDetailsCodeTab } from './TaskDetailsCodeTab'
import { TaskDetailsRunsTab } from './TaskDetailsRunsTab'
import { TaskDetailsTaskTab } from './TaskDetailsTaskTab'
import { TaskDetailsTriggersTab } from './TaskDetailsTriggersTab'
import { TaskDetailsPageParams, TaskDetailsTab } from './types'
import { VersionPicker } from './VersionPicker'

const defaultSort: Sort = {
  key: 'created_at',
  direction: Sort_Direction.DESCENDING,
} as Sort

export const TaskDetailsPage = () => {
  const { selectedTab, setSelectedTab } = useSelectedTab(
    TaskDetailsTab.RUNS,
    TaskDetailsTab,
  )
  const params = useParams<TaskDetailsPageParams>()
  const searchParams = useSearchParams()
  const { setLatestTask } = useLatestTasks()

  // Fetch all versions of the task
  const allVersionsQuery = useListTaskVersions({
    taskName: params.name,
    sort: {
      sortBy: defaultSort,
      sortForQueryKey: getSortParamForQueryKey(defaultSort),
    },
  })
  const latestVersion =
    allVersionsQuery.data?.pages?.[0]?.versions?.[0]?.version

  // Use the version from params if provided, otherwise use latest version for display
  const versionToRender = params.version || latestVersion
  const org = useOrg()

  // Fetch latest version details for header metadata only
  const latestTaskDetails = useTaskDetails({
    name: params.name,
    version: latestVersion!,
    project: params.project,
    domain: params.domain,
    org,
    enabled: !!latestVersion,
  })

  useEffect(() => {
    if (latestTaskDetails.data?.details) {
      setLatestTask(latestTaskDetails.data?.details)
    }
  }, [latestTaskDetails.data?.details, setLatestTask])

  // Fetch selected version details for tabs content
  const selectedTaskDetails = useTaskDetails({
    name: params.name,
    version: versionToRender!,
    project: params.project,
    domain: params.domain,
    org,
    enabled: !!versionToRender,
  })

  // Use latest version data for metadata
  const taskData = latestTaskDetails.data?.details
  // Use selected version data for tabs
  const selectedTaskData = selectedTaskDetails.data?.details

  // For run button: when "all versions" is selected, use latest version's spec
  // Otherwise use the selected version's spec
  const runButtonTaskSpec = params.version
    ? selectedTaskData?.spec
    : latestTaskDetails.data?.details?.spec
  const runButtonVersion = params.version || latestVersion

  const tabs: TabType<TaskDetailsTab>[] = useMemo(() => {
    const baseTabs: TabType<TaskDetailsTab>[] = [
      {
        label: 'Runs',
        content: (
          <TaskDetailsRunsTab version={params.version ?? latestVersion ?? ''} />
        ),
        icon: <LogsIcon width={16} />,
        path: TaskDetailsTab.RUNS,
      },
      {
        label: 'Task',
        content: (
          <TaskDetailsTaskTab
            latestVersion={latestVersion}
            version={params.version}
          />
        ),
        icon: <TaskIcon width={14} />,
        path: TaskDetailsTab.TASK,
      },
      {
        label: 'Triggers',
        icon: <LogsIcon width={16} />,
        content: (
          <TaskDetailsTriggersTab
            version={params.version || undefined}
            latestVersion={latestVersion}
          />
        ),
        path: TaskDetailsTab.TRIGGERS,
      },
      {
        label: 'Code',
        icon: <CodeIcon width={14} />,
        content: (
          <TaskDetailsCodeTab
            latestVersion={latestVersion}
            version={params.version}
          />
        ),
        path: TaskDetailsTab.CODE,
      },
    ]

    return baseTabs
  }, [latestVersion, params.version])

  const triggersCount = getActiveTaskTriggersCount(
    taskData?.metadata?.triggersSummary,
  )
  const metadata = useMemo(() => {
    // Use task details if available, otherwise fall back to latest task data

    const shortDescription = taskData?.spec?.documentation?.shortDescription
    return {
      title: { value: taskData?.metadata?.shortName ?? '' },
      subtitle: [
        {
          label: '',
          value:
            taskData &&
            'spec' in taskData &&
            taskData.spec?.taskTemplate?.id?.name
              ? taskData.spec?.taskTemplate?.id?.name
              : (taskData?.metadata?.environmentName ?? ''),
        },
        shortDescription && {
          label: 'Description:',
          value: (
            <Tooltip
              placement="bottom"
              contentClassName="py-1.5 px-4 shadow-[0px_8px_8px_0px_rgba(0,0,0,0.4)] dark:!bg-(--system-gray-1)"
              content={shortDescription}
              disabled={shortDescription.length <= 40}
            >
              <span
                className={
                  shortDescription.length <= 40 ? '' : 'cursor-pointer'
                }
              >
                {shortDescription.slice(0, 40)}
                {shortDescription.length > 40 ? '...' : null}
              </span>
            </Tooltip>
          ),
          disableCopy: true,
        },
      ].filter(Boolean),
      dataList: [
        {
          label: 'Deployed',
          value: toDateFormat({
            includeSeconds: false,
            timestamp: taskData?.metadata?.deployedAt,
          }),
        },
        {
          label: 'Source',
          value: taskData?.spec?.documentation?.sourceCode?.link ? (
            <Link
              href={taskData?.spec?.documentation?.sourceCode?.link ?? ''}
              target="blank"
              className="flex h-5 items-center gap-1 rounded-lg bg-(--system-gray-2) px-1.25 text-nowrap"
            >
              <GithubIcon width={14} height={14} />
              <span>GitHub</span>
              <ArrowUpRightIcon className="size-4 dark:text-(--system-gray-5)" />
            </Link>
          ) : (
            '-'
          ),
        },
        {
          label: 'Trigger',
          value: isNumber(triggersCount) ? (
            <Link
              href={{
                pathname: `/domain/${params.domain}/project/${params.project}/tasks/${params.name}`,
                query: {
                  ...Object.fromEntries(searchParams?.entries() || []),
                  tab: 'triggers',
                },
              }}
            >
              <StatusDotBadge
                color={triggersCount ? 'green' : 'gray'}
                displayText={`${triggersCount} active`}
                filled={!!triggersCount}
              />
            </Link>
          ) : (
            '-'
          ),
        },
      ],
    } as DetailsMetadata
  }, [
    taskData,
    triggersCount,
    params.domain,
    params.name,
    params.project,
    searchParams,
  ])

  return (
    <LaunchFormStateProvider buttonText="Run">
      <DetailsLayout
        metadata={metadata}
        actionBtn={
          <TaskRunButton
            taskSpec={runButtonTaskSpec}
            version={runButtonVersion!}
          />
        }
      >
        <div className="flex min-h-0 w-full min-w-0 bg-(--system-gray-2)">
          <Tabs<TaskDetailsTab>
            tabs={tabs}
            currentTab={selectedTab}
            onClickTab={setSelectedTab}
            scrollShadow
            widgets={[
              {
                key: 'version-picker',
                content: (
                  <VersionPicker
                    versionsQuery={allVersionsQuery}
                    currentVersion={params.version}
                    latestVersion={latestVersion}
                  />
                ),
              },
            ]}
          />
        </div>
      </DetailsLayout>
    </LaunchFormStateProvider>
  )
}
