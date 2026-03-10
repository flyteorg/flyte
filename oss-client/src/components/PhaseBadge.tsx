import { StatusIcon } from '@/components/StatusIcons'
import { ActionPhase } from '@/gen/flyteidl2/common/phase_pb'
import { getPhaseClass, getPhaseString } from '@/lib/phaseUtils'

interface PhaseBadgeProps {
  phase: ActionPhase | undefined
}

export function PhaseBadge({ phase }: PhaseBadgeProps) {
  return (
    <div
      className={`flex items-center gap-2 rounded-lg px-2 py-1 ${getPhaseClass(phase)}`}
    >
      <StatusIcon disableTooltip={true} phase={phase} />
      <span className="text-2xs leading-4 font-semibold">
        {getPhaseString(phase)}
      </span>
    </div>
  )
}
