import { PhaseTransition } from '@/gen/flyteidl2/workflow/run_definition_pb'
import { ActionPhase } from '@/gen/flyteidl2/common/phase_pb'
import { useAccent } from '@/hooks/usePalette'
import { getRunningTime, toDateFormat } from '@/lib/dateUtils'
import { getColorsByPhase } from '@/lib/getColorByPhase'
import { mapPhaseToDisplayString } from '@/lib/mapPhaseToDisplayString'
import { getEarliestLatest } from './helpers'
import { TooltipPhaseSection, TooltipSection } from './types'

export const makeTooltipSection = (
  phaseTransition: PhaseTransition,
  colorPhase?: ActionPhase,
): TooltipSection => {
  return {
    colorPhase: colorPhase || phaseTransition.phase,
    phase: phaseTransition.phase,
    duration: getRunningTime({
      timestamp: phaseTransition.startTime,
      endTimestamp: phaseTransition.endTime,
      showSubSecondPrecision: true,
    }),
    phaseTransition,
    type: 'phase',
    key: mapPhaseToDisplayString[phaseTransition.phase],
  }
}

const PhaseTimelineTooltip = ({
  tooltipPhaseSection: { key, duration, phase },
}: {
  tooltipPhaseSection: TooltipPhaseSection
}) => {
  const color = getColorsByPhase(phase)
  const accent = useAccent(color)
  return (
    <div className="mb-1.5 text-2xs">
      <span style={{ color: accent }}>{key}:</span>
      <span className="ml-1 dark:text-(--system-gray-5)">{duration}</span>
    </div>
  )
}

const SummarySection = ({
  tooltipSections,
}: {
  tooltipSections: TooltipPhaseSection[]
}) => {
  if (!tooltipSections.length) return null
  const phaseTransitions = tooltipSections.map(
    (section) => section.phaseTransition,
  )
  const range = getEarliestLatest(phaseTransitions)
  const duration = getRunningTime({
    endTimestamp: range.endTime,
    timestamp: range.startTime,
    showSubSecondPrecision: true,
  })

  return (
    <div className="text-2xs dark:text-(--system-gray-5)">
      <div>Total Duration: {duration}</div>
      {range.startTime && (
        <div>Start Time: {toDateFormat({ timestamp: range.startTime })}</div>
      )}
      {range.endTime && (
        <div>Stop Time: {toDateFormat({ timestamp: range.endTime })}</div>
      )}
    </div>
  )
}

export const TimelineTooltip = ({
  tooltipSections,
}: {
  tooltipSections: TooltipSection[]
}) => {
  const genericSections = tooltipSections.filter((p) => p.type === 'generic')
  const phaseSections = tooltipSections.filter((p) => p.type === 'phase')

  return (
    <div className="rounded-lg bg-(--system-white) p-2 text-zinc-500 shadow-[0px_8px_8px_0px_#00000066] dark:bg-(--system-gray-1)">
      {genericSections.map((s) => (
        <span key={s.key}>{s.content}</span>
      ))}
      {phaseSections.length > 0 ? (
        <div className="p-2">
          {phaseSections.map((s) => (
            <PhaseTimelineTooltip key={s.key} tooltipPhaseSection={s} />
          ))}
          <SummarySection tooltipSections={phaseSections} />
        </div>
      ) : null}
    </div>
  )
}