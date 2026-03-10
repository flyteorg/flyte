import { SearchBar } from '@/components/SearchBar'
import { useSearchTerm } from '@/hooks/useQueryParamState'

export const ListRunsSearch = () => {
  const { searchTermInput, setSearchTerm } = useSearchTerm()

  return (
    <SearchBar
      placeholder="Search runs"
      value={searchTermInput ?? undefined}
      onChange={(e) => setSearchTerm(e.target.value)}
    />
  )
}
