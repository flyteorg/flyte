import { XMarkIcon } from '@heroicons/react/16/solid'
import { PlusIcon } from '@heroicons/react/20/solid'
import { useCallback, useEffect, useMemo, useState } from 'react'
import { Button } from './Button'
import { ErrorText } from './ErrorText'
import { Input, InputGroup } from './Input'
import { KVPair } from './LaunchForm/Tabs/types'

export const KeyValueForm = ({
  state = [],
  updateState,
}: {
  state?: KVPair[]
  updateState: (kvPairs: KVPair[]) => void
}) => {
  const [kvPairs, setKVPairs] = useState<KVPair[]>([...(state || [])])

  const shouldEnableAddButton = useMemo(() => {
    return kvPairs.every((pair) => pair.key)
  }, [kvPairs])

  // Track duplicate keys
  const duplicateKeys = useMemo(() => {
    const keyCounts = new Map<string, number>()
    kvPairs.forEach((pair) => {
      const trimmedKey = pair.key?.trim()
      if (trimmedKey) {
        keyCounts.set(trimmedKey, (keyCounts.get(trimmedKey) || 0) + 1)
      }
    })
    const duplicates = new Set<string>()
    keyCounts.forEach((count, key) => {
      if (count > 1) {
        duplicates.add(key)
      }
    })
    return duplicates
  }, [kvPairs])

  const updateKVPair = useCallback(
    ({ index, key, value }: UpdateKVPairProps) => {
      const copy = [...kvPairs]
      copy[index] = { key, value }
      setKVPairs(copy)
    },
    [kvPairs],
  )

  const removeKVPair = useCallback(
    (index: number) => {
      const updated = kvPairs.filter((_, i) => i !== index)
      setKVPairs(updated)
    },
    [kvPairs],
  )

  useEffect(() => {
    updateState(kvPairs)
  }, [kvPairs, updateState])

  useEffect(() => {
    if (state.length === 0) {
      setKVPairs([{ key: '', value: '' }])
    }
  }, [state.length])

  return (
    <div className="flex flex-col gap-3">
      {kvPairs.map((kv, i) => (
        <KeyValueInput
          key={i}
          index={i}
          keyInput={kv.key}
          value={kv.value}
          removeKVPair={removeKVPair}
          updateKVPair={updateKVPair}
          isDuplicate={
            kv.key?.trim() ? duplicateKeys.has(kv.key.trim()) : false
          }
        />
      ))}
      <Button
        disabled={!shouldEnableAddButton}
        onClick={() => {
          setKVPairs([...kvPairs, { key: '', value: '' }])
        }}
        className="w-[fit-content]"
        outline
      >
        <PlusIcon /> Add
      </Button>
    </div>
  )
}

type UpdateKVPairProps = { index: number; key: string; value: string }

const KeyValueInput = ({
  index,
  keyInput,
  removeKVPair,
  updateKVPair,
  value,
  isDuplicate,
}: {
  index: number
  keyInput: string
  updateKVPair: (args: UpdateKVPairProps) => void
  removeKVPair: (i: number) => void
  value: string
  isDuplicate: boolean
}) => {
  const [shouldShowError, setShouldShowError] = useState(false)
  const trimmedKey = keyInput?.trim()
  const trimmedValue = value?.trim() ?? ''
  const hasSpaces = trimmedKey ? /\s/.test(trimmedKey) : false
  const showDuplicateError = isDuplicate && trimmedKey !== ''
  const showKeyMissingError =
    (shouldShowError && !trimmedKey) || (!trimmedKey && !!trimmedValue)
  const showSpacesError = hasSpaces && trimmedKey !== ''

  return (
    <div>
      <InputGroup className="flex gap-2">
        <Input
          onBlur={(e) => {
            if (!e.target.value?.trim()) {
              // delay to avoid flashing validation error on delete
              setTimeout(() => {
                setShouldShowError(true)
              }, 50)
            }
          }}
          onChange={(e) => {
            updateKVPair({ index, key: e.target.value, value })
            if (e.target.value?.trim()) {
              setShouldShowError(false)
            }
          }}
          placeholder="Key"
          value={keyInput}
        />
        <Input
          onChange={(e) => {
            updateKVPair({ index, key: keyInput, value: e.target.value })
          }}
          placeholder="Value"
          value={value}
        />
        <Button size="sm" onClick={() => removeKVPair(index)} outline>
          <XMarkIcon className="dark:!text-(--system-white)" />
        </Button>
      </InputGroup>
      {showKeyMissingError && <ErrorText>Key is required</ErrorText>}
      {showSpacesError && <ErrorText>Key cannot contain spaces</ErrorText>}
      {showDuplicateError && !showSpacesError && (
        <ErrorText>Duplicate key: &quot;{trimmedKey}&quot;</ErrorText>
      )}
    </div>
  )
}
