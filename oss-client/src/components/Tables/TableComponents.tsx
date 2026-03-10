import { PropsWithChildren, Ref } from 'react'

export const TableRow = ({
  children,
  className,
  onClick = () => null,
  style,
  ref,
  dataIndex,
}: PropsWithChildren<{
  className?: string
  onClick?: () => void
  style?: React.CSSProperties
  ref?: Ref<HTMLTableRowElement>
  dataIndex?: string | number
}>) => {
  return (
    <tr
      className={`${className}`}
      onClick={onClick}
      style={style}
      ref={ref}
      data-index={dataIndex}
    >
      {children}
    </tr>
  )
}

export const TableHead = ({
  children,
  className,
  ...props
}: React.ThHTMLAttributes<HTMLTableCellElement> & PropsWithChildren) => {
  return (
    <th
      className={`py-2 text-left text-sm font-semibold text-zinc-500 dark:text-zinc-400 ${className}`}
      {...props}
    >
      {children}
    </th>
  )
}

export const TableCell = ({
  children,
  className,
  style,
}: PropsWithChildren<{
  className?: string
  style?: React.CSSProperties
}>) => (
  <td className={`py-2 ${className}`} style={style}>
    {children}
  </td>
)
