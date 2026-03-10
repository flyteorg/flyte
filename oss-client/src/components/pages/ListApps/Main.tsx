'use client'

import { Header } from '@/components/Header'
import { NavPanelLayout } from '@/components/NavPanel/NavPanelLayout'
import { useOrg } from '@/hooks/useOrg'
import { useParams } from 'next/navigation'
import { Suspense } from 'react'
import { ProjectDomainParams } from '../RunDetails/types'
import { useListApps } from '@/hooks/useApps'
import { ListAppsContent } from './ListAppsContent'
import { ListAppsSearch } from './components'
import { useSearchTerm } from '@/hooks/useQueryParamState'

export function ListAppsPage() {
  const params = useParams<ProjectDomainParams>()
  const { searchTerm } = useSearchTerm()
  const org = useOrg()

  const appsQuery = useListApps({
    domain: params.domain,
    org,
    projectId: params.project,
    search: searchTerm,
  })

  return (
    <Suspense>
      <main className="bg-primary flex h-full min-h-0 w-full">
        <NavPanelLayout initialSize="wide" mode="embedded">
          <div className="flex h-full min-h-0 w-full flex-col">
            <Header showSearch={true} />

            <div className="flex items-center justify-between gap-2 px-10 pt-6 pb-3">
              <div className="flex flex-col">
                <h1 className="text-xl font-medium">Apps</h1>
              </div>
              <ListAppsSearch />
            </div>

            <ListAppsContent
              listAppsQuery={appsQuery}
              searchQuery={searchTerm}
            />
          </div>
        </NavPanelLayout>
      </main>
    </Suspense>
  )
}
