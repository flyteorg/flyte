import { Timestamp } from '@/gen/google/protobuf/timestamp_pb'
import { ActionPhase } from '@/gen/flyteidl2/common/phase_pb'
import { useActionBarMetrics } from './actionBarHelpers'
import { useAccent } from '@/hooks/usePalette'
import { getColorsByPhase } from '@/lib/getColorByPhase'

export const ActionBar = ({
  endTime,
  isGroup,
  phase,
  startTime,
}: {
  endTime: Timestamp | undefined
  isGroup: boolean
  phase: ActionPhase | undefined
  startTime: Timestamp | undefined
}) => {
  const colorString = getColorsByPhase(phase)
  let color = useAccent(colorString)
  if (isGroup) {
    color = `var(--accent-gray)`
  }

  const { offsetPercent, widthPercent } = useActionBarMetrics({
    startTime,
    endTime,
  })
  if (phase === ActionPhase.UNSPECIFIED) {
    return null
  }

  return (
    <>
      <div className="relative h-[4px] w-full bg-transparent">
        <div
          className="absolute top-0 h-[4px] rounded-full"
          style={{
            left: `${offsetPercent}%`,
            width: `${widthPercent}%`,
            background: color,
          }}
        />
      </div>
    </>
  )
}
