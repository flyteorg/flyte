import { KeyValueForm } from '@/components/KeyValueSet'
import { useCallback, useEffect } from 'react'
import { useFormContext } from 'react-hook-form'
import { TabLayout } from './LaunchFormTabs'
import { KVPair, LaunchFormState } from './types'
import { getKvPairsValidationError } from './validation'

export const LaunchFormLabels = () => {
  const { setValue, watch, setError, clearErrors } =
    useFormContext<LaunchFormState>()
  const state = watch('labels')

  useEffect(() => {
    const message = getKvPairsValidationError(state ?? [])
    if (message) {
      setError('labels', { message, type: 'validation' })
    } else {
      clearErrors('labels')
    }
  }, [state, setError, clearErrors])

  const updateState = useCallback(
    (kvPairs: KVPair[]) => {
      setValue('labels', kvPairs)
    },
    [setValue],
  )

  return (
    <TabLayout>
      <div className="my-3 flex flex-col">
        <div className="text-sm font-semibold text-(--system-white)">
          Labels
        </div>
        <div className="text-xs text-(--system-gray-5)">
          Labels to apply to the run resource
        </div>
      </div>
      <KeyValueForm state={state} updateState={updateState} />
    </TabLayout>
  )
}
