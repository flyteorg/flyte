import { useCopyToClipboard } from '@/components/CopyButton'
import { PopoverMenu } from '@/components/Popovers'
import { StatusIcon } from '@/components/StatusIcons/StatusIcon'
import { createColumnHelper } from '@tanstack/react-table'
import { useRouter } from 'next/navigation'
import { type RunsTableRow } from './types'

const helper = createColumnHelper<RunsTableRow>()

const plainRunId = helper.accessor('runId', {
  cell: (info) => (
    <div className="flex min-w-0 items-center gap-1">
      <StatusIcon phase={info.getValue().status} isActive={false} />
      <div className="min-w-0 truncate text-[13px] font-medium text-(--system-gray-7) dark:text-gray-300">
        {info.getValue().id}
      </div>
    </div>
  ),
  header: 'Run',
  minSize: 250,
})

const complexRunId = helper.accessor('runId', {
  cell: (info) => {
    const taskName =
      info.row.original.name?.shortName ||
      info.row.original.name?.fullName ||
      ''
    const runId = info.getValue().id || ''
    return (
      <div className="flex min-w-0 items-center gap-1">
        <StatusIcon phase={info.getValue().status} isActive={false} />
        <div className="flex min-w-0 flex-col">
          <div className="min-w-0 truncate text-[13px] font-medium text-zinc-950 dark:text-white">
            {taskName}
          </div>
          <div className="min-w-0 truncate text-[11px] font-medium text-(--system-gray-5) text-gray-500">
            Run ID: {runId}
          </div>
        </div>
      </div>
    )
  },
  header: 'Run',
  minSize: 250,
})

const environment = helper.accessor('environment', {
  header: '',
  cell: (info) => (
    <div className="min-w-0 truncate text-[13px] font-medium text-(--system-gray-7) dark:text-gray-300">
      {info.getValue() || '-'}
    </div>
  ),
  minSize: 150,
  size: 150,
})

const name = helper.accessor('name', {
  header: 'Run',
  cell: (info) => (
    <div className="flex min-w-0 flex-col">
      <div className="min-w-0 truncate text-sm/5 font-semibold text-zinc-950 dark:text-white">
        {info.getValue().shortName}
      </div>
      <div className="min-w-0 truncate text-xs/4 font-medium text-zinc-400">
        {info.getValue().fullName}
      </div>
    </div>
  ),
  minSize: 350,
})

const runTime = helper.accessor('runTime', {
  header: () => <div className="w-full">Duration</div>,
  cell: (info) => (
    <div className="text-xs/5 font-medium text-gray-500 dark:text-[#CACACA]">
      {info.getValue()}
    </div>
  ),
  minSize: 100,
  size: 100,
})

const startTime = helper.accessor('startTime', {
  header: () => 'Start time',
  cell: (info) => (
    <div className="text-xs/5 font-medium whitespace-nowrap text-gray-500 dark:text-[#CACACA]">
      {info.getValue()}
    </div>
  ),
  minSize: 180,
  size: 180,
})

const endTime = helper.accessor('endTime', {
  header: () => 'End time',
  cell: (info) => (
    <div className="text-xs/5 font-medium whitespace-nowrap text-gray-500 dark:text-[#CACACA]">
      {info.getValue()}
    </div>
  ),
  minSize: 180,
  size: 180,
})

type SetLaunchFormOpen = () => void

const getActions = (setIsLaunchFormOpen: SetLaunchFormOpen) => {
  return helper.accessor('actions', {
    cell: (info) => (
      <ActionsCell
        value={info.getValue()}
        setIsLaunchFormOpen={setIsLaunchFormOpen}
      />
    ),
    header: '',
    minSize: 50,
    size: 50,
  })
}

type ActionsCellProps = {
  value: RunsTableRow['actions']
  setIsLaunchFormOpen: SetLaunchFormOpen
}
const ActionsCell = ({ value, setIsLaunchFormOpen }: ActionsCellProps) => {
  const { handleCopy } = useCopyToClipboard({})
  const router = useRouter()
  return (
    <div
      className="ml-auto w-max"
      onClick={(e) => {
        // prevent navigation
        e.preventDefault()
        e.stopPropagation()
      }}
    >
      <PopoverMenu
        portal
        variant="overflow"
        items={[
          {
            id: 'url',
            label: 'View run details',
            type: 'item',
            onClick: () => router.push(value.url),
          },
          {
            id: 'divider-1',
            type: 'divider',
          },
          {
            id: 'rerun',
            label: 'Rerun task',
            type: 'item',
            onClick: () => setIsLaunchFormOpen(),
          },
          {
            id: 'divider-2',
            type: 'divider',
          },
          {
            id: 'copy-run-name',
            type: 'item',
            label: 'Copy run name',
            onClick: (e) => {
              handleCopy(e, value.runId ?? '')
            },
          },
          {
            id: 'copy-latest-version',
            type: 'item',
            label: 'Copy latest version number',
            onClick: (e) => handleCopy(e, value.latestVersion ?? ''),
          },
          {
            id: 'copy-action-id',
            type: 'item',
            label: 'Copy action ID',
            onClick: (e) => handleCopy(e, value.actionId ?? ''),
          },
        ]}
      />
    </div>
  )
}

const trigger = helper.accessor('trigger', {
  header: () => 'Trigger',
  cell: (info) => {
    const value = info.getValue()
    return (
      <div className="flex min-w-0 flex-col">
        <div className="min-w-0 truncate text-[13px] font-medium text-zinc-950 dark:text-white">
          {value.name || '-'}
        </div>
        {value.type && (
          <div className="min-w-0 truncate text-[11px] font-medium text-gray-500 dark:text-gray-400">
            {value.type}
          </div>
        )}
      </div>
    )
  },
  minSize: 180,
  size: 180,
})

export const listRunsColumns = {
  plainRunId,
  complexRunId,
  environment,
  trigger,
  name,
  runTime,
  startTime,
  endTime,
  getActions,
}
