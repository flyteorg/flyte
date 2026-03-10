import { useRef } from 'react'
import { useAccent, usePalette } from '@/hooks/usePalette'
import { PhaseTransition } from '@/gen/flyteidl2/workflow/run_definition_pb'
import { ActionPhase } from '@/gen/flyteidl2/common/phase_pb'
import { getColorsByPhase } from '@/lib/getColorByPhase'
import { mapPhaseToDisplayString } from '@/lib/mapPhaseToDisplayString'
import { LiveTimestamp } from '@/components/LiveTimestamp'
import { useGlobalNow } from '../state/GlobalTimestamp'
import { StatusIcon } from '@/components/StatusIcons'
import { type TimelineObject } from './types'
import { getIsRunning, getEarliestLatest, durationBetween } from './helpers'
import { makeTooltipSection } from './ActionPhaseTooltip'

const SetupAnnotation = ({
  color,
  status,
}: {
  color?: string
  status?: string | undefined
}) => (
  <div className="flex items-center gap-1 text-2xs">
    <span style={{ color: color }}>Setup: </span>
    {status && <span className="text-zinc-500">{status}</span>}
  </div>
)

export const useMakeTimelineData = ({
  phaseTransitions,
}: {
  phaseTransitions: PhaseTransition[]
}): TimelineObject[] => {
  const palette = usePalette()
  const isRunning = getIsRunning(phaseTransitions)
  const globalNow = useGlobalNow((s) => (isRunning ? s.now : null))
  const nowRef = useRef(Date.now())

  // If not running, keep last known time frozen
  if (globalNow) {
    nowRef.current = globalNow
  }

  const totalBounds = getEarliestLatest(phaseTransitions)
  const endTime = isRunning ? nowRef.current : totalBounds.endTime
  const totalDuration = durationBetween(totalBounds.startTime, endTime)

  const preRunPhases = phaseTransitions.filter((p) => p.phase < ActionPhase.RUNNING)
  const preRunBounds = getEarliestLatest(preRunPhases)
  const preRunDuration = durationBetween(
    preRunBounds.startTime,
    preRunBounds.endTime,
  )

  // If the total duration is 0, fall back to a default visual ratio.
  // Otherwise, calculate preRunPercentage proportionally, ensuring a minimum width.
  let preRunPercentage: number
  if (totalDuration <= 1) {
    preRunPercentage = 85
  } else {
    const computed = Math.floor((preRunDuration / totalDuration) * 100)
    preRunPercentage = Math.max(computed, 15)
  }

  const runningPhase = phaseTransitions.find((p) => p.phase === ActionPhase.RUNNING)

  const terminalPhase = phaseTransitions.find((p) => p.phase > ActionPhase.RUNNING)
  const terminalColor = getColorsByPhase(terminalPhase?.phase)
  const terminalAccentColor = useAccent(terminalColor)

  // 1. Unspecified
  if (
    phaseTransitions.length === 1 &&
    phaseTransitions[0].phase === ActionPhase.UNSPECIFIED
  ) {
    return [
      {
        leftAnnotation: 'Setup',
        rightAnnotation: '',
        accentColor: palette.accent.gray,
        percentage: '85%',
        tooltipSections: [
          { type: 'generic', content: 'Setup', key: 'generic' },
        ],
      },
      {
        leftAnnotation: 'Run',
        rightAnnotation: '',
        accentColor: palette.accent.gray,
        percentage: '15%',
        tooltipSections: [
          { type: 'generic', content: 'Not started', key: 'generic' },
        ],
      },
    ]
  }

  // 2. Has not reached RUNNING yet
  if (!phaseTransitions.some((p) => p.phase >= ActionPhase.RUNNING)) {
    const latestPrerunPhase = preRunPhases[preRunPhases.length - 1]
    return [
      {
        leftAnnotation: (
          <SetupAnnotation
            status={mapPhaseToDisplayString[latestPrerunPhase.phase]}
          ></SetupAnnotation>
        ),
        rightAnnotation: (
          <LiveTimestamp
            className="!text-right"
            minWidth={25}
            timestamp={preRunBounds.startTime}
            endTimestamp={preRunBounds.endTime}
          />
        ),
        accentColor: palette.accent.purple,
        percentage: '85%',
        tooltipSections: preRunPhases.map(makeTooltipSection),
      },
      {
        leftAnnotation: 'Run',
        rightAnnotation: '',
        accentColor: palette.accent.gray,
        percentage: '15%',
        tooltipSections: [
          { type: 'generic', content: 'Not started', key: 'generic' },
        ],
      },
    ]
  }

  // 3. RUNNING
  if (runningPhase && !terminalPhase) {
    const runningPhasePercent = 100 - preRunPercentage
    return [
      {
        leftAnnotation: 'Setup',
        rightAnnotation: (
          <LiveTimestamp
            className="!text-right"
            minWidth={25}
            timestamp={preRunBounds.startTime}
            endTimestamp={preRunBounds.endTime}
          />
        ),
        accentColor: palette.accent.purple,
        percentage: `${preRunPercentage}%`,
        tooltipSections: preRunPhases.map(makeTooltipSection),
      },
      {
        leftAnnotation: runningPhasePercent > 15 ? 'Run' : '',
        rightAnnotation: (
          <div className="flex items-center gap-x-1">
            <LiveTimestamp
              className="!text-right"
              minWidth={25}
              timestamp={runningPhase.startTime}
              endTimestamp={runningPhase.endTime}
            />
            <StatusIcon phase={ActionPhase.RUNNING} />
          </div>
        ),
        accentColor: palette.accent.blue,
        percentage: `${runningPhasePercent}%`,
        tooltipSections: [makeTooltipSection(runningPhase)],
      },
    ]
  }

  // 4. TERMINAL
  if (terminalPhase && runningPhase) {
    const terminalPhasePercent = Math.max(100 - preRunPercentage, 12)
    return [
      {
        leftAnnotation: 'Setup',
        rightAnnotation: preRunBounds.startTime ? (
          <LiveTimestamp
            className="!text-right"
            minWidth={25}
            timestamp={preRunBounds.startTime}
            endTimestamp={preRunBounds.endTime}
          />
        ) : (
          ''
        ),
        accentColor: palette.accent.purple,
        percentage: `${preRunPercentage}%`,
        tooltipSections: preRunPhases.map(makeTooltipSection),
      },
      {
        leftAnnotation:
          terminalPhasePercent > 15
            ? mapPhaseToDisplayString[terminalPhase.phase]
            : '',
        rightAnnotation: runningPhase ? (
          <div className="flex items-center gap-x-1">
            <LiveTimestamp
              className="!text-right"
              minWidth={25}
              timestamp={runningPhase.startTime}
              endTimestamp={runningPhase.endTime}
            />
            <StatusIcon phase={terminalPhase.phase} />
          </div>
        ) : (
          ''
        ),
        accentColor: terminalAccentColor,
        percentage: `${terminalPhasePercent}%`,
        tooltipSections: [
          makeTooltipSection(runningPhase, terminalPhase.phase),
        ],
      },
    ]
  }
  // Handles case where no prerun phases are returned
  if (phaseTransitions.length === 1) {
    const phase = phaseTransitions[0].phase
    const color = getColorsByPhase(phase)
    const accent = palette.accent[color]
    return [
      {
        leftAnnotation: mapPhaseToDisplayString[phase],
        rightAnnotation: '',
        accentColor: accent,
        percentage: '100%',
        tooltipSections: [
          {
            type: 'generic',
            content: `${mapPhaseToDisplayString[phase]}`,
            key: 'generic',
          },
        ],
      },
    ]
  } else if (terminalPhase) {
    const terminalPhasePercent = Math.max(100 - preRunPercentage, 12)
    return [
      {
        leftAnnotation: 'Setup',
        rightAnnotation: preRunBounds.startTime ? (
          <LiveTimestamp
            className="!text-right"
            minWidth={25}
            timestamp={preRunBounds.startTime}
            endTimestamp={preRunBounds.endTime}
          />
        ) : (
          ''
        ),
        accentColor: palette.accent.purple,
        percentage: `${preRunPercentage}%`,
        tooltipSections: preRunPhases.map(makeTooltipSection),
      },
      {
        leftAnnotation:
          terminalPhasePercent > 15
            ? mapPhaseToDisplayString[terminalPhase.phase]
            : '',
        rightAnnotation: runningPhase ? (
          <div className="flex items-center gap-x-1">
            <LiveTimestamp
              className="!text-right"
              minWidth={25}
              timestamp={runningPhase.startTime}
              endTimestamp={runningPhase.endTime}
            />
            <StatusIcon phase={terminalPhase.phase} />
          </div>
        ) : (
          mapPhaseToDisplayString[terminalPhase.phase]
        ),
        accentColor: terminalAccentColor,
        percentage: `${terminalPhasePercent}%`,
        tooltipSections: [makeTooltipSection(terminalPhase)],
      },
    ]
  }

  return []
}