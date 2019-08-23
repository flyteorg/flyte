import {
    negativeTextColor,
    positiveTextColor,
    secondaryTextColor,
    statusColors
} from 'components/Theme';
import { TaskType } from 'models';
import {
    NodeExecutionPhase,
    TaskExecutionPhase,
    WorkflowExecutionPhase
} from 'models/Execution/enums';
import { ExecutionPhaseConstants, NodeExecutionDisplayType } from './types';

export const executionRefreshIntervalMs = 10000;
export const noLogsFoundString = 'No logs found';

/** Shared values for color/text/etc for each execution phase */
export const workflowExecutionPhaseConstants: {
    [key in WorkflowExecutionPhase]: ExecutionPhaseConstants
} = {
    [WorkflowExecutionPhase.ABORTED]: {
        badgeColor: statusColors.SKIPPED,
        text: 'Aborted',
        textColor: negativeTextColor
    },
    [WorkflowExecutionPhase.FAILING]: {
        badgeColor: statusColors.FAILURE,
        text: 'Failing',
        textColor: negativeTextColor
    },
    [WorkflowExecutionPhase.FAILED]: {
        badgeColor: statusColors.FAILURE,
        text: 'Failed',
        textColor: negativeTextColor
    },
    [WorkflowExecutionPhase.QUEUED]: {
        badgeColor: statusColors.QUEUED,
        text: 'Queued',
        textColor: secondaryTextColor
    },
    [WorkflowExecutionPhase.RUNNING]: {
        badgeColor: statusColors.RUNNING,
        text: 'Running',
        textColor: secondaryTextColor
    },
    [WorkflowExecutionPhase.SUCCEEDED]: {
        badgeColor: statusColors.SUCCESS,
        text: 'Succeeded',
        textColor: positiveTextColor
    },
    [WorkflowExecutionPhase.SUCCEEDING]: {
        badgeColor: statusColors.SUCCESS,
        text: 'Succeeding',
        textColor: positiveTextColor
    },
    [WorkflowExecutionPhase.TIMED_OUT]: {
        badgeColor: statusColors.FAILURE,
        text: 'Timed Out',
        textColor: negativeTextColor
    },
    [WorkflowExecutionPhase.UNDEFINED]: {
        badgeColor: statusColors.UNKNOWN,
        text: 'Unknown',
        textColor: secondaryTextColor
    }
};

/** Shared values for color/text/etc for each node execution phase */
export const nodeExecutionPhaseConstants: {
    [key in NodeExecutionPhase]: ExecutionPhaseConstants
} = {
    [NodeExecutionPhase.ABORTED]: {
        badgeColor: statusColors.FAILURE,
        text: 'Aborted',
        textColor: negativeTextColor
    },
    [NodeExecutionPhase.FAILING]: {
        badgeColor: statusColors.FAILURE,
        text: 'Failing',
        textColor: negativeTextColor
    },
    [NodeExecutionPhase.FAILED]: {
        badgeColor: statusColors.FAILURE,
        text: 'Failed',
        textColor: negativeTextColor
    },
    [NodeExecutionPhase.QUEUED]: {
        badgeColor: statusColors.RUNNING,
        text: 'Queued',
        textColor: secondaryTextColor
    },
    [NodeExecutionPhase.RUNNING]: {
        badgeColor: statusColors.RUNNING,
        text: 'Running',
        textColor: secondaryTextColor
    },
    [NodeExecutionPhase.SUCCEEDED]: {
        badgeColor: statusColors.SUCCESS,
        text: 'Succeeded',
        textColor: positiveTextColor
    },
    [NodeExecutionPhase.TIMED_OUT]: {
        badgeColor: statusColors.FAILURE,
        text: 'Timed Out',
        textColor: negativeTextColor
    },
    [NodeExecutionPhase.SKIPPED]: {
        badgeColor: statusColors.UNKNOWN,
        text: 'Skipped',
        textColor: secondaryTextColor
    },
    [NodeExecutionPhase.UNDEFINED]: {
        badgeColor: statusColors.UNKNOWN,
        text: 'Unknown',
        textColor: secondaryTextColor
    }
};

/** Shared values for color/text/etc for each node execution phase */
export const taskExecutionPhaseConstants: {
    [key in TaskExecutionPhase]: ExecutionPhaseConstants
} = {
    [TaskExecutionPhase.ABORTED]: {
        badgeColor: statusColors.FAILURE,
        text: 'Aborted',
        textColor: negativeTextColor
    },
    [TaskExecutionPhase.FAILED]: {
        badgeColor: statusColors.FAILURE,
        text: 'Failed',
        textColor: negativeTextColor
    },
    [TaskExecutionPhase.QUEUED]: {
        badgeColor: statusColors.RUNNING,
        text: 'Queued',
        textColor: secondaryTextColor
    },
    [TaskExecutionPhase.RUNNING]: {
        badgeColor: statusColors.RUNNING,
        text: 'Running',
        textColor: secondaryTextColor
    },
    [TaskExecutionPhase.SUCCEEDED]: {
        badgeColor: statusColors.SUCCESS,
        text: 'Succeeded',
        textColor: positiveTextColor
    },
    [TaskExecutionPhase.UNDEFINED]: {
        badgeColor: statusColors.UNKNOWN,
        text: 'Unknown',
        textColor: secondaryTextColor
    }
};

export const taskTypeToNodeExecutionDisplayType: {
    [k in TaskType]: NodeExecutionDisplayType
} = {
    [TaskType.ARRAY]: NodeExecutionDisplayType.ArrayTask,
    [TaskType.BATCH_HIVE]: NodeExecutionDisplayType.BatchHiveTask,
    [TaskType.DYNAMIC]: NodeExecutionDisplayType.DynamicTask,
    [TaskType.HIVE]: NodeExecutionDisplayType.HiveTask,
    [TaskType.PYTHON]: NodeExecutionDisplayType.PythonTask,
    [TaskType.SIDECAR]: NodeExecutionDisplayType.SidecarTask,
    [TaskType.SPARK]: NodeExecutionDisplayType.SparkTask,
    [TaskType.UNKNOWN]: NodeExecutionDisplayType.UnknownTask,
    [TaskType.WAITABLE]: NodeExecutionDisplayType.WaitableTask
};
