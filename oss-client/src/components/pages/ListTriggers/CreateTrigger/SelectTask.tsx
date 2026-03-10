import { useState } from 'react'
import { SearchBar } from '@/components/SearchBar'
import { Checkbox } from '@/components/Checkbox'
import { TasksTable } from './SelectTaskTable'
import { TaskDetails } from './types'
import { useDebounce } from 'react-use'

export const SelectTask = ({
  onSelectTask,
}: {
  onSelectTask: (taskDetails: TaskDetails) => void
}) => {
  const [searchTermInput, setSearchTermInput] = useState('')
  const [debouncedSearchTerm, setDebouncedSearchTerm] = useState('')
  const [shouldFilterUserTasks, setShouldFilterTasks] = useState(false)

  useDebounce(() => setDebouncedSearchTerm(searchTermInput), 200, [
    searchTermInput,
  ])

  return (
    <>
      <div className="p-4">
        <h2 className="text-sm font-bold">Select task</h2>
        <p className="text-2xs leading-tight font-medium dark:text-(--system-gray-5)">
          Select the task to receive the trigger. Triggers will always attach to
          the latest version of the task. Please note, triggers created through
          the UI will be deleted on new deploys of the task. To create a durable
          trigger you must define it within your task code.
        </p>
        <div className="mt-5 flex justify-between pr-4">
          <SearchBar
            onChange={(e) => setSearchTermInput(e.target.value)}
            placeholder="Search tasks & Environment"
            value={searchTermInput ?? undefined}
          />
          <div className="flex items-center gap-2 text-2xs font-medium dark:text-(--system-gray-5)">
            <Checkbox
              checked={shouldFilterUserTasks}
              onClick={() => {
                setShouldFilterTasks(!shouldFilterUserTasks)
              }}
            />
            Only my tasks
          </div>
        </div>
      </div>
      <TasksTable
        onSelectTask={onSelectTask}
        searchTerm={debouncedSearchTerm}
        shouldFilterUserTasks={shouldFilterUserTasks}
      />
    </>
  )
}
