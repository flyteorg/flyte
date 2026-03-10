import { Button } from '@/components/Button'
import { ClearAllFiltersButton } from '@/components/ClearAllFiltersButton'
import { CopyButton } from '@/components/CopyButton'
import { RelativeTimeFilter } from '@/components/RelativeTimeFilter'
import { TableState } from '@/components/Tables'
import { TriggerAutomationTypeFilter } from '@/components/TriggerAutomationTypeFilter'
import { ListTasksResponse } from '@/gen/flyteidl2/task/task_service_pb'
import { useListTasks } from '@/hooks/useListTasks'
import { TableSort } from '@/hooks/useQueryParamSort'
import { getLocation } from '@/lib/windowUtils'
import { ProjectDomainPageParams } from '@/types/pageParams'
import { python } from '@codemirror/lang-python'
import { ArrowUpRightIcon } from '@heroicons/react/24/outline'
import { vscodeDark, vscodeLight } from '@uiw/codemirror-theme-vscode'
import CodeMirror, { EditorView } from '@uiw/react-codemirror'
import { useTheme } from 'next-themes'
import { useParams } from 'next/navigation'
import { TasksTable } from '../TasksTable/Main'
import { TaskTableRowWithHighlights } from '../TasksTable/types'

const useCodeSnippets = () => {
  const params = useParams<ProjectDomainPageParams>()
  const { hostname } = getLocation()
  return [
    {
      id: '1',
      label:
        'If you don’t have a local flyte configuration, first create it with:',
      code: `flyte create config \\
    --endpoint ${hostname} \\
    --project ${params.project} \\
    --domain ${params.domain} \\
    --builder remote`,
    },
    {
      id: '2',
      label: 'Then create a simple script called hello_world.py:',
      code: `import flyte

env = flyte.TaskEnvironment(name="hello_world")

@env.task
def fn(x: int) -> int:
    slope, intercept = 2, 5
    return slope * x + intercept

@env.task
def main(n: int) -> float:
    y_list = list(flyte.map(fn, range(n)))
    return sum(y_list) / len(y_list)`,
    },
    {
      id: '3',
      label: 'Then deploy the task environment:',
      code: `flyte deploy hello_world.py env`,
    },
  ]
}

export const TasksPageContent = ({
  tasksQuery,
  searchTerm,
  tableSort,
  hasActiveFilters,
  clearAllFilters,
}: {
  tasksQuery: ReturnType<typeof useListTasks>
  searchTerm?: string
  tableSort: TableSort<TaskTableRowWithHighlights>
  hasActiveFilters: boolean
  clearAllFilters: () => void
}) => {
  const { resolvedTheme } = useTheme()
  const codeSnippets = useCodeSnippets()

  const allTasks =
    tasksQuery.data?.pages?.flatMap(
      (page: ListTasksResponse) => page.tasks ?? [],
    ) ?? []

  return (
    <div className="flex h-full min-h-0 flex-grow flex-col">
      <div className="flex items-center gap-2 overflow-x-auto px-10 pb-6">
        <div className={'flex shrink-0 flex-nowrap items-center gap-2'}>
          <RelativeTimeFilter label="Last updated" queryKey="updated_at" />

          <TriggerAutomationTypeFilter />

          {hasActiveFilters && (
            <ClearAllFiltersButton onClick={clearAllFilters} />
          )}
        </div>
      </div>

      <TableState
        dataLabel="deployed tasks"
        data={allTasks}
        isError={tasksQuery.isError}
        isLoading={tasksQuery.isLoading}
        subtitle="You must deploy a task environment for its tasks to appear here"
        searchQuery={searchTerm}
        content={
          <>
            <Button
              plain
              href="https://www.union.ai/docs/v2/byoc/api-reference/flyte-cli/#flyte-deploy"
              target="_blank"
            >
              <span className="text-2xs">How to deploy a task environment</span>
              <ArrowUpRightIcon
                data-slot="icon"
                className="!size-3 dark:!text-[#828282]"
              />
            </Button>
            <div className="max-w-[600px] min-w-[550px]">
              {codeSnippets.map(({ code, id, label }) => (
                <div key={id} className="mt-6 w-full">
                  <p className="text-sm font-bold">{label}</p>
                  <div className="relative mt-2 w-full text-[11px] [&_.cm-editor]:!bg-transparent [&_.cm-focused]:!outline-none [&_.cm-gutters]:!bg-transparent [&_.cm-scroller]:!rounded-2xl [&_.cm-scroller]:!border [&_.cm-scroller]:!border-(--system-white)/14 [&_.cm-scroller>:where(.cm-content)]:!p-5">
                    <div className="pointer-events-auto absolute top-3 right-3 z-20">
                      <CopyButton value={code} />
                    </div>
                    <CodeMirror
                      readOnly
                      editable={false}
                      theme={
                        resolvedTheme === 'dark' ? vscodeDark : vscodeLight
                      }
                      extensions={[python(), EditorView.lineWrapping]}
                      basicSetup={{
                        lineNumbers: false,
                        foldGutter: false,
                        highlightActiveLine: false,
                        highlightActiveLineGutter: false,
                      }}
                      value={code}
                    />
                  </div>
                </div>
              ))}
            </div>
          </>
        }
      >
        {(data) => (
          <TasksTable
            tasks={data}
            tasksQuery={tasksQuery}
            searchTerm={searchTerm}
            tableSort={tableSort}
          />
        )}
      </TableState>
    </div>
  )
}
