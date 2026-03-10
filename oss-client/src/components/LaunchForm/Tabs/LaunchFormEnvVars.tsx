import { useCallback, useEffect } from 'react'
import { useFormContext } from 'react-hook-form'
import { KeyValueForm } from '@/components/KeyValueSet'
import { TabLayout } from './LaunchFormTabs'
import { KVPair, LaunchFormState } from './types'
import { getKvPairsValidationError } from './validation'

export const LaunchFormEnvVars = () => {
  const { setValue, watch, setError, clearErrors } =
    useFormContext<LaunchFormState>()
  const envs = watch('envs')

  useEffect(() => {
    const message = getKvPairsValidationError(envs ?? [])
    if (message) {
      setError('envs', { message, type: 'validation' })
    } else {
      clearErrors('envs')
    }
  }, [envs, setError, clearErrors])

  const updateState = useCallback(
    (kvPairs: KVPair[]) => {
      setValue('envs', kvPairs)
    },
    [setValue],
  )

  return (
    <TabLayout>
      <div className="my-3">
        <div className="text-sm font-semibold text-(--system-white)">
          Environment variables
        </div>
      </div>
      <KeyValueForm state={envs} updateState={updateState} />
    </TabLayout>
  )
}
