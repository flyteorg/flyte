import { useCopyToClipboard } from '@/components/CopyButton'
import { MenuItem, PopoverMenu } from '@/components/Popovers'
import { StatusDot } from '@/components/StatusDotBadge'
import { StatusIcon } from '@/components/StatusIcons/StatusIcon'
import { createColumnHelper } from '@tanstack/react-table'
import { useRouter, useSearchParams } from 'next/navigation'
import { useMemo } from 'react'
import { TaskTableRowWithHighlights } from './types'

export const useTaskTableColumns = () => {
  const searchParams = useSearchParams()
  const router = useRouter()
  const helper = createColumnHelper<TaskTableRowWithHighlights>()
  const { handleCopy } = useCopyToClipboard({})

  return useMemo(() => {
    return [
      helper.accessor('name', {
        cell: (info) => {
          const { fullName, shortDescription } = info.getValue()
          return (
            <div className="flex flex-col truncate">
              <div className="truncate text-sm font-semibold text-zinc-950 dark:text-white">
                {fullName}
              </div>
              {shortDescription && (
                <div className="truncate text-[11px] font-medium text-zinc-950 dark:text-(--system-gray-6)">
                  {shortDescription}
                </div>
              )}
            </div>
          )
        },
        header: 'Name',
        minSize: 450,
        enableSorting: true,
      }),
      helper.accessor('trigger', {
        cell: (info) => (
          <div className="flex items-center gap-2">
            {info.getValue() ? (
              <>
                <StatusDot
                  color={info.getValue()?.active ? 'green' : 'gray'}
                  filled={info.getValue()?.active}
                />
                <div className="flex flex-col">
                  <p className="text-xs/4 font-medium text-(--system-gray-7)">
                    {info.getValue()?.title ?? '-'}
                  </p>
                  {info.getValue()?.subtitle ? (
                    <p className="line-clamp-2 text-[11px] font-medium text-zinc-400 dark:text-(--system-gray-6)">
                      {info.getValue()?.subtitle}
                    </p>
                  ) : null}
                </div>
              </>
            ) : (
              <span className="ml-4">-</span>
            )}
          </div>
        ),
        header: 'Trigger type',
        minSize: 200,
        size: 200,
        enableSorting: false,
      }),
      helper.accessor('lastRun', {
        cell: (info) => {
          const value = info.getValue()
          if (!value) {
            return (
              <span className="text-sm text-zinc-400 dark:text-zinc-500">
                -
              </span>
            )
          }
          return (
            <div className="flex min-w-0 items-center gap-2">
              <StatusIcon phase={value.phase} isActive={false} />
              <div className="flex min-w-0 flex-1 flex-col">
                <div className="truncate text-xs/4 font-medium text-(--system-gray-7)">
                  {value.time}
                </div>
                {value.runId && (
                  <div className="truncate text-[11px] font-medium text-zinc-400 dark:text-(--system-gray-6)">
                    {value.runId}
                  </div>
                )}
              </div>
            </div>
          )
        },
        header: 'Last run',
        size: 250,
        minSize: 250,
        enableSorting: false,
      }),
      helper.accessor('createdAt', {
        cell: (info) => (
          <div className="flex min-w-0 flex-col">
            <div className="truncate text-xs/4 font-medium text-(--system-gray-7)">
              {info.getValue().date}
            </div>
            <div className="truncate text-[11px] font-medium text-zinc-400 dark:text-(--system-gray-6)">
              {info.getValue().version}
            </div>
          </div>
        ),
        header: 'Last deployed',
        size: 200,
        minSize: 200,
        enableSorting: true,
      }),

      helper.accessor('copyAction', {
        cell: (info) => {
          const task = info.getValue()
          const lastRun = task.taskSummary?.latestRun
          const newSearchParams = new URLSearchParams(searchParams)
          newSearchParams.set('launchTab', 'inputs')

          return (
            <div
              className="ml-auto w-max"
              onClick={(e) => {
                // prevent navigation
                e.stopPropagation()
                e.preventDefault()
              }}
            >
              <PopoverMenu
                portal
                overflowProps={{ orientation: 'horizontal' }}
                variant="overflow"
                menuClassName="w-auto data-open:outline-offset-2 data-open:outline-2 data-open:outline-solid data-open:outline-indigo-500"
                items={[
                  {
                    id: 'view-task-details',
                    type: 'item',
                    label: 'View task details',
                    onClick: () =>
                      router.push(
                        `/domain/${task.taskId?.domain}/project/${task.taskId?.project}/tasks/${task.taskId?.name}`,
                      ),
                  },
                  ...(lastRun
                    ? [
                        {
                          id: 'view-run-details',
                          type: 'item',
                          label: 'View last run details',
                          onClick: () =>
                            router.push(
                              `/domain/${lastRun?.runId?.domain}/project/${lastRun?.runId?.project}/runs/${lastRun?.runId?.name}`,
                            ),
                        } as MenuItem,
                      ]
                    : []),
                  {
                    id: 'divider-01',
                    type: 'divider',
                  },
                  {
                    id: 'run-task',
                    type: 'item',
                    label: 'Run task',
                    onClick: () =>
                      router.push(
                        `/domain/${task.taskId?.domain}/project/${task.taskId?.project}/tasks/${task.taskId?.name}/?${newSearchParams.toString()}`,
                      ),
                  },
                  {
                    id: 'divider-02',
                    type: 'divider',
                  },
                  {
                    id: 'copy-task-name',
                    type: 'item',
                    label: 'Copy full task name',
                    onClick: (e) => {
                      handleCopy(e, task.taskId?.name ?? '')
                    },
                  },
                  {
                    id: 'copy-env-name',
                    type: 'item',
                    label: 'Copy environment name',
                    onClick: (e) => {
                      handleCopy(e, task.metadata?.environmentName ?? '')
                    },
                  },
                  {
                    id: 'copy-version',
                    type: 'item',
                    label: 'Copy latest version string',
                    onClick: (e) => handleCopy(e, task.taskId?.version ?? ''),
                  },
                ]}
              />
            </div>
          )
        },
        header: '',
        size: 50,
        minSize: 50,
        enableSorting: false,
      }),
    ]
  }, [handleCopy, helper, router, searchParams])
}
