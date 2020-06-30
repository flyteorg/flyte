import {
    Execution,
    ExecutionData,
    getExecution,
    LiteralMap,
    terminateWorkflowExecution,
    WorkflowExecutionIdentifier
} from 'models';

import { useAPIContext } from 'components/data/apiContext';
import { maxBlobDownloadSizeBytes } from 'components/Literals/constants';
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

/** A hook for fetching the inputs object associated with an Execution. Will
 * handle both the legacy (`computedInputs`) and current (externally stored) formats
 */
export function useWorkflowExecutionInputs(execution: Execution) {
    const { getExecutionData, getRemoteLiteralMap } = useAPIContext();
    return useFetchableData<LiteralMap, WorkflowExecutionIdentifier>(
        {
            debugName: 'ExecutionInputs',
            defaultValue: { literals: {} } as LiteralMap,
            doFetch: async () => {
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
            }
        },
        execution.id
    );
}
