import clsx from 'clsx'

export const SearchFilterButton = ({
  displayText,
  isSelected,
  onClick,
}: {
  displayText: string | React.ReactNode
  isSelected: boolean
  onClick: () => void
}) => {
  const label =
    typeof displayText === 'string'
      ? isSelected
        ? `Remove ${displayText} filter`
        : `Filter by ${displayText}`
      : undefined

  return (
    <button
      aria-label={label}
      aria-pressed={isSelected}
      className={clsx(
        'cursor-pointer rounded-lg border-[1.5px] px-[10px] py-1 text-[13px] leading-none font-medium focus:outline-hidden',
        isSelected
          ? 'border-(--system-gray-4) bg-(--system-gray-3) text-white'
          : 'border-(--system-gray-4) bg-transparent text-(--system-gray-5)',
      )}
      onClick={onClick}
    >
      {displayText}
    </button>
  )
}
