'use client'

import React from 'react'
import { TagIcon } from '@heroicons/react/20/solid'
import { ActionMetadata } from '@/gen/flyteidl2/workflow/run_definition_pb'

interface TaskTypeProps {
  type?: string
  withIcon?: boolean
}

export const TaskType = ({ type, withIcon = true }: TaskTypeProps) => {
  return (
    <div className="inline-flex items-center gap-1">
      {withIcon && <TagIcon className="h-4 w-4" />}
      <span className="text-[14px] leading-5 font-normal tracking-[0px]">
        {type}
      </span>
    </div>
  )
}

export const TaskIcon = ({ metadata }: { metadata?: ActionMetadata }) => {
  if (metadata?.spec.case === 'task') {
    return (
      <div className="flex h-6 items-center gap-1 text-xs font-medium opacity-40 dark:text-white">
        <svg
          xmlns="http://www.w3.org/2000/svg"
          width="14"
          height="14"
          viewBox="0 0 14 14"
          fill="none"
        >
          <path
            d="M5.71429 2.71429H6.57143M4.42857 4V1.85714C4.42857 1.38376 4.81233 1 5.28571 1H8.71429C9.18769 1 9.57143 1.38376 9.57143 1.85714V6.14286C9.57143 6.61624 9.18769 7 8.71429 7H5.28571C4.81233 7 4.42857 7.38376 4.42857 7.85714V12.1429C4.42857 12.6163 4.81233 13 5.28571 13H8.71429C9.18769 13 9.57143 12.6163 9.57143 12.1429V10M7.42857 4.42857H1.85714C1.38376 4.42857 1 4.81233 1 5.28571V9.57143C1 10.0448 1.38376 10.4286 1.85714 10.4286H4.42857M6.57143 9.57143H12.1429C12.6163 9.57143 13 9.18769 13 8.71429V4.42857C13 3.95519 12.6163 3.57143 12.1429 3.57143H9.57143M7.42857 11.2857H8.28571"
            stroke="white"
            strokeWidth="1.4"
          />
        </svg>
        Python
      </div>
    )
  }
  return ''
}
