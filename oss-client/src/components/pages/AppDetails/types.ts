export type ProjectDomainParams = {
  domain: string
  project: string
}

export type AppDetailsParams = ProjectDomainParams & {
  appId: string
}
