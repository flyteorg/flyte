'use client'

import { SortDirection, SortingState } from '@tanstack/react-table'
import { Sort, Sort_Direction } from '@/gen/flyteidl2/common/list_pb'
import { camelCase, snakeCase } from 'lodash'
import { useCallback, useEffect, useMemo, useState } from 'react'
import { getLocation, getWindow } from '@/lib/windowUtils'

export const getSortParamForQueryKey = (sortParam: Sort) =>
  `sort:${sortParam.key}=${sortParam.direction}`

export type TableSort<T extends object> = {
  state: SortingState
  onToggleColumnSort: ToggleSortColumnByFn<T>
}
export type QuerySort = {
  sortBy: Sort
  sortForQueryKey: string
}

export type ToggleSortColumnByFn<T> = (sortBy: {
  columnId: keyof T
  direction: SortDirection // always "asc" | "desc" as "enableSortingRemoval" is set to false in VirtualizedTable
}) => void

export function useQueryParamSort<T extends object>({
  defaultSort,
}: {
  defaultSort: Sort
}): {
  querySort: QuerySort
  tableSort: TableSort<T>
} {
  const [sortParam, setSortParam] = useState<Sort>(defaultSort)
  const { href } = getLocation()

  useEffect(() => {
    const url = new URL(href)
    const sortParams = [...url.searchParams.entries()].reduce<
      { key: string; direction: Sort_Direction }[]
    >((acc, [key, value]) => {
      if (value === 'asc' || value === 'desc') {
        acc.push({
          key: snakeCase(key),
          direction:
            value === 'asc'
              ? Sort_Direction.ASCENDING
              : Sort_Direction.DESCENDING,
        })
      }
      return acc
    }, [])
    if (sortParams.length > 0) {
      setSortParam(sortParams[0] as Sort)
    }
  }, [href])

  const toggleSortColumnBy = useCallback<ToggleSortColumnByFn<T>>(
    ({ columnId, direction }) => {
      const w = getWindow()
      if (!w) return
      const url = new URL(href)
      const snakeColumnId = snakeCase(columnId as string)

      // remove existing sort param
      url.searchParams.delete(sortParam.key)

      // set new sort param
      url.searchParams.set(snakeColumnId, direction)
      setSortParam({
        key: snakeColumnId,
        direction:
          direction === 'asc'
            ? Sort_Direction.ASCENDING
            : Sort_Direction.DESCENDING,
      } as Sort)

      w.history.pushState({}, '', url)
    },
    [sortParam, href],
  )

  // sortParams converted to pass to Table
  const tableSort: SortingState = useMemo(
    () => [
      {
        id: camelCase(sortParam.key),
        desc: sortParam.direction === Sort_Direction.DESCENDING,
      },
    ],
    [sortParam],
  )

  return {
    querySort: {
      sortBy: sortParam,
      sortForQueryKey: getSortParamForQueryKey(sortParam),
    },
    tableSort: {
      state: tableSort,
      onToggleColumnSort: toggleSortColumnBy,
    },
  }
}
