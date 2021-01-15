import { LargeLoadingSpinner } from 'components/common/LoadingSpinner';
import { WaitForQuery } from 'components/common/WaitForQuery';
import { withRouteParams } from 'components/common/withRouteParams';
import { useConditionalQuery } from 'components/hooks/useConditionalQuery';
import { TaskExecution, TaskExecutionIdentifier } from 'models/Execution/types';
import * as React from 'react';
import { executionRefreshIntervalMs } from '../constants';
import { makeTaskExecutionQuery } from '../taskExecutionQueries';
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
    const taskExecutionQuery = useConditionalQuery(
        {
            ...makeTaskExecutionQuery(taskExecutionId),
            refetchInterval: executionRefreshIntervalMs
        },
        execution => !taskExecutionIsTerminal(execution)
    );

    const renderContent = (taskExecution: TaskExecution) => (
        <>
            <TaskExecutionDetailsAppBarContent taskExecution={taskExecution} />
            <TaskExecutionNodes taskExecution={taskExecution} />
        </>
    );

    return (
        <WaitForQuery
            loadingComponent={LargeLoadingSpinner}
            query={taskExecutionQuery}
        >
            {renderContent}
        </WaitForQuery>
    );
};

export const TaskExecutionDetails = withRouteParams<
    TaskExecutionDetailsRouteParams
>(TaskExecutionDetailsContainer);
