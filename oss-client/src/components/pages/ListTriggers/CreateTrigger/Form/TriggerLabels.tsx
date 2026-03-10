import { useCallback } from 'react'
import { useFormContext } from 'react-hook-form'
import { KeyValueForm } from '@/components/KeyValueSet'
import { CreateTriggerState, KVPair } from './types'
import { HelpText } from './styleHelpers'

export const TriggerLabels = () => {
  const { setValue, watch } = useFormContext<CreateTriggerState>()
  const updateState = useCallback(
    (kvPairs: KVPair[]) => {
      setValue('labels', kvPairs)
    },
    [setValue],
  )

  const state = watch('labels')

  return (
    <div>
      <div className="my-3 flex flex-col">
        <HelpText>
          Optional labels to attach to the trigger, default is None.
        </HelpText>
      </div>
      <KeyValueForm state={state} updateState={updateState} />
    </div>
  )
}
