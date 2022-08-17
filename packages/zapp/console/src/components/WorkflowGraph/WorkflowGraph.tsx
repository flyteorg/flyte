import { dNode } from 'models/Graph/types';
import { Workflow } from 'models/Workflow/types';
import * as React from 'react';
import ReactFlowGraphComponent from 'components/flytegraph/ReactFlow/ReactFlowGraphComponent';
import { Error } from 'models/Common/types';
import { NonIdealState } from 'components/common/NonIdealState';
import { DataError } from 'components/Errors/DataError';
import { WaitForQuery } from 'components/common/WaitForQuery';
import { useQuery } from 'react-query';
import { makeNodeExecutionDynamicWorkflowQuery } from 'components/Workflow/workflowQueries';
import { createDebugLogger } from 'common/log';
import { CompiledNode } from 'models/Node/types';
import { TaskExecutionPhase } from 'models/Execution/enums';
import { NodeExecutionsByIdContext } from 'components/Executions/contexts';
import { useContext } from 'react';
import { checkForDynamicExecutions } from 'components/common/utils';
import { transformerWorkflowToDag } from './transformerWorkflowToDag';

export interface WorkflowGraphProps {
  onNodeSelectionChanged: (selectedNodes: string[]) => void;
  onPhaseSelectionChanged: (phase: TaskExecutionPhase) => void;
  selectedPhase?: TaskExecutionPhase;
  isDetailsTabClosed: boolean;
  workflow: Workflow;
}

interface PrepareDAGResult {
  dag: dNode | null;
  staticExecutionIdsMap?: any;
  error?: Error;
}

const debug = createDebugLogger('@WorkflowGraph');

function workflowToDag(workflow: Workflow): PrepareDAGResult {
  try {
    if (!workflow.closure) {
      throw new Error('Workflow has no closure');
    }
    if (!workflow.closure.compiledWorkflow) {
      throw new Error('Workflow closure missing a compiled workflow');
    }
    const { compiledWorkflow } = workflow.closure;
    const { dag, staticExecutionIdsMap } = transformerWorkflowToDag(compiledWorkflow);

    debug('workflowToDag:dag', dag);

    return { dag, staticExecutionIdsMap };
  } catch (e) {
    return {
      dag: null,
      error: e as Error,
    };
  }
}

export interface DynamicWorkflowMapping {
  rootGraphNodeId: CompiledNode;
  dynamicWorkflow: any;
  dynamicExecutions: any[];
}
export const WorkflowGraph: React.FC<WorkflowGraphProps> = (props) => {
  const {
    onNodeSelectionChanged,
    onPhaseSelectionChanged,
    selectedPhase,
    isDetailsTabClosed,
    workflow,
  } = props;
  const nodeExecutionsById = useContext(NodeExecutionsByIdContext);
  const { dag, staticExecutionIdsMap, error } = workflowToDag(workflow);

  const dynamicParents = checkForDynamicExecutions(nodeExecutionsById, staticExecutionIdsMap);
  const dynamicWorkflowQuery = useQuery(makeNodeExecutionDynamicWorkflowQuery(dynamicParents));
  const renderReactFlowGraph = (dynamicWorkflows) => {
    debug('DynamicWorkflows:', dynamicWorkflows);
    let mergedDag = dag;
    for (const dynamicId in dynamicWorkflows) {
      if (staticExecutionIdsMap[dynamicId]) {
        if (workflow.closure?.compiledWorkflow) {
          const dynamicWorkflow = transformerWorkflowToDag(
            workflow.closure?.compiledWorkflow,
            dynamicWorkflows,
          );
          mergedDag = dynamicWorkflow.dag;
        }
      }
    }
    const merged = mergedDag;
    return (
      <ReactFlowGraphComponent
        dynamicWorkflows={dynamicWorkflows}
        data={merged}
        onNodeSelectionChanged={onNodeSelectionChanged}
        onPhaseSelectionChanged={onPhaseSelectionChanged}
        selectedPhase={selectedPhase}
        isDetailsTabClosed={isDetailsTabClosed}
      />
    );
  };

  if (error) {
    return <NonIdealState title="Cannot render Workflow graph" description={error.message} />;
  } else {
    return (
      <WaitForQuery errorComponent={DataError} query={dynamicWorkflowQuery}>
        {renderReactFlowGraph}
      </WaitForQuery>
    );
  }
};
