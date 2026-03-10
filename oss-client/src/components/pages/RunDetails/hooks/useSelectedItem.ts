import { useCallback, useMemo } from 'react'
import { useQueryState } from 'nuqs'
import { Action } from '@/gen/flyteidl2/workflow/run_definition_pb'
import { useRunStore } from '../state/RunStore'
import { GROUP_SEPARATOR } from '../state/flattenTree'

type SelectedItem = {
  type: 'action' | 'group'
  id: string
}

type UseSelectedItemReturn = {
  selectedItem: SelectedItem | null
  setSelectedItem: (id: string) => void
}

const getDefaultValue = (action: Action | undefined): SelectedItem | null => {
  if (!action || !action?.id?.name) {
    return null
  }
  return {
    type: 'action',
    id: action.id.name,
  }
}

/*
 * Hook for tracking the currently selected item (action or group).
 * Stores the selection in query params (?i=...) so it survives reloads / share links.
 */
export const useSelectedItem = (): UseSelectedItemReturn => {
  const [params, setParams] = useQueryState('i')
  const run = useRunStore((s) => s.run?.action)
  const actions = useRunStore((s) => s.actions)

  const setSelectedItem = useCallback(
    (uniqueId: string) => {
      const encoded = encodeURIComponent(uniqueId)
      setParams((prev) => (encoded === prev ? prev : encoded))
    },
    [setParams],
  )

  const selectedItem = useMemo<SelectedItem | null>(() => {
    if (!params) return getDefaultValue(run)

    const decoded = decodeURIComponent(params)
    const isGroup = decoded.includes(GROUP_SEPARATOR)

    // Return default (null) if:
    // 1. No param is present in queryparams
    // 2. Not a group
    // 3. Actions have been loaded (length > 0)
    // 4. The requested action doesn't exist in the store

    if (
      !decoded &&
      !isGroup &&
      Object.keys(actions).length > 0 &&
      !actions[decoded]
    ) {
      return getDefaultValue(run)
    }

    return {
      type: isGroup ? 'group' : 'action',
      id: decoded,
    }
  }, [params, actions, run])

  return {
    selectedItem,
    setSelectedItem,
  }
}

export const useSelectedActionId = () => {
  const { selectedItem } = useSelectedItem()
  return selectedItem?.type === 'action' ? selectedItem.id : null
}

export const useSelectedGroup = () => {
  const { selectedItem } = useSelectedItem()
  if (selectedItem?.type !== 'group') return null
  const [groupName] = selectedItem.id.split('::').slice(-1)
  return {
    id: selectedItem.id,
    name: groupName,
  }
}
