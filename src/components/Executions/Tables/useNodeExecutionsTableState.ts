import { useState } from 'react';
import { DetailedNodeExecution } from '../types';
import { NodeExecutionsTableProps } from './NodeExecutionsTable';
import { NodeExecutionsTableState } from './types';

export function useNodeExecutionsTableState(
    props: NodeExecutionsTableProps
): NodeExecutionsTableState {
    const { value: executions } = props;

    const [
        selectedExecution,
        setSelectedExecution
    ] = useState<DetailedNodeExecution | null>(null);

    return {
        executions,
        selectedExecution,
        setSelectedExecution
    };
}
