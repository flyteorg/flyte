import React from 'react'
import { ActionPhase } from '@/gen/flyteidl2/common/phase_pb'
import {
  NotStartedFolderIcon,
  QueuedFolder,
  InitializingFolderIcon,
  RunningFolderIcon,
  TimedOutFolderIcon,
  AbortedFolderIcon,
  ErrorFolderIcon,
  NoStatusFolder,
  SuccessStatusFolder,
} from './icons/GroupFolderIcon'

export const FolderGroup = ({
  phase,
  props,
}: {
  phase?: ActionPhase | null
  props?: React.ComponentPropsWithoutRef<'svg'>
}) => {
  switch (phase) {
    case ActionPhase.UNSPECIFIED: {
      return <NotStartedFolderIcon {...props} />
    }
    case ActionPhase.INITIALIZING: {
      return <InitializingFolderIcon {...props} />
    }
    case ActionPhase.QUEUED:
    case ActionPhase.WAITING_FOR_RESOURCES: {
      return <QueuedFolder {...props} />
    }

    case ActionPhase.RUNNING: {
      return <RunningFolderIcon {...props} />
    }
    case ActionPhase.TIMED_OUT: {
      return <TimedOutFolderIcon {...props} />
    }
    case ActionPhase.ABORTED: {
      return <AbortedFolderIcon {...props} />
    }
    case ActionPhase.FAILED: {
      return <ErrorFolderIcon {...props} />
    }
    case ActionPhase.SUCCEEDED:
      return <SuccessStatusFolder {...props} />
    default: {
      return <NoStatusFolder {...props} />
    }
  }
}