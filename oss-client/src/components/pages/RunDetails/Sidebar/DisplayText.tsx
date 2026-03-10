import React from 'react'
import clsx from 'clsx'
import { ActionPhase } from '@/gen/flyteidl2/common/phase_pb'

export const DisplayText = ({
  displayText,
  isSelected,
  maxWidth,
  phase,
}: {
  displayText: string | undefined
  isSelected: boolean
  maxWidth?: number
  phase: ActionPhase | null | undefined
}) => (
  <span
    className={clsx(
      'min-w-0 overflow-hidden text-sm text-ellipsis whitespace-nowrap',
      'ml-1',
      {
        'text-(--accent-red)': phase === ActionPhase.FAILED,
        'font-extrabold text-zinc-900 dark:text-white':
          phase !== ActionPhase.FAILED && isSelected,
        'font-medium text-zinc-700 dark:text-[#E2E2E2]':
          phase !== ActionPhase.FAILED && !isSelected,
      },
    )}
    data-testid="timeline-action-label"
    style={{
      maxWidth: `${maxWidth}px`,
      width: 'fit-content',
    }}
  >
    {displayText}
  </span>
)
