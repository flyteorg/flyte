import { useEffect, useState } from 'react'
import { useSearchTerm } from '@/hooks/useQueryParamState'
import { LoadingSpinner } from '@/components/LoadingSpinner'
import { useQueryFilters } from '@/hooks/useQueryFilters'

const DELAY = 300

export const EmptySidebar = () => {
  const { searchTerm } = useSearchTerm()
  const { filters } = useQueryFilters()

  const [show, setShow] = useState(false)

  useEffect(() => {
    const timer = setTimeout(() => setShow(true), DELAY)
    return () => clearTimeout(timer)
  }, [])

  if (show && (filters.status || searchTerm)) {
    return (
      <div className="flex h-full flex-col items-center justify-center gap-2">
        <div className="text-lg">No actions found</div>
        <div className="text-zinc-600">
          {' '}
          No actions found for the specified filters
        </div>
      </div>
    )
  }
  return <LoadingSpinner delay={DELAY} />
}
