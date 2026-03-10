import { useMemo, forwardRef, useCallback } from 'react'
import clsx from 'clsx'
import { AnimatePresence, motion } from 'motion/react'
import { Button } from '@/components/Button'
import { useQueryFilters } from '@/hooks/useQueryFilters'
import { ActionPhase } from '@/gen/flyteidl2/common/phase_pb'
import { useAccent } from '@/hooks/usePalette'
import { filterConfigs } from '@/components/StatusFilter'
import { StatusIcon } from '@/components/StatusIcons'
import { getColorsByPhase } from '@/lib/getColorByPhase'
import { formatNumber } from '@/lib/numberUtils'
import { Tooltip } from '@/components/Tooltip'
import {
  getPhaseEnumValue,
  getPhaseFroπmEnum,
  getPhaseString,
  type PhaseKey,
} from '@/lib/phaseUtils'

export const StatusBadgeFilter = forwardRef<
  HTMLDivElement,
  {
    displayText?: string
    forceVisible?: boolean
    handleClickCb?: () => void
    phaseCount: number
    phase: ActionPhase | undefined
  }
>(({ displayText, forceVisible, handleClickCb, phaseCount, phase }, ref) => {
  const filterConfig = filterConfigs.find((f) => f.phase === phase)
  const filterConfigValue = filterConfig?.value || getPhaseFroπmEnum(phase)
  const { addStatusFilters, filters, clearFilter, toggleFilter } =
    useQueryFilters()
  const color = getColorsByPhase(phase)
  const accent = useAccent(color)

  const isFiltered = filterConfigValue
    ? filters.status?.includes(filterConfigValue)
    : false
  // Toggle only this filter: if others are active and this isn't, clear and activate this.
  // If others are active and this is, clear this one.
  // If no filters are active, toggle it.
  const toggleFilterCb = useCallback(() => {
    if (!filterConfig) return
    const isActive = filters.status?.includes(filterConfig.value)
    if (filters.status?.length) {
      if (isActive) {
        // Remove only this filter
        toggleFilter({ type: 'status', status: filterConfig.value })
      } else {
        // Clear all and activate only this filter
        clearFilter()
        addStatusFilters([filterConfig.value])
      }
    } else {
      // No filters active, toggle this one
      toggleFilter({ type: 'status', status: filterConfig.value })
    }
  }, [addStatusFilters, clearFilter, filterConfig, filters, toggleFilter])

  const handleClick = handleClickCb || toggleFilterCb

  const shouldShow = useMemo(() => {
    if (forceVisible) return true
    if (isFiltered) return true
    if (phaseCount > 0) return true
    return false
  }, [forceVisible, isFiltered, phaseCount])

  if (!shouldShow) {
    return null
  }

  return (
    <AnimatePresence>
      {
        <motion.div
          ref={ref}
          layout
          key={filterConfig?.value}
          initial={{ opacity: 0, scale: 0.95 }}
          animate={{
            opacity: 1,
            scale: 1,
          }}
          exit={{ opacity: 0, scale: 0.95 }}
          transition={{
            duration: 0.2,
            opacity: { duration: 0.15 },
            scale: { duration: 0.15 },
          }}
          style={{ overflow: 'hidden', minWidth: 65 }}
        >
          <Button
            className={clsx(
              'flex h-[26px] items-center !p-1',
              // Active state: subtle gray bg in light, transparent in dark (same look inactive had)
              isFiltered &&
                'rounded-md [--btn-bg:var(--system-gray-2)] [--btn-border:var(--system-gray-4)] dark:border-transparent dark:[--btn-bg:transparent] dark:[--btn-border:transparent]',
            )}
            onClick={handleClick}
            plain={!isFiltered}
            size="sm"
          >
            <StatusIcon disableTooltip={true} phase={phase} />
            <span style={{ color: accent, display: 'inline-block' }}>
              {displayText}
            </span>
            <span
              className="font-mono"
              style={{ color: accent, display: 'inline-block' }}
            >
              {formatNumber(phaseCount)}
            </span>
          </Button>
        </motion.div>
      }
    </AnimatePresence>
  )
})

const SetupItem = ({
  count,
  label,
  onClick,
  phase,
}: {
  count: number
  label: string
  onClick: () => void
  phase: ActionPhase
}) => (
  <div
    className="h-6 cursor-pointer px-2 text-[12px] hover:bg-(--system-gray-2) dark:hover:bg-(--system-gray-4)"
    onClick={onClick}
  >
    <div className="flex items-center justify-between">
      <div className="flex items-center gap-1">
        <StatusIcon phase={phase} isStatic disableTooltip />
        <span className={clsx('text-(--system-gray-5)')}>{label}</span>
      </div>
      <div className="text-(--accent-purple)">{count}</div>
    </div>
  </div>
)

StatusBadgeFilter.displayName = 'StatusBadgeFilter'

const setupStatuses: PhaseKey[] = [
  'QUEUED',
  'INITIALIZING',
  'WAITING_FOR_RESOURCES',
]

export const SetupBadgeFilter = ({
  phaseCounts,
}: {
  phaseCounts: Record<ActionPhase, number>
}) => {
  const { addStatusFilters, filters, removeStatusFilters, toggleFilter } =
    useQueryFilters()
  const allsetupCounts = useMemo(
    () =>
      phaseCounts[ActionPhase.INITIALIZING] +
      phaseCounts[ActionPhase.QUEUED] +
      phaseCounts[ActionPhase.WAITING_FOR_RESOURCES],
    [phaseCounts],
  )

  const isAnySetupActive = filters.status?.some((s) =>
    setupStatuses.includes(s as PhaseKey),
  )

  const areAllSetupActive = setupStatuses.every((s) => {
    if (!filters || !filters.status) return false
    const statusArr = filters.status.filter((x): x is PhaseKey => x != null)
    return statusArr.includes(s)
  })

  const { displayString, count, phase } = useMemo(() => {
    const { status = [] } = filters
    if (areAllSetupActive) {
      return {
        count: allsetupCounts,
        displayString: '',
        phase: ActionPhase.QUEUED,
      }
    }
    const activePhase = setupStatuses.find((s) => s && status?.includes(s))
    const activePhaseEnum = getPhaseEnumValue(activePhase)
    return {
      count: activePhaseEnum ? phaseCounts[activePhaseEnum] : allsetupCounts,
      displayString: activePhaseEnum ? getPhaseString(activePhaseEnum) : '',
      phase: activePhaseEnum,
    }
  }, [allsetupCounts, areAllSetupActive, filters, phaseCounts])

  const toggleAllSetupFilters = useCallback(() => {
    if (isAnySetupActive) {
      removeStatusFilters(setupStatuses)
    } else {
      addStatusFilters(setupStatuses)
    }
  }, [addStatusFilters, isAnySetupActive, removeStatusFilters])

  const tooltipContent = (
    <div className="flex w-54 flex-col rounded-md bg-(--system-gray-1)">
      <SetupItem
        count={allsetupCounts}
        label="Set up"
        onClick={toggleAllSetupFilters}
        phase={ActionPhase.QUEUED}
      />
      <hr className="mt-2 text-(--system-gray-2)" />
      <SetupItem
        count={phaseCounts[ActionPhase.QUEUED]}
        label="Queued"
        onClick={() => toggleFilter({ type: 'status', status: 'QUEUED' })}
        phase={ActionPhase.QUEUED}
      />

      <SetupItem
        count={phaseCounts[ActionPhase.INITIALIZING]}
        label="Initializing"
        onClick={() => toggleFilter({ type: 'status', status: 'INITIALIZING' })}
        phase={ActionPhase.INITIALIZING}
      />
      <SetupItem
        count={phaseCounts[ActionPhase.WAITING_FOR_RESOURCES]}
        label="Waiting for resources"
        onClick={() =>
          toggleFilter({ type: 'status', status: 'WAITING_FOR_RESOURCES' })
        }
        phase={ActionPhase.WAITING_FOR_RESOURCES}
      />
    </div>
  )

  return (
    <div>
      <Tooltip
        content={tooltipContent}
        disabled={isAnySetupActive && !areAllSetupActive}
        placement="bottom-start"
        offsetProp={{ crossAxis: -5, mainAxis: -1 }}
      >
        <StatusBadgeFilter
          displayText={displayString}
          forceVisible={allsetupCounts > 0}
          handleClickCb={toggleAllSetupFilters}
          phase={phase || ActionPhase.QUEUED}
          phaseCount={count}
        />
      </Tooltip>
    </div>
  )
}
