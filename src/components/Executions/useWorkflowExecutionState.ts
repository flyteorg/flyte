import {
    useDataRefresher,
    useFetchableData,
    useNodeExecutions
} from 'components/hooks';
import { every } from 'lodash';
import {
    Execution,
    executionSortFields,
    FilterOperation,
    limits,
    SortDirection,
    Workflow,
    WorkflowId
} from 'models';
import { useContext } from 'react';
import {
    executionIsTerminal,
    executionRefreshIntervalMs,
    nodeExecutionIsTerminal
} from '.';
import { ExecutionDataCacheContext } from './contexts';

/** Using a custom fetchable to make sure the related workflow is fetched
 * using an ExecutionDataCache, ensuring that the extended details for NodeExecutions
 * can be found.
 */
function useCachedWorkflow(id: WorkflowId) {
    const dataCache = useContext(ExecutionDataCacheContext);
    return useFetchableData<Workflow, WorkflowId>(
        {
            debugName: 'Workflow',
            defaultValue: {} as Workflow,
            doFetch: id => dataCache.getWorkflow(id)
        },
        id
    );
}

/** Fetches both the workflow and nodeExecutions for a given WorkflowExecution.
 * Will also map node details to the node executions.
 */
export function useWorkflowExecutionState(
    execution: Execution,
    filter: FilterOperation[] = []
) {
    const sort = {
        key: executionSortFields.createdAt,
        direction: SortDirection.ASCENDING
    };
    const nodeExecutionsRequestConfig = {
        filter,
        sort,
        limit: limits.NONE
    };
    const nodeExecutions = useNodeExecutions(
        execution.id,
        nodeExecutionsRequestConfig
    );

    const workflow = useCachedWorkflow(execution.closure.workflowId);

    // We will continue to refresh the node executions list as long
    // as either the parent execution or any child is non-terminal
    useDataRefresher(execution.id, nodeExecutions, {
        interval: executionRefreshIntervalMs,
        valueIsFinal: executions =>
            every(executions, nodeExecutionIsTerminal) &&
            executionIsTerminal(execution)
    });

    return { workflow, nodeExecutions, nodeExecutionsRequestConfig };
}
