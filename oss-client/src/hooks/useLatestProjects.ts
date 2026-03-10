import { Domain, Project } from '@/gen/flyteidl2/project/project_service_pb'
import { useCallback } from 'react'
import useLocalStorage from 'use-local-storage'

type ProjectDomain = {
  project: Project
  domain: Domain
}

// Save last 30 project/domain pairs to local storage to
export const useLatestProjectDomainPairs = () => {
  const [latestProjectDomainPairs, setLatestProjectDomainPairs] =
    useLocalStorage<ProjectDomain[]>('latestProjectDomainPairs', [])
  const setLatestProjectDomain = useCallback(
    ({ domain, project }: { domain: Domain; project: Project }) => {
      setLatestProjectDomainPairs((prev: ProjectDomain[] | undefined) => {
        const previousProjectDomains: ProjectDomain[] = prev || []
        const withoutDuplicate: ProjectDomain[] = previousProjectDomains.filter(
          (pd: ProjectDomain) =>
            !(pd.project.id === project.id && pd.domain.id === domain.id),
        )
        const updated: ProjectDomain[] = [
          { project, domain },
          ...withoutDuplicate,
        ]
        return updated.slice(0, 30)
      })
    },
    [setLatestProjectDomainPairs],
  )
  const removeProjectDomain = useCallback(
    ({ domain, project }: { domain: Domain; project: Project }) => {
      setLatestProjectDomainPairs((prev: ProjectDomain[] | undefined) => {
        const previousProjectDomains: ProjectDomain[] = prev || []
        const updated: ProjectDomain[] = previousProjectDomains.filter(
          (pd: ProjectDomain) =>
            !(pd.project.id === project.id && pd.domain.id === domain.id),
        )
        return updated
      })
    },
    [setLatestProjectDomainPairs],
  )

  const removeProjectById = useCallback(
    (projectId: string) => {
      setLatestProjectDomainPairs((prev: ProjectDomain[] | undefined) => {
        const previousProjectDomains: ProjectDomain[] = prev || []
        return previousProjectDomains.filter(
          (pd: ProjectDomain) => pd.project.id !== projectId,
        )
      })
    },
    [setLatestProjectDomainPairs],
  )

  return {
    latestProjectDomainPairs,
    removeProjectById,
    removeProjectDomain,
    setLatestProjectDomain,
  }
}
