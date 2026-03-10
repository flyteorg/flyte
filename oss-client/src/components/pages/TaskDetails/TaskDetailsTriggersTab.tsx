import { useTriggerAPIFilters } from '@/hooks/filters/useTriggerAPIFilters'
import { useOrg } from '@/hooks/useOrg'
import { useSearchTerm } from '@/hooks/useQueryParamState'
import { useListTriggers } from '@/hooks/useTriggers'
import { useParams, useRouter } from 'next/navigation'
import React from 'react'
import { ListTriggersPageContent } from '../ListTriggers/ListTriggersPageContent/ListTriggersPageContent'
import { NotLatestVersionMessage } from './NotLatestVersionMessage'
import { TaskDetailsPageParams } from './types'
import { getLocation } from '@/lib/windowUtils'

export const TaskDetailsTriggersTab: React.FC<{
  version?: string
  latestVersion?: string
}> = ({ version, latestVersion }) => {
  const params = useParams<TaskDetailsPageParams>()
  const router = useRouter()
  const org = useOrg()
  const { searchTerm } = useSearchTerm()
  const { filters, hasActiveFilters, clearAllFilters } = useTriggerAPIFilters()

  const isLatestVersion = !version || version === latestVersion
  const hasNoVersion = !version

  const triggersQuery = useListTriggers({
    org,
    domain: params.domain,
    projectId: params.project,
    search: searchTerm || '',
    filters,
    taskName: params.name,
    taskVersion: version,
    enabled: isLatestVersion || hasNoVersion,
  })
  const { search } = getLocation()

  const handleViewLatestVersion = () => {
    const queryString = search
    if (latestVersion) {
      router.push(
        `/domain/${params.domain}/project/${params.project}/tasks/${params.name}/${latestVersion}${queryString}`,
      )
    } else {
      router.push(
        `/domain/${params.domain}/project/${params.project}/tasks/${params.name}${queryString}`,
      )
    }
  }

  return (
    <div className="flex h-full min-h-0 w-full min-w-0 flex-1 flex-col gap-2 px-8 pb-8">
      {!isLatestVersion && !hasNoVersion ? (
        <NotLatestVersionMessage
          onViewLatestVersion={handleViewLatestVersion}
        />
      ) : (
        <div className="flex h-full min-h-0 w-full min-w-0 flex-1 flex-col rounded-lg border border-zinc-200 bg-white pt-4 dark:border-zinc-700 dark:bg-(--system-black)">
          <ListTriggersPageContent
            searchTerm={''}
            triggersQuery={triggersQuery}
            className=""
            filtersClassName="pl-4 !pb-4"
            taskName={params.name}
            hideColumns={['task']}
            hasActiveFilters={hasActiveFilters}
            clearAllFilters={clearAllFilters}
          />
        </div>
      )}
    </div>
  )
}
