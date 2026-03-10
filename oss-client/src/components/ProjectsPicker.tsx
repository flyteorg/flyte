import { useLatestProjectDomainPairs } from '@/hooks/useLatestProjects'
import { useProjectData } from '@/hooks/useProjects'
import { useDomainStore } from '@/lib/DomainStore'
import { getUpdatedProjectPath } from '@/lib/urlUtils'
import { useParams, usePathname, useRouter } from 'next/navigation'
import { useMemo } from 'react'
import { CopyButton } from './CopyButton'

import { ProjectState } from '@/gen/flyteidl2/project/project_service_pb'
import { PopoverMenu, type MenuItem } from './Popovers'

export const ProjectsPicker = ({
  shouldDisableNavigation,
}: {
  shouldDisableNavigation?: boolean
}) => {
  const params = useParams<{ project: string }>()
  const selectedDomain = useDomainStore((s) => s.selectedDomain)
  const { latestProjectDomainPairs } = useLatestProjectDomainPairs()
  const projectData = useProjectData({ id: params.project })
  const router = useRouter()
  const pathname = usePathname()
  const projectName = projectData.data?.name || 'Project'

  const projectNavItems: MenuItem[] = useMemo(() => {
    return latestProjectDomainPairs
      .filter((pd) => pd.domain.id === selectedDomain?.id)
      .filter(
        (pd) =>
          pd.project.state !== ProjectState.ARCHIVED &&
          pd.project.state !== ProjectState.SYSTEM_ARCHIVED,
      )
      .filter((pd) => pd.project.id !== params.project)
      .slice(0, 3)
      .map((pd) => ({
        id: pd.project.id,
        label: pd.project.name,
        onClick: () => {
          const updatedPathname = getUpdatedProjectPath(pathname, pd.project.id)
          router.push(updatedPathname)
        },
      }))
  }, [
    latestProjectDomainPairs,
    params.project,
    pathname,
    router,
    selectedDomain?.id,
  ])

  return (
    <div className="flex flex-col">
      <PopoverMenu
        items={[
          {
            id: 'project-label',
            type: 'item',
            label: (
              <div>
                Copy project ID{' '}
                <CopyButton size="sm" value={projectData.data?.id || ''} />
              </div>
            ),
          },
          { id: 'divider', type: 'divider' },
          ...(projectNavItems.length > 0 && !shouldDisableNavigation
            ? [
                ...projectNavItems,
                { id: 'divider', type: 'divider' } as MenuItem,
              ]
            : []),
          {
            id: 'switcher',
            type: 'item',
            label: 'All projects',
            onClick: () => router.push(`/projects`),
          },
        ]}
        label={<div className="text-xs font-semibold">{projectName}</div>}
        showCheckboxes={false}
        portal
        placement="bottom-start"
        size="sm"
        variant="dropdown"
        triggerClassName="text-(--system-gray-7) border-(--system-gray-4) !py-0 whitespace-nowrap dark:hover:text-white hover:text-black hover:bg-transparent"
      />
    </div>
  )
}
