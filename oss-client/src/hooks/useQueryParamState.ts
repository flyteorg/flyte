'use client'

import { useQueryState, parseAsString } from 'nuqs'
import { useCallback, useEffect, useState } from 'react'
import { useDebounce } from 'react-use'
import { getLocation } from '@/lib/windowUtils'

type ValueOf<T> = T[keyof T]

export function useSelectedTab<Enum extends Record<string, string>>(
  defaultValue: ValueOf<Enum>,
  allValues: Enum,
) {
  type V = ValueOf<Enum>
  const allowedValues = Object.values(allValues)

  const [tab, setTab] = useQueryState<V>('tab', {
    defaultValue,
    parse: (value) =>
      allowedValues.includes(value) ? (value as V) : defaultValue,
    serialize: (value) => value,
  })

  const setSelectedTab = useCallback(
    (newTab: V) => {
      setTab(newTab)
    },
    [setTab],
  )

  return { selectedTab: tab, setSelectedTab }
}

export const useSearchTerm = () => {
  const [searchTermDebounced, setSearchTermDebounced] = useQueryState(
    's',
    parseAsString.withDefault('').withOptions({ shallow: false }),
  )
  const location = getLocation()

  // Always use a string (never null) so that consumers passing
  // `value={searchTermInput}` keep inputs controlled at all times.
  // An uncontrolled input (value={undefined}) retains its stale DOM
  // value even after React state updates.
  const [searchTermInput, setSearchTermInput] = useState<string>(
    () => new URLSearchParams(location.search).get('s') ?? '',
  )

  // On mount, re-check the actual URL. During navigation the useState
  // initializer may capture a stale URL (before Next.js updates it).
  // By the time this effect runs the URL reflects the new route, so
  // we can correct the local state.
  useEffect(() => {
    const urlSearchTerm = new URLSearchParams(location.search).get('s')
    if (!urlSearchTerm && searchTermInput) {
      setSearchTermInput('')
      setSearchTermDebounced(null)
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const setSearchTerm = useCallback(
    (newValue: string | null) => {
      setSearchTermInput(newValue || '')
    },
    [setSearchTermInput],
  )

  useDebounce(() => setSearchTermDebounced(searchTermInput || null), 1000, [
    searchTermInput,
  ])

  return {
    searchTermInput,
    searchTerm: searchTermDebounced,
    setSearchTerm,
  }
}

const DELIMITER = ';'

export const useSelectedUsers = () => {
  const [selectedUsers, setSelectedUsers] = useQueryState('user')
  const toggleUser = useCallback(
    (newValue: string | null) => {
      setSelectedUsers((currentSelectedUsers) => {
        if (!newValue) {
          return null
        } else if (!currentSelectedUsers) {
          return `${newValue}${DELIMITER}`
        } else if (currentSelectedUsers.includes(newValue)) {
          return (
            currentSelectedUsers.replace(`${newValue}${DELIMITER}`, '') || null
          )
        }
        return currentSelectedUsers.concat(`${newValue}${DELIMITER}`)
      })
    },
    [setSelectedUsers],
  )

  return {
    selectedUsers,
    toggleUser: toggleUser,
  }
}

export const useShowArchived = () => {
  const [shouldShowArchived, setShouldShowArchived] = useQueryState('archived')

  const toggleShouldShowArchived = useCallback(
    (newState: boolean) => {
      if (newState) {
        setShouldShowArchived('true')
      } else {
        setShouldShowArchived(null)
      }
    },
    [setShouldShowArchived],
  )

  return {
    shouldShowArchived,
    toggleShouldShowArchived,
  }
}
