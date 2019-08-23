import {
    limits,
    listNodeExecutions,
    listTaskExecutionChildren,
    NodeExecution,
    RequestConfig,
    TaskExecutionIdentifier,
    WorkflowExecutionIdentifier
} from 'models';
import { useFetchableData } from './useFetchableData';

interface NodeExecutionsFetchData {
    id: WorkflowExecutionIdentifier;
    config: RequestConfig;
}

interface TaskExecutionChildrenFetchData {
    taskExecutionId: TaskExecutionIdentifier;
    config: RequestConfig;
}

const doFetchNodeExecutions = async ({
    id,
    config
}: NodeExecutionsFetchData) => {
    const { entities } = await listNodeExecutions(id, {
        ...config,
        limit: limits.NONE
    });
    return entities;
};

/** Fetches all the child NodeExecutions of a WorkflowExecution, according to a
 * passed RequestConfig.
 */
export function useNodeExecutions(
    id: WorkflowExecutionIdentifier,
    config: RequestConfig
) {
    return useFetchableData<NodeExecution[], NodeExecutionsFetchData>(
        {
            debugName: 'NodeExecutions',
            defaultValue: [],
            doFetch: doFetchNodeExecutions
        },
        { id, config }
    );
}

const doFetchTaskExecutionChildren = async ({
    taskExecutionId,
    config
}: TaskExecutionChildrenFetchData) => {
    const { entities } = await listTaskExecutionChildren(taskExecutionId, {
        ...config,
        limit: limits.NONE
    });
    return entities;
};

/** Fetches the child `NodeExecutions` belonging to a `TaskExecution` */
export function useTaskExecutionChildren(
    taskExecutionId: TaskExecutionIdentifier,
    config: RequestConfig
) {
    return useFetchableData<NodeExecution[], TaskExecutionChildrenFetchData>(
        {
            debugName: 'TaskExecutionChildren',
            defaultValue: [],
            doFetch: doFetchTaskExecutionChildren
        },
        { taskExecutionId, config }
    );
}
