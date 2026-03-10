'use client'

import { Header } from '@/components/Header'
import { NavItems, NavPanelLayout } from '@/components/NavPanel'
import { type NavLink as NavLinkType } from '@/components/NavPanel/types'
import { ProjectsPage } from '@/components/pages/Projects'
import { Project, ProjectState } from '@/gen/flyteidl2/project/project_service_pb'
import { useLatestProjectDomainPairs } from '@/hooks/useLatestProjects'
import { useDomainStore } from '@/lib/DomainStore'
import { FolderIcon } from '@heroicons/react/20/solid'
import { Suspense, useMemo } from 'react'

const ProjectDomainPair = ({ project }: { project: Project }) => {
  return (
    <div className="group relative flex w-full">
      <div className="flex w-full items-center justify-between">
        <div className="text-md max-w-42 truncate font-semibold text-(--system-gray-6)">
          {project.name}
        </div>
      </div>
    </div>
  )
}

export function ProjectsClient() {
  const { latestProjectDomainPairs, setLatestProjectDomain } =
    useLatestProjectDomainPairs()
  const { selectedDomain } = useDomainStore()

  const navLinks: NavLinkType[] = useMemo(() => {
    return latestProjectDomainPairs
      .filter((pd) => pd.domain.id === selectedDomain?.id)
      .filter(
        (pd) =>
          pd.project.state !== ProjectState.ARCHIVED &&
          pd.project.state !== ProjectState.SYSTEM_ARCHIVED,
      )
      .map(({ project, domain }) => ({
        displayComponent: <ProjectDomainPair project={project} />,
        displayText: `${project.name}-${domain.id}`,
        makeHref: () => `/domain/${domain.id}/project/${project.id}/runs`,
        onClick: () => {
          setLatestProjectDomain({ domain, project })
        },
        type: 'link' as const,
        icon: <FolderIcon className="text-(--system-gray-5)" />,
        shouldHideIconOnCollapse: true,
      }))
      .slice(0, 12) // prevent overflow in menu
  }, [latestProjectDomainPairs, selectedDomain?.id, setLatestProjectDomain])

  return (
    <Suspense>
      <div className="flex h-full w-full">
        <NavPanelLayout
          navItems={[NavItems.ProjectsHeader, ...navLinks]}
          initialSize="wide"
          mode="embedded"
        >
          <main className="flex h-full w-full flex-col">
            <Header showSearch={true} />
            <ProjectsPage />
          </main>
        </NavPanelLayout>
      </div>
    </Suspense>
  )
}
