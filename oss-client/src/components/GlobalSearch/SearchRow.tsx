import { Link } from '../Link'
import clsx from 'clsx'

export const SearchRow = ({
  children,
  href,
  isSelected,
}: {
  children: React.ReactNode
  href: string
  isSelected?: boolean
}) => {
  return (
    <Link href={`/v2/${href}`}>
      <div
        className={clsx(
          'cursor-pointer bg-(--system-gray-1) px-4 py-2 hover:bg-(--system-gray-3)',
          isSelected && 'bg-(--system-gray-3)',
        )}
      >
        {children}
      </div>
    </Link>
  )
}
