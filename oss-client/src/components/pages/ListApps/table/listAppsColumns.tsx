import { LinkArrow } from '@/components/Link'
import { createColumnHelper } from '@tanstack/react-table'
import { AppStatusBadge } from '../components/AppStatusBadge'
import { ListAppsOverflowActions } from './ListAppsOverflowActions'
import { AppTableItem } from './types'

const helper = createColumnHelper<AppTableItem>()

export const baseColumns = [
  helper.accessor('status', {
    cell: (info) => <AppStatusBadge status={info.getValue()} />,
    header: 'Status',
    minSize: 140,
    size: 140,
  }),

  helper.accessor('replicas', {
    cell: (info) => (
      <span className="truncate overflow-hidden text-sm whitespace-nowrap dark:text-(--system-gray-7)">
        {info.getValue().min} of {info.getValue().max}
      </span>
    ),
    header: 'Replicas',
    minSize: 85,
    size: 85,
  }),

  helper.accessor('name', {
    cell: (info) => (
      <div className="truncate overflow-hidden whitespace-nowrap">
        <div className="truncate overflow-hidden text-sm leading-[16px] font-normal whitespace-nowrap">
          {info.getValue().displayText}
        </div>
        <div className="flex items-center truncate overflow-hidden text-xs leading-[16px] text-nowrap dark:text-(--system-gray-5)">
          Endpoint:{' '}
          <LinkArrow
            displayText={info.getValue().endpoint}
            href={info.getValue().endpoint}
          />
        </div>
      </div>
    ),
    header: 'Name',
    minSize: 205,
  }),
  helper.accessor('type', {
    cell: (info) => (
      <span className="truncate text-sm dark:text-(--system-gray-7)">
        {info.getValue() || '-'}
      </span>
    ),
    minSize: 150,
    size: 150,
    header: 'Type',
  }),
  helper.accessor('lastDeployed', {
    cell: (info) => (
      <div className="truncate text-sm text-(--system-gray-7)">
        {info.getValue().relativeTime}
      </div>
    ),
    header: () => <span className="text-nowrap">Last Deployed</span>,
    minSize: 130,
    size: 130,
  }),

  helper.accessor('actions', {
    cell: (info) => <ListAppsOverflowActions app={info.getValue()} />,
    header: '',
    minSize: 50,
    size: 50,
  }),
]
