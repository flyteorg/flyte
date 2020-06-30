import { DetailsPanel } from 'components/common';
import { WorkflowGraph } from 'components/WorkflowGraph';
import { keyBy } from 'lodash';
import { endNodeId, NodeExecution, startNodeId } from 'models';
import { Workflow } from 'models/Workflow';
import * as React from 'react';
import { NodeExecutionsContext } from '../contexts';
import { useDetailedNodeExecutions } from '../useDetailedNodeExecutions';
import { NodeExecutionDetails } from './NodeExecutionDetails';
import { TaskExecutionNodeRenderer } from './TaskExecutionNodeRenderer/TaskExecutionNodeRenderer';

export interface ExecutionWorkflowGraphProps {
    nodeExecutions: NodeExecution[];
    workflow: Workflow;
}

/** Wraps a WorkflowGraph, customizing it to also show execution statuses */
export const ExecutionWorkflowGraph: React.FC<ExecutionWorkflowGraphProps> = ({
    nodeExecutions,
    workflow
}) => {
    const detailedNodeExecutions = useDetailedNodeExecutions(nodeExecutions);
    const nodeExecutionsById = React.useMemo(
        () => keyBy(detailedNodeExecutions, 'id.nodeId'),
        [detailedNodeExecutions]
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

    const selectedExecution = selectedNodes.length
        ? nodeExecutionsById[selectedNodes[0]]
        : null;
    const onCloseDetailsPanel = () => setSelectedNodes([]);

    return (
        <>
            <NodeExecutionsContext.Provider value={nodeExecutionsById}>
                <WorkflowGraph
                    onNodeSelectionChanged={onNodeSelectionChanged}
                    nodeRenderer={TaskExecutionNodeRenderer}
                    selectedNodes={selectedNodes}
                    workflow={workflow}
                />
            </NodeExecutionsContext.Provider>
            <DetailsPanel
                open={selectedExecution !== null}
                onClose={onCloseDetailsPanel}
            >
                {selectedExecution && (
                    <NodeExecutionDetails
                        onClose={onCloseDetailsPanel}
                        execution={selectedExecution}
                    />
                )}
            </DetailsPanel>
        </>
    );
};
