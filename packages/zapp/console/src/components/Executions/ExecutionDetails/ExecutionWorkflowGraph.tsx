import { DetailsPanel } from 'components/common/DetailsPanel';
import { WaitForQuery } from 'components/common/WaitForQuery';
import { DataError } from 'components/Errors/DataError';
import { makeWorkflowQuery } from 'components/Workflow/workflowQueries';
import { WorkflowGraph } from 'components/WorkflowGraph/WorkflowGraph';
import { TaskExecutionPhase } from 'models/Execution/enums';
import { endNodeId, startNodeId } from 'models/Node/constants';
import { Workflow, WorkflowId } from 'models/Workflow/types';
import * as React from 'react';
import { useContext, useEffect, useState } from 'react';
import { useQuery, useQueryClient } from 'react-query';
import { NodeExecutionsByIdContext } from '../contexts';
import { NodeExecutionDetailsPanelContent } from './NodeExecutionDetailsPanelContent';

export interface ExecutionWorkflowGraphProps {
  workflowId: WorkflowId;
}

/** Wraps a WorkflowGraph, customizing it to also show execution statuses */
export const ExecutionWorkflowGraph: React.FC<ExecutionWorkflowGraphProps> = ({ workflowId }) => {
  const queryClient = useQueryClient();
  const workflowQuery = useQuery<Workflow, Error>(makeWorkflowQuery(queryClient, workflowId));

  const [selectedNodes, setSelectedNodes] = useState<string[]>([]);
  const nodeExecutionsById = useContext(NodeExecutionsByIdContext);

  // Note: flytegraph allows multiple selection, but we only support showing
  // a single item in the details panel
  const selectedExecution = selectedNodes.length
    ? nodeExecutionsById[selectedNodes[0]]
      ? nodeExecutionsById[selectedNodes[0]].id
      : {
          nodeId: selectedNodes[0],
          executionId: nodeExecutionsById[Object.keys(nodeExecutionsById)[0]].id.executionId,
        }
    : null;

  const [selectedPhase, setSelectedPhase] = useState<TaskExecutionPhase | undefined>(undefined);
  const [isDetailsTabClosed, setIsDetailsTabClosed] = useState<boolean>(!selectedExecution);

  useEffect(() => {
    setIsDetailsTabClosed(!selectedExecution);
  }, [selectedExecution]);

  const onNodeSelectionChanged = (newSelection: string[]) => {
    const validSelection = newSelection.filter((nodeId) => {
      if (nodeId === startNodeId || nodeId === endNodeId) {
        return false;
      }
      return true;
    });
    setSelectedNodes(validSelection);
  };

  const onCloseDetailsPanel = () => {
    setSelectedPhase(undefined);
    setIsDetailsTabClosed(true);
    setSelectedNodes([]);
  };

  const renderGraph = (workflow: Workflow) => (
    <WorkflowGraph
      onNodeSelectionChanged={onNodeSelectionChanged}
      selectedPhase={selectedPhase}
      onPhaseSelectionChanged={setSelectedPhase}
      isDetailsTabClosed={isDetailsTabClosed}
      workflow={workflow}
    />
  );

  return (
    <>
      <WaitForQuery errorComponent={DataError} query={workflowQuery}>
        {renderGraph}
      </WaitForQuery>
      <DetailsPanel open={!!selectedExecution} onClose={onCloseDetailsPanel}>
        {selectedExecution && (
          <NodeExecutionDetailsPanelContent
            onClose={onCloseDetailsPanel}
            phase={selectedPhase}
            nodeExecutionId={selectedExecution}
          />
        )}
      </DetailsPanel>
    </>
  );
};
