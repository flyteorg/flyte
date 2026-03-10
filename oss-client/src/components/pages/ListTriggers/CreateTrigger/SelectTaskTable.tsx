import { useMemo } from 'react'
import { useListTasks } from '@/hooks/useListTasks'
import { useQueryParamSort } from '@/hooks/useQueryParamSort'
import {
  Filter,
  Filter_Function,
  Sort,
  Sort_Direction,
} from '@/gen/flyteidl2/common/list_pb'
import { Task } from '@/gen/flyteidl2/task/task_definition_pb'
import { VirtualizedTable } from '@/components/Tables'
import { TableFooterLoadMore } from '@/components/Tables'
import { timestampToMillis } from '@/lib/dateUtils'
import { useIdentity } from '@/hooks/useIdentity'
import { useColumns } from './useSelectTaskTableColumns'
import { TableTask, TaskDetails } from './types'

const transformTask = (task: Task): TableTask => ({
  name: {
    shortName: task.metadata?.shortName,
    fullName: task.taskId?.name,
    shortDescription: task.metadata?.shortDescription,
  },
  lastDeployed: {
    version: task.taskId?.version,
    date: timestampToMillis(task.metadata?.deployedAt),
  },
  createdUser: task.metadata?.deployedBy,
  actions: task,
})

const defaultSort: Sort = {
  key: 'created_at',
  direction: Sort_Direction.DESCENDING,
} as Sort

export const TasksTable = ({
  onSelectTask,
  searchTerm,
  shouldFilterUserTasks,
}: {
  onSelectTask: (taskDetails: TaskDetails) => void
  searchTerm: string
  shouldFilterUserTasks: boolean
}) => {
  const { querySort, tableSort } = useQueryParamSort<TableTask>({ defaultSort })
  const user = useIdentity()

  const filters = useMemo(() => {
    if (!shouldFilterUserTasks) return []
    if (!user.data?.subject) return []
    return [
      {
        function: Filter_Function.VALUE_IN,
        field: 'deployed_by',
        values: [user.data.subject],
      } as Filter,
    ]
  }, [shouldFilterUserTasks, user.data?.subject])

  const query = useListTasks({
    filters: filters,
    name: searchTerm,
    limit: 100,
    sort: querySort,
  })

  const data = useMemo(() => {
    if (query.data?.pages.length) {
      const tasks = query.data.pages.map((res) => res.tasks).flat()
      const formatted = tasks.map(transformTask)
      return formatted
    }
    return []
  }, [query.data?.pages])

  const columns = useColumns()

  return (
    <VirtualizedTable
      columns={columns}
      data={data}
      footerRow={<TableFooterLoadMore query={query} label="tasks" />}
      onRowClick={({ actions: { taskId } }) => {
        if (taskId?.name && taskId.version) {
          onSelectTask({ taskId: taskId.name, taskVersion: taskId.version })
        }
      }}
      rowHeight={60}
      overscan={10}
      sorting={tableSort}
    />
  )
}
