import { WaitForQuery } from 'components/common/WaitForQuery';
import { DataError } from 'components/Errors/DataError';
import * as React from 'react';
import { useAllChildNodeExecutionGroupsQuery } from '../nodeExecutionQueries';
import { NodeExecutionsRequestConfigContext } from '../contexts';
import { ExecutionWorkflowGraph } from './ExecutionWorkflowGraph';
import { NodeExecution } from 'models/Execution/types';

export const ExecutionChildrenLoader = ({ nodeExecutions, workflowId }) => {
    const requestConfig = React.useContext(NodeExecutionsRequestConfigContext);
    const childGroupsQuery = useAllChildNodeExecutionGroupsQuery(
        nodeExecutions,
        requestConfig
    );

    const renderGraphComponent = childGroups => {
        const output: any[] = [];
        for (let i = 0; i < childGroups.length; i++) {
            for (let j = 0; j < childGroups[i].length; j++) {
                for (
                    let k = 0;
                    k < childGroups[i][j].nodeExecutions.length;
                    k++
                ) {
                    output.push(
                        childGroups[i][j].nodeExecutions[k] as NodeExecution
                    );
                }
            }
        }
        const executions: NodeExecution[] = output.concat(nodeExecutions);
        return nodeExecutions.length > 0 ? (
            <ExecutionWorkflowGraph
                nodeExecutions={executions}
                workflowId={workflowId}
            />
        ) : null;
    };

    return (
        <WaitForQuery errorComponent={DataError} query={childGroupsQuery}>
            {renderGraphComponent}
        </WaitForQuery>
    );
};
