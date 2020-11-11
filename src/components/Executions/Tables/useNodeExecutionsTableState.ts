import { useContext, useMemo, useState } from 'react';
import { ExecutionDataCacheContext } from '../contexts';
import { DetailedNodeExecution } from '../types';
import { NodeExecutionsTableProps } from './NodeExecutionsTable';
import { NodeExecutionsTableState } from './types';

export function useNodeExecutionsTableState({
    value: executions
}: NodeExecutionsTableProps): NodeExecutionsTableState {
    const dataCache = useContext(ExecutionDataCacheContext);
    const [selectedExecutionKey, setSelectedExecutionKey] = useState<
        string | null
    >(null);

    const selectedExecution = selectedExecutionKey
        ? dataCache.getNodeExecution(selectedExecutionKey)
        : null;

    const setSelectedExecution = useMemo(
        () => (newValue: DetailedNodeExecution | null) =>
            setSelectedExecutionKey(newValue?.cacheKey ?? null),
        [setSelectedExecutionKey]
    );

    return {
        executions,
        selectedExecution,
        setSelectedExecution
    };
}
