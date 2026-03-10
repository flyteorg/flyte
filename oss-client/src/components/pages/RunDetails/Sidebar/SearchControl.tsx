import { SearchBar } from '@/components/SearchBar'
import { useSearchTerm } from '@/hooks/useQueryParamState'

export const SearchControl = () => {
  const { searchTermInput, setSearchTerm } = useSearchTerm()

  return (
    <div className="flex h-[28px] items-center justify-between">
      <form
        className="min-w-0 flex-1"
        onSubmit={(e) => {
          e.preventDefault()
          setSearchTerm(searchTermInput)
        }}
      >
        <SearchBar
          value={searchTermInput ?? undefined}
          onChange={(e) => setSearchTerm(e.target.value ?? '')}
          className="w-full"
        />
      </form>
    </div>
  )
}
