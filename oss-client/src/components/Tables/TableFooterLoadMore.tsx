import { InfiniteData, UseInfiniteQueryResult } from '@tanstack/react-query'
import { useCallback, useEffect, useRef } from 'react'
import { useInView } from 'react-intersection-observer'

type TableFooterLoadMoreProps<T> = {
  label: string
  query?: UseInfiniteQueryResult<InfiniteData<T, unknown>, Error>
}

export function TableFooterLoadMore<T>({
  label,
  query,
}: TableFooterLoadMoreProps<T>) {
  // Track if we're already requesting the next page
  const isRequestingRef = useRef(false)

  // Intersection observer for infinite scrolling
  const { ref, inView } = useInView({
    threshold: 0,
    rootMargin: '100px',
  })

  // Memoize the fetch function to prevent unnecessary re-renders
  const fetchNextPage = useCallback(() => {
    if (
      query?.hasNextPage &&
      !query?.isFetchingNextPage &&
      !isRequestingRef.current
    ) {
      // Mark that we're requesting
      isRequestingRef.current = true

      // Fetch the next page
      query.fetchNextPage().finally(() => {
        // Mark that we're done requesting
        isRequestingRef.current = false
      })
    }
  }, [query])

  // Fetch next page when the invisible marker comes into view
  useEffect(() => {
    if (inView) {
      fetchNextPage()
    }
  }, [inView, fetchNextPage])

  return query?.hasNextPage ? (
    <div ref={ref} className="h-4 w-full" aria-label={`Load more ${label}`}>
      {query.isFetchingNextPage && (
        <div className="flex justify-center py-2">
          <div className="text-sm text-gray-500">Loading more {label}...</div>
        </div>
      )}
    </div>
  ) : null
}
