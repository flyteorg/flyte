import { VirtualizedTable } from '@/components/Tables'
import { TableFooterLoadMore } from '@/components/Tables'
import { Task } from '@/gen/flyteidl2/task/task_definition_pb'
import { useListTasks } from '@/hooks/useListTasks'
import { TableSort } from '@/hooks/useQueryParamSort'
import { useMemo } from 'react'
import { useTaskTableColumns } from './getTaskTableColumns'
import { type TaskTableRowWithHighlights } from './types'
import { formatForTable, formatHighlights } from './util'

export const TasksTable = ({
  tasks,
  tasksQuery,
  searchTerm,
  tableSort,
}: {
  tasks: Task[]
  searchTerm?: string
  tasksQuery?: ReturnType<typeof useListTasks>
  tableSort: TableSort<TaskTableRowWithHighlights>
}) => {
  const columns = useTaskTableColumns()

  const formattedTasks = useMemo(() => {
    return formatHighlights(
      tasks?.map((t) => formatForTable(t)),
      searchTerm,
    )
  }, [tasks, searchTerm])

  return (
    <VirtualizedTable<TaskTableRowWithHighlights>
      columns={columns}
      data={formattedTasks || []}
      footerRow={<TableFooterLoadMore query={tasksQuery} label="tasks" />}
      rowHeight={60}
      overscan={10}
      getRowHref={(taskRow) =>
        `/domain/${taskRow.original.taskId?.domain}/project/${taskRow.original.taskId?.project}/tasks/${taskRow.original.taskId?.name}`
      }
      sorting={tableSort}
    />
  )
}
