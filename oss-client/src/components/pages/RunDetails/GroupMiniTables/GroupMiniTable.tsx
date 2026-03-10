import { ChevronRightIcon } from '@/components/icons/ChevronRightIcon'
import { MiniTable, TableRow } from '@/components/Tables'
import { useMemo, useState } from 'react'
import { useSelectedItem } from '../hooks/useSelectedItem'
import { type GroupTableItem } from './types'
import { useGroupColumns } from './useGroupColumns'

const INITIAL_ROW_COUNT = 3

const MiniTableFooter = ({
  actionType,
  handleClick,
  totalItems,
}: {
  actionType: string
  handleClick: () => void
  totalItems: number
}) => (
  <tfoot className="border-t-1 border-(--system-gray-2) dark:border-(--system-gray-2)">
    <TableRow
      className="flex cursor-pointer items-center px-6 py-2 text-xs font-normal dark:text-(--system-gray-5)"
      onClick={handleClick}
    >
      View all {actionType} actions ({totalItems}){' '}
      <ChevronRightIcon width={22} />
    </TableRow>
  </tfoot>
)

export const GroupMiniTable = ({
  actionType,
  data,
}: {
  data: GroupTableItem[] | null
  actionType: string
}) => {
  const [hasExpandedAllActions, setHasExpandedAllActions] = useState(false)
  const { setSelectedItem } = useSelectedItem()
  const columns = useGroupColumns({ durationDisplayText: 'Duration' })

  const dataToDisplay = useMemo(() => {
    if (hasExpandedAllActions) {
      return data || []
    }
    return data?.slice(0, INITIAL_ROW_COUNT) || []
  }, [data, hasExpandedAllActions])

  if (!data) return null
  const shouldShowFooter =
    !hasExpandedAllActions && dataToDisplay?.length !== data.length

  return (
    <MiniTable
      className="mb-2 max-h-[400px] w-full overflow-y-auto [&_th]:sticky [&_th]:top-0 [&_th]:bg-(--system-black)"
      onRowClick={(val) => {
        const actionId = val.original.action?.id?.name
        if (actionId) {
          setSelectedItem(actionId)
        }
      }}
      {...(shouldShowFooter && {
        footer: (
          <MiniTableFooter
            actionType={actionType}
            handleClick={() => setHasExpandedAllActions(true)}
            totalItems={data.length}
          />
        ),
      })}
      columns={columns}
      data={dataToDisplay}
    />
  )
}
