import { PopoverMenu } from '@/components/Popovers'
import {
  App,
  Status_DeploymentStatus,
} from '@/gen/flyteidl2/app/app_definition_pb'
import { getStatus } from '@/lib/appUtils'
import { useStartApp, useStopApp } from '@/hooks/useApps'
import { useRouter } from 'next/navigation'
import { getLocation } from '@/lib/windowUtils'

export const ListAppsOverflowActions = ({ app }: { app: App }) => {
  const status = getStatus(app.status?.conditions)
  const router = useRouter()
  const isUnspecifiedOrStopped =
    status === Status_DeploymentStatus.UNASSIGNED ||
    status === Status_DeploymentStatus.STOPPED

  const startApp = useStartApp({ app })
  const stopApp = useStopApp({ app })

  const startStop = isUnspecifiedOrStopped
    ? { id: 'start-app', label: 'Start app', onClick: () => startApp.mutate() }
    : { id: 'stop-app', label: 'Stop app', onClick: () => stopApp.mutate() }
  const { pathname } = getLocation()

  return (
    <div
      onClick={(e) => {
        e.preventDefault()
      }}
    >
      <PopoverMenu
        items={[
          {
            id: 'app-details',
            label: 'View app details',
            onClick: () => {
              const route = pathname.replace('/v2', '')
              router.push(`${route}/${app.metadata?.id?.name}`)
            },
          },
          {
            id: 'divider-01',
            type: 'divider',
          },
          startStop,
          {
            id: 'divider-02',
            type: 'divider',
          },
          {
            id: 'copy-name',
            label: 'Copy app name',
            onClick: () =>
              navigator.clipboard.writeText(app.metadata?.id?.name || '-'),
          },
          {
            id: 'copy-endpoint',
            label: 'Copy endpoint link',
            onClick: () =>
              navigator.clipboard.writeText(
                app.status?.ingress?.publicUrl || '-',
              ),
          },
        ]}
        variant="overflow"
        overflowProps={{ orientation: 'horizontal' }}
      />
    </div>
  )
}
