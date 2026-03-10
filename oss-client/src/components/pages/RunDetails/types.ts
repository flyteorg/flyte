export type ProjectDomainParams = {
  domain: string
  project: string
}

export type RunDetailsPageParams = ProjectDomainParams & {
  runId: string
}

export function isSafeParams(
  p: RunDetailsPageParams | null,
): p is RunDetailsPageParams {
  return (p as RunDetailsPageParams).domain !== undefined
}

export enum RunDetailsTab {
  OVERVIEW = 'overview',
  LOGS = 'logs',
  REPORTS = 'reports',
  TASK = 'task',
  ACCESS = 'access',
  METRICS = 'metrics',
  CODE = 'code',
}

export enum RunLogType {
  RUN = 'RUN',
  K8S = 'K8S',
}

export type AvailableRunLogTypes = Record<RunLogType, boolean>

export type ChildPhaseCounts = { [key: number]: number }

export type ExternalLinkUrlProps = {
  name: string
  url: string
  icon?: React.FC<{ className?: string }>
  ready?: boolean
}
