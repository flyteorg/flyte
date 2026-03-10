import { TableCell, TableHead, TableRow } from '@/components/Tables'
import {
  ColumnDef,
  flexRender,
  getCoreRowModel,
  useReactTable,
} from '@tanstack/react-table'
import Link from 'next/link'

export function Table<T extends object>({
  columns,
  data,
  getRowHref,
  onRowClick,
}: {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  columns: ColumnDef<T, any>[]
  data: T[]
  getRowHref?: (args: T) => string // if passed, each row will be a link
  onRowClick?: (args: T) => void
}) {
  const table = useReactTable({
    data,
    columns,
    getCoreRowModel: getCoreRowModel(),
  })

  return (
    <div className="min-h-0 w-full overflow-auto border-t border-(--system-gray-3) align-middle">
      <table className="min-w-full border-collapse">
        {/* shadow added as border; standard bottom border is not working with sticky header */}
        <thead className="bg-primary sticky top-0 z-10 shadow-[0_1px_0_0_#F4F4F5] dark:shadow-[0_1px_0_0_#27272A]">
          {table.getHeaderGroups().map((headerGroup) => (
            <TableRow key={headerGroup.id}>
              {headerGroup.headers.map((header) => {
                return (
                  <TableHead
                    className="px-6"
                    key={header.id}
                    style={{ width: `${header.getSize()}px` }}
                  >
                    {header.isPlaceholder
                      ? null
                      : flexRender(
                          header.column.columnDef.header,
                          header.getContext(),
                        )}
                  </TableHead>
                )
              })}
            </TableRow>
          ))}
        </thead>
        <tbody>
          {table.getRowModel().rows.map((row) => {
            const href = getRowHref?.(row.original)
            const innerRow = (
              <TableRow
                className="cursor-pointer border-b border-(--system-gray-3) hover:bg-(--system-gray-3)"
                key={row.id}
                onClick={() => {
                  onRowClick?.(row.original)
                }}
              >
                {row.getVisibleCells().map((cell) => (
                  <TableCell
                    key={cell.id}
                    className="px-6"
                    style={{
                      width: `${cell.column.getSize()}px`,
                      // maxWidth is not set as exact passed column.maxSize, but proportionally to whole row
                      // maxWidth needed for text-ellipsis
                      maxWidth: `${cell.column.columnDef.maxSize}px`,
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
    </div>
  )
}
