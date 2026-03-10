import { AppsIcon } from '../icons/AppsIcon'
import { RunsIcon } from '../icons/RunsIcon'
import { StatusIcon } from '../StatusIcons'
import { ShareIcon as TaskIcon } from '@heroicons/react/24/outline'
import { ActionPhase } from '@/gen/flyteidl2/common/phase_pb'
import { Status_DeploymentStatus } from '@/gen/flyteidl2/app/app_definition_pb'
import { AppStatusBadge } from '../pages/ListApps/components/AppStatusBadge'
import { formatSearchResultDate } from '@/lib/dateUtils'

const primaryText = 'text-[13px] font-semibold text-white'
const subText = 'text-[11px] font-medium text-(--system-gray-5)'

const DateWidget = ({ sortByDate }: { sortByDate: number | undefined }) => (
  <span className={`${subText} shrink-0`}>
    {formatSearchResultDate(sortByDate)}
  </span>
)

export const AppResult = ({
  name,
  sortByDate,
  status,
}: {
  isLocalStorage?: boolean
  name: string | undefined
  status: Status_DeploymentStatus
  sortByDate: number | undefined
}) => {
  if (!name) return null
  return (
    <div className="flex justify-between gap-4">
      <div className="flex min-w-0 items-center gap-2">
        <AppsIcon className="size-4 shrink-0 text-(--system-gray-5)" />
        <span className={`${primaryText} truncate`}>{name}</span>
        <AppStatusBadge className="shrink-0" status={status} />
      </div>
      <DateWidget sortByDate={sortByDate} />
    </div>
  )
}

export const RunResult = ({
  env,
  phase,
  runFnName,
  runId,
  sortByDate,
}: {
  env: string | undefined
  isLocalStorage?: boolean
  phase: ActionPhase | undefined
  runId: string | undefined
  runFnName: string | undefined
  sortByDate: number | undefined
}) => {
  if (!runFnName || !runId) return null
  return (
    <div className="flex justify-between gap-4">
      <div className="flex min-w-0 items-center gap-2">
        <RunsIcon className="size-4 shrink-0 text-(--system-gray-5)" />
        <span className={`${primaryText} text-nowrap`}>{runId}</span>
        <StatusIcon iconSize="md" phase={phase} />
        <span className={`${subText} truncate`}>
          {env}.{runFnName}
        </span>
      </div>
      <DateWidget sortByDate={sortByDate} />
    </div>
  )
}

export const TaskResult = ({
  environmentName,
  name,
  shortName,
  sortByDate,
  version,
}: {
  environmentName?: string
  name: string | undefined
  shortName?: string
  sortByDate: number | undefined
  version: string | undefined
}) => {
  if (!name) return null
  return (
    <div className="flex justify-between gap-4">
      <div className="flex min-w-0 items-center gap-2">
        <TaskIcon className="size-4 shrink-0 text-(--system-gray-5)" />
        <span className={`${primaryText} truncate`}>
          {environmentName}.{shortName || name}
        </span>
        <span className={`${subText} shrink-0`}>{version}</span>
      </div>
      <DateWidget sortByDate={sortByDate} />
    </div>
  )
}
