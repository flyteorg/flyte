'use client'

import { DatePickerPopover, quickRanges30Days } from '@/components/DatePicker'
import { Header } from '@/components/Header'
import { listRunsColumns, ListRunsContent } from '@/components/ListRuns'
import { ListRunsSearch } from '@/components/ListRuns/filters/ListRunsSearch'
import { NavPanelLayout } from '@/components/NavPanel/NavPanelLayout'
import { Suspense } from 'react'

export function ListRunsPage() {
  const { environment, complexRunId, trigger, runTime, startTime, endTime } =
    listRunsColumns
  return (
    <Suspense>
      <main className="bg-primary flex h-full min-h-0 w-full">
        <NavPanelLayout initialSize="wide" mode="embedded">
          <div className="flex h-full min-h-0 w-full flex-col">
            <Header showSearch={true} />

            <div className="flex items-center justify-between gap-2 px-10 pt-6 pb-3">
              <div className="flex flex-col gap-1">
                <h1 className="text-xl font-medium">Runs</h1>
                <DatePickerPopover
                  labelPrefix="Runs"
                  maxDaysBack={30}
                  quickRanges={quickRanges30Days}
                />
              </div>
              <ListRunsSearch />
            </div>

            <ListRunsContent
              listTableColumns={[
                complexRunId,
                environment,
                trigger,
                runTime,
                startTime,
                endTime,
              ]}
              className="bg-primary min-w-0 flex-1 gap-2 [&:first-child]:px-10 [&:first-child]:pb-6"
              showGroupToggle={true}
            />
          </div>
        </NavPanelLayout>
      </main>
    </Suspense>
  )
}

export default ListRunsPage
