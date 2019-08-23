import { WaitForData } from 'components/common';
import {
    RefreshConfig,
    useDataRefresher,
    useWorkflowExecution
} from 'components/hooks';
import { isEqual } from 'lodash';
import { Execution, NodeExecutionClosure, WorkflowNodeMetadata } from 'models';
import * as React from 'react';
import { executionIsTerminal, executionRefreshIntervalMs } from '..';
import { ExecutionContext } from '../ExecutionDetails/contexts';
import { DetailedNodeExecution } from '../types';
import { useDetailedTaskExecutions } from '../useDetailedTaskExecutions';
import {
    useTaskExecutions,
    useTaskExecutionsRefresher
} from '../useTaskExecutions';
import { generateRowSkeleton } from './generateRowSkeleton';
import { NoExecutionsContent } from './NoExecutionsContent';
import { useColumnStyles } from './styles';
import { generateColumns as generateTaskExecutionColumns } from './taskExecutionColumns';
import { TaskExecutionRow } from './TaskExecutionRow';
import { generateColumns as generateWorkflowExecutionColumns } from './workflowExecutionColumns';
import { WorkflowExecutionRow } from './WorkflowExecutionRow';

export interface NodeExecutionChildrenProps {
    execution: DetailedNodeExecution;
}

const TaskNodeExecutionChildren: React.FC<NodeExecutionChildrenProps> = ({
    execution: nodeExecution
}) => {
    const taskExecutions = useDetailedTaskExecutions(
        useTaskExecutions(nodeExecution.id)
    );
    useTaskExecutionsRefresher(nodeExecution, taskExecutions);

    const columnStyles = useColumnStyles();
    // Memoizing columns so they won't be re-generated unless the styles change
    const { columns, Skeleton } = React.useMemo(
        () => {
            const columns = generateTaskExecutionColumns(columnStyles);
            return { columns, Skeleton: generateRowSkeleton(columns) };
        },
        [columnStyles]
    );
    return (
        <WaitForData
            spinnerVariant="medium"
            loadingComponent={Skeleton}
            {...taskExecutions}
        >
            {taskExecutions.value.length ? (
                taskExecutions.value.map(taskExecution => (
                    <TaskExecutionRow
                        columns={columns}
                        key={taskExecution.cacheKey}
                        execution={taskExecution}
                        nodeExecution={nodeExecution}
                    />
                ))
            ) : (
                <NoExecutionsContent />
            )}
        </WaitForData>
    );
};

interface WorkflowNodeExecution extends DetailedNodeExecution {
    closure: NodeExecutionClosure & {
        workflowNodeMetadata: WorkflowNodeMetadata;
    };
}
interface WorkflowNodeExecutionChildrenProps
    extends NodeExecutionChildrenProps {
    execution: WorkflowNodeExecution;
}

const executionRefreshConfig: RefreshConfig<Execution> = {
    interval: executionRefreshIntervalMs,
    valueIsFinal: executionIsTerminal
};

const WorkflowNodeExecutionChildren: React.FC<
    WorkflowNodeExecutionChildrenProps
> = ({ execution }) => {
    const { executionId } = execution.closure.workflowNodeMetadata;
    const workflowExecution = useWorkflowExecution(executionId).fetchable;
    useDataRefresher(executionId, workflowExecution, executionRefreshConfig);

    const columnStyles = useColumnStyles();
    // Memoizing columns so they won't be re-generated unless the styles change
    const { columns, Skeleton } = React.useMemo(
        () => {
            const columns = generateWorkflowExecutionColumns(columnStyles);
            return { columns, Skeleton: generateRowSkeleton(columns) };
        },
        [columnStyles]
    );
    return (
        <WaitForData
            spinnerVariant="medium"
            loadingComponent={Skeleton}
            {...workflowExecution}
        >
            {workflowExecution.value ? (
                <WorkflowExecutionRow
                    columns={columns}
                    execution={workflowExecution.value}
                />
            ) : (
                <NoExecutionsContent />
            )}
        </WaitForData>
    );
};

/** Renders a nested list of row items for children of a NodeExecution */
export const NodeExecutionChildren: React.FC<
    NodeExecutionChildrenProps
> = props => {
    const { workflowNodeMetadata } = props.execution.closure;
    const { execution: topExecution } = React.useContext(ExecutionContext);

    // Nested NodeExecutions will sometimes have `workflowNodeMetadata` that
    // points to the parent WorkflowExecution. We only want to expand workflow
    // nodes that point to a different workflow execution
    if (
        workflowNodeMetadata &&
        !isEqual(workflowNodeMetadata.executionId, topExecution.id)
    ) {
        return (
            <WorkflowNodeExecutionChildren
                {...props}
                execution={props.execution as WorkflowNodeExecution}
            />
        );
    }
    return <TaskNodeExecutionChildren {...props} />;
};
