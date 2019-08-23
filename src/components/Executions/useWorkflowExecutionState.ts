import {
    useDataRefresher,
    useNodeExecutions,
    useWorkflow
} from 'components/hooks';
import { every } from 'lodash';
import {
    Execution,
    executionSortFields,
    FilterOperation,
    limits,
    SortDirection
} from 'models';
import {
    executionIsTerminal,
    executionRefreshIntervalMs,
    nodeExecutionIsTerminal
} from '.';
import { useDetailedNodeExecutions } from './useDetailedNodeExecutions';

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
    const rawNodeExecutions = useNodeExecutions(execution.id, {
        filter,
        sort,
        limit: limits.NONE
    });
    const workflow = useWorkflow(execution.closure.workflowId);
    const nodeExecutions = useDetailedNodeExecutions(
        rawNodeExecutions,
        workflow
    );

    // We will continue to refresh the node executions list as long
    // as either the parent execution or any child is non-terminal
    useDataRefresher(execution.id, nodeExecutions, {
        interval: executionRefreshIntervalMs,
        valueIsFinal: nodeExecutions =>
            every(nodeExecutions, nodeExecutionIsTerminal) &&
            executionIsTerminal(execution)
    });

    return { workflow, nodeExecutions };
}
