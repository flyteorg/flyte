import { Domain, Project } from '@/gen/flyteidl2/project/project_service_pb'

export function getUniqueDomains(projects: Project[]): Domain[] {
  const domainMap = new Map<string, Domain>()
  projects
    .flatMap((project) => project.domains)
    .forEach((domain) => {
      if (!domainMap.has(domain.id)) {
        domainMap.set(domain.id, domain)
      }
    })

  return Array.from(domainMap.values())
}
