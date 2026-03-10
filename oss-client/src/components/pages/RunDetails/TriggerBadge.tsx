import { ChevronRightIcon } from '@/components/icons/ChevronRightIcon'
import { TriggersIcon } from '@/components/icons/TriggersIcon'
import {
  Action,
  ActionDetails,
} from '@/gen/flyteidl2/workflow/run_definition_pb'
import { getTriggerTypeString } from '@/lib/triggerUtils'
import Link from 'next/link'

type TriggerBadgeProps = {
  action: Action | ActionDetails | null | undefined
}

export function TriggerBadge({ action }: TriggerBadgeProps) {
  const triggerType = action?.metadata?.triggerType?.type
  const triggerTypeString = getTriggerTypeString(triggerType)
  const triggerName =
    action?.metadata?.triggerName ||
    action?.metadata?.triggerId?.name?.name ||
    undefined
  const taskName =
    action?.metadata?.triggerId?.name?.taskName ||
    (action?.metadata?.spec?.case === 'task'
      ? action?.metadata?.spec?.value?.id?.name
      : undefined)
  const domain = action?.id?.run?.domain
  const project = action?.id?.run?.project

  const triggerDetailsUrl =
    triggerName && taskName && domain && project
      ? `/domain/${domain}/project/${project}/triggers/${taskName}/${triggerName}`
      : undefined

  if (triggerName && triggerDetailsUrl) {
    return (
      <Link
        className="flex max-w-[270px] items-center gap-1 rounded-lg bg-(--system-gray-2) px-2"
        href={triggerDetailsUrl}
      >
        <TriggersIcon className="size-3 text-gray-500" />
        <span className="truncate text-sm font-medium text-zinc-950 dark:text-white">
          {triggerName}
        </span>
        <ChevronRightIcon className="ml-1 text-gray-500" />
      </Link>
    )
  }

  if (triggerTypeString) {
    return (
      <span className="truncate text-sm font-medium text-zinc-950 dark:text-white">
        {triggerTypeString}
      </span>
    )
  }

  return '-'
}
