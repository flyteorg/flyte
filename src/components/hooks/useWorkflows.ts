import { useAPIContext } from 'components/data/apiContext';
import {
    IdentifierScope,
    NamedEntityIdentifier,
    RequestConfig,
    ResourceType,
    Workflow
} from 'models';
import { usePagination } from './usePagination';

/** A hook for fetching a paginated list of workflows */
export function useWorkflows(scope: IdentifierScope, config: RequestConfig) {
    const { listWorkflows } = useAPIContext();
    return usePagination<Workflow, IdentifierScope>(
        { ...config, cacheItems: true, fetchArg: scope },
        listWorkflows
    );
}

/** A hook for fetching a paginated list of workflow ids */
export function useWorkflowIds(scope: IdentifierScope, config: RequestConfig) {
    const { listIdentifiers } = useAPIContext();
    return usePagination<NamedEntityIdentifier, IdentifierScope>(
        { ...config, fetchArg: scope },
        (scope, requestConfig) =>
            listIdentifiers(
                { scope, type: ResourceType.WORKFLOW },
                requestConfig
            )
    );
}
