import { ChartIcon } from '@/components/icons/ChartIcon'
import { LoadingSpinner } from '@/components/LoadingSpinner'
import clsx from 'clsx'

const classes = {
  dataLabel: 'text-lg font-medium text-(--system-white)',
  subtitle: 'text-sm text-(--system-gray-5) ',
}

type TableEmptyProps<T> = {
  data?: T[]
  dataLabel: string
  subtitle: string
  content?: React.ReactNode
  searchQuery?: string | null
  isError?: boolean
  isLoading?: boolean
  children: (data: T[]) => React.ReactNode
  className?: string
}

export function TableState<T extends object>({
  dataLabel,
  subtitle,
  data = [],
  content,
  searchQuery,
  isError,
  isLoading,
  children,
  className,
}: TableEmptyProps<T>) {
  let renderContent = undefined
  switch (true) {
    case isLoading:
      renderContent = <LoadingSpinner color="union" />
      break
    case isError:
      renderContent = (
        <>
          <div className={classes.dataLabel}>Error</div>
          <div className={classes.subtitle}>
            We&apos;re having trouble loading your {dataLabel}
          </div>
        </>
      )
      break
    case Boolean(searchQuery):
      renderContent = (
        <>
          <div className={classes.dataLabel}>No {dataLabel} found</div>
          <div className={classes.subtitle}>
            No {dataLabel} found matching{' '}
            <span className="font-bold text-(--system-gray-5)">
              {searchQuery}
            </span>
          </div>
        </>
      )
      break
    default:
      // displayed when data is empty
      renderContent = (
        <>
          <div className="flex gap-2 text-(--system-gray-5)">
            <ChartIcon className="mt-0.5 size-5" />
            <span className="text-sm">No {dataLabel}</span>
          </div>
          <div className={`${classes.subtitle} max-w-106 pb-2 text-center`}>
            {subtitle}
          </div>
          {content}
        </>
      )
  }

  return data.length > 0 ? (
    <>{children(data)}</>
  ) : (
    <div
      className={clsx(
        'max-h-100vh flex min-h-[200px] flex-1 flex-col items-center overflow-auto border-t border-zinc-200 p-4 dark:border-zinc-700',
        className,
      )}
    >
      <div className="flex flex-1 flex-col items-center justify-center gap-2">
        {renderContent}
      </div>
    </div>
  )
}
