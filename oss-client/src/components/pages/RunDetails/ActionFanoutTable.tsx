import { useEffect, useState } from 'react'
import isEqual from 'lodash/isEqual'
import { ChildPhaseCounts } from './types'
import { getChildCount } from './Sidebar/util'
import { ActionPhase } from '@/gen/flyteidl2/common/phase_pb'
import { mapPhaseToDisplayString } from '@/lib/mapPhaseToDisplayString'

const Cell = ({
  label,
  value,
  valueColor,
  percent,
}: {
  label: string
  value: number
  valueColor: string
  percent?: number | null
}) => {
  return (
    <div className="flex justify-between" data-testid="subaction-cell">
      <div className="text-(--system-gray-5)">{label}</div>
      <div className="flex gap-1">
        {value ? (
          <span style={{ color: valueColor }}>{value?.toLocaleString()}</span>
        ) : (
          <span>-</span>
        )}
        {percent != null && <span>({percent.toFixed(1)}%)</span>}
      </div>
    </div>
  )
}

const Divider = () => <hr className="text-(--system-gray-3)" />

export const ActionFanoutTable = ({
  childPhaseCounts,
  isRunTerminal,
}: {
  childPhaseCounts: ChildPhaseCounts
  isRunTerminal?: boolean
}) => {
  const [displayCounts, setDisplayCounts] = useState(childPhaseCounts)
  const total = getChildCount(childPhaseCounts)

  useEffect(() => {
    if (isRunTerminal) {
      // run is done → show final counts immediately & stop
      setDisplayCounts(childPhaseCounts)
      return
    }

    const id = setInterval(() => {
      setDisplayCounts((prev) => {
        if (!isEqual(prev, childPhaseCounts)) {
          return childPhaseCounts
        }
        return prev
      })
    }, 50)

    return () => clearInterval(id)
  }, [childPhaseCounts, isRunTerminal])

  const percentComplete =
    (childPhaseCounts[ActionPhase.SUCCEEDED] / total) * 100 || null
  const percentFailed =
    (childPhaseCounts[ActionPhase.FAILED] / total) * 100 || null
  const percentRunning =
    (childPhaseCounts[ActionPhase.RUNNING] / total) * 100 || null

  const shouldShowRareStatusCol =
    childPhaseCounts[ActionPhase.TIMED_OUT] ||
    childPhaseCounts[ActionPhase.ABORTED]

  return (
    <div className="flex flex-col gap-2 px-5 py-4 text-xs font-semibold text-(--system-gray-5)">
      <div className="flex justify-between">
        <div>Total</div>

        {total.toLocaleString()}
      </div>
      <Divider />
      <div className="flex gap-5">
        {/* col-init */}
        <div className="flex-1">
          <div className="flex flex-col gap-1">
            <Cell
              label={mapPhaseToDisplayString[ActionPhase.INITIALIZING]}
              value={displayCounts[ActionPhase.INITIALIZING]}
              valueColor="var(--accent-purple)"
            />
            <Divider />
            <Cell
              label={mapPhaseToDisplayString[ActionPhase.WAITING_FOR_RESOURCES]}
              value={displayCounts[ActionPhase.WAITING_FOR_RESOURCES]}
              valueColor="var(--accent-purple)"
            />
            <Divider />
            <Cell
              label={mapPhaseToDisplayString[ActionPhase.QUEUED]}
              value={displayCounts[ActionPhase.QUEUED]}
              valueColor="var(--accent-purple)"
            />
          </div>
        </div>
        {/* col-running-completed-failed */}
        <div className="flex-1">
          <div className="flex flex-col gap-1">
            <Cell
              label={mapPhaseToDisplayString[ActionPhase.RUNNING]}
              percent={percentRunning}
              valueColor="var(--accent-blue)"
              value={displayCounts[ActionPhase.RUNNING]}
            />
            <Divider />
            <Cell
              label={mapPhaseToDisplayString[ActionPhase.SUCCEEDED]}
              percent={percentComplete}
              valueColor="var(--accent-green)"
              value={displayCounts[ActionPhase.SUCCEEDED]}
            />
            <Divider />
            <Cell
              label={mapPhaseToDisplayString[ActionPhase.FAILED]}
              percent={percentFailed}
              valueColor="var(--accent-red)"
              value={displayCounts[ActionPhase.FAILED]}
            />
          </div>
        </div>
        {shouldShowRareStatusCol && (
          <div className="flex-1">
            <div className="flex flex-col gap-1">
              <Cell
                label={mapPhaseToDisplayString[ActionPhase.TIMED_OUT]}
                valueColor="var(--accent-yellow)"
                value={displayCounts[ActionPhase.TIMED_OUT]}
              />
              <Divider />
              <Cell
                label={mapPhaseToDisplayString[ActionPhase.ABORTED]}
                valueColor="var(--accent-orange)"
                value={displayCounts[ActionPhase.ABORTED]}
              />
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
