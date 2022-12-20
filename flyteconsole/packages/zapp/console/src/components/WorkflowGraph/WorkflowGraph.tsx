import * as React from 'react';
import ReactFlowGraphComponent from 'components/flytegraph/ReactFlow/ReactFlowGraphComponent';
import { Error } from 'models/Common/types';
import { NonIdealState } from 'components/common/NonIdealState';
import { CompiledNode } from 'models/Node/types';
import { TaskExecutionPhase } from 'models/Execution/enums';
import { dNode } from 'models/Graph/types';
import t from './strings';

export interface WorkflowGraphProps {
  onNodeSelectionChanged: (selectedNodes: string[]) => void;
  onPhaseSelectionChanged: (phase: TaskExecutionPhase) => void;
  selectedPhase?: TaskExecutionPhase;
  isDetailsTabClosed: boolean;
  mergedDag: any;
  error: Error | null;
  dynamicWorkflows: any;
  initialNodes: dNode[];
}
export interface DynamicWorkflowMapping {
  rootGraphNodeId: CompiledNode;
  dynamicWorkflow: any;
  dynamicExecutions: any[];
}
export const WorkflowGraph: React.FC<WorkflowGraphProps> = ({
  onNodeSelectionChanged,
  onPhaseSelectionChanged,
  selectedPhase,
  isDetailsTabClosed,
  mergedDag,
  error,
  dynamicWorkflows,
  initialNodes,
}) => {
  if (error) {
    return <NonIdealState title={t('graphErrorTitle')} description={error.message} />;
  }

  // If the dag is empty, show the message, instead of trying to display it
  if (!mergedDag) {
    return <NonIdealState title={t('graphErrorTitle')} description={t('graphErrorEmptyGraph')} />;
  }

  return (
    <ReactFlowGraphComponent
      dynamicWorkflows={dynamicWorkflows}
      data={mergedDag}
      onNodeSelectionChanged={onNodeSelectionChanged}
      onPhaseSelectionChanged={onPhaseSelectionChanged}
      selectedPhase={selectedPhase}
      isDetailsTabClosed={isDetailsTabClosed}
      initialNodes={initialNodes}
    />
  );
};
