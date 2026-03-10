// import { PopoverMenu } from '@/components/Popovers'
import {
  BulkSelectionColumnProps,
  createBulkSelectionColumn,
} from '@/components/Tables/BulkSelectionColumn'
import { createColumnHelper } from '@tanstack/react-table'
import Link from 'next/link'
import { useMemo } from 'react'
import { TriggerSwitchCell } from './TriggersSwitchCell'
import { TriggerTableRowWithHighlights } from './types'

export const useTriggersTableColumns = (
  bulkSelectionProps?: BulkSelectionColumnProps<TriggerTableRowWithHighlights>['bulkSelection'],
  hideColumns?: string[],
) => {
  const helper = createColumnHelper<TriggerTableRowWithHighlights>()

  return useMemo(() => {
    const columns = []

    // Add bulk selection column if bulk selection props are provided
    if (bulkSelectionProps) {
      columns.push(createBulkSelectionColumn(bulkSelectionProps))
    }

    // Add the existing columns
    columns.push(
      helper.accessor('active', {
        cell: (info) => (
          <TriggerSwitchCell trigger={info.row.original.actions} />
        ),
        header: 'Status',
        minSize: 120,
        size: 120,
      }),
      helper.accessor('name', {
        cell: (info) => {
          const nextRun = info.getValue().nextRun
          return (
            <div className="truncate">
              <div className="truncate text-sm font-semibold text-zinc-500 dark:text-zinc-400">
                {info.getValue().name}
              </div>
              <div className="truncate text-sm font-medium text-zinc-950 dark:text-white">
                {info.getValue().schedule}
              </div>
              {nextRun && (
                <span className="block truncate text-[11px] text-zinc-500 dark:text-zinc-400">
                  Next Run: {info.getValue().nextRun}
                </span>
              )}
            </div>
          )
        },
        header: 'Trigger',
        minSize: 350,
      }),
      helper.accessor('task', {
        cell: (info) => (
          <Link
            onClick={(e: React.MouseEvent) => {
              // prevent navigation
              e.stopPropagation()
            }}
            className="truncate rounded-md pr-1 pl-1 hover:bg-(--system-gray-2) hover:dark:bg-(--system-black)"
            href={`/domain/${info.getValue()?.domain}/project/${info.getValue()?.project}/tasks/${info.getValue()?.taskName}`}
          >
            <span className="block truncate text-sm font-medium text-(--system-white)">
              {info.getValue()?.taskName}
            </span>
          </Link>
        ),
        header: 'Associated Task',
        minSize: 250,
      }),
      helper.accessor('triggered', {
        cell: (info) => (
          <div className="text-sm font-medium text-(--system-white)">
            {info.getValue().date}
          </div>
        ),
        header: 'Last run',
        minSize: 200,
        size: 200,
      }),
      helper.accessor('updated', {
        cell: (info) => {
          const principal = info.getValue().updatedBy?.principal
          const isUser = principal?.case === 'user'
          return (
            <div className="flex flex-col font-medium text-zinc-500 dark:text-zinc-400">
              <span className="text-xs">{info.getValue().date}</span>
              <span className="text-[11px]">
                {isUser
                  ? `by ${principal.value?.spec?.firstName} ${principal.value?.spec?.lastName}`
                  : ''}
              </span>
            </div>
          )
        },
        header: 'Last updated',
        minSize: 200,
        size: 200,
      }),
    )

    const filteredColumns = columns.filter((column) => {
      // Filter out columns based on hideColumns using accessor name (id/accessorKey)
      if (hideColumns && hideColumns.length > 0) {
        let accessor = ''
        const colWithAccessorKey = column as { accessorKey?: unknown }
        if (typeof colWithAccessorKey.accessorKey === 'string') {
          accessor = colWithAccessorKey.accessorKey
        } else {
          const colWithId = column as { id?: unknown }
          if (typeof colWithId.id === 'string') {
            accessor = colWithId.id
          }
        }
        return !hideColumns.includes(accessor)
      }
      return true
    })

    return filteredColumns
  }, [helper, bulkSelectionProps, hideColumns])
}
