'use client'

import { Header } from '@/components/Header'
import { NavPanelLayout } from '@/components/NavPanel'
import { useTriggerAPIFilters } from '@/hooks/filters/useTriggerAPIFilters'
import { useOrg } from '@/hooks/useOrg'
import { useSearchTerm } from '@/hooks/useQueryParamState'
import { useQueryParamSort } from '@/hooks/useQueryParamSort'
import { useListTriggers } from '@/hooks/useTriggers'
import { Sort, Sort_Direction } from '@/gen/flyteidl2/common/list_pb'
import { type ProjectDomainPageParams } from '@/types/pageParams'
import { useParams } from 'next/navigation'
import { Suspense, useState } from 'react'
import { CreateTriggerDrawer } from './CreateTrigger/Main'
import { ListTriggersPageContent } from './ListTriggersPageContent/ListTriggersPageContent'
import { TriggersToolbar } from './TriggersToolbar/TriggersToolbar'
import { TriggerTableRowWithHighlights } from './TriggersTable/types'

const defaultSort: Sort = {
  key: 'active',
  direction: Sort_Direction.DESCENDING,
} as Sort

export const ListTriggersPage = () => {
  const { searchTerm } = useSearchTerm()
  const { querySort, tableSort } =
    useQueryParamSort<TriggerTableRowWithHighlights>({ defaultSort })
  const { filters, hasActiveFilters, clearAllFilters } = useTriggerAPIFilters()
  const [showCreateTrigger, setShowCreateTrigger] = useState(false)
  const org = useOrg()
  const params = useParams<ProjectDomainPageParams>()

  const triggersQuery = useListTriggers({
    org,
    domain: params.domain,
    projectId: params.project,
    search: searchTerm || '',
    filters,
    sort: querySort,
  })

  return (
    <Suspense>
      <div className="flex h-full w-full">
        <NavPanelLayout mode="embedded" initialSize="wide">
          <main className="bg-primary flex h-full w-full flex-col">
            <Header showSearch={true} />
            <TriggersToolbar triggersQuery={triggersQuery} />

            <ListTriggersPageContent
              triggersQuery={triggersQuery}
              searchTerm={searchTerm}
              hasActiveFilters={hasActiveFilters}
              clearAllFilters={clearAllFilters}
              tableSort={tableSort}
              setShowCreateTrigger={() => setShowCreateTrigger(true)}
            />
          </main>
          <CreateTriggerDrawer
            isOpen={showCreateTrigger}
            setIsOpen={setShowCreateTrigger}
          />
        </NavPanelLayout>
      </div>
    </Suspense>
  )
}
