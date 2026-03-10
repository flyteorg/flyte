export type ProjectDomainParams = {
  domain: string
  project: string
}

export type TriggerDetailsPageParams = ProjectDomainParams & {
  taskName: string
  version: string
  name: string
}

export enum TriggerDetailsTab {
  RUNS = 'runs',
  SPEC = 'spec',
  ACTIVITY = 'activity',
}
