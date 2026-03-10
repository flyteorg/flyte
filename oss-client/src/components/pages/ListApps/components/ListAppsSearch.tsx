import { useSearchTerm } from '@/hooks/useQueryParamState'
import { SearchBar } from '@/components/SearchBar'

export const ListAppsSearch = () => {
  const { searchTermInput, setSearchTerm } = useSearchTerm()
  return (
    <SearchBar
      placeholder="Search apps"
      value={searchTermInput ?? undefined}
      onChange={(e) => setSearchTerm(e.target.value)}
    />
  )
}
