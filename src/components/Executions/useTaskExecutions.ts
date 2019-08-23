import { every } from 'lodash';
import {
    ExecutionData,
    getTaskExecutionData,
    limits,
    listTaskExecutions,
    NodeExecution,
    NodeExecutionIdentifier,
    TaskExecution,
    TaskExecutionIdentifier
} from 'models';
import { useDataRefresher } from '../hooks';
import { FetchableData } from '../hooks/types';
import { useFetchableData } from '../hooks/useFetchableData';
import { executionRefreshIntervalMs } from './constants';
import { nodeExecutionIsTerminal, taskExecutionIsTerminal } from './utils';

const doFetchTaskExecutions = async (id: NodeExecutionIdentifier) => {
    const { entities } = await listTaskExecutions(id, {
        limit: limits.NONE
    });
    return entities;
};

/** A hook for fetching the list of TaskExecutions associated with a
 * NodeExecution
 */
export function useTaskExecutions(
    id: NodeExecutionIdentifier
): FetchableData<TaskExecution[]> {
    return useFetchableData<TaskExecution[], NodeExecutionIdentifier>(
        {
            debugName: 'TaskExecutions',
            defaultValue: [],
            doFetch: doFetchTaskExecutions
        },
        id
    );
}

/** Fetches the signed URLs for TaskExecution data (inputs/outputs) */
export function useTaskExecutionData(
    id: TaskExecutionIdentifier
): FetchableData<ExecutionData> {
    return useFetchableData<ExecutionData, TaskExecutionIdentifier>(
        {
            debugName: 'TaskExecutionData',
            defaultValue: {} as ExecutionData,
            doFetch: id => getTaskExecutionData(id)
        },
        id
    );
}

export function useTaskExecutionsRefresher(
    nodeExecution: NodeExecution,
    taskExecutionsFetchable: ReturnType<typeof useTaskExecutions>
) {
    return useDataRefresher(nodeExecution.id, taskExecutionsFetchable, {
        interval: executionRefreshIntervalMs,
        valueIsFinal: taskExecutions =>
            every(taskExecutions, taskExecutionIsTerminal) &&
            nodeExecutionIsTerminal(nodeExecution)
    });
}
