import { DetailsPanel } from 'components/common/DetailsPanel';
import { WaitForQuery } from 'components/common/WaitForQuery';
import { DataError } from 'components/Errors/DataError';
import { makeWorkflowQuery } from 'components/Workflow/workflowQueries';
import { WorkflowGraph } from 'components/WorkflowGraph/WorkflowGraph';
import { keyBy } from 'lodash';
import { NodeExecution } from 'models/Execution/types';
import { endNodeId, startNodeId } from 'models/Node/constants';
import { Workflow, WorkflowId } from 'models/Workflow/types';
import * as React from 'react';
import { useQuery, useQueryClient } from 'react-query';
import { NodeExecutionsContext } from '../contexts';
import { NodeExecutionDetailsPanelContent } from './NodeExecutionDetailsPanelContent';

export interface ExecutionWorkflowGraphProps {
    nodeExecutions: NodeExecution[];
    workflowId: WorkflowId;
}

/** Wraps a WorkflowGraph, customizing it to also show execution statuses */
export const ExecutionWorkflowGraph: React.FC<ExecutionWorkflowGraphProps> = ({
    nodeExecutions,
    workflowId
}) => {
    const workflowQuery = useQuery<Workflow, Error>(
        makeWorkflowQuery(useQueryClient(), workflowId)
    );
    const nodeExecutionsById = React.useMemo(
        () => keyBy(nodeExecutions, 'id.nodeId'),
        [nodeExecutions]
    );

    const [selectedNodes, setSelectedNodes] = React.useState<string[]>([]);
    const onNodeSelectionChanged = (newSelection: string[]) => {
        const validSelection = newSelection.filter(nodeId => {
            if (nodeId === startNodeId || nodeId === endNodeId) {
                return false;
            }
            const execution = nodeExecutionsById[nodeId];
            if (!execution) {
                return false;
            }
            return true;
        });
        setSelectedNodes(validSelection);
    };

    // Note: flytegraph allows multiple selection, but we only support showing
    // a single item in the details panel
    const selectedExecution = selectedNodes.length
        ? nodeExecutionsById[selectedNodes[0]].id
        : null;

    const onCloseDetailsPanel = () => setSelectedNodes([]);

    const renderGraph = (workflow: Workflow) => (
        <WorkflowGraph
            onNodeSelectionChanged={onNodeSelectionChanged}
            nodeExecutionsById={nodeExecutionsById}
            workflow={workflow}
        />
    );

    return (
        <>
            <NodeExecutionsContext.Provider value={nodeExecutionsById}>
                <WaitForQuery errorComponent={DataError} query={workflowQuery}>
                    {renderGraph}
                </WaitForQuery>
            </NodeExecutionsContext.Provider>
            <DetailsPanel
                open={selectedExecution !== null}
                onClose={onCloseDetailsPanel}
            >
                {selectedExecution && (
                    <NodeExecutionDetailsPanelContent
                        onClose={onCloseDetailsPanel}
                        nodeExecutionId={selectedExecution}
                    />
                )}
            </DetailsPanel>
        </>
    );
};
