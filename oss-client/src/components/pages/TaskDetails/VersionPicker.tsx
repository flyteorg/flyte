import { ChevronRightIcon } from '@/components/icons/ChevronRightIcon'
import { PopoverMenu, type MenuItem } from '@/components/Popovers'
import { ListVersionsResponse } from '@/gen/flyteidl2/task/task_service_pb'
import { formatDateWithConditionalYear } from '@/lib/dateUtils'
import { getLocation } from '@/lib/windowUtils'
import { useParams, useRouter } from 'next/navigation'
import { useMemo, useState } from 'react'
import { TaskDetailsPageParams } from './types'

export const VersionPicker = ({
  versionsQuery,
  currentVersion,
  latestVersion,
}: {
  versionsQuery: {
    data?: {
      pages?: ListVersionsResponse[]
    }
  }
  currentVersion?: string
  latestVersion?: string
}) => {
  const params = useParams<TaskDetailsPageParams>()
  const router = useRouter()
  const [isOpen, setIsOpen] = useState(false)

  // Get all versions with their dates (already sorted by date from API)
  const tasksWithVersions = useMemo(() => {
    const allVersions =
      versionsQuery.data?.pages?.flatMap((page) => page.versions ?? []) ?? []

    return allVersions.map((versionResponse) => ({
      version: versionResponse.version,
      deployedAt: versionResponse.deployedAt,
    }))
  }, [versionsQuery.data])

  const menuItems: MenuItem[] = useMemo(() => {
    const items: MenuItem[] = []

    // Add header with version count
    items.push({
      id: 'header',
      type: 'custom',
      component: (
        <div className="px-2 py-0.5 text-sm font-semibold text-(--system-gray-7)">
          Versions ({tasksWithVersions.length})
        </div>
      ),
    })

    // Add divider
    items.push({
      id: 'divider',
      type: 'divider',
    })

    // Add "All versions" as the first option
    items.push({
      id: 'all-versions',
      type: 'item',
      label: 'All versions',
      selected: !currentVersion,
      onClick: () => {
        // Navigate to task details page without version parameter
        // Get current query string from window.location to ensure it's up to date
        const queryString = getLocation().search
        const url = `/domain/${params.domain}/project/${params.project}/tasks/${params.name}${queryString}`
        router.push(url)
      },
    })

    // Add each version as a menu item with date
    tasksWithVersions.forEach(({ version, deployedAt }) => {
      const dateStr = formatDateWithConditionalYear(deployedAt)

      const isSelected = version === currentVersion
      items.push({
        id: `version-${version}`,
        type: 'item',
        label: (
          <div className="flex w-full items-center whitespace-nowrap">
            <span className="max-w-[170px] min-w-0 flex-1 truncate">
              {version}
            </span>
            <span
              className={`ml-4 flex-shrink-0 text-right text-xs ${
                isSelected ? 'text-(--system-white)' : 'text-(--system-gray-5)'
              }`}
              style={{ minWidth: '120px' }}
            >
              {dateStr}
            </span>
          </div>
        ),
        selected: isSelected,
        onClick: () => {
          // Navigate to task details page with this version
          // Get current query string from window.location to ensure it's up to date
          const queryString = getLocation().search
          const url = `/domain/${params.domain}/project/${params.project}/tasks/${params.name}/${version}${queryString}`
          router.push(url)
        },
      })
    })

    return items
  }, [
    tasksWithVersions,
    currentVersion,
    params.domain,
    params.project,
    params.name,
    router,
  ])

  if (tasksWithVersions.length === 0) return null

  const getDisplayVersion = () => {
    if (!currentVersion) return 'All versions'
    const isLatest = currentVersion === latestVersion
    return isLatest
      ? `Latest version (${currentVersion})`
      : `Version (${currentVersion})`
  }

  const displayVersion = getDisplayVersion()

  return (
    <PopoverMenu
      items={menuItems}
      showChevron={false}
      showCheckboxes={false}
      noSelectedBackground={true}
      outline={false}
      size="sm"
      variant="dropdown"
      open={isOpen}
      onOpenChange={setIsOpen}
      itemClassName="!py-1 -my-1"
    >
      <div className="flex cursor-pointer items-center gap-1 rounded-lg border border-(--system-gray-6) px-2 py-1 text-sm text-(--system-gray-5) hover:text-(--system-white)">
        {displayVersion}
        <div className="ml-1 flex h-4 w-4 items-center justify-center">
          <ChevronRightIcon className="rotate-90 transition-all" width={8} />
        </div>
      </div>
    </PopoverMenu>
  )
}
