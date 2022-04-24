import {
  negativeTextColor,
  positiveTextColor,
  secondaryTextColor,
  statusColors,
} from 'components/Theme/constants';
import {
  CatalogCacheStatus,
  NodeExecutionPhase,
  TaskExecutionPhase,
  WorkflowExecutionPhase,
} from 'models/Execution/enums';
import { TaskType } from 'models/Task/constants';
import { ExecutionPhaseConstants, NodeExecutionDisplayType } from './types';

export const executionRefreshIntervalMs = 10000;
export const noLogsFoundString = 'No logs found';

/** Shared values for color/text/etc for each execution phase */
export const workflowExecutionPhaseConstants: {
  [key in WorkflowExecutionPhase]: ExecutionPhaseConstants;
} = {
  [WorkflowExecutionPhase.ABORTED]: {
    badgeColor: statusColors.SKIPPED,
    text: 'Aborted',
    textColor: negativeTextColor,
  },
  [WorkflowExecutionPhase.ABORTING]: {
    badgeColor: statusColors.SKIPPED,
    text: 'Aborting',
    textColor: negativeTextColor,
  },
  [WorkflowExecutionPhase.FAILING]: {
    badgeColor: statusColors.FAILURE,
    text: 'Failing',
    textColor: negativeTextColor,
  },
  [WorkflowExecutionPhase.FAILED]: {
    badgeColor: statusColors.FAILURE,
    text: 'Failed',
    textColor: negativeTextColor,
  },
  [WorkflowExecutionPhase.QUEUED]: {
    badgeColor: statusColors.QUEUED,
    text: 'Queued',
    textColor: secondaryTextColor,
  },
  [WorkflowExecutionPhase.RUNNING]: {
    badgeColor: statusColors.RUNNING,
    text: 'Running',
    textColor: secondaryTextColor,
  },
  [WorkflowExecutionPhase.SUCCEEDED]: {
    badgeColor: statusColors.SUCCESS,
    text: 'Succeeded',
    textColor: positiveTextColor,
  },
  [WorkflowExecutionPhase.SUCCEEDING]: {
    badgeColor: statusColors.SUCCESS,
    text: 'Succeeding',
    textColor: positiveTextColor,
  },
  [WorkflowExecutionPhase.TIMED_OUT]: {
    badgeColor: statusColors.FAILURE,
    text: 'Timed Out',
    textColor: negativeTextColor,
  },
  [WorkflowExecutionPhase.UNDEFINED]: {
    badgeColor: statusColors.UNKNOWN,
    text: 'Unknown',
    textColor: secondaryTextColor,
  },
};

/** Shared values for color/text/etc for each node execution phase */
export const nodeExecutionPhaseConstants: {
  [key in NodeExecutionPhase]: ExecutionPhaseConstants;
} = {
  [NodeExecutionPhase.ABORTED]: {
    badgeColor: statusColors.FAILURE,
    text: 'Aborted',
    textColor: negativeTextColor,
  },
  [NodeExecutionPhase.FAILING]: {
    badgeColor: statusColors.FAILURE,
    text: 'Failing',
    textColor: negativeTextColor,
  },
  [NodeExecutionPhase.FAILED]: {
    badgeColor: statusColors.FAILURE,
    text: 'Failed',
    textColor: negativeTextColor,
  },
  [NodeExecutionPhase.QUEUED]: {
    badgeColor: statusColors.RUNNING,
    text: 'Queued',
    textColor: secondaryTextColor,
  },
  [NodeExecutionPhase.RUNNING]: {
    badgeColor: statusColors.RUNNING,
    text: 'Running',
    textColor: secondaryTextColor,
  },
  [NodeExecutionPhase.DYNAMIC_RUNNING]: {
    badgeColor: statusColors.RUNNING,
    text: 'Running',
    textColor: secondaryTextColor,
  },
  [NodeExecutionPhase.SUCCEEDED]: {
    badgeColor: statusColors.SUCCESS,
    text: 'Succeeded',
    textColor: positiveTextColor,
  },
  [NodeExecutionPhase.TIMED_OUT]: {
    badgeColor: statusColors.FAILURE,
    text: 'Timed Out',
    textColor: negativeTextColor,
  },
  [NodeExecutionPhase.SKIPPED]: {
    badgeColor: statusColors.UNKNOWN,
    text: 'Skipped',
    textColor: secondaryTextColor,
  },
  [NodeExecutionPhase.RECOVERED]: {
    badgeColor: statusColors.SUCCESS,
    text: 'Recovered',
    textColor: positiveTextColor,
  },
  [NodeExecutionPhase.UNDEFINED]: {
    badgeColor: statusColors.UNKNOWN,
    text: 'Unknown',
    textColor: secondaryTextColor,
  },
};

/** Shared values for color/text/etc for each node execution phase */
export const taskExecutionPhaseConstants: {
  [key in TaskExecutionPhase]: ExecutionPhaseConstants;
} = {
  [TaskExecutionPhase.ABORTED]: {
    badgeColor: statusColors.FAILURE,
    text: 'Aborted',
    textColor: negativeTextColor,
  },
  [TaskExecutionPhase.FAILED]: {
    badgeColor: statusColors.FAILURE,
    text: 'Failed',
    textColor: negativeTextColor,
  },
  [TaskExecutionPhase.WAITING_FOR_RESOURCES]: {
    badgeColor: statusColors.RUNNING,
    text: 'Waiting',
    textColor: secondaryTextColor,
  },
  [TaskExecutionPhase.QUEUED]: {
    badgeColor: statusColors.RUNNING,
    text: 'Queued',
    textColor: secondaryTextColor,
  },
  [TaskExecutionPhase.INITIALIZING]: {
    badgeColor: statusColors.RUNNING,
    text: 'Initializing',
    textColor: secondaryTextColor,
  },
  [TaskExecutionPhase.RUNNING]: {
    badgeColor: statusColors.RUNNING,
    text: 'Running',
    textColor: secondaryTextColor,
  },
  [TaskExecutionPhase.SUCCEEDED]: {
    badgeColor: statusColors.SUCCESS,
    text: 'Succeeded',
    textColor: positiveTextColor,
  },
  [TaskExecutionPhase.UNDEFINED]: {
    badgeColor: statusColors.UNKNOWN,
    text: 'Unknown',
    textColor: secondaryTextColor,
  },
};

export const taskTypeToNodeExecutionDisplayType: {
  [k in TaskType]: NodeExecutionDisplayType;
} = {
  [TaskType.ARRAY]: NodeExecutionDisplayType.MapTask,
  [TaskType.BATCH_HIVE]: NodeExecutionDisplayType.BatchHiveTask,
  [TaskType.DYNAMIC]: NodeExecutionDisplayType.DynamicTask,
  [TaskType.HIVE]: NodeExecutionDisplayType.HiveTask,
  [TaskType.PYTHON]: NodeExecutionDisplayType.PythonTask,
  [TaskType.SIDECAR]: NodeExecutionDisplayType.SidecarTask,
  [TaskType.SPARK]: NodeExecutionDisplayType.SparkTask,
  [TaskType.UNKNOWN]: NodeExecutionDisplayType.UnknownTask,
  [TaskType.WAITABLE]: NodeExecutionDisplayType.WaitableTask,
  [TaskType.MPI]: NodeExecutionDisplayType.MpiTask,
  [TaskType.ARRAY_AWS]: NodeExecutionDisplayType.ARRAY_AWS,
  [TaskType.ARRAY_K8S]: NodeExecutionDisplayType.ARRAY_K8S,
};

export const cacheStatusMessages: { [k in CatalogCacheStatus]: string } = {
  [CatalogCacheStatus.CACHE_DISABLED]: 'Caching was disabled for this execution.',
  [CatalogCacheStatus.CACHE_HIT]: 'Output for this execution was read from cache.',
  [CatalogCacheStatus.CACHE_LOOKUP_FAILURE]: 'Failed to lookup cache information.',
  [CatalogCacheStatus.CACHE_MISS]: 'No cached output was found for this execution.',
  [CatalogCacheStatus.CACHE_POPULATED]: 'The result of this execution was written to cache.',
  [CatalogCacheStatus.CACHE_PUT_FAILURE]: 'Failed to write output for this execution to cache.',
};
export const unknownCacheStatusString = 'Cache status is unknown';
export const viewSourceExecutionString = 'View source execution';
