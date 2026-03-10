export enum SearchFilter {
  task = 'Taskname',
  app = 'App',
  none = 'None',
  run = 'Run',
}

export type SearchResult = {
  href: string
  id: string
  isLocalStorage?: boolean
  displayText: string
  secondaryText?: string
  content: React.ReactNode
  dateVisited?: string
  sortByDate: number
  status?: React.ReactNode
}
