import { useQueryState } from 'nuqs'
import { useCallback, useMemo } from 'react'

type OwnerId = string

interface OwnerFilter {
  selectedOwners: OwnerId[]
  toggleOwner: (ownerId: OwnerId) => void
  clearFilter: () => void
}

const DELIMITER = ';'

export const useOwnerFilter = (queryKey: string): OwnerFilter => {
  const [ownerValue, setOwnerValue] = useQueryState<string | null>(queryKey, {
    parse: (value) => {
      if (!value) return null
      return value
    },
    serialize: (value) => value || '',
    defaultValue: null,
  })

  const selectedOwners = useMemo(() => {
    if (!ownerValue) return []
    return ownerValue.split(DELIMITER).filter(Boolean)
  }, [ownerValue])

  const toggleOwner = useCallback(
    (ownerId: OwnerId) => {
      setOwnerValue((current) => {
        const currentOwners = current
          ? current.split(DELIMITER).filter(Boolean)
          : []

        const newOwners = currentOwners.includes(ownerId)
          ? currentOwners.filter((id) => id !== ownerId)
          : [...currentOwners, ownerId]

        return newOwners.length > 0
          ? newOwners.join(DELIMITER) + DELIMITER
          : null
      })
    },
    [setOwnerValue],
  )

  const clearFilter = useCallback(() => {
    setOwnerValue(null)
  }, [setOwnerValue])

  return useMemo(
    () => ({
      selectedOwners,
      toggleOwner,
      clearFilter,
    }),
    [selectedOwners, toggleOwner, clearFilter],
  )
}

// Backward compatibility
export const useTriggerOwnerFilter = (): OwnerFilter => {
  return useOwnerFilter('owner')
}
