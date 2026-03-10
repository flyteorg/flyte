'use client'

import React from 'react'
import { Run } from '@/gen/flyteidl2/workflow/run_definition_pb'
import { ActionPhase } from '@/gen/flyteidl2/common/phase_pb'
import type { Timestamp } from '@bufbuild/protobuf/wkt'
import dayjs from 'dayjs'
import { getPhaseClass, getPhaseString } from '@/lib/phaseUtils'
import clsx from 'clsx'
import { Tooltip } from '@/components/Tooltip'
import { getRunningTime } from '@/lib/dateUtils'

interface RunExecutionTimelineProps {
  runs: Run[]
  height?: number
  minRuns?: number
}

const getTimestamp = (timestamp: Timestamp | undefined): number => {
  if (!timestamp) return 0
  return dayjs.unix(Number(timestamp.seconds)).valueOf()
}

export const RunExecutionTimeline: React.FC<RunExecutionTimelineProps> = ({
  runs,
  height = 40,
  minRuns = 100,
}) => {
  // Create a local copy of runs, reverse them, and take only minRuns
  const localRuns = [...runs].reverse().slice(0, minRuns)

  // Find the maximum duration for scaling
  const maxDuration =
    Math.max(
      ...localRuns.map((run) => {
        const start = getTimestamp(run.action?.status?.startTime)
        const end = getTimestamp(run.action?.status?.endTime) || Date.now()
        return end - start
      }),
    ) || 1 // Prevent division by zero

  // Calculate how many placeholder runs we need
  const placeholderCount = Math.max(0, minRuns - localRuns.length)
  const placeholderRuns = Array(placeholderCount).fill(null)

  return (
    <div className="bg-component w-full rounded-xl p-6">
      <div className="mb-2 text-sm text-gray-400">
        Last {minRuns} executions
      </div>
      <div
        className="flex items-end space-x-1"
        style={{ height: `${height}px` }}
      >
        {placeholderRuns.map((_, index) => (
          <Tooltip
            key={`placeholder-${index}`}
            content={<div className="text-gray-300">No execution data</div>}
          >
            <div
              className={clsx(
                getPhaseClass(ActionPhase.UNSPECIFIED),
                'relative flex-1 cursor-default opacity-20',
              )}
              style={{
                height: '10%',
                minWidth: '4px',
              }}
            />
          </Tooltip>
        ))}
        {localRuns.map((run, index: number) => {
          const start = getTimestamp(run.action?.status?.startTime)
          const end = getTimestamp(run.action?.status?.endTime) || Date.now()
          const duration = end - start

          const heightPercent = Math.max(15, (duration / maxDuration) * 100) // Minimum 15% height
          const displayDuration = getRunningTime({
            timestamp: run.action?.status?.startTime,
            endTimestamp: run.action?.status?.endTime,
            maxUnits: 2,
          })
          const phase = run.action?.status?.phase || ActionPhase.UNSPECIFIED
          const phaseClass = getPhaseClass(phase)
          const runName = run.action?.id?.run?.name || 'Unnamed Run'

          return (
            <Tooltip
              key={`run-${index}`}
              content={
                <div className="px-4 py-2">
                  <div className="mb-1 font-medium dark:text-white">
                    {runName}
                  </div>
                  <div className="dark:text-zinc-100">
                    Duration: {displayDuration}
                  </div>
                  <div className="dark:text-zinc-100">
                    Phase: {getPhaseString(phase)}
                  </div>
                </div>
              }
              placement="bottom"
            >
              <div
                className={clsx(
                  phaseClass,
                  'relative flex-1 cursor-pointer transition-opacity hover:opacity-80',
                )}
                style={{
                  height: `${heightPercent}%`,
                  minWidth: '4px',
                }}
              />
            </Tooltip>
          )
        })}
      </div>
    </div>
  )
}