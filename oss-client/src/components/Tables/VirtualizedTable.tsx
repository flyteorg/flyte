import { Button } from '@/components/Button'
import { StatusIcon } from '@/components/StatusIcons'
import { TableCell, TableHead, TableRow } from '@/components/Tables'
import { ActionPhase } from '@/gen/flyteidl2/common/phase_pb'
import { ToggleSortColumnByFn } from '@/hooks/useQueryParamSort'
import { ChevronUpIcon, XMarkIcon } from '@heroicons/react/16/solid'
import {
  ColumnDef,
  flexRender,
  getCoreRowModel,
  SortDirection,
  SortingState,
  useReactTable,
} from '@tanstack/react-table'
import { useVirtualizer } from '@tanstack/react-virtual'
import clsx from 'clsx'
import Link from 'next/link'
import { useCallback, useEffect, useRef, useState } from 'react'
import { type BulkAction, type NotificationProps } from './types'

export type ScrollToIndexFn = (
  index: number,
  options?: {
    align?: 'start' | 'center' | 'end'
    behavior?: 'auto' | 'smooth'
  },
) => void

// Virtualized table component for large datasets
export function VirtualizedTable<T extends object>({
  bulkAction,
  columns,
  data,
  getRowHref,
  onRowClick,
  footerRow,
  rowHeight = 60, // Default row height
  getRowHeight, // Optional function to get row height based on row data
  overscan = 5, // Number of items to render outside viewport
  className,
  getRowClassName,
  onVirtualizerReady,
  sorting,
  hideLastRowBorder,
}: {
  bulkAction?: {
    bulkActions: BulkAction<T>[]
    selectedItems: T[]
    clearSelection: () => void
  }
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  columns: ColumnDef<T, any>[]
  data: T[]
  getRowHref?: (args: T) => string // if passed, each row will be a link
  onRowClick?: (args: T) => void
  footerRow?: React.ReactNode // Optional footer row for infinite scrolling
  rowHeight?: number // Height of each row
  getRowHeight?: (row: T, index: number) => number // Optional function to get row height based on row data
  overscan?: number // Number of items to render outside viewport
  className?: string // Additional classes for the table
  getRowClassName?: (row: T) => string // Optional function to get custom className for a row
  onVirtualizerReady?: (scrollToIndex: ScrollToIndexFn) => void
  sorting?: {
    state: SortingState
    onToggleColumnSort: ToggleSortColumnByFn<T>
  }
  /** When true, the last data row has no bottom border */
  hideLastRowBorder?: boolean
}) {
  const [notification, setNotification] = useState<Omit<
    NotificationProps,
    'durationMs'
  > | null>(null)
  const notificationTimeoutRef = useRef<NodeJS.Timeout | null>(null)
  const table = useReactTable({
    data,
    columns: columns,
    getCoreRowModel: getCoreRowModel(),
    state: {
      sorting: sorting?.state,
    },
    enableSorting: !!sorting,
    enableSortingRemoval: false,
    enableMultiSort: false,
  })

  const parentRef = useRef<HTMLDivElement>(null)

  const rowVirtualizer = useVirtualizer({
    count: data.length,
    getScrollElement: () => parentRef.current,
    estimateSize: (index) =>
      getRowHeight ? getRowHeight(data[index], index) : rowHeight,
    measureElement: (element) => element?.getBoundingClientRect().height,
    overscan,
  })

  // Expose scrollToIndex method to parent component
  useEffect(() => {
    if (onVirtualizerReady) {
      onVirtualizerReady(
        (
          index: number,
          options?: {
            align?: 'start' | 'center' | 'end'
            behavior?: 'auto' | 'smooth'
          },
        ) => {
          rowVirtualizer.scrollToIndex(index, {
            align: options?.align || 'start',
            behavior:
              options?.behavior ||
              // Always default to smooth scrolling
              'smooth',
          })
        },
      )
    }
  }, [onVirtualizerReady, rowVirtualizer])

  const clearNotification = useCallback(() => {
    setNotification(null)
    if (notificationTimeoutRef.current) {
      clearTimeout(notificationTimeoutRef.current)
      notificationTimeoutRef.current = null
    }
  }, [setNotification])

  const showNotification = useCallback(
    (props: NotificationProps) => {
      setNotification({
        message: props.message,
        undoCallback: props.undoCallback,
      })
      notificationTimeoutRef.current = setTimeout(() => {
        setNotification(null)
        notificationTimeoutRef.current = null
      }, props.durationMs)
    },
    [setNotification],
  )

  return (
    <div
      ref={parentRef}
      className={clsx(
        'bg-primary relative min-h-0 flex-1 overflow-auto border-t border-zinc-200 align-middle dark:border-zinc-700',
        // scroll
        '[scrollbar-gutter:stable] [&::-webkit-scrollbar]:size-2 [&::-webkit-scrollbar-corner]:bg-transparent [&::-webkit-scrollbar-track]:bg-transparent',
        // scroll thumb
        '[&::-webkit-scrollbar-thumb]:rounded-full [&::-webkit-scrollbar-thumb]:bg-zinc-400 dark:[&::-webkit-scrollbar-thumb]:bg-zinc-600 [&::-webkit-scrollbar-thumb:hover]:bg-zinc-500 dark:[&::-webkit-scrollbar-thumb:hover]:bg-zinc-500',
        className,
      )}
    >
      {notification && (
        <div className="absolute top-0 left-1/2 z-50 flex -translate-x-1/2 transform items-center justify-center gap-2 rounded-sm px-4 py-2 text-xs shadow-xs dark:bg-(--system-gray-1) dark:text-(--system-gray-6)">
          {notification.variant === 'warn' ? (
            <StatusIcon phase={ActionPhase.FAILED} isStatic />
          ) : (
            <StatusIcon phase={ActionPhase.SUCCEEDED} isStatic />
          )}
          {notification.message}
          {notification.undoCallback && (
            <Button onClick={notification.undoCallback} plain>
              <span className="font-normal text-(--union)">Undo</span>
            </Button>
          )}
        </div>
      )}
      <table className="grid">
        <thead className="bg-primary sticky top-0 z-10 shadow-[0_1px_0_0_#F4F4F5] dark:shadow-[0_1px_0_0_#27272A]">
          {table.getHeaderGroups().map((headerGroup) => (
            <TableRow key={headerGroup.id} className="flex">
              {headerGroup.headers.map((header) => {
                const isSortable = header.column.getCanSort()
                const sortDirection = header.column.getIsSorted()
                const isSorted = !!header.column.getIsSorted()
                return (
                  <TableHead
                    className={clsx(
                      'px-4 last:pr-6',
                      bulkAction ? 'first:pr-0 first:pl-2' : 'first:pl-10',
                      header.column.columnDef.size ===
                        header.column.columnDef.minSize
                        ? ''
                        : 'flex-1',
                    )}
                    key={header.id}
                    style={{
                      width:
                        header.column.columnDef.size ===
                        header.column.columnDef.minSize
                          ? `${header.column.columnDef.size}px`
                          : undefined,
                      minWidth: `${header.column.columnDef.minSize}px`,
                    }}
                  >
                    <div
                      className={clsx(
                        'flex items-center gap-1',
                        isSortable && 'cursor-pointer select-none',
                        isSorted && 'dark:text-(--system-gray-7)',
                      )}
                      onClick={() => {
                        if (!isSortable) return
                        sorting?.onToggleColumnSort({
                          columnId: header.column.id as keyof T,
                          direction:
                            header.column.getNextSortingOrder() as SortDirection, // always "asc" | "desc"
                        })
                      }}
                      title={
                        isSortable
                          ? header.column.getNextSortingOrder() === 'asc'
                            ? 'Sort ascending'
                            : 'Sort descending'
                          : undefined
                      }
                    >
                      {flexRender(
                        header.column.columnDef.header,
                        header.getContext(),
                      )}
                      {isSorted ? (
                        <ChevronUpIcon
                          className={clsx(
                            'size-6 min-w-6 transition-transform duration-300',
                            sortDirection === 'asc' && 'rotate-0',
                            sortDirection === 'desc' && 'rotate-180',
                          )}
                        />
                      ) : null}
                    </div>
                  </TableHead>
                )
              })}
            </TableRow>
          ))}
          {bulkAction && bulkAction.selectedItems.length > 0 && (
            <tr className="absolute top-0 left-0 flex h-11 w-[101%] items-center bg-(--system-gray-2)">
              <td className="flex w-full items-center justify-between pr-5">
                <div className="flex gap-2 p-4">
                  {bulkAction.selectedItems.length} selected
                  <Button
                    className="!px-1 !py-0"
                    plain
                    onClick={bulkAction.clearSelection}
                  >
                    <XMarkIcon width={20} />
                  </Button>
                </div>
                {bulkAction.bulkActions.map((action) => (
                  <Button
                    key={action.makeLabel(bulkAction.selectedItems)}
                    className="!py-0"
                    outline
                    onClick={() =>
                      action.onClick(bulkAction.selectedItems, {
                        clearNotification,
                        clearSelection: bulkAction.clearSelection,
                        showNotification,
                      })
                    }
                  >
                    {action.makeLabel(bulkAction.selectedItems)}
                  </Button>
                ))}
              </td>
            </tr>
          )}
        </thead>

        {/* Virtualized content */}
        <tbody
          style={{
            height: `${rowVirtualizer.getTotalSize()}px`,
            position: 'relative',
          }}
        >
          {/* Virtualized rows using the official TanStack pattern */}
          {rowVirtualizer.getVirtualItems().map((virtualRow) => {
            const row = table.getRowModel().rows[virtualRow.index]
            if (!row) return null
            const href = getRowHref?.(row.original)
            const customRowClassName = getRowClassName?.(row.original) || ''
            const isLastRow =
              hideLastRowBorder && virtualRow.index === data.length - 1
            const innerRow = (
              <TableRow
                dataIndex={row.id}
                ref={rowVirtualizer.measureElement}
                className={clsx(
                  'group absolute flex w-full',
                  // Only add cursor-pointer and hover if not disabled by custom className
                  !customRowClassName.includes('no-hover') && 'cursor-pointer',
                  // border (after as bottom border for safari)
                  'after:absolute after:right-0 after:bottom-0 after:left-0 after:block after:h-[1px] after:bg-(--system-gray-3)',
                  isLastRow && 'after:hidden [&_td]:after:hidden',
                  // hover (hover over td for safari) - only if not disabled
                  !customRowClassName.includes('no-hover') &&
                    'hover:bg-(--system-gray-1) hover:[&>td]:bg-(--system-gray-1) dark:hover:[&>td]:bg-(--system-gray-2)',
                  customRowClassName,
                )}
                key={row.id}
                onClick={() => {
                  onRowClick?.(row.original)
                }}
                style={{
                  height: `${virtualRow.size}px`,
                  transform: `translateY(${virtualRow.start}px)`,
                }}
              >
                {row.getVisibleCells().map((cell) => (
                  <TableCell
                    key={cell.id}
                    className={clsx(
                      'flex items-center px-4 last:pr-6',
                      bulkAction ? 'first:pr-0 first:pl-2' : 'first:pl-10',
                      cell.column.columnDef.size ===
                        cell.column.columnDef.minSize
                        ? ''
                        : 'flex-1',
                      // border (after as bottom border for safari)
                      'relative after:absolute after:right-0 after:bottom-0 after:left-0 after:block after:h-[1px] after:bg-(--system-gray-3)',
                    )}
                    style={{
                      width:
                        cell.column.columnDef.size ===
                        cell.column.columnDef.minSize
                          ? `${cell.column.columnDef.size}px`
                          : undefined,
                      minWidth: `${cell.column.columnDef.minSize}px`,
                    }}
                  >
                    {flexRender(cell.column.columnDef.cell, cell.getContext())}
                  </TableCell>
                ))}
              </TableRow>
            )
            return href ? (
              <Link key={row.id} href={href} className="contents">
                {innerRow}
              </Link>
            ) : (
              innerRow
            )
          })}
        </tbody>
      </table>

      {/* Footer row for infinite scrolling - inside scrollable area so it only appears when scrolled to bottom */}
      {footerRow && <div className="w-full">{footerRow}</div>}
    </div>
  )
}
