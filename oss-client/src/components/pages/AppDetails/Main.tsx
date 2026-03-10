'use client'

import { CopyButtonWithTooltip } from '@/components/CopyButtonWithTooltip'
import { DetailsLayout } from '@/components/DetailsLayout'
import { DetailsMetadata } from '@/components/DetailsLayout/DetailsLayout'
import { Link } from '@/components/Link'
import { Tabs, TabType } from '@/components/Tabs'
import { BarChartIcon } from '@/components/icons/BarChartIcon'
import { LogsIcon } from '@/components/icons/LogsIcon'
import { TaskIcon } from '@/components/icons/TaskIcon'
import { useAppDetails } from '@/hooks/useApps'
import { useLatestApps } from '@/hooks/useLatestResources'
import { useOrg } from '@/hooks/useOrg'
import { useSelectedTab } from '@/hooks/useQueryParamState'
import { getLastDeployedData, getStatus } from '@/lib/appUtils'
import { getLocation } from '@/lib/windowUtils'
import { useParams } from 'next/navigation'
import { useEffect, useMemo } from 'react'
import { AppStatusBadge } from '../ListApps/components/AppStatusBadge'
import { AppLogsTab } from './AppLogsTab'
import { AppMetricsTab } from './AppMetricsTab'
import { AppSpecTab } from './AppSpecTab'
import { AppStartStopModal } from './AppStartStopModal'
import { AppDetailsParams } from './types'
import { AppDetailsCodeTab } from './AppDetailsCodeTab'
import { CodeIcon } from '@/components/icons/CodeIcon'

export enum AppDetailsTab {
  METRICS = 'metrics',
  LOGS = 'logs',
  APP = 'app',
  DEPLOYMENTS = 'deployments',
  CODE = 'code',
}

const TabLayout = ({ children }: { children: React.ReactNode }) => (
  <div className="flex h-full min-h-0 w-full min-w-0 flex-1 flex-col gap-2 px-8 pb-8">
    {children}
  </div>
)

export const AppDetailsPage = () => {
  const org = useOrg()
  const { setLatestApp } = useLatestApps()
  const params = useParams<AppDetailsParams>()
  const { selectedTab, setSelectedTab } = useSelectedTab(
    AppDetailsTab.LOGS,
    AppDetailsTab,
  )

  const { data } = useAppDetails({
    domain: params.domain,
    name: params.appId,
    org,
    projectId: params.project,
  })

  useEffect(() => {
    if (data?.app) {
      setLatestApp(data.app)
    }
  }, [data?.app, setLatestApp])

  const metadata: DetailsMetadata = useMemo(() => {
    const appName = data?.app?.metadata?.id?.name || '-'
    const replicas = `${data?.app?.status?.currentReplicas || 0} of ${data?.app?.spec?.autoscaling?.replicas?.max || 0}`
    const type = data?.app?.spec?.profile?.type || '-'
    const lastDeployed = getLastDeployedData(data?.app)
    const status = getStatus(data?.app?.status?.conditions)
    const { href } = getLocation()

    return {
      dataList: [
        { label: 'Replicas', value: replicas },
        { label: 'Type', value: type },
        { label: 'Last deployed', value: lastDeployed.relativeTime },
      ],
      title: {
        value: appName,
        badge: (
          <>
            <CopyButtonWithTooltip
              icon="chain"
              textInitial="Copy app URL"
              textCopied="App URL copied to clipboard"
              value={href}
            />
            <AppStatusBadge status={status} />
          </>
        ),
      },
      subtitle: [
        {
          copyValue: data?.app?.status?.ingress?.publicUrl || '-',
          label: '',
          value: data?.app?.status?.ingress?.publicUrl ? (
            <Link
              href={data?.app?.status?.ingress?.publicUrl}
              target="blank"
              rel="noopener noreferrer"
            >
              {data?.app?.status?.ingress?.publicUrl || '-'}
            </Link>
          ) : (
            '-'
          ),
          disableCopy: false,
        },
      ],
    }
  }, [data?.app])

  const tabs: TabType<AppDetailsTab>[] = useMemo(() => {
    const baseTabs: TabType<AppDetailsTab>[] = [
      {
        content: (
          <TabLayout>
            <AppLogsTab app={data?.app} />
          </TabLayout>
        ),
        icon: <LogsIcon />,
        label: 'Logs',
        path: AppDetailsTab.LOGS,
      },
      {
        content: (
          <TabLayout>
            <AppMetricsTab appId={data?.app?.metadata?.id?.name} />
          </TabLayout>
        ),
        icon: <BarChartIcon />,
        label: 'Metrics',
        path: AppDetailsTab.METRICS,
      },
      {
        content: (
          <TabLayout>
            <AppSpecTab app={data?.app} />
          </TabLayout>
        ),
        icon: <TaskIcon />,
        label: 'App',
        path: AppDetailsTab.APP,
      },
      {
        content: (
          <TabLayout>
            <AppDetailsCodeTab />
          </TabLayout>
        ),
        icon: <CodeIcon width={14} />,
        label: 'Code',
        path: AppDetailsTab.CODE,
      },
    ]

    return baseTabs
  }, [data?.app])

  return (
    <>
      <DetailsLayout
        metadata={metadata}
        actionBtn={data?.app && <AppStartStopModal app={data.app} />}
      >
        <div className="flex min-h-0 w-full min-w-0 bg-(--system-gray-2)">
          <Tabs<AppDetailsTab>
            currentTab={selectedTab}
            onClickTab={setSelectedTab}
            tabs={tabs}
          />
        </div>
      </DetailsLayout>
    </>
  )
}
