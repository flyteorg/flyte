import { App, Condition } from '@/gen/flyteidl2/app/app_definition_pb'
import { Status_DeploymentStatus } from '@/gen/flyteidl2/app/app_definition_pb'
import { getRelativeDate } from '@/lib/dateUtils'

export const getStatus = (appConditions: Condition[] | undefined) => {
  return (
    appConditions?.[appConditions.length - 1].deploymentStatus ||
    Status_DeploymentStatus.UNSPECIFIED
  )
}

export const getLastDeployed = (appConditions: Condition[] | undefined) => {
  return appConditions?.find(
    (condition) =>
      condition.deploymentStatus === Status_DeploymentStatus.ACTIVE,
  )
}

export const getLastDeployedData = (app: App | undefined) => {
  const lastDeployment = getLastDeployed(app?.status?.conditions)
  let relativeTime = '-'
  if (lastDeployment?.lastTransitionTime) {
    relativeTime = getRelativeDate(lastDeployment?.lastTransitionTime, 'always')
  }
  return {
    deployedTimestamp: lastDeployment?.lastTransitionTime,
    relativeTime,
    version: Number(lastDeployment?.revision).toString() || '-',
  }
}
