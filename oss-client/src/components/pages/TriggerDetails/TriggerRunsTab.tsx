import { DatePickerPopover } from '@/components/DatePicker'
import { listRunsColumns, ListRunsContent } from '@/components/ListRuns'
import { Filter_Function } from '@/gen/flyteidl2/common/list_pb'
import { getFilter } from '@/lib/filterUtils'
import { useParams } from 'next/navigation'
import React, { useMemo } from 'react'
import { TriggerDetailsPageParams } from './types'

export const TriggerRunsTab: React.FC = () => {
  const params = useParams<TriggerDetailsPageParams>()

  const decodedName = useMemo(
    () => (params.name ? decodeURIComponent(params.name) : ''),
    [params.name],
  )

  const filters = useMemo(() => {
    const list = [
      getFilter({
        function: Filter_Function.EQUAL,
        field: 'task_name',
        values: [params.taskName],
      }),
      getFilter({
        function: Filter_Function.EQUAL,
        field: 'trigger_name',
        values: [decodedName],
      }),
    ]
    if (params.version) {
      list.push(
        getFilter({
          function: Filter_Function.EQUAL,
          field: 'task_version',
          values: [params.version],
        }),
      )
    }
    return list
  }, [params.version, decodedName, params.taskName])

  const { plainRunId, runTime, startTime, endTime } = listRunsColumns

  return (
    <>
      <div className="flex max-h-full min-h-0 w-full min-w-0 flex-1 flex-col gap-2 px-8 pb-8">
        <ListRunsContent
          additionalFilters={filters}
          datePickerPopover={<DatePickerPopover labelPrefix="Runs" />}
          listTableColumns={[plainRunId, runTime, startTime, endTime]}
          className="overflow-hidden rounded-lg border border-(--system-gray-3) bg-(--system-black) [&>*:first-child]:px-6 [&>*:first-child]:py-2.5"
          noRowsMessage="This trigger has not created any runs"
          hideLastRowBorder
        />
      </div>
    </>
  )
}
