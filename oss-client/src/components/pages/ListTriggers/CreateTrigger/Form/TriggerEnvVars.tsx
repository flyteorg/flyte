import { useCallback } from 'react'
import { useFormContext } from 'react-hook-form'
import { KeyValueForm } from '@/components/KeyValueSet'
import { KVPair, CreateTriggerState } from './types'
import { HelpText } from './styleHelpers'

export const TriggerEnvVars = () => {
  const { setValue, watch } = useFormContext<CreateTriggerState>()
  const updateState = useCallback(
    (kvPairs: KVPair[]) => {
      setValue('envVars', kvPairs)
    },
    [setValue],
  )

  const envs = watch('envVars')

  return (
    <div>
      <div className="my-3">
        <HelpText>
          Optional environment variables for the trigger, default is None.
        </HelpText>
        <HelpText>
          If provided, will replace the environment variables set in the config
          of the task.
        </HelpText>
      </div>
      <KeyValueForm state={envs} updateState={updateState} />
    </div>
  )
}
