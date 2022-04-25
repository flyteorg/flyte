import { transformerWorkflowToDag } from 'components/WorkflowGraph/transformerWorkflowToDag';
import { getTaskDisplayType } from 'components/Executions/utils';
import { NodeExecutionDetails, NodeExecutionDisplayType } from 'components/Executions/types';
import { Workflow } from 'models/Workflow/types';
import { Identifier } from 'models/Common/types';
import { CompiledTask } from 'models/Task/types';
import { dNode } from 'models/Graph/types';
import { isEndNode, isStartNode } from 'components/WorkflowGraph/utils';
import { UNKNOWN_DETAILS } from './types';

interface NodeExecutionInfo extends NodeExecutionDetails {
  scopedId?: string;
}

export interface CurrentExecutionDetails {
  executionId: Identifier;
  nodes: NodeExecutionInfo[];
}

function convertToPlainNodes(nodes: dNode[], level = 0): dNode[] {
  const result: dNode[] = [];
  if (!nodes || nodes.length === 0) {
    return result;
  }
  nodes.forEach((node) => {
    if (isStartNode(node) || isEndNode(node)) {
      return;
    }
    result.push({ ...node, level });
    if (node.nodes.length > 0) {
      result.push(...convertToPlainNodes(node.nodes, level + 1));
    }
  });
  return result;
}

const getNodeDetails = (node: dNode, tasks: CompiledTask[]): NodeExecutionInfo => {
  if (node.value.taskNode) {
    const templateName = node.value.taskNode.referenceId.name ?? node.name;
    const task = tasks.find((t) => t.template.id.name === templateName);
    const taskType = getTaskDisplayType(task?.template.type);

    return {
      scopedId: node.scopedId,
      displayId: node.value.id ?? node.id,
      displayName: templateName,
      displayType: taskType,
      taskTemplate: task?.template,
    };
  }

  if (node.value.workflowNode) {
    const workflowNode = node.value.workflowNode;
    const info = workflowNode.launchplanRef ?? workflowNode.subWorkflowRef;
    return {
      scopedId: node.scopedId,
      displayId: node.value.id ?? node.id,
      displayName: node.name ?? info?.name ?? 'N/A',
      displayType: NodeExecutionDisplayType.Workflow,
    };
  }

  // TODO: https://github.com/flyteorg/flyteconsole/issues/274
  if (node.value.branchNode) {
    return {
      scopedId: node.scopedId,
      displayId: node.value.id ?? node.id,
      displayName: 'branchNode',
      displayType: NodeExecutionDisplayType.BranchNode,
    };
  }

  return UNKNOWN_DETAILS;
};

export function createExecutionDetails(workflow: Workflow): CurrentExecutionDetails {
  const result: CurrentExecutionDetails = {
    executionId: workflow.id,
    nodes: [],
  };

  if (!workflow.closure?.compiledWorkflow) {
    return result;
  }

  const compiledWorkflow = workflow.closure?.compiledWorkflow;
  const { tasks = [] } = compiledWorkflow;

  let dNodes = transformerWorkflowToDag(compiledWorkflow).dag.nodes ?? [];
  dNodes = convertToPlainNodes(dNodes);

  dNodes.forEach((n) => {
    const details = getNodeDetails(n, tasks);
    result.nodes.push({
      ...details,
    });
  });

  return result;
}
