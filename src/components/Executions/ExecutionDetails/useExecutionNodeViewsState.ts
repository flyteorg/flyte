import { useConditionalQuery } from 'components/hooks/useConditionalQuery';
import {
    Execution,
    executionSortFields,
    FilterOperation,
    limits,
    NodeExecution,
    SortDirection
} from 'models';
import { useQueryClient } from 'react-query';
import { executionRefreshIntervalMs } from '../constants';
import { makeNodeExecutionListQuery } from '../nodeExecutionQueries';
import { executionIsTerminal, nodeExecutionIsTerminal } from '../utils';

export function useExecutionNodeViewsState(
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

    const shouldEnableQuery = (executions: NodeExecution[]) =>
        !executionIsTerminal(execution) ||
        executions.some(ne => !nodeExecutionIsTerminal(ne));

    const nodeExecutionsQuery = useConditionalQuery(
        {
            ...makeNodeExecutionListQuery(
                useQueryClient(),
                execution.id,
                nodeExecutionsRequestConfig
            ),
            refetchInterval: executionRefreshIntervalMs
        },
        shouldEnableQuery
    );

    return { nodeExecutionsQuery, nodeExecutionsRequestConfig };
}
