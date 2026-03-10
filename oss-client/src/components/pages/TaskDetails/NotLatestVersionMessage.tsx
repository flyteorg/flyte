import { ArrowTopRightIcon } from '@/components/icons/ArrowTopRightIcon'
import { ChartIcon } from '@/components/icons/ChartIcon'

export const NotLatestVersionMessage = ({
  onViewLatestVersion,
}: {
  onViewLatestVersion?: () => void
}) => {
  return (
    <div className="flex h-full min-h-0 w-full min-w-0 flex-1 flex-col rounded-lg border border-zinc-200 bg-white pt-4 dark:border-zinc-700 dark:bg-(--system-black)">
      <div className="max-h-100vh flex min-h-[200px] flex-1 flex-col items-center overflow-auto p-4">
        <div className="flex flex-1 flex-col items-center justify-center gap-2">
          <div className="flex gap-2 text-(--system-gray-5)">
            <ChartIcon className="mt-0.5 size-5" />
            <span className="text-sm">
              Triggers apply to latest task version only
            </span>
          </div>
          <div className="text-sm text-zinc-600 dark:text-zinc-400">
            This task version has no active triggers. All trigger configurations
            live in the latest version.
          </div>
          {onViewLatestVersion && (
            <button
              onClick={onViewLatestVersion}
              className="mt-2 flex w-fit items-center gap-2 text-xs font-medium text-(--system-black) hover:underline dark:text-(--system-white)"
            >
              View latest version
              <ArrowTopRightIcon className="h-2 w-2 text-(--system-gray-5)" />
            </button>
          )}
        </div>
      </div>
    </div>
  )
}
