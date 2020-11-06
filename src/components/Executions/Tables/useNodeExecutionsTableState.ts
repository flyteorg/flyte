import { useContext, useMemo, useState } from 'react';
import { ExecutionDataCacheContext } from '../contexts';
import { DetailedNodeExecution } from '../types';
import { mapNodeExecutionDetails } from '../utils';
import { NodeExecutionsTableProps } from './NodeExecutionsTable';
import { NodeExecutionsTableState } from './types';

export function useNodeExecutionsTableState({
    value
}: NodeExecutionsTableProps): NodeExecutionsTableState {
    const dataCache = useContext(ExecutionDataCacheContext);
    const executions = useMemo(
        () => mapNodeExecutionDetails(value, dataCache),
        [value]
    );

    const [selectedExecutionKey, setSelectedExecutionKey] = useState<
        string | null
    >(null);

    const selectedExecution = useMemo(
        () =>
            executions.find(
                ({ cacheKey }) => cacheKey === selectedExecutionKey
            ) || null,
        [executions, selectedExecutionKey]
    );

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
