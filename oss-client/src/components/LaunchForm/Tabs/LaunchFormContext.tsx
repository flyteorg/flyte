import { useCallback, useEffect, useMemo } from 'react'
import { useFormContext } from 'react-hook-form'
import { KeyValueForm } from '@/components/KeyValueSet'
import { TabLayout } from './LaunchFormTabs'
import { KVPair, LaunchFormState } from './types'
import { getKvPairsValidationError } from './validation'
import {
  KeyValuePair,
  KeyValuePairSchema,
} from '@/gen/flyteidl2/core/literals_pb'
import { create } from '@bufbuild/protobuf'

export const LaunchFormContext = () => {
  const { setValue, watch, setError, clearErrors } =
    useFormContext<LaunchFormState>()

  // Convert KeyValuePair[] to KVPair[] for the form
  const contextKeyValuePairs = watch('context')

  // Sync validation to form state so the Context tab shows the error icon
  useEffect(() => {
    const message = getKvPairsValidationError(contextKeyValuePairs ?? [])
    if (message) {
      setError('context', { message, type: 'validation' })
    } else {
      clearErrors('context')
    }
  }, [contextKeyValuePairs, setError, clearErrors])
  const kvPairs = useMemo(
    () =>
      (contextKeyValuePairs || []).map((kv) => ({
        key: kv.key || '',
        value: kv.value || '',
      })),
    [contextKeyValuePairs],
  )

  const updateState = useCallback(
    (kvPairs: KVPair[]) => {
      // Convert KVPair[] back to KeyValuePair[]
      const keyValuePairs: KeyValuePair[] = kvPairs.map((kv) =>
        create(KeyValuePairSchema, {
          key: kv.key,
          value: kv.value,
        }),
      )
      setValue('context', keyValuePairs)
    },
    [setValue],
  )

  return (
    <TabLayout>
      <div className="my-3 flex flex-col">
        <div className="text-sm font-semibold text-(--system-white)">
          Context
        </div>
        <div className="text-xs text-(--system-gray-5)">
          Context values for the run
        </div>
      </div>
      <KeyValueForm state={kvPairs} updateState={updateState} />
    </TabLayout>
  )
}
