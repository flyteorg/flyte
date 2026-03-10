'use client'

import { CopyButtonWithTooltip } from '@/components/CopyButtonWithTooltip'
import { DetailsLayout } from '@/components/DetailsLayout'
import { DetailsMetadata } from '@/components/DetailsLayout/DetailsLayout'
import { BarChartIcon } from '@/components/icons/BarChartIcon'
import { CodeIcon } from '@/components/icons/CodeIcon'
import { LogsIcon } from '@/components/icons/LogsIcon'
import { ReportsIcon } from '@/components/icons/ReportsIcon'
import { SummaryIcon } from '@/components/icons/SummaryIcon'
import { TaskIcon } from '@/components/icons/TaskIcon'
import { LiveTimestamp } from '@/components/LiveTimestamp'
import { LoadingSpinner } from '@/components/LoadingSpinner'
import { DebugRerunControls } from '@/components/pages/RunDetails/DebugRerunControls'
import { PhaseBadge } from '@/components/PhaseBadge'
import { RunButton } from '@/components/RunButton'
import { Tabs, TabType, TabWidget } from '@/components/Tabs'
import { useLatestRuns } from '@/hooks/useLatestResources'
import { LaunchFormStateProvider } from '@/hooks/useLaunchFormState'
import { useOrg } from '@/hooks/useOrg'
import { useSelectedTab } from '@/hooks/useQueryParamState'
import { useRunDetailState } from '@/hooks/useRunDetailState'
import { useTaskDetails } from '@/hooks/useTaskDetails'
import { useWatchActionDetails } from '@/hooks/useWatchActionDetails'
import { getActionDisplayString, getTaskType } from '@/lib/actionUtils'
import { toDateFormat } from '@/lib/dateUtils'
import { getLocation } from '@/lib/windowUtils'
import clsx from 'clsx'
import dynamic from 'next/dynamic'
import Link from 'next/link'
import { useParams, useSearchParams } from 'next/navigation'
import { useEffect, useMemo, useState } from 'react'
import { AttemptPicker } from './AttemptPicker'
import { useSyncSelectedAttempt } from './hooks/useSelectedAttempt'
import { useSelectedItem } from './hooks/useSelectedItem'
import { RunDetailsGroupTab } from './RunDetailsGroupTab'
import { RunDetailsSummaryTab } from './RunDetailsSummaryTab'
import { RunInfoButton, RunInfoDrawer } from './RunInfo'
import { Sidebar } from './Sidebar'
import { GlobalNowTicker } from './state/GlobalTimestamp'
import { useLayoutStore } from './state/LayoutStore'
import { useRunStore } from './state/RunStore'
import { TriggerBadge } from './TriggerBadge'
import { RunDetailsPageParams, RunDetailsTab } from './types'

const RunDetailsLogsTab = dynamic(
  () =>
    import('./RunDetailsLogsTab').then((mod) => ({
      default: mod.RunDetailsLogsTab,
    })),
  { ssr: false, loading: () => <TabLoadingFallback /> },
)

const RunMetricsTab = dynamic(
  () =>
    import('./RunMetricsTab').then((mod) => ({
      default: mod.RunMetricsTab,
    })),
  { ssr: false, loading: () => <TabLoadingFallback /> },
)

const RunDetailsReportsTab = dynamic(
  () =>
    import('./Reports/RunDetailsReportsTab').then((mod) => ({
      default: mod.RunDetailsReportsTab,
    })),
  { ssr: false, loading: () => <TabLoadingFallback /> },
)

const RunDetailsTaskTab = dynamic(
  () =>
    import('./RunDetailsTaskTab').then((mod) => ({
      default: mod.RunDetailsTaskTab,
    })),
  { ssr: false, loading: () => <TabLoadingFallback /> },
)

const RunDetailsCodeTab = dynamic(
  () =>
    import('./RunDetailsCodeTab').then((mod) => ({
      default: mod.RunDetailsCodeTab,
    })),
  { ssr: false, loading: () => <TabLoadingFallback /> },
)

function TabLoadingFallback() {
  return (
    <div className="flex h-32 items-center justify-center">
      <LoadingSpinner />
    </div>
  )
}

export const RunDetailsPage = () => {
  const searchParams = useSearchParams()
  const params = useParams<RunDetailsPageParams>()
  const { selectedTab, setSelectedTab } = useSelectedTab(
    RunDetailsTab.OVERVIEW,
    RunDetailsTab,
  )
  const { selectedItem } = useSelectedItem()
  const selectedActionId =
    selectedItem?.type === 'action' ? selectedItem.id : null
  const selectedActionDetails = useWatchActionDetails(selectedActionId)

  const run = useRunStore((s) => s.run?.action)
  const selectedAction = useRunStore((s) => s.getAction(selectedActionId || ''))
  const taskType = getTaskType(selectedAction?.action?.metadata)
  const totalActionAttempts = selectedAction?.action?.status?.attempts || 0
  const shouldShowAttemptsPicker = totalActionAttempts > 1
  const setRun = useRunStore((s) => s.setRun)
  const mode = useLayoutStore((s) => s.mode)
  const { setLatestRun } = useLatestRuns()
  const [isRunInfoOpen, setIsRunInfoOpen] = useState(false)

  useEffect(() => {
    if (run?.id) {
      setLatestRun(run)
    }
  }, [run, setLatestRun])

  useSyncSelectedAttempt(selectedActionDetails?.data)

  const validTabs = useMemo(() => {
    if (selectedItem?.type === 'group' || taskType === 'trace') {
      return []
    }
    const tabs: TabType<RunDetailsTab>[] = [
      {
        label: 'Summary',
        content: (
          <RunDetailsSummaryTab
            selectedActionDetailsQuery={selectedActionDetails}
          />
        ),
        icon: <SummaryIcon width={14} />,
        path: RunDetailsTab.OVERVIEW,
      },
      {
        content: <RunDetailsLogsTab />,
        icon: <LogsIcon width={16} />,
        label: 'Logs',
        path: RunDetailsTab.LOGS,
      },
      {
        content: <RunMetricsTab />,
        icon: <BarChartIcon />,
        label: 'Metrics',
        path: RunDetailsTab.METRICS,
      },
      {
        label: 'Reports',
        icon: <ReportsIcon width={14} />,
        content: <RunDetailsReportsTab />,
        path: RunDetailsTab.REPORTS,
      },
      {
        label: 'Task',
        icon: <TaskIcon width={14} />,
        content: <RunDetailsTaskTab />,
        path: RunDetailsTab.TASK,
      },
      {
        label: 'Code',
        icon: <CodeIcon width={14} />,
        content: <RunDetailsCodeTab />,
        path: RunDetailsTab.CODE,
      },
    ]
    return tabs
  }, [selectedItem, taskType, selectedActionDetails])

  const orgId = useOrg()
  useRunDetailState({
    domain: params?.domain || '',
    projectId: params?.project || '',
    orgId,
    runId: params?.runId || '',
  })

  // Task identifier from run action (for header task link existence check)
  const headerTaskSpec =
    run?.metadata?.spec?.case === 'task' ? run.metadata.spec.value : null
  const headerTaskId = headerTaskSpec?.id
  const headerTaskDetailsQuery = useTaskDetails({
    project: headerTaskId?.project ?? '',
    domain: headerTaskId?.domain ?? '',
    name: headerTaskId?.name ?? '',
    version: headerTaskId?.version ?? '',
    org: headerTaskId?.org || orgId,
    enabled: !!(
      headerTaskId?.domain &&
      headerTaskId?.project &&
      headerTaskId?.name &&
      headerTaskId?.version
    ),
  })
  const headerTaskExists = !!headerTaskDetailsQuery.data?.details

  // right-aligned taskbar items
  const widgets: TabWidget[] = useMemo(() => {
    const defaultWidgets: TabWidget[] = [
      {
        key: 'vscodeLog',
        content: (
          <DebugRerunControls actionDetailsQuery={selectedActionDetails} />
        ),
      },
    ]

    return shouldShowAttemptsPicker
      ? [
          {
            key: 'attemptsPicker',
            content: (
              <AttemptPicker actionDetails={selectedActionDetails?.data} />
            ),
          },
          ...defaultWidgets,
        ]
      : defaultWidgets
  }, [selectedActionDetails, shouldShowAttemptsPicker])

  const { startTime, endTime, phase } = run?.status || {}

  const { origin, pathname } = getLocation()

  const metadata = useMemo(() => {
    const newSearchParams = new URLSearchParams(searchParams)
    newSearchParams.set('i', run?.id?.name ?? '')
    const url = `${origin}${pathname}?${newSearchParams.toString()}`
    const runDisplayName = getActionDisplayString(run)
    const taskSpec =
      run?.metadata?.spec?.case === 'task' ? run.metadata.spec.value : null
    const taskId = taskSpec?.id
    const taskDisplayName = taskSpec?.id?.name ?? ''
    const taskDetailsHref =
      taskId?.domain && taskId?.project && taskId?.name && taskId?.version
        ? `/domain/${taskId.domain}/project/${taskId.project}/tasks/${taskId.name}/${taskId.version}`
        : null
    return {
      title: {
        value: runDisplayName,
        badge: (
          <>
            <CopyButtonWithTooltip
              icon="chain"
              textInitial="Copy run URL"
              textCopied="Run URL copied to clipboard"
              value={url}
              classNameBtn="-mx-2"
            />
            <PhaseBadge phase={phase} />
          </>
        ),
      },
      subtitle: [
        { label: 'Run:', value: run?.id?.run?.name ?? '' },
        ...(taskDisplayName
          ? [
              {
                label: 'Task:',
                value:
                  taskDetailsHref && headerTaskExists ? (
                    <Link href={taskDetailsHref} className="hover:underline">
                      {taskDisplayName}
                    </Link>
                  ) : (
                    taskDisplayName
                  ),
                copyValue: taskDisplayName,
              },
            ]
          : []),
      ],
      dataList: [
        {
          label: 'Duration',
          value: (
            <LiveTimestamp
              className="text-sm font-medium"
              endTimestamp={endTime}
              timestamp={startTime}
            />
          ),
        },
        { label: 'Start Time', value: toDateFormat({ timestamp: startTime }) },
        {
          label: 'Trigger',
          value: <TriggerBadge action={run} />,
        },
      ],
    } as DetailsMetadata
  }, [
    endTime,
    headerTaskExists,
    phase,
    run,
    searchParams,
    startTime,
    origin,
    pathname,
  ])

  // reset run when page unmounts
  useEffect(() => {
    return () => setRun(null)
  }, [setRun])

  return (
    <>
      <GlobalNowTicker />
      <LaunchFormStateProvider buttonText="Rerun">
        <DetailsLayout
          metadata={metadata}
          actionBtn={
            <div className="flex items-center gap-3">
              <RunInfoButton onClick={() => setIsRunInfoOpen(true)} />
              <RunButton />
            </div>
          }
          classes={{
            childrenContainer: clsx(mode === 'default' && 'after:left-[24px]'),
            metadataContainer: 'pr-4 pl-6',
          }}
          navPanelLayoutProps={{ initialSize: 'thin', mode: 'overlay' }}
        >
          <Sidebar />
          {mode !== 'full-action-log' && (
            <div className="sticky z-10 flex min-h-0 w-full min-w-0 justify-between border-l-1 border-(--system-gray-2) bg-(--system-gray-2) dark:border-(--system-gray-3)">
              {taskType === 'trace' ? (
                <div className="relative z-0 min-h-0 flex-1 overflow-auto [scrollbar-gutter:stable_both-edges] [&::-webkit-scrollbar]:w-2 [&::-webkit-scrollbar-corner]:bg-transparent [&::-webkit-scrollbar-thumb]:rounded-full [&::-webkit-scrollbar-thumb]:bg-zinc-400 dark:[&::-webkit-scrollbar-thumb]:bg-zinc-600 [&::-webkit-scrollbar-thumb:hover]:bg-zinc-500 dark:[&::-webkit-scrollbar-thumb:hover]:bg-zinc-500 [&::-webkit-scrollbar-track]:bg-transparent">
                  <RunDetailsSummaryTab
                    selectedActionDetailsQuery={selectedActionDetails}
                  />
                </div>
              ) : selectedItem?.type === 'group' ? (
                <RunDetailsGroupTab />
              ) : (
                <Tabs<RunDetailsTab>
                  tabs={validTabs}
                  currentTab={selectedTab}
                  onClickTab={setSelectedTab}
                  widgets={widgets}
                  scrollShadow
                />
              )}
            </div>
          )}
          <RunInfoDrawer isOpen={isRunInfoOpen} setIsOpen={setIsRunInfoOpen} />
        </DetailsLayout>
      </LaunchFormStateProvider>
    </>
  )
}
