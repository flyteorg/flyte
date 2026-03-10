import { App } from '@/gen/flyteidl2/app/app_definition_pb'
import { AppTableItem } from './types'
import { getLastDeployedData, getStatus } from '@/lib/appUtils'

export const formatAppForTable = (app: App): AppTableItem => ({
  id: app.metadata?.id,
  status: getStatus(app.status?.conditions),
  replicas: {
    min: app.status?.currentReplicas || 0,
    max: app.spec?.autoscaling?.replicas?.max || 0,
  },
  name: {
    displayText: app.metadata?.id?.name || '-',
    endpoint: app.status?.ingress?.publicUrl || '-',
  },
  type: app.spec?.profile?.type || '-',
  lastDeployed: getLastDeployedData(app),
  deployedBy: app.spec?.creator,
  actions: app,
  original: app,
})
