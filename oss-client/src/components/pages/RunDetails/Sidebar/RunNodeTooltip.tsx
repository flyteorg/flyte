import { Timestamp } from '@bufbuild/protobuf/wkt'
import clsx from 'clsx'
import { ActionPhase } from '@/gen/flyteidl2/common/phase_pb'
import { ArrowsRightIcon } from '@/components/icons/ArrowsRightIcon'
import { toDateFormat } from '@/lib/dateUtils'
import { mapPhaseToDisplayString } from '@/lib/mapPhaseToDisplayString'
import { useAccent } from '@/hooks/usePalette'
import { getColorsByPhase } from '@/lib/getColorByPhase'
import { CacheStatusIcon, StatusIcon } from '@/components/StatusIcons'
import {
  CacheErrorStatuses,
  mapCacheStatusToDisplayString,
} from '@/components/StatusIcons/CacheStatusIcon'
import { useMemo } from 'react'

import { useEffect, useState } from 'react'
import { ActionWithChildren } from '../state/types'
import { usePhaseCounts } from './usePhaseCounts'
import { ActionBadge } from './ActionBadge'
import { getChildCount } from './util'
import { CatalogCacheStatus } from '@/gen/flyteidl2/core/catalog_pb'

function useDelayedVisibility(value: number, delayMs: number) {
  const [visible, setVisible] = useState(value > 0)

  useEffect(() => {
    if (value > 0) {
      setVisible(true)
      return
    }

    const timer = setTimeout(() => setVisible(false), delayMs)
    return () => clearTimeout(timer)
  }, [value, delayMs])

  return visible
}

const SecondaryText = ({ children }: { children: React.ReactNode }) => (
  <div className="dark:text-(--system-gray-5)">{children}</div>
)

const PhaseCountDisplay = ({
  count,
  phase,
}: {
  count: number
  phase: ActionPhase
}) => {
  const color = getColorsByPhase(phase)
  const accentColor = useAccent(color)
  const isVisible = useDelayedVisibility(count, 3000)

  return (
    <div
      className={clsx(
        'flex items-center gap-1 overflow-hidden transition-all duration-500',
        isVisible ? 'max-h-6 opacity-100' : 'max-h-0 opacity-0',
      )}
      style={{ color: accentColor }}
    >
      <StatusIcon phase={phase} /> {count} {mapPhaseToDisplayString[phase]}
    </div>
  )
}

const CacheStatusDisplay = ({
  cacheStatus,
}: {
  cacheStatus: CatalogCacheStatus
}) => {
  const cacheStatusText = useMemo(
    () => `Cache ${mapCacheStatusToDisplayString[cacheStatus]}`,
    [cacheStatus],
  )
  const cacheClassName = useMemo(
    () =>
      CacheErrorStatuses.includes(cacheStatus)
        ? 'text-orange-400'
        : 'text-[#777777]',
    [cacheStatus],
  )

  return (
    <div
      className={`flex flex-row items-center gap-x-1 text-2xs ${cacheClassName}`}
    >
      <CacheStatusIcon size="md" status={cacheStatus} />
      {cacheStatusText}
    </div>
  )
}

export const RunNodeTooltip = ({
  attempts,
  endTime,
  startTime,
  taskName = 'task',
  cacheStatus,
  node,
}: {
  attempts: number | undefined
  endTime: Timestamp | undefined
  startTime: Timestamp | undefined
  taskName: string | undefined
  cacheStatus: CatalogCacheStatus | undefined
  node: ActionWithChildren
}) => {
  const childrenPhaseCounts = usePhaseCounts({ node })
  return (
    <div className="rounded-md bg-white px-4 py-2 text-black dark:bg-(--system-gray-1) dark:text-white">
      <div className="py-1">{taskName}</div>
      <hr className="w-full border-b-1 border-(--system-gray-3)" />
      <div className="flex flex-col gap-1 py-2">
        <div className="flex items-center gap-2">
          <ActionBadge childCount={getChildCount(childrenPhaseCounts)} />
          Total Actions
        </div>
        {Object.entries(childrenPhaseCounts).map(([phaseKey, count]) => {
          const phase = Number(phaseKey) as ActionPhase
          return (
            <PhaseCountDisplay
              count={count}
              key={mapPhaseToDisplayString[phase]}
              phase={phase}
            />
          )
        })}
        {typeof attempts === 'number' && attempts > 1 && (
          <div className="flex items-center text-[#999]">
            {' '}
            <ArrowsRightIcon color="#999" />
            {attempts} attempts
          </div>
        )}
        {cacheStatus !== undefined && (
          <CacheStatusDisplay cacheStatus={cacheStatus} />
        )}
      </div>
      {startTime && (
        <SecondaryText>
          Start Time: {toDateFormat({ timestamp: startTime })}
        </SecondaryText>
      )}
      {endTime && (
        <SecondaryText>
          End Time: {toDateFormat({ timestamp: endTime })}
        </SecondaryText>
      )}
    </div>
  )
}
