import { APIContextValue, useAPIContext } from 'components/data/apiContext';
import {
    limits,
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

/** Fetches a list of `NodeExecution`s which are children of a WorkflowExecution.
 * This function is meant to be consumed by hooks which are composing data.
 * If you're calling it from a component, consider using `useNodeExecutions` instead.
 */
export const fetchNodeExecutions = async (
    { id, config }: NodeExecutionsFetchData,
    apiContext: APIContextValue
) => {
    const { listNodeExecutions } = apiContext;
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
    const apiContext = useAPIContext();
    return useFetchableData<NodeExecution[], NodeExecutionsFetchData>(
        {
            debugName: 'NodeExecutions',
            defaultValue: [],
            doFetch: data => fetchNodeExecutions(data, apiContext)
        },
        { id, config }
    );
}

/** Fetches a list of `NodeExecution`s which are children of the given `TaskExecution`.
 * This function is meant to be consumed by hooks which are composing data.
 * If you're calling it from a component, consider using `useTaskExecutionChildren` instead.
 */
export const fetchTaskExecutionChildren = async (
    { taskExecutionId, config }: TaskExecutionChildrenFetchData,
    apiContext: APIContextValue
) => {
    const { listTaskExecutionChildren } = apiContext;
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
    const apiContext = useAPIContext();
    return useFetchableData<NodeExecution[], TaskExecutionChildrenFetchData>(
        {
            debugName: 'TaskExecutionChildren',
            defaultValue: [],
            doFetch: data => fetchTaskExecutionChildren(data, apiContext)
        },
        { taskExecutionId, config }
    );
}
