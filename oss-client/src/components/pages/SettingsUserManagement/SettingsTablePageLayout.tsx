import { SearchBar } from '@/components/SearchBar'
import { TableState } from '@/components/Tables'
import { ReactNode } from 'react'

interface SettingsTablePageLayoutProps<T extends object> {
  count: number
  countLabel: string
  searchPlaceholder: string
  searchValue?: string
  onSearchChange: (e: React.ChangeEvent<HTMLInputElement>) => void
  data: T[]
  dataLabel: string
  subtitle: string
  isError: boolean
  isLoading: boolean
  searchQuery?: string
  children: (data: T[]) => ReactNode
  headerActions?: ReactNode
}

export const SettingsTablePageLayout = <T extends object>({
  count,
  countLabel,
  searchPlaceholder,
  searchValue,
  onSearchChange,
  data,
  dataLabel,
  subtitle,
  isError,
  isLoading,
  searchQuery,
  children,
  headerActions,
}: SettingsTablePageLayoutProps<T>) => {
  return (
    <div className="mx-10 flex h-full min-h-0 grow flex-col overflow-hidden rounded-xl border border-(--system-gray-3) bg-white dark:bg-(--system-black)">
      <div className="flex items-center justify-between gap-4 px-4 py-2.5">
        <div className="flex items-center gap-4">
          <div className="text-sm text-zinc-500 dark:text-(--system-gray-5)">
            {countLabel} {count} total
          </div>
          {headerActions}
        </div>
        <SearchBar
          placeholder={searchPlaceholder}
          value={searchValue}
          onChange={onSearchChange}
        />
      </div>

      <div className="flex min-h-0 flex-1 flex-col">
        <TableState
          dataLabel={dataLabel}
          subtitle={subtitle}
          data={data}
          isError={isError}
          isLoading={isLoading}
          searchQuery={searchQuery}
        >
          {children}
        </TableState>
      </div>
    </div>
  )
}
