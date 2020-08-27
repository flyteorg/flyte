import { APIContextValue, useAPIContext } from 'components/data/apiContext';
import { maxBlobDownloadSizeBytes } from 'components/Literals/constants';
import {
    Execution,
    ExecutionData,
    LiteralMap,
    terminateWorkflowExecution,
    WorkflowExecutionIdentifier
} from 'models';
import { FetchableData, FetchableExecution } from '../hooks/types';
import { useFetchableData } from '../hooks/useFetchableData';
import { ExecutionDataCache } from './types';

/** A hook for fetching a WorkflowExecution */
export function useWorkflowExecution(
    id: WorkflowExecutionIdentifier,
    dataCache: ExecutionDataCache
): FetchableExecution {
    const fetchable = useFetchableData<Execution, WorkflowExecutionIdentifier>(
        {
            debugName: 'Execution',
            defaultValue: {} as Execution,
            doFetch: id => dataCache.getWorkflowExecution(id)
        },
        id
    );

    const terminateExecution = async (cause: string) => {
        await terminateWorkflowExecution(id, cause);
        await fetchable.fetch();
    };

    return { fetchable, terminateExecution };
}

/** Fetches the signed URLs for NodeExecution data (inputs/outputs) */
export function useWorkflowExecutionData(
    id: WorkflowExecutionIdentifier
): FetchableData<ExecutionData> {
    const { getExecutionData } = useAPIContext();
    return useFetchableData<ExecutionData, WorkflowExecutionIdentifier>(
        {
            debugName: 'ExecutionData',
            defaultValue: {} as ExecutionData,
            doFetch: id => getExecutionData(id)
        },
        id
    );
}

/** Fetches the inputs object for a given WorkflowExecution.
 * This function is meant to be consumed by hooks which are composing data.
 * If you're calling it from a component, consider using `useTaskExecutions` instead.
 */
export const fetchWorkflowExecutionInputs = async (
    execution: Execution,
    apiContext: APIContextValue
) => {
    const { getExecutionData, getRemoteLiteralMap } = apiContext;
    if (execution.closure.computedInputs) {
        return execution.closure.computedInputs;
    }
    const { inputs } = await getExecutionData(execution.id);
    if (
        !inputs.url ||
        !inputs.bytes ||
        inputs.bytes.gt(maxBlobDownloadSizeBytes)
    ) {
        return { literals: {} };
    }
    return getRemoteLiteralMap(inputs.url);
};

/** A hook for fetching the inputs object associated with an Execution. Will
 * handle both the legacy (`computedInputs`) and current (externally stored) formats
 */
export function useWorkflowExecutionInputs(execution: Execution) {
    const apiContext = useAPIContext();
    return useFetchableData<LiteralMap, WorkflowExecutionIdentifier>(
        {
            debugName: 'ExecutionInputs',
            defaultValue: { literals: {} } as LiteralMap,
            doFetch: async () =>
                fetchWorkflowExecutionInputs(execution, apiContext)
        },
        execution.id
    );
}
