export type ProjectDomainParams = {
  domain: string
  project: string
}

export type TaskDetailsPageParams = ProjectDomainParams & {
  version: string
  name: string
}

export enum TaskDetailsTab {
  RUNS = 'runs',
  TASK = 'task',
  TRIGGERS = 'triggers',
  CODE = 'code',
}
