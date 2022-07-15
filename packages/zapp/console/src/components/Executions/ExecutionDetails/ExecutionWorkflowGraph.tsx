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
import { fetchTaskExecutionList } from '../taskExecutionQueries';
import { getGroupedLogs } from '../TaskExecutionsList/utils';
import { NodeExecutionDetailsPanelContent } from './NodeExecutionDetailsPanelContent';

export interface ExecutionWorkflowGraphProps {
  nodeExecutions: NodeExecution[];
  workflowId: WorkflowId;
}

interface WorkflowNodeExecution extends NodeExecution {
  logsByPhase?: LogsByPhase;
}

/** Wraps a WorkflowGraph, customizing it to also show execution statuses */
export const ExecutionWorkflowGraph: React.FC<ExecutionWorkflowGraphProps> = ({
  nodeExecutions,
  workflowId,
}) => {
  const queryClient = useQueryClient();
  const workflowQuery = useQuery<Workflow, Error>(makeWorkflowQuery(queryClient, workflowId));

  const [nodeExecutionsWithResources, setNodeExecutionsWithResources] = useState<
    WorkflowNodeExecution[]
  >([]);
  const [selectedNodes, setSelectedNodes] = useState<string[]>([]);

  const nodeExecutionsById = useMemo(
    () => keyBy(nodeExecutionsWithResources, 'scopedId'),
    [nodeExecutionsWithResources],
  );
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
    let isCurrent = true;
    async function fetchData(nodeExecutions, queryClient) {
      const newValue = await Promise.all(
        nodeExecutions.map(async (nodeExecution) => {
          const taskExecutions = await fetchTaskExecutionList(queryClient, nodeExecution.id);

          const useNewMapTaskView = taskExecutions.every((taskExecution) => {
            const {
              closure: { taskType, metadata, eventVersion = 0 },
            } = taskExecution;
            return isMapTaskV1(
              eventVersion,
              metadata?.externalResources?.length ?? 0,
              taskType ?? undefined,
            );
          });
          const externalResources: ExternalResource[] = taskExecutions
            .map((taskExecution) => taskExecution.closure.metadata?.externalResources)
            .flat()
            .filter((resource): resource is ExternalResource => !!resource);

          const logsByPhase: LogsByPhase = getGroupedLogs(externalResources);

          return {
            ...nodeExecution,
            ...(useNewMapTaskView && logsByPhase.size > 0 && { logsByPhase }),
          };
        }),
      );

      if (isCurrent) {
        setNodeExecutionsWithResources(newValue);
      }
    }

    fetchData(nodeExecutions, queryClient);

    return () => {
      isCurrent = false;
    };
  }, [nodeExecutions]);

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
