import { CopyButtonWithTooltip } from '@/components/CopyButtonWithTooltip'
import { TabSection } from '@/components/TabSection'
import { AnimatePresence, motion } from 'motion/react'
import { ActionFanoutTable } from './ActionFanoutTable'
import { useSelectedGroup } from './hooks/useSelectedItem'
import { useGetFanoutData } from './hooks/useGetFanoutData'
import { GroupMiniTable } from './GroupMiniTables'
import { useGroupAggregation } from './GroupMiniTables/useGroupAggregation'
import { getLocation } from '@/lib/windowUtils'

const GroupDataHeading = ({ displayText }: { displayText: string }) => (
  <div className="mb-1 flex gap-2">
    <span>{displayText}</span>
    <CopyButtonWithTooltip
      icon="chain"
      textInitial="Copy group URL"
      textCopied="Group URL copied to clipboard"
      value={getLocation().href}
      classNameBtn="-ml-1"
    />
  </div>
)

export const RunDetailsGroupTab = () => {
  const groupData = useSelectedGroup()
  const childPhaseCounts = useGetFanoutData()
  const { failed, longest } = useGroupAggregation()

  if (!groupData) {
    return (
      <div className="flex h-15 w-full bg-(--bg-red) p-5 px-10 text-(--accent-red)">
        Something went wrong. It looks like you are trying to access a group
        that does not exist.
      </div>
    )
  }

  return (
    <div className="h-full min-h-0 px-10 py-5">
      <AnimatePresence mode="wait">
        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          exit={{ opacity: 0 }}
          transition={{ duration: 0.3, ease: 'easeInOut' }}
          className="absolute inset-0 mt-5 flex h-full min-h-0 flex-col gap-6 overflow-auto px-8 [scrollbar-gutter:stable_both-edges] [&>*:last-child]:pb-8"
        >
          {childPhaseCounts && (
            <TabSection
              heading={<GroupDataHeading displayText={groupData.name} />}
            >
              <ActionFanoutTable childPhaseCounts={childPhaseCounts} />
            </TabSection>
          )}

          {failed.length > 0 && (
            <TabSection heading="Failed actions">
              <GroupMiniTable actionType="failed" data={failed} />
            </TabSection>
          )}

          {longest.length > 2 && (
            <TabSection heading="Longest running times">
              <GroupMiniTable actionType="longest running" data={longest} />
            </TabSection>
          )}
        </motion.div>
      </AnimatePresence>
    </div>
  )
}
