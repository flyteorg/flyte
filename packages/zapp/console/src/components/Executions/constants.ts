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
import t from './strings';
import { ExecutionPhaseConstants, NodeExecutionDisplayType } from './types';

export const executionRefreshIntervalMs = 10000;
export const noLogsFoundString = t('noLogsFoundString');

/** Shared values for color/text/etc for each execution phase */
export const workflowExecutionPhaseConstants: {
  [key in WorkflowExecutionPhase]: ExecutionPhaseConstants;
} = {
  [WorkflowExecutionPhase.ABORTED]: {
    badgeColor: statusColors.SKIPPED,
    text: t('aborted'),
    textColor: negativeTextColor,
  },
  [WorkflowExecutionPhase.ABORTING]: {
    badgeColor: statusColors.SKIPPED,
    text: t('aborting'),
    textColor: negativeTextColor,
  },
  [WorkflowExecutionPhase.FAILING]: {
    badgeColor: statusColors.FAILURE,
    text: t('failing'),
    textColor: negativeTextColor,
  },
  [WorkflowExecutionPhase.FAILED]: {
    badgeColor: statusColors.FAILURE,
    text: t('failed'),
    textColor: negativeTextColor,
  },
  [WorkflowExecutionPhase.QUEUED]: {
    badgeColor: statusColors.QUEUED,
    text: t('queued'),
    textColor: secondaryTextColor,
  },
  [WorkflowExecutionPhase.RUNNING]: {
    badgeColor: statusColors.RUNNING,
    text: t('running'),
    textColor: secondaryTextColor,
  },
  [WorkflowExecutionPhase.SUCCEEDED]: {
    badgeColor: statusColors.SUCCESS,
    text: t('succeeded'),
    textColor: positiveTextColor,
  },
  [WorkflowExecutionPhase.SUCCEEDING]: {
    badgeColor: statusColors.SUCCESS,
    text: t('succeeding'),
    textColor: positiveTextColor,
  },
  [WorkflowExecutionPhase.TIMED_OUT]: {
    badgeColor: statusColors.FAILURE,
    text: t('timedOut'),
    textColor: negativeTextColor,
  },
  [WorkflowExecutionPhase.UNDEFINED]: {
    badgeColor: statusColors.UNKNOWN,
    text: t('unknown'),
    textColor: secondaryTextColor,
  },
};

/** Shared values for color/text/etc for each node execution phase */
export const nodeExecutionPhaseConstants: {
  [key in NodeExecutionPhase]: ExecutionPhaseConstants;
} = {
  [NodeExecutionPhase.ABORTED]: {
    badgeColor: statusColors.FAILURE,
    text: t('aborted'),
    textColor: negativeTextColor,
  },
  [NodeExecutionPhase.FAILING]: {
    badgeColor: statusColors.FAILURE,
    text: t('failing'),
    textColor: negativeTextColor,
  },
  [NodeExecutionPhase.FAILED]: {
    badgeColor: statusColors.FAILURE,
    text: t('failed'),
    textColor: negativeTextColor,
  },
  [NodeExecutionPhase.QUEUED]: {
    badgeColor: statusColors.RUNNING,
    text: t('queued'),
    textColor: secondaryTextColor,
  },
  [NodeExecutionPhase.RUNNING]: {
    badgeColor: statusColors.RUNNING,
    text: t('running'),
    textColor: secondaryTextColor,
  },
  [NodeExecutionPhase.DYNAMIC_RUNNING]: {
    badgeColor: statusColors.RUNNING,
    text: t('running'),
    textColor: secondaryTextColor,
  },
  [NodeExecutionPhase.SUCCEEDED]: {
    badgeColor: statusColors.SUCCESS,
    text: t('succeeded'),
    textColor: positiveTextColor,
  },
  [NodeExecutionPhase.TIMED_OUT]: {
    badgeColor: statusColors.FAILURE,
    text: t('timedOut'),
    textColor: negativeTextColor,
  },
  [NodeExecutionPhase.SKIPPED]: {
    badgeColor: statusColors.UNKNOWN,
    text: t('skipped'),
    textColor: secondaryTextColor,
  },
  [NodeExecutionPhase.RECOVERED]: {
    badgeColor: statusColors.SUCCESS,
    text: t('recovered'),
    textColor: positiveTextColor,
  },
  [NodeExecutionPhase.UNDEFINED]: {
    badgeColor: statusColors.UNKNOWN,
    text: t('unknown'),
    textColor: secondaryTextColor,
  },
};

/** Shared values for color/text/etc for each node execution phase */
export const taskExecutionPhaseConstants: {
  [key in TaskExecutionPhase]: ExecutionPhaseConstants;
} = {
  [TaskExecutionPhase.ABORTED]: {
    badgeColor: statusColors.FAILURE,
    text: t('aborted'),
    textColor: negativeTextColor,
  },
  [TaskExecutionPhase.FAILED]: {
    badgeColor: statusColors.FAILURE,
    text: t('failed'),
    textColor: negativeTextColor,
  },
  [TaskExecutionPhase.WAITING_FOR_RESOURCES]: {
    badgeColor: statusColors.RUNNING,
    text: t('waiting'),
    textColor: secondaryTextColor,
  },
  [TaskExecutionPhase.QUEUED]: {
    badgeColor: statusColors.RUNNING,
    text: t('queued'),
    textColor: secondaryTextColor,
  },
  [TaskExecutionPhase.INITIALIZING]: {
    badgeColor: statusColors.RUNNING,
    text: t('initializing'),
    textColor: secondaryTextColor,
  },
  [TaskExecutionPhase.RUNNING]: {
    badgeColor: statusColors.RUNNING,
    text: t('running'),
    textColor: secondaryTextColor,
  },
  [TaskExecutionPhase.SUCCEEDED]: {
    badgeColor: statusColors.SUCCESS,
    text: t('succeeded'),
    textColor: positiveTextColor,
  },
  [TaskExecutionPhase.UNDEFINED]: {
    badgeColor: statusColors.UNKNOWN,
    text: t('unknown'),
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
  [CatalogCacheStatus.CACHE_DISABLED]: t('cacheDisabledMessage'),
  [CatalogCacheStatus.CACHE_HIT]: t('cacheHitMessage'),
  [CatalogCacheStatus.CACHE_LOOKUP_FAILURE]: t('cacheLookupFailureMessage'),
  [CatalogCacheStatus.CACHE_MISS]: t('cacheMissMessage'),
  [CatalogCacheStatus.CACHE_POPULATED]: t('cachePopulatedMessage'),
  [CatalogCacheStatus.CACHE_PUT_FAILURE]: t('cachePutFailure'),
  [CatalogCacheStatus.MAP_CACHE]: t('mapCacheMessage'),
};
export const unknownCacheStatusString = t('unknownCacheStatusString');
export const viewSourceExecutionString = t('viewSourceExecutionString');
