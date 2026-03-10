import { Button } from '@/components/Button'
import { ClearAllFiltersButton } from '@/components/ClearAllFiltersButton'
import { CopyButton } from '@/components/CopyButton'
import { RelativeTimeFilter } from '@/components/RelativeTimeFilter'
import { TableState } from '@/components/Tables'
import { TriggerAutomationTypeFilter } from '@/components/TriggerAutomationTypeFilter'
import { TriggerStatusFilter } from '@/components/TriggerStatusFilter'
import { Action } from '@/gen/flyteidl2/common/authorization_pb'
import { useIsAuthorized } from '@/hooks/useAuthorize'
import { TableSort } from '@/hooks/useQueryParamSort'
import { useListTriggers } from '@/hooks/useTriggers'
import { getLocation } from '@/lib/windowUtils'
import { ProjectDomainPageParams } from '@/types/pageParams'
import { python } from '@codemirror/lang-python'
import { PlusIcon } from '@heroicons/react/24/outline'
import { vscodeDark, vscodeLight } from '@uiw/codemirror-theme-vscode'
import CodeMirror, { EditorView } from '@uiw/react-codemirror'
import clsx from 'clsx'
import { useTheme } from 'next-themes'
import { useParams } from 'next/navigation'
import { TriggersTable } from '../TriggersTable/Main'
import { TriggerTableRowWithHighlights } from '../TriggersTable/types'

const useCodeSnippets = (taskName: string = 'my_triggered_task') => {
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
      label: 'Then create a simple script called task_trigger.py:',
      code: `import flyte
from datetime import datetime

env = flyte.TaskEnvironment(name="${taskName}")

# trigger every hour
trigger = flyte.Trigger.hourly(trigger_time_input_key="trigger_time")

@env.task(triggers=trigger)
def example_task(trigger_time: datetime, x: int = 1) -> str:
    return f"Task executed at {trigger_time.isoformat()} with x={x}"`,
    },
    {
      id: '3',
      label: 'Then deploy the task environment with the trigger:',
      code: `flyte deploy task_trigger.py env`,
    },
  ]
}

export const ListTriggersPageContent = ({
  triggersQuery,
  searchTerm,
  tableSort,
  className,
  filtersClassName,
  taskName,
  hideColumns,
  hasActiveFilters,
  clearAllFilters,
  setShowCreateTrigger,
}: {
  triggersQuery: ReturnType<typeof useListTriggers>
  searchTerm: string | null
  tableSort?: TableSort<TriggerTableRowWithHighlights>
  className?: string
  filtersClassName?: string
  taskName?: string
  hideColumns?: string[]
  hasActiveFilters: boolean
  clearAllFilters: () => void
  setShowCreateTrigger?: () => void // not all callers allow create trigger
}) => {
  const { resolvedTheme } = useTheme()
  const codeSnippets = useCodeSnippets(taskName)
  const canCreateTrigger = useIsAuthorized({
    action: Action.REGISTER_FLYTE_INVENTORY,
  })
  const shouldShowCreateTrigger = canCreateTrigger && setShowCreateTrigger
  return (
    <div className={clsx('flex h-full min-h-0 flex-grow flex-col', className)}>
      <div
        className={clsx(
          'flex items-center gap-2 overflow-x-auto px-10 pb-6',
          filtersClassName,
        )}
      >
        <div className={'flex w-full items-center justify-between'}>
          <div className="flex flex-nowrap items-center gap-2">
            <TriggerStatusFilter />

            <TriggerAutomationTypeFilter />

            <RelativeTimeFilter label="Last run" queryKey="triggered_at" />
            <RelativeTimeFilter label="Last updated" queryKey="updated_at" />

            {hasActiveFilters && (
              <ClearAllFiltersButton onClick={clearAllFilters} />
            )}
          </div>
          {shouldShowCreateTrigger && (
            <Button color="union" size="xs" onClick={setShowCreateTrigger}>
              <PlusIcon className="!size-3.5" />
              Add trigger
            </Button>
          )}
        </div>
      </div>

      <TableState
        dataLabel="triggers"
        data={triggersQuery.data?.triggers}
        isError={triggersQuery.isError}
        isLoading={triggersQuery.isLoading}
        subtitle="Triggers let you automate task runs on a schedule or event. Create one in the UI or use the example code below to get started."
        searchQuery={searchTerm}
        content={
          <>
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
          <TriggersTable
            triggers={data}
            triggersQuery={triggersQuery}
            searchTerm={searchTerm || ''}
            className={className}
            hideColumns={hideColumns}
            tableSort={tableSort}
          />
        )}
      </TableState>
    </div>
  )
}
