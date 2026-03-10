import { Switch } from '@/components/Switch'
import { Trigger } from '@/gen/flyteidl2/trigger/trigger_definition_pb'
import { useOrg } from '@/hooks/useOrg'
import { useSearchTerm } from '@/hooks/useQueryParamState'
import {
  useUpdateTriggers,
  type UpdateTriggersProps,
} from '@/hooks/useTriggers'
import { type ProjectDomainPageParams } from '@/types/pageParams'
import { useParams } from 'next/navigation'
import { useCallback, useEffect, useState } from 'react'
import { getTriggerName } from './util'

export const TriggerSwitchCell = ({ trigger }: { trigger: Trigger }) => {
  const [enabled, setEnabled] = useState(trigger.active)
  const { searchTerm } = useSearchTerm()
  const params = useParams<ProjectDomainPageParams>()

  // Sync local state with trigger prop when it changes
  useEffect(() => {
    setEnabled(trigger.active)
  }, [trigger.active])

  const org = useOrg()
  const { mutate } = useUpdateTriggers({
    domain: params.domain,
    projectId: params.project,
    search: searchTerm || '',
    org,
    taskNames: [trigger.id?.name?.taskName || ''],
  })

  const onChangeCallback = useCallback(
    (props: UpdateTriggersProps) => {
      mutate(props)
    },
    [mutate],
  )
  return (
    <div className="text-sm">
      <Switch
        size="sm"
        color="green"
        checked={enabled}
        onChange={() => {
          const newValue = !enabled
          setEnabled(newValue)
          const triggerName = getTriggerName(trigger)
          onChangeCallback({
            active: newValue,
            triggerNames: [triggerName],
          })
        }}
      />
    </div>
  )
}
