import { APIContextValue } from 'components/data/apiContext';
import {
    limits,
    RequestConfig,
    TaskExecutionIdentifier,
    WorkflowExecutionIdentifier
} from 'models';

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
