import { useMemo } from 'react'
import { createColumnHelper } from '@tanstack/react-table'
import { type GroupTableItem } from './types'
import { StatusIcon } from '@/components/StatusIcons'

export const useGroupColumns = ({
  durationDisplayText,
}: {
  durationDisplayText: string
}) => {
  const helper = createColumnHelper<GroupTableItem>()
  return useMemo(() => {
    return [
      helper.accessor('name', {
        cell: (info) => (
          <div className="flex items-center gap-3">
            <StatusIcon iconSize="md" phase={info.getValue().phase} />

            <div>
              <div className="h-4.5 font-semibold dark:text-(--system-gray-5)">
                {info.getValue().taskName}
              </div>
              <div className="font-semibold">
                Action ID: {info.getValue().actionId}
              </div>
            </div>
          </div>
        ),
        header: 'Name',
        size: 450,
      }),
      helper.accessor('duration', {
        cell: (info) => (
          <div className="font-medium dark:text-(--system-gray-5)">
            {info.getValue()}
          </div>
        ),
        header: durationDisplayText,
        size: 130,
      }),
      helper.accessor('startTime', {
        cell: (info) => (
          <div className="font-medium dark:text-(--system-gray-5)">
            {info.getValue()}
          </div>
        ),
        header: 'Start time',
        size: 130,
      }),
    ]
  }, [durationDisplayText, helper])
}
