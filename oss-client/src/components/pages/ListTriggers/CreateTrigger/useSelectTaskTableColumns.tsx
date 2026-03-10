'use client'

// import { useRouter } from 'next/navigation'
// import { useCallback, useMemo } from 'react'
import { toDateFormat } from '@/lib/dateUtils'
import { createColumnHelper } from '@tanstack/react-table'
import { type TableTask } from './types'
// import { PopoverMenu } from '@/components/Popovers'
// import { Task } from '@/gen/flyteidl2/task/task_definition_pb'

const helper = createColumnHelper<TableTask>()

export const useColumns = () => {
  return [
    helper.accessor('name', {
      cell: (info) => {
        const { shortName, shortDescription } = info.getValue()
        return (
          <div className="flex flex-col truncate">
            <div className="truncate text-sm font-semibold text-zinc-950 dark:text-white">
              {shortName}
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
      minSize: 500,
      enableSorting: false,
    }),
    helper.accessor('lastDeployed', {
      cell: (info) => (
        <div className="flex min-w-0 flex-col">
          <div className="truncate text-xs/4 font-medium text-(--system-gray-7)">
            {toDateFormat({ timestamp: info.getValue().date })}
          </div>
          <div className="truncate text-[11px] font-medium text-zinc-400 dark:text-(--system-gray-6)">
            {info.getValue().version}
          </div>
        </div>
      ),
      header: 'Last deployed',
      size: 50,
      enableSorting: true,
    }),
  ]
}
