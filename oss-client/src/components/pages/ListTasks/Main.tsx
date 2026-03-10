'use client'

import { Header } from '@/components/Header'
import { NavPanelLayout } from '@/components/NavPanel'
import { SearchBar } from '@/components/SearchBar'
import {
  Filter_Function,
  Sort,
  Sort_Direction,
} from '@/gen/flyteidl2/common/list_pb'
import { useTaskAPIFilters } from '@/hooks/filters/useTaskAPIFilters'
import { useListTasks } from '@/hooks/useListTasks'
import { useSearchTerm } from '@/hooks/useQueryParamState'
import { useQueryParamSort } from '@/hooks/useQueryParamSort'
import { Suspense } from 'react'
import { TasksPageContent } from './components/TasksPageContent'
import { TaskTableRowWithHighlights } from './TasksTable/types'

const defaultSort: Sort = {
  key: 'created_at',
  direction: Sort_Direction.DESCENDING,
} as Sort

export function ListTasksPage() {
  const { searchTermInput, searchTerm, setSearchTerm } = useSearchTerm()
  const { querySort, tableSort } =
    useQueryParamSort<TaskTableRowWithHighlights>({ defaultSort })
  const { filters, hasActiveFilters, clearAllFilters } = useTaskAPIFilters()

  const query = useListTasks({
    name: searchTerm,
    filterFunction: Filter_Function.CONTAINS_CASE_INSENSITIVE,
    sort: querySort,
    filters,
  })

  return (
    <Suspense>
      <div className="flex h-full w-full">
        <NavPanelLayout mode="embedded" initialSize="wide">
          <main className="flex h-full w-full flex-col">
            <Header showSearch={true} />
            <div className="flex items-center justify-between gap-2 px-10 pt-6 pb-6">
              <div className="flex flex-col">
                <h1 className="text-xl font-medium">Tasks</h1>
              </div>
              <SearchBar
                placeholder="Search tasks"
                value={searchTermInput ?? undefined}
                onChange={(e) => setSearchTerm(e.target.value)}
              />
            </div>

            <div className="bg-primary flex h-full min-h-0 flex-grow flex-col gap-3">
              <TasksPageContent
                tasksQuery={query}
                searchTerm={searchTerm}
                tableSort={tableSort}
                hasActiveFilters={hasActiveFilters}
                clearAllFilters={clearAllFilters}
              />
            </div>
          </main>
        </NavPanelLayout>
      </div>
    </Suspense>
  )
}
