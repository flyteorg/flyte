import { RequestConfig } from 'models/AdminEntity/types';
import { IdentifierScope } from 'models/Common/types';
import { listExecutions } from 'models/Execution/api';
import { Execution } from 'models/Execution/types';
import { usePagination } from './usePagination';

/** A hook for fetching a paginated list of workflow executions */
export function useWorkflowExecutions(scope: IdentifierScope, config: RequestConfig) {
  return usePagination<Execution, IdentifierScope>(
    { ...config, cacheItems: true, fetchArg: scope },
    listExecutions,
  );
}
