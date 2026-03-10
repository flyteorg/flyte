import { VirtualizedTable, type BulkAction } from '@/components/Tables'
import { useBulkSelection } from '@/components/Tables/useBulkSelection'
import { Trigger } from '@/gen/flyteidl2/trigger/trigger_definition_pb'
import { useOrg } from '@/hooks/useOrg'
import { TableSort } from '@/hooks/useQueryParamSort'
import { useListTriggers, useUpdateTriggers } from '@/hooks/useTriggers'
import { type ProjectDomainPageParams } from '@/types/pageParams'
import { useParams } from 'next/navigation'
import { useMemo } from 'react'
import { useTriggersTableColumns } from './getTriggersTableColumns'
import { type TriggerTableRowWithHighlights } from './types'
import { formatForTable, formatHighlights, getTriggerName } from './util'

export const TriggersTable = ({
  triggers,
  searchTerm,
  className,
  hideColumns,
  tableSort,
}: {
  triggers: Trigger[]
  searchTerm: string
  triggersQuery?: ReturnType<typeof useListTriggers>
  className?: string
  hideColumns?: string[]
  tableSort?: TableSort<TriggerTableRowWithHighlights>
}) => {
  const formattedTriggers = useMemo(() => {
    return formatHighlights(
      triggers?.map((t) => formatForTable(t)),
      searchTerm,
    )
  }, [triggers, searchTerm])
  const params = useParams<ProjectDomainPageParams>()

  // Bulk selection logic
  const getId = (item: TriggerTableRowWithHighlights) =>
    item.actions.id?.name?.name || ''
  const bulk = useBulkSelection(formattedTriggers || [], getId)

  const columns = useTriggersTableColumns(bulk, hideColumns)

  const org = useOrg()
  const { mutate } = useUpdateTriggers({
    org,
    domain: params.domain,
    projectId: params.project,
    search: searchTerm || '',
    taskNames: triggers.map((t) => t.id?.name?.taskName || ''),
  })

  const bulkActions: BulkAction<TriggerTableRowWithHighlights>[] = useMemo(
    () => [
      {
        makeLabel: (selectedItems) => {
          return selectedItems.every((i) => i.active.status === true)
            ? 'Deactivate'
            : 'Activate'
        },

        onClick: async (selectedItems, actions) => {
          const newActiveState = !selectedItems.every(
            (i) => i.active.status === true,
          )
          const newActiveStateDisplayString =
            newActiveState === false ? 'deactivated' : 'activated'
          const triggerNames = selectedItems.map((item) =>
            getTriggerName(item.actions),
          )
          // keep track of triggers that actually toggled for undo correctness
          const undoTriggerNames = selectedItems
            .filter((item) => item.actions.active !== newActiveState)
            .map((item) => getTriggerName(item.actions))
          try {
            console.log('about to await mutation')
            await mutate({
              active: newActiveState,
              triggerNames,
            })
            const undo = () => {
              mutate({
                active: !newActiveState,
                triggerNames: undoTriggerNames,
              })
              actions.clearNotification()
            }
            actions.showNotification({
              message: `${selectedItems.length} task${selectedItems.length > 1 ? 's' : ''} ${newActiveStateDisplayString}`,
              durationMs: 4000,
              undoCallback: undo,
            })
          } catch (e) {
            console.log('Error updating triggers', e)
            actions.showNotification({
              message: `There was an error updating triggers`,
              durationMs: 4000,
              variant: 'warn',
            })
          }
        },
      },
    ],
    [mutate],
  )

  return (
    <VirtualizedTable<TriggerTableRowWithHighlights>
      bulkAction={{
        bulkActions,
        selectedItems: bulk.selectedItems,
        clearSelection: bulk.clearSelection,
      }}
      columns={columns}
      data={formattedTriggers || []}
      rowHeight={80}
      overscan={10}
      getRowHref={(triggerRow) =>
        `/domain/${triggerRow.actions.id?.name?.domain}/project/${triggerRow.actions.id?.name?.project}/triggers/${triggerRow.actions.id?.name?.taskName}/${triggerRow.actions.id?.name?.name}`
      }
      className={className}
      sorting={tableSort}
    />
  )
}
