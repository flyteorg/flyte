import {
  ColumnDef,
  flexRender,
  getCoreRowModel,
  useReactTable,
} from '@tanstack/react-table'
import { TableCell, TableHead, TableRow } from './TableComponents'

type MiniTableProps<T> = {
  className?: string
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  columns: ColumnDef<T, any>[]
  data: T[]
  footer?: React.ReactNode
  onRowClick?: (args: T) => void
}

export function MiniTable<T extends object>(props: MiniTableProps<T>) {
  const table = useReactTable({
    data: props.data,
    columns: props.columns,
    getCoreRowModel: getCoreRowModel(),
  })
  return (
    <table className={props.className}>
      <thead>
        {table.getHeaderGroups().map((headerGroup) => (
          <TableRow
            key={headerGroup.id}
            className="border-b-1 border-(--system-gray-2) dark:border-(--system-gray-2)"
          >
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

      <tbody className="divide-y divide-(--system-gray-2) dark:divide-(--system-gray-2)">
        {table.getRowModel().rows.map((row) => (
          <TableRow
            className="cursor-pointer hover:bg-(--system-gray-2)"
            key={row.id}
            onClick={() => props.onRowClick?.(row.original)}
          >
            {row.getVisibleCells().map((cell) => (
              <TableCell
                key={cell.id}
                className="px-6 text-xs"
                style={{ width: `${cell.column.getSize()}px` }}
              >
                {flexRender(cell.column.columnDef.cell, cell.getContext())}
              </TableCell>
            ))}
          </TableRow>
        ))}
      </tbody>
      {props.footer ?? props.footer}
    </table>
  )
}
