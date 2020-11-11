import { useFetchableData } from 'components/hooks';
import { RequestConfig, WorkflowExecutionIdentifier } from 'models';
import { useContext } from 'react';
import { ExecutionDataCacheContext } from './contexts';
import { DetailedNodeExecution } from './types';

export interface UseDetailedNodeExecutionsData {
    executionId: WorkflowExecutionIdentifier;
    requestConfig: RequestConfig;
}

export function useDetailedNodeExecutions(
    executionId: WorkflowExecutionIdentifier,
    requestConfig: RequestConfig
) {
    const dataCache = useContext(ExecutionDataCacheContext);
    return useFetchableData<
        DetailedNodeExecution[],
        UseDetailedNodeExecutionsData
    >(
        {
            debugName: 'DetailedNodeExecutions',
            defaultValue: [],
            doFetch: data =>
                dataCache.getNodeExecutions(
                    data.executionId,
                    data.requestConfig
                )
        },
        { executionId, requestConfig }
    );
}
