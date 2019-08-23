import {
    Execution,
    ExecutionData,
    getExecution,
    getExecutionData,
    terminateWorkflowExecution,
    WorkflowExecutionIdentifier
} from 'models';

import { FetchableData, FetchableExecution } from './types';
import { useFetchableData } from './useFetchableData';

/** A hook for fetching a WorkflowExecution */
export function useWorkflowExecution(
    id: WorkflowExecutionIdentifier
): FetchableExecution {
    const fetchable = useFetchableData<Execution, WorkflowExecutionIdentifier>(
        {
            debugName: 'Execution',
            defaultValue: {} as Execution,
            doFetch: id => getExecution(id)
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
    return useFetchableData<ExecutionData, WorkflowExecutionIdentifier>(
        {
            debugName: 'ExecutionData',
            defaultValue: {} as ExecutionData,
            doFetch: id => getExecutionData(id)
        },
        id
    );
}
