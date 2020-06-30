import { WaitForData, withRouteParams } from 'components/common';
import {
    RefreshConfig,
    useDataRefresher,
    useTaskExecution
} from 'components/hooks';
import { TaskExecution, TaskExecutionIdentifier } from 'models';
import * as React from 'react';
import { executionRefreshIntervalMs } from '../constants';
import { ExecutionDataCacheContext } from '../contexts';
import { useExecutionDataCache } from '../useExecutionDataCache';
import { taskExecutionIsTerminal } from '../utils';
import { TaskExecutionDetailsAppBarContent } from './TaskExecutionDetailsAppBarContent';
import { TaskExecutionNodes } from './TaskExecutionNodes';

export interface TaskExecutionDetailsRouteParams {
    domainId: string;
    executionName: string;
    nodeId: string;
    projectId: string;
    retryAttempt: string;
    taskDomain: string;
    taskName: string;
    taskProject: string;
    taskVersion: string;
}
export type TaskExecutionDetailsProps = TaskExecutionDetailsRouteParams;

const refreshConfig: RefreshConfig<TaskExecution> = {
    interval: executionRefreshIntervalMs,
    valueIsFinal: taskExecutionIsTerminal
};

function routeParamsToTaskExecutionId(
    params: TaskExecutionDetailsRouteParams
): TaskExecutionIdentifier {
    let retryAttempt = Number.parseInt(params.retryAttempt, 10);
    if (Number.isNaN(retryAttempt)) {
        retryAttempt = 0;
    }

    return {
        retryAttempt,
        nodeExecutionId: {
            executionId: {
                domain: params.domainId,
                name: params.executionName,
                project: params.projectId
            },
            nodeId: params.nodeId
        },
        taskId: {
            domain: params.taskDomain,
            name: params.taskName,
            project: params.taskProject,
            version: params.taskVersion
        }
    };
}

export const TaskExecutionDetailsContainer: React.FC<TaskExecutionDetailsProps> = props => {
    const taskExecutionId = routeParamsToTaskExecutionId(props);
    const dataCache = useExecutionDataCache();
    const taskExecution = useTaskExecution(taskExecutionId);

    useDataRefresher(taskExecutionId, taskExecution, refreshConfig);

    return (
        <WaitForData {...taskExecution}>
            <TaskExecutionDetailsAppBarContent
                taskExecution={taskExecution.value}
            />
            <ExecutionDataCacheContext.Provider value={dataCache}>
                <TaskExecutionNodes taskExecution={taskExecution.value} />
            </ExecutionDataCacheContext.Provider>
        </WaitForData>
    );
};

export const TaskExecutionDetails = withRouteParams<
    TaskExecutionDetailsRouteParams
>(TaskExecutionDetailsContainer);
