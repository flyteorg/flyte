import {
    ExecutionData,
    getNodeExecution,
    getNodeExecutionData,
    NodeExecution,
    NodeExecutionIdentifier
} from 'models';

import { FetchableData } from './types';
import { useFetchableData } from './useFetchableData';

/** A hook for fetching a NodeExecution */
export function useNodeExecution(
    id: NodeExecutionIdentifier
): FetchableData<NodeExecution> {
    return useFetchableData<NodeExecution, NodeExecutionIdentifier>(
        {
            debugName: 'NodeExecution',
            defaultValue: {} as NodeExecution,
            doFetch: id => getNodeExecution(id)
        },
        id
    );
}

/** Fetches the signed URLs for NodeExecution data (inputs/outputs) */
export function useNodeExecutionData(
    id: NodeExecutionIdentifier
): FetchableData<ExecutionData> {
    return useFetchableData<ExecutionData, NodeExecutionIdentifier>(
        {
            debugName: 'NodeExecutionData',
            defaultValue: {} as ExecutionData,
            doFetch: id => getNodeExecutionData(id)
        },
        id
    );
}
