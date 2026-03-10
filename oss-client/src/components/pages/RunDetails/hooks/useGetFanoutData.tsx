import { useSelectedItem } from './useSelectedItem'
import { useRunStore } from '../state/RunStore'
import { usePhaseCounts } from '../Sidebar/usePhaseCounts'

export const useGetFanoutData = () => {
  const { selectedItem } = useSelectedItem()
  const flatItems = useRunStore((s) => s.flatItems)
  const groupItem = flatItems.find(
    (flatItem) => flatItem.id === selectedItem?.id,
  )
  const phaseCounts = usePhaseCounts({ node: groupItem?.node })
  return phaseCounts
}
