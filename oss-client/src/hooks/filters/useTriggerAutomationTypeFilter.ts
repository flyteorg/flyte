import { useQueryState } from 'nuqs'
import { useCallback, useMemo } from 'react'

type AutomationType = 'TYPE_SCHEDULE' | 'TYPE_NONE'

interface TriggerAutomationTypeFilter {
  selectedTypes: AutomationType[]
  toggleType: (type: AutomationType) => void
  clearFilter: () => void
}

const DELIMITER = ';'

export const useTriggerAutomationTypeFilter =
  (): TriggerAutomationTypeFilter => {
    const [typeValue, setTypeValue] = useQueryState<string | null>('type', {
      parse: (value) => {
        if (!value) return null
        return value
      },
      serialize: (value) => value || '',
      defaultValue: null,
    })

    const selectedTypes = useMemo(() => {
      if (!typeValue) return []
      return typeValue
        .split(DELIMITER)
        .filter(
          (v) => v === 'TYPE_SCHEDULE' || v === 'TYPE_NONE',
        ) as AutomationType[]
    }, [typeValue])

    const toggleType = useCallback(
      (type: AutomationType) => {
        setTypeValue((current) => {
          const currentTypes = current
            ? current.split(DELIMITER).filter(Boolean)
            : []

          const newTypes = currentTypes.includes(type)
            ? currentTypes.filter((t) => t !== type)
            : [...currentTypes, type]

          return newTypes.length > 0
            ? newTypes.join(DELIMITER) + DELIMITER
            : null
        })
      },
      [setTypeValue],
    )

    const clearFilter = useCallback(() => {
      setTypeValue(null)
    }, [setTypeValue])

    return useMemo(
      () => ({
        selectedTypes,
        toggleType,
        clearFilter,
      }),
      [selectedTypes, toggleType, clearFilter],
    )
  }
