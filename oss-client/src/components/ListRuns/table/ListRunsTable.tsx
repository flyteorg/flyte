import { TableFooterLoadMore, VirtualizedTable } from '@/components/Tables'
import { Run } from '@/gen/flyteidl2/workflow/run_definition_pb'
import { useWatchRuns } from '@/hooks/useWatchRuns'
import { ColumnDef } from '@tanstack/react-table'
import { useMemo } from 'react'
import { type RunsTableRow } from './types'
import { formatForTable } from './util'

interface ListRunsTableProps {
  // 'any' because of complex type
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  columns: ColumnDef<RunsTableRow, any>[]
  runs: Run[]
  runsQuery: ReturnType<typeof useWatchRuns>
  showFooter?: boolean
  hideLastRowBorder?: boolean
}

export const ListRunsTable = ({
  columns,
  runs,
  runsQuery,
  showFooter = true,
  hideLastRowBorder = false,
}: ListRunsTableProps) => {
  const formattedRuns = useMemo(() => {
    return runs?.map(formatForTable)
  }, [runs])

  return (
    <VirtualizedTable<RunsTableRow>
      columns={columns}
      data={formattedRuns || []}
      getRowHref={(runRow) =>
        `/domain/${runRow.original.action?.id?.run?.domain}/project/${runRow.original.action?.id?.run?.project}/runs/${runRow.original.action?.id?.run?.name}`
      }
      footerRow={
        showFooter ? (
          <TableFooterLoadMore
            query={runsQuery}
            label="runs"
            key={'more-runs'}
          />
        ) : undefined
      }
      hideLastRowBorder={hideLastRowBorder}
      rowHeight={60}
      overscan={10}
    />
  )
}
