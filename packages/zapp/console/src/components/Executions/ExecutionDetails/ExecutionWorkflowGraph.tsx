import { DetailsPanel } from 'components/common/DetailsPanel';
import { WaitForQuery } from 'components/common/WaitForQuery';
import { DataError } from 'components/Errors/DataError';
import { makeWorkflowQuery } from 'components/Workflow/workflowQueries';
import { WorkflowGraph } from 'components/WorkflowGraph/WorkflowGraph';
import { keyBy } from 'lodash';
import { TaskExecutionPhase } from 'models/Execution/enums';
import { ExternalResource, LogsByPhase, NodeExecution } from 'models/Execution/types';
import { endNodeId, startNodeId } from 'models/Node/constants';
import { isMapTaskV1 } from 'models/Task/utils';
import { Workflow, WorkflowId } from 'models/Workflow/types';
import * as React from 'react';
import { useEffect, useMemo, useState } from 'react';
import { useQuery, useQueryClient } from 'react-query';
import { NodeExecutionsContext } from '../contexts';
import { getGroupedLogs } from '../TaskExecutionsList/utils';
import { useTaskExecutions, useTaskExecutionsRefresher } from '../useTaskExecutions';
import { NodeExecutionDetailsPanelContent } from './NodeExecutionDetailsPanelContent';

export interface ExecutionWorkflowGraphProps {
  nodeExecutions: NodeExecution[];
  workflowId: WorkflowId;
}

/** Wraps a WorkflowGraph, customizing it to also show execution statuses */
export const ExecutionWorkflowGraph: React.FC<ExecutionWorkflowGraphProps> = ({
  nodeExecutions,
  workflowId,
}) => {
  const workflowQuery = useQuery<Workflow, Error>(makeWorkflowQuery(useQueryClient(), workflowId));

  const nodeExecutionsWithResources = nodeExecutions.map((nodeExecution) => {
    const taskExecutions = useTaskExecutions(nodeExecution.id);
    useTaskExecutionsRefresher(nodeExecution, taskExecutions);

    const useNewMapTaskView = taskExecutions.value.every((taskExecution) => {
      const {
        closure: { taskType, metadata, eventVersion = 0 },
      } = taskExecution;
      return isMapTaskV1(
        eventVersion,
        metadata?.externalResources?.length ?? 0,
        taskType ?? undefined,
      );
    });
    const externalResources: ExternalResource[] = taskExecutions.value
      .map((taskExecution) => taskExecution.closure.metadata?.externalResources)
      .flat()
      .filter((resource): resource is ExternalResource => !!resource);

    const logsByPhase: LogsByPhase = getGroupedLogs(externalResources);

    return {
      ...nodeExecution,
      ...(useNewMapTaskView && logsByPhase.size > 0 && { logsByPhase }),
    };
  });

  const nodeExecutionsById = useMemo(
    () => keyBy(nodeExecutionsWithResources, 'scopedId'),
    [nodeExecutionsWithResources],
  );

  const [selectedNodes, setSelectedNodes] = useState<string[]>([]);
  const onNodeSelectionChanged = (newSelection: string[]) => {
    const validSelection = newSelection.filter((nodeId) => {
      if (nodeId === startNodeId || nodeId === endNodeId) {
        return false;
      }
      return true;
    });
    setSelectedNodes(validSelection);
  };

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

  const onCloseDetailsPanel = () => {
    setSelectedPhase(undefined);
    setIsDetailsTabClosed(true);
    setSelectedNodes([]);
  };

  const [selectedPhase, setSelectedPhase] = useState<TaskExecutionPhase | undefined>(undefined);
  const [isDetailsTabClosed, setIsDetailsTabClosed] = useState<boolean>(!selectedExecution);

  useEffect(() => {
    setIsDetailsTabClosed(!selectedExecution);
  }, [selectedExecution]);

  const renderGraph = (workflow: Workflow) => (
    <WorkflowGraph
      onNodeSelectionChanged={onNodeSelectionChanged}
      selectedPhase={selectedPhase}
      onPhaseSelectionChanged={setSelectedPhase}
      isDetailsTabClosed={isDetailsTabClosed}
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
