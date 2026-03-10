import { Button } from '@/components/Button'
import { CollapseIcon } from '@/components/icons/CollapseIcon'
import { ExpandIcon } from '@/components/icons/ExpandIcon'
import { MenuItem, PopoverMenu } from '@/components/Popovers'
import { useStatusFilterMenuItems } from '@/components/StatusFilter'
import { Tooltip } from '@/components/Tooltip'
import { ActionPhase } from '@/gen/flyteidl2/common/phase_pb'
import { EnrichedAction } from '@/gen/flyteidl2/workflow/run_definition_pb'
import { AnimatePresence, motion } from 'motion/react'
import { useEffect, useMemo } from 'react'
import { useRunStore } from '../state/RunStore'
import { SearchControl } from './SearchControl'
import { SetupBadgeFilter, StatusBadgeFilter } from './StatusBadgeFilter'

const getPhaseCount = (
  action: EnrichedAction | null | undefined,
  phase: ActionPhase,
) => {
  if (!action) return 0
  const parentCount = action.action?.status?.phase === phase ? 1 : 0
  const childCount = action.childrenPhaseCounts[phase] || 0
  return childCount + parentCount
}

export const SidebarControls = () => {
  const run = useRunStore((s) => s.run)

  const flatItems = useRunStore((s) => s.flatItems)
  const toggleExpandAll = useRunStore((s) => s.toggleCollapseAll)
  const collapsedItems = useRunStore((s) => s.collapsedItems)
  const hasCollapsedItems = collapsedItems.size > 0
  const shouldDisableToggle = flatItems.length === 0
  const expandCollapseLabel = useMemo(
    () => `${collapsedItems.size > 0 ? 'Expand' : 'Collapse'} All Items`,
    [collapsedItems],
  )
  const statusMenuItems = useStatusFilterMenuItems()

  const items: MenuItem[] = useMemo(() => {
    return [
      {
        disabled: shouldDisableToggle,
        id: 'expand-collapse',
        type: 'item',
        label: expandCollapseLabel,
        onClick: toggleExpandAll,
      },

      ...statusMenuItems,
    ]
  }, [
    expandCollapseLabel,
    shouldDisableToggle,
    statusMenuItems,
    toggleExpandAll,
  ])

  const phaseCounts: Record<ActionPhase, number> = useMemo(
    () => ({
      [ActionPhase.UNSPECIFIED]: 0,
      [ActionPhase.ABORTED]: getPhaseCount(run, ActionPhase.ABORTED),
      [ActionPhase.FAILED]: getPhaseCount(run, ActionPhase.FAILED),
      [ActionPhase.INITIALIZING]: getPhaseCount(run, ActionPhase.INITIALIZING),
      [ActionPhase.QUEUED]: getPhaseCount(run, ActionPhase.QUEUED),
      [ActionPhase.RUNNING]: getPhaseCount(run, ActionPhase.RUNNING),
      [ActionPhase.SUCCEEDED]: getPhaseCount(run, ActionPhase.SUCCEEDED),
      [ActionPhase.TIMED_OUT]: getPhaseCount(run, ActionPhase.TIMED_OUT),
      [ActionPhase.WAITING_FOR_RESOURCES]: getPhaseCount(
        run,
        ActionPhase.WAITING_FOR_RESOURCES,
      ),
    }),
    [run],
  )

  useEffect(() => {
    return () => {
      Object.keys(phaseCounts).forEach((key) => {
        phaseCounts[key as unknown as ActionPhase] = 0
      })
    }
  }, [phaseCounts])

  const priorityOrder = [
    ActionPhase.SUCCEEDED,
    ActionPhase.FAILED,
    ActionPhase.ABORTED,
    ActionPhase.TIMED_OUT,
    ActionPhase.RUNNING,
  ]

  return (
    <div className="flex flex-col gap-3">
      <SearchControl />
      <div className="flex items-center justify-between">
        <div className="flex h-7 flex-1 gap-1 text-xs font-semibold text-(--system-gray-6)">
          Filter:
          <div className="flex flex-1">
            <motion.div className="flex items-center gap-1" layout="position">
              <AnimatePresence>
                {priorityOrder.map((phase) => (
                  <StatusBadgeFilter
                    key={phase}
                    phaseCount={phaseCounts[phase] ?? 0}
                    phase={phase}
                  />
                ))}
                <SetupBadgeFilter phaseCounts={phaseCounts} />
              </AnimatePresence>
            </motion.div>
          </div>
          <div>
            <Tooltip
              content={
                <div className="px-2 py-1 text-2xs/4">
                  {expandCollapseLabel}
                </div>
              }
              placement="bottom"
            >
              <Button
                {...(shouldDisableToggle && {
                  'data-disabled': shouldDisableToggle,
                })}
                disabled={shouldDisableToggle}
                size="xxs"
                plain={true}
                onClick={toggleExpandAll}
                className="items-center !p-0 text-sm !text-(--system-gray-1) dark:!text-(--system-gray-5)"
              >
                {hasCollapsedItems ? (
                  <ExpandIcon height={17} />
                ) : (
                  <CollapseIcon height={17} />
                )}
              </Button>
            </Tooltip>
            <PopoverMenu
              closeOnItemClick={false}
              items={items}
              overflowProps={{ orientation: 'horizontal' }}
              variant="overflow"
            />
          </div>
        </div>
      </div>
    </div>
  )
}
