import {
    Execution,
    IdentifierScope,
    listExecutions,
    RequestConfig
} from 'models';

import { usePagination } from './usePagination';

/** A hook for fetching a paginated list of workflow executions */
export function useWorkflowExecutions(
    scope: IdentifierScope,
    config: RequestConfig
) {
    return usePagination<Execution, IdentifierScope>(
        { ...config, cacheItems: true, fetchArg: scope },
        listExecutions
    );
}
