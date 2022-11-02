import {
  graphStatusColors,
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
    text: t('aborted'),
    value: 'ABORTED',
    badgeColor: statusColors.SKIPPED,
    nodeColor: graphStatusColors.ABORTED,
    textColor: negativeTextColor,
  },
  [WorkflowExecutionPhase.ABORTING]: {
    text: t('aborting'),
    value: 'ABORTING',
    badgeColor: statusColors.SKIPPED,
    nodeColor: graphStatusColors.ABORTED,
    textColor: negativeTextColor,
  },
  [WorkflowExecutionPhase.FAILING]: {
    text: t('failing'),
    value: 'FAILING',
    badgeColor: statusColors.FAILURE,
    nodeColor: graphStatusColors.FAILING,
    textColor: negativeTextColor,
  },
  [WorkflowExecutionPhase.FAILED]: {
    text: t('failed'),
    value: 'FAILED',
    badgeColor: statusColors.FAILURE,
    nodeColor: graphStatusColors.FAILED,
    textColor: negativeTextColor,
  },
  [WorkflowExecutionPhase.QUEUED]: {
    text: t('queued'),
    value: 'QUEUED',
    badgeColor: statusColors.QUEUED,
    nodeColor: graphStatusColors.QUEUED,
    textColor: secondaryTextColor,
  },
  [WorkflowExecutionPhase.RUNNING]: {
    text: t('running'),
    value: 'RUNNING',
    badgeColor: statusColors.RUNNING,
    nodeColor: graphStatusColors.RUNNING,
    textColor: secondaryTextColor,
  },
  [WorkflowExecutionPhase.SUCCEEDED]: {
    text: t('succeeded'),
    value: 'SUCCEEDED',
    badgeColor: statusColors.SUCCESS,
    nodeColor: graphStatusColors.SUCCEEDED,
    textColor: positiveTextColor,
  },
  [WorkflowExecutionPhase.SUCCEEDING]: {
    text: t('succeeding'),
    value: 'SUCCEEDING',
    badgeColor: statusColors.SUCCESS,
    nodeColor: graphStatusColors.SUCCEEDED,
    textColor: positiveTextColor,
  },
  [WorkflowExecutionPhase.TIMED_OUT]: {
    text: t('timedOut'),
    value: 'TIMED_OUT',
    badgeColor: statusColors.FAILURE,
    nodeColor: graphStatusColors.FAILED,
    textColor: negativeTextColor,
  },
  [WorkflowExecutionPhase.UNDEFINED]: {
    text: t('unknown'),
    value: 'UNKNOWN',
    badgeColor: statusColors.UNKNOWN,
    nodeColor: graphStatusColors.UNDEFINED,
    textColor: secondaryTextColor,
  },
};

/** Shared values for color/text/etc for each node execution phase */
export const nodeExecutionPhaseConstants: {
  [key in NodeExecutionPhase]: ExecutionPhaseConstants;
} = {
  [NodeExecutionPhase.ABORTED]: {
    text: t('aborted'),
    value: 'ABORTED',
    badgeColor: statusColors.FAILURE,
    nodeColor: graphStatusColors.ABORTED,
    textColor: negativeTextColor,
  },
  [NodeExecutionPhase.FAILING]: {
    text: t('failing'),
    value: 'FAILING',
    badgeColor: statusColors.FAILURE,
    nodeColor: graphStatusColors.FAILING,
    textColor: negativeTextColor,
  },
  [NodeExecutionPhase.FAILED]: {
    text: t('failed'),
    value: 'FAILED',
    badgeColor: statusColors.FAILURE,
    nodeColor: graphStatusColors.FAILED,
    textColor: negativeTextColor,
  },
  [NodeExecutionPhase.QUEUED]: {
    text: t('queued'),
    value: 'QUEUED',
    badgeColor: statusColors.RUNNING,
    nodeColor: graphStatusColors.QUEUED,
    textColor: secondaryTextColor,
  },
  [NodeExecutionPhase.RUNNING]: {
    text: t('running'),
    value: 'RUNNING',
    badgeColor: statusColors.RUNNING,
    nodeColor: graphStatusColors.RUNNING,
    textColor: secondaryTextColor,
  },
  [NodeExecutionPhase.DYNAMIC_RUNNING]: {
    text: t('running'),
    value: 'RUNNING',
    badgeColor: statusColors.RUNNING,
    nodeColor: graphStatusColors.RUNNING,
    textColor: secondaryTextColor,
  },
  [NodeExecutionPhase.SUCCEEDED]: {
    text: t('succeeded'),
    value: 'SUCCEEDED',
    badgeColor: statusColors.SUCCESS,
    nodeColor: graphStatusColors.SUCCEEDED,
    textColor: positiveTextColor,
  },
  [NodeExecutionPhase.TIMED_OUT]: {
    text: t('timedOut'),
    value: 'TIMED_OUT',
    badgeColor: statusColors.FAILURE,
    nodeColor: graphStatusColors.FAILED,
    textColor: negativeTextColor,
  },
  [NodeExecutionPhase.SKIPPED]: {
    text: t('skipped'),
    value: 'SKIPPED',
    badgeColor: statusColors.UNKNOWN,
    nodeColor: graphStatusColors.UNDEFINED,
    textColor: secondaryTextColor,
  },
  [NodeExecutionPhase.RECOVERED]: {
    text: t('recovered'),
    value: 'RECOVERED',
    badgeColor: statusColors.SUCCESS,
    nodeColor: graphStatusColors.SUCCEEDED,
    textColor: positiveTextColor,
  },
  [NodeExecutionPhase.PAUSED]: {
    text: t('paused'),
    value: 'PAUSED',
    badgeColor: statusColors.PAUSED,
    nodeColor: graphStatusColors.PAUSED,
    textColor: secondaryTextColor,
  },
  [NodeExecutionPhase.UNDEFINED]: {
    text: t('unknown'),
    value: 'UNKNOWN',
    badgeColor: statusColors.UNKNOWN,
    nodeColor: graphStatusColors.UNDEFINED,
    textColor: secondaryTextColor,
  },
};

/** Shared values for color/text/etc for each node execution phase */
export const taskExecutionPhaseConstants: {
  [key in TaskExecutionPhase]: ExecutionPhaseConstants;
} = {
  [TaskExecutionPhase.ABORTED]: {
    text: t('aborted'),
    value: 'ABORTED',
    badgeColor: statusColors.FAILURE,
    nodeColor: graphStatusColors.ABORTED,
    textColor: negativeTextColor,
  },
  [TaskExecutionPhase.FAILED]: {
    text: t('failed'),
    value: 'FAILED',
    badgeColor: statusColors.FAILURE,
    nodeColor: graphStatusColors.FAILED,
    textColor: negativeTextColor,
  },
  [TaskExecutionPhase.WAITING_FOR_RESOURCES]: {
    text: t('waiting'),
    value: 'WAITING',
    badgeColor: statusColors.RUNNING,
    nodeColor: graphStatusColors.RUNNING,
    textColor: secondaryTextColor,
  },
  [TaskExecutionPhase.QUEUED]: {
    text: t('queued'),
    value: 'QUEUED',
    badgeColor: statusColors.RUNNING,
    nodeColor: graphStatusColors.QUEUED,
    textColor: secondaryTextColor,
  },
  [TaskExecutionPhase.INITIALIZING]: {
    text: t('initializing'),
    value: 'INITIALIZING',
    badgeColor: statusColors.RUNNING,
    nodeColor: graphStatusColors.RUNNING,
    textColor: secondaryTextColor,
  },
  [TaskExecutionPhase.RUNNING]: {
    text: t('running'),
    value: 'RUNNING',
    badgeColor: statusColors.RUNNING,
    nodeColor: graphStatusColors.RUNNING,
    textColor: secondaryTextColor,
  },
  [TaskExecutionPhase.SUCCEEDED]: {
    text: t('succeeded'),
    value: 'SUCCEEDED',
    badgeColor: statusColors.SUCCESS,
    nodeColor: graphStatusColors.SUCCEEDED,
    textColor: positiveTextColor,
  },
  [TaskExecutionPhase.UNDEFINED]: {
    text: t('unknown'),
    value: 'UNKNOWN',
    badgeColor: statusColors.UNKNOWN,
    nodeColor: graphStatusColors.UNDEFINED,
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
