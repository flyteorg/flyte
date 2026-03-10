import { DOCS_BYOC_USER_GUIDE_URL } from '@/lib/constants'
import { useDomainStore } from '@/lib/DomainStore'
import { getProfileHref } from '@/lib/navigationUtils'
import { FolderIcon } from '@heroicons/react/16/solid'
import { DocumentTextIcon } from '@heroicons/react/20/solid'
import { ShareIcon } from '@heroicons/react/24/outline'
import { Cog6ToothIcon } from '@heroicons/react/24/solid'
import { AppsIcon } from '../icons/AppsIcon'
import { RunsIcon } from '../icons/RunsIcon'
import { TriggersIcon } from '../icons/TriggersIcon'
import { NavPanelWidth, NavWidget, type NavLink as NavLinkType } from './types'

export const SettingsLink: NavLinkType = {
  displayText: 'Settings',
  makeHref: ({ pathname }) => getProfileHref(pathname),
  icon: <Cog6ToothIcon className="size-4" />,
  type: 'link',
}

export const ProjectsLink: NavLinkType = {
  className: 'semibold text-white',
  displayText: 'Projects',
  makeHref: () => `/projects`,
  icon: <FolderIcon />,
  type: 'link',
  shouldHideIconOnCollapse: true,
}

const ProjectsHeaderWidget = ({ size }: { size: NavPanelWidth }) => {
  const { selectedDomain } = useDomainStore()
  if (!selectedDomain || size === 'thin') {
    return null
  }
  return (
    <span className="pl-2 text-2xs font-semibold dark:text-(--system-gray-5)">
      Recently viewed in {selectedDomain.name}
    </span>
  )
}

export const ProjectsHeader: NavWidget = {
  displayText: 'projectsHeader',
  type: 'widget',
  widget: (size) => <ProjectsHeaderWidget size={size} />,
}

export const RunsLink: NavLinkType = {
  displayText: 'Runs',
  makeHref: ({ project, domain }) =>
    `/domain/${domain}/project/${project}/runs`,
  icon: <RunsIcon className="size-4" />,
  type: 'link',
}

export const TasksLink: NavLinkType = {
  displayText: 'Tasks',
  makeHref: ({ project, domain }) =>
    `/domain/${domain}/project/${project}/tasks`,
  icon: <ShareIcon className="size-4 min-w-4" />,
  type: 'link',
}

export const TriggersLink: NavLinkType = {
  displayText: 'Triggers',
  makeHref: ({ project, domain }) =>
    `/domain/${domain}/project/${project}/triggers`,
  icon: <TriggersIcon />,
  type: 'link',
}

export const AppsLink: NavLinkType = {
  displayText: 'Apps',
  makeHref: ({ project, domain }) =>
    `/domain/${domain}/project/${project}/apps`,
  icon: <AppsIcon />,
  type: 'link',
}

export const useDefaultItems = () => {
  return [RunsLink, TriggersLink, TasksLink, AppsLink].filter(
    Boolean,
  ) as NavLinkType[]
}

export const DocumentationLink: NavLinkType = {
  displayText: 'Documentation',
  makeHref: () => `${DOCS_BYOC_USER_GUIDE_URL}/flyte-2/`,
  icon: <DocumentTextIcon className="h-4" />,
  type: 'link',
  target: '_blank',
}

export const useDefaultOrgItems = () => [SettingsLink, DocumentationLink]
