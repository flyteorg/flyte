import { useCallback } from 'react'
import { useFormContext } from 'react-hook-form'
import { KeyValueForm } from '@/components/KeyValueSet'
import { KVPair, CreateTriggerState } from './types'
import { HelpText } from './styleHelpers'

export const TriggerAnnotations = () => {
  const { setValue, watch } = useFormContext<CreateTriggerState>()
  const updateState = useCallback(
    (kvPairs: KVPair[]) => {
      setValue('annotations', kvPairs)
    },
    [setValue],
  )

  const annotations = watch('annotations')

  return (
    <div>
      <div className="my-3">
        <HelpText>
          Optional annotations to attach to the trigger, default is None.
        </HelpText>
      </div>
      <KeyValueForm state={annotations} updateState={updateState} />
    </div>
  )
}
