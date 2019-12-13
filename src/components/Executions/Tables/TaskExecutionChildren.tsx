import { WaitForData } from 'components/common';
import { useTaskExecutionChildren } from 'components/hooks';
import {
    executionSortFields,
    limits,
    SortDirection,
    TaskExecution
} from 'models';
import * as React from 'react';
import { useDetailedNodeExecutions } from '../useDetailedNodeExecutions';
import { generateRowSkeleton } from './generateRowSkeleton';
import { generateColumns } from './nodeExecutionColumns';
import { NodeExecutionRow } from './NodeExecutionRow';
import { NoExecutionsContent } from './NoExecutionsContent';
import { useColumnStyles } from './styles';

export interface TaskExecutionChildrenProps {
    taskExecution: TaskExecution;
}

/** Renders a nested list of row items for children of a NodeExecution */
export const TaskExecutionChildren: React.FC<TaskExecutionChildrenProps> = ({
    taskExecution
}) => {
    const sort = {
        key: executionSortFields.createdAt,
        direction: SortDirection.ASCENDING
    };
    const nodeExecutions = useDetailedNodeExecutions(
        useTaskExecutionChildren(taskExecution.id, {
            sort,
            limit: limits.NONE
        })
    );

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
            {...nodeExecutions}
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
