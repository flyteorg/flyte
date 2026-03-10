import { useState, useCallback, useMemo } from 'react'

export function useBulkSelection<T>(
  items: T[],
  getId: (item: T) => string,
  onChange?: (selected: T[]) => void,
) {
  const [selectedIds, setSelectedIds] = useState<Set<string>>(new Set())

  const isSelected = useCallback(
    (item: T) => selectedIds.has(getId(item)),
    [selectedIds, getId],
  )

  const toggleItem = useCallback(
    (item: T) => {
      const id = getId(item)
      setSelectedIds((prev) => {
        const next = new Set(prev)
        if (next.has(id)) {
          next.delete(id)
        } else {
          next.add(id)
        }
        const selected = items.filter((i) => next.has(getId(i)))
        onChange?.(selected)
        return next
      })
    },
    [getId, items, onChange],
  )

  const selectAll = useCallback(() => {
    const all = new Set(items.map(getId))
    setSelectedIds(all)
    onChange?.(items)
  }, [getId, items, onChange])

  const unselectAll = useCallback(() => {
    setSelectedIds(new Set())
  }, [])

  const clearSelection = useCallback(() => {
    setSelectedIds(new Set())
    onChange?.([])
  }, [onChange])

  const selectedItems = useMemo(
    () => items.filter((i) => selectedIds.has(getId(i))),
    [items, selectedIds, getId],
  )

  const count = selectedIds.size

  const areAllItemsSelected = useMemo(() => {
    return items.length === selectedIds.size
  }, [items, selectedIds])

  return {
    areAllItemsSelected,
    selectedItems,
    isSelected,
    toggleItem,
    selectAll,
    clearSelection,
    count,
    unselectAll,
  }
}
