'use client'

import { memo, useMemo } from 'react'
import {
  CheckCircleIcon,
  XCircleIcon,
  ExclamationCircleIcon,
  ExclamationTriangleIcon,
} from '@heroicons/react/20/solid'
import { ActionPhase } from '@/gen/flyteidl2/common/phase_pb'
import { CircularProgressIcon } from '@/components/icons/CircularProgressIcon'
import { CircleQuestionSolidIcon } from '@/components/icons/CircleQuestionSolidIcon'
import { Tooltip } from '../Tooltip'
import { MultiEllipsisIcon } from '@/components/icons/MultiEllipsisIcon'
import { IconSize, iconSizeMap } from './iconSize'
import { mapPhaseToDisplayString } from '@/lib/mapPhaseToDisplayString'
import { getColorsByPhase } from '@/lib/getColorByPhase'
import { useAccent } from '@/hooks/usePalette'
import { Dot } from '../Dot'

interface StatusIconProps {
  className?: string
  iconSize?: IconSize
  phase?: ActionPhase
  isActive?: boolean
  taskType?: string | undefined
  disableTooltip?: boolean
  isStatic?: boolean
}

export const StatusIconComponent = ({
  className = '',
  iconSize = 'md',
  phase,
  isActive = false,
  taskType,
  disableTooltip,
  isStatic = false, // without animation
}: StatusIconProps) => {
  const iconSizeClass = iconSizeMap[iconSize]
  const iconClassName = `${iconSizeClass} ${className}`
  const color = getColorsByPhase(phase)
  const accent = useAccent(color)

  const innerContent = useMemo(() => {
    const wrapperClass = isActive
      ? `relative rounded flex items-center justify-center ${iconSizeClass}`
      : `flex items-center justify-center ${iconClassName} bg-[unset]`

    if (taskType === 'trace') {
      return (
        <div className={`${iconSizeClass} flex items-center justify-center`}>
          <Dot color={accent} />
        </div>
      )
    }

    switch (phase) {
      case ActionPhase.QUEUED:
        return (
          <div className={wrapperClass} style={{ color: accent }}>
            <MultiEllipsisIcon className={iconClassName} />
          </div>
        )
      case ActionPhase.WAITING_FOR_RESOURCES:
        return (
          <div className={wrapperClass} style={{ color: accent }}>
            <CircularProgressIcon
              className={iconClassName}
              isStatic={isStatic}
            />
          </div>
        )
      case ActionPhase.INITIALIZING:
        return (
          <div className={wrapperClass} style={{ color: accent }}>
            <CircularProgressIcon
              className={iconClassName}
              isStatic={isStatic}
            />
          </div>
        )
      case ActionPhase.RUNNING:
        return (
          <div className={wrapperClass} style={{ color: accent }}>
            <CircularProgressIcon
              className={iconClassName}
              isStatic={isStatic}
            />
          </div>
        )
      case ActionPhase.SUCCEEDED:
        return (
          <div className={wrapperClass} style={{ color: accent }}>
            <CheckCircleIcon className={iconClassName} />
          </div>
        )
      case ActionPhase.FAILED:
        return (
          <div className={wrapperClass} style={{ color: accent }}>
            <ExclamationCircleIcon className={iconClassName} />
          </div>
        )
      case ActionPhase.ABORTED:
        return (
          <div className={wrapperClass} style={{ color: accent }}>
            <XCircleIcon className={iconClassName} />
          </div>
        )
      case ActionPhase.TIMED_OUT:
        return (
          <div className={wrapperClass} style={{ color: accent }}>
            <ExclamationTriangleIcon className={iconClassName} />
          </div>
        )
      case ActionPhase.UNSPECIFIED:
      default:
        return (
          <div className={wrapperClass} style={{ color: accent }}>
            <CircleQuestionSolidIcon className={iconClassName} />
          </div>
        )
    }
  }, [
    accent,
    iconSizeClass,
    iconClassName,
    isActive,
    isStatic,
    phase,
    taskType,
  ])

  if (disableTooltip) {
    return innerContent
  }

  return phase ? (
    <Tooltip
      content={mapPhaseToDisplayString[phase]}
      openDelay={700}
      placement="bottom"
    >
      {innerContent}
    </Tooltip>
  ) : (
    innerContent
  )
}

const areEqual = (prevProps: StatusIconProps, nextProps: StatusIconProps) =>
  prevProps.phase === nextProps.phase

export const StatusIcon = memo(StatusIconComponent, areEqual)