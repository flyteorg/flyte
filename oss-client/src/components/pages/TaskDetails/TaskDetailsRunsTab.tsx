import { DatePickerPopover } from '@/components/DatePicker'
import { listRunsColumns, ListRunsContent } from '@/components/ListRuns'
import { Filter_Function } from '@/gen/flyteidl2/common/list_pb'
import { useLaunchFormState } from '@/hooks/useLaunchFormState'
import { getFilter } from '@/lib/filterUtils'
import { useParams } from 'next/navigation'
import React, { useMemo } from 'react'
import { TaskDetailsPageParams } from './types'

export const TaskDetailsRunsTab: React.FC<{ version: string }> = ({
  version,
}) => {
  const params = useParams<TaskDetailsPageParams>()

  const filters = useMemo(() => {
    const list = [
      getFilter({
        function: Filter_Function.EQUAL,
        field: 'task_name',
        values: [params.name],
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
  }, [params.version, params.name])

  const { plainRunId, runTime, startTime, endTime, getActions, trigger } =
    listRunsColumns

  const { setIsOpen } = useLaunchFormState()

  const noRowsMessage = version
    ? 'This task version does not have any runs'
    : 'Get started by triggering a run with flyte from the CLI'

  return (
    <div className="flex max-h-full min-h-0 w-full min-w-0 flex-1 flex-col gap-2 px-8 pb-8">
      <ListRunsContent
        datePickerPopover={<DatePickerPopover labelPrefix="Runs" />}
        additionalFilters={filters}
        listTableColumns={[
          plainRunId,
          trigger,
          runTime,
          startTime,
          endTime,
          getActions(() => setIsOpen(true)),
        ]}
        className="overflow-hidden rounded-lg border border-(--system-gray-3) bg-(--system-black) [&>*:first-child]:px-6 [&>*:first-child]:py-2.5"
        noRowsMessage={noRowsMessage}
        hideLastRowBorder
      />
    </div>
  )
}
