import clsx from 'clsx'

interface ClearAllFiltersButtonProps {
  onClick: () => void
  className?: string
}

export const ClearAllFiltersButton = ({
  onClick,
  className,
}: ClearAllFiltersButtonProps) => {
  return (
    <button
      onClick={onClick}
      className={clsx(
        'inline-flex h-6 cursor-pointer items-center gap-1 rounded-lg px-2 py-0.5 transition-colors focus:outline-none',
        'border-[1.5px] border-(--system-gray-2)',
        'text-xs font-medium text-(--system-white)',
        'hover:bg-(--system-gray-3)',
        className,
      )}
    >
      Clear all
    </button>
  )
}
