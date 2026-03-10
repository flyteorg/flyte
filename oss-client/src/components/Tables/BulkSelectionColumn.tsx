import { createColumnHelper } from '@tanstack/react-table'
import clsx from 'clsx'
import { Checkbox } from '../Checkbox'

export interface BulkSelectionColumnProps<T> {
  bulkSelection: {
    areAllItemsSelected: boolean
    isSelected: (item: T) => boolean
    toggleItem: (item: T) => void
    selectAll: () => void
    unselectAll: () => void
  }
}

export function createBulkSelectionColumn<T extends object>(
  bulkSelection: BulkSelectionColumnProps<T>['bulkSelection'],
) {
  const helper = createColumnHelper<T>()

  return helper.display({
    id: 'bulk-selection',
    header: () => (
      <div className="mt-1 flex h-full w-full items-center justify-center text-center">
        <div className="flex items-center justify-center">
          <Checkbox
            checked={bulkSelection.areAllItemsSelected}
            onChange={() => {
              if (bulkSelection.areAllItemsSelected) {
                bulkSelection.unselectAll()
              } else {
                bulkSelection.selectAll()
              }
            }}
            className="h-4 w-4 data-checked:border-(--system-gray-4) data-checked:bg-transparent"
          />
        </div>
      </div>
    ),
    cell: ({ row }) => (
      <div className="flex h-full w-full items-center justify-center text-center">
        <div className="flex items-center justify-center">
          <Checkbox
            checked={bulkSelection.isSelected(row.original)}
            onClick={(e) => e.stopPropagation()}
            onChange={() => bulkSelection.toggleItem(row.original)}
            className={clsx(
              'h-4 w-4 cursor-pointer transition-opacity data-checked:border-(--system-gray-4) data-checked:bg-transparent',
              bulkSelection.isSelected(row.original)
                ? 'opacity-100'
                : 'opacity-0 group-hover:opacity-100',
            )}
          />
        </div>
      </div>
    ),
    minSize: 30,
    size: 30,
    enableSorting: false,
    enableHiding: false,
  })
}
