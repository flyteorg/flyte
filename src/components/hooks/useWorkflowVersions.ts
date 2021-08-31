import { IdentifierScope } from 'models/Common/types';
import { RequestConfig } from 'models/AdminEntity/types';
import { listWorkflows } from 'models/Workflow/api';
import { Workflow } from 'models/Workflow/types';
import { usePagination } from './usePagination';

/**
 * A hook for fetching a paginated list of workflow versions.
 * @param scope
 * @param config
 */
export function useWorkflowVersions(
    scope: IdentifierScope,
    config: RequestConfig
) {
    return usePagination<Workflow, IdentifierScope>(
        { ...config, cacheItems: true, fetchArg: scope },
        listWorkflows
    );
}
