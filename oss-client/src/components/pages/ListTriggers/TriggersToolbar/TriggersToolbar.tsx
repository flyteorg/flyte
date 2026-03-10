import { SearchBar } from '@/components/SearchBar'
import { useSearchTerm } from '@/hooks/useQueryParamState'
import { useListTriggers } from '@/hooks/useTriggers'

type TriggersToolbarProps = {
  triggersQuery: ReturnType<typeof useListTriggers>
}

export const TriggersToolbar = ({ triggersQuery }: TriggersToolbarProps) => {
  const { searchTermInput, setSearchTerm } = useSearchTerm()
  return (
    <>
      <div className="flex items-center justify-between gap-2 px-10 pt-6 pb-6">
        <div className="flex flex-col">
          <h1 className="text-xl font-medium">Triggers</h1>
          <span className="text-2xs font-semibold dark:text-[#898989]">
            {triggersQuery.data?.triggers.length} total
          </span>
        </div>

        <SearchBar
          placeholder="Search triggers"
          value={searchTermInput ?? undefined}
          onChange={(e) => setSearchTerm(e.target.value)}
        />
      </div>
    </>
  )
}
