import { WaitForData } from 'components/common';
import { waitForAllFetchables } from 'components/hooks';
import { Execution } from 'models';
import * as React from 'react';
import { useWorkflowExecutionState } from '../useWorkflowExecutionState';
import { generateRowSkeleton } from './generateRowSkeleton';
import { generateColumns } from './nodeExecutionColumns';
import { NodeExecutionRow } from './NodeExecutionRow';
import { NoExecutionsContent } from './NoExecutionsContent';
import { useColumnStyles } from './styles';

export interface WorkflowExecutionChildrenProps {
    execution: Execution;
}

/** Renders a nested list of row items for children of a WorkflowExecution */
export const WorkflowExecutionChildren: React.FC<WorkflowExecutionChildrenProps> = ({
    execution
}) => {
    // Note: Not applying a filter to nested node executions
    const { workflow, nodeExecutions } = useWorkflowExecutionState(execution);

    const columnStyles = useColumnStyles();
    // Memoizing columns so they won't be re-generated unless the styles change
    const { columns, Skeleton } = React.useMemo(() => {
        const columns = generateColumns(columnStyles);
        return { columns, Skeleton: generateRowSkeleton(columns) };
    }, [columnStyles]);
    return (
        <WaitForData
            spinnerVariant="medium"
            loadingComponent={Skeleton}
            {...waitForAllFetchables([workflow, nodeExecutions])}
        >
            {nodeExecutions.value.length ? (
                nodeExecutions.value.map(execution => (
                    <NodeExecutionRow
                        columns={columns}
                        key={execution.cacheKey}
                        execution={execution}
                    />
                ))
            ) : (
                <NoExecutionsContent />
            )}
        </WaitForData>
    );
};
