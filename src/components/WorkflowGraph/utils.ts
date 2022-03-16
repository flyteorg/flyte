import { Identifier } from 'models/Common/types';
import { endNodeId, startNodeId } from 'models/Node/constants';
import { CompiledWorkflow, Workflow } from 'models/Workflow/types';
import { CompiledNode, TaskNode } from 'models/Node/types';
import { CompiledTask, TaskTemplate } from 'models/Task/types';
import { dTypes, dNode } from 'models/Graph/types';
import { transformerWorkflowToDag } from './transformerWorkflowToDag';
/**
 * @TODO these are dupes for testing, remove once tests fixed
 */
export const DISPLAY_NAME_START = 'start';
export const DISPLAY_NAME_END = 'end';

export const isStartOrEndNode = (node: any) => {
  return node.id === startNodeId || node.id === endNodeId;
};

export function isStartNode(node: any) {
  return node.id === startNodeId;
}

export function isEndNode(node: any) {
  return node.id === endNodeId;
}

export function isExpanded(node: any) {
  return !!node.expanded;
}

/**
 * Utility funciton assumes (loose) parity between [a]->[b] if matching
 * keys have matching values.
 * @param a     object
 * @param b     object
 * @returns     boolean
 */
export const checkIfObjectsAreSame = (a, b) => {
  for (const k in a) {
    if (a[k] != b[k]) {
      return false;
    }
  }
  return true;
};

/**
 * Returns a display name from either workflows or nodes
 * @param context input can be either CompiledWorkflow or CompiledNode
 * @returns Display name
 */
export const getDisplayName = (context: any, truncate = true): string => {
  let displayName = '';
  if (context.metadata) {
    // Compiled Node with Meta
    displayName = context.metadata.name;
  } else if (context.displayId) {
    // NodeExecutionDetails
    displayName = context.displayId;
  } else if (context.template?.id?.name) {
    // CompiledWorkflow
    displayName = context.template.id.name;
  } else if (context.id) {
    // Compiled Node (start/end)
    displayName = context.id;
  }

  if (displayName == startNodeId) {
    return DISPLAY_NAME_START;
  } else if (displayName == endNodeId) {
    return DISPLAY_NAME_END;
  } else if (displayName.indexOf('.') > 0 && truncate) {
    /* Note: for displaying truncated task name */
    return displayName.substring(displayName.lastIndexOf('.') + 1, displayName.length);
  }

  return displayName;
};

/**
 * Returns the id for CompiledWorkflow
 * @param context   will find id for this entity
 * @returns id
 */
export const getWorkflowId = (workflow: CompiledWorkflow): string => {
  return workflow.template.id.name;
};

export const createWorkflowNodeFromDynamic = (dw) => {
  return {
    subWorkflowRef: {
      domain: 'development',
      name: '',
      project: '',
    },
  };
};

export const getNodeTypeFromCompiledNode = (node: CompiledNode): dTypes => {
  if (isStartNode(node)) {
    return dTypes.start;
  } else if (isEndNode(node)) {
    return dTypes.end;
  } else if (node.branchNode) {
    return dTypes.subworkflow;
  } else if (node.workflowNode) {
    return dTypes.subworkflow;
  } else {
    return dTypes.task;
  }
};

export const getSubWorkflowFromId = (id, workflow) => {
  const { subWorkflows } = workflow;
  /* Find current matching entitity from subWorkflows */
  for (const k in subWorkflows) {
    const subWorkflowId = subWorkflows[k].template.id;
    if (checkIfObjectsAreSame(subWorkflowId, id)) {
      return subWorkflows[k];
    }
  }
  return false;
};

export const getTaskTypeFromCompiledNode = (taskNode: TaskNode, tasks: CompiledTask[]) => {
  for (let i = 0; i < tasks.length; i++) {
    const compiledTask: CompiledTask = tasks[i];
    const taskTemplate: TaskTemplate = compiledTask.template;
    const templateId: Identifier = taskTemplate.id;
    if (checkIfObjectsAreSame(templateId, taskNode.referenceId)) {
      return compiledTask;
    }
  }
  return null;
};

export const getNodeNameFromDag = (dagData: dNode, nodeId: string) => {
  const id = nodeId.slice(nodeId.lastIndexOf('-') + 1);
  const node = dagData[id];

  return getNodeTemplateName(node);
};

export const getNodeTemplateName = (node: dNode) => {
  const value = node?.value;
  if (value?.workflowNode) {
    const { launchplanRef, subWorkflowRef } = node.value.workflowNode;
    const identifier = (launchplanRef ? launchplanRef : subWorkflowRef) as Identifier;
    return identifier.name;
  }

  if (value?.taskNode) {
    return value.taskNode.referenceId.name;
  }

  return '';
};

export const transformWorkflowToKeyedDag = (workflow: Workflow) => {
  if (!workflow.closure?.compiledWorkflow) return {};

  const { dag } = transformerWorkflowToDag(workflow.closure?.compiledWorkflow);
  const data = {};
  dag.nodes.forEach((node) => {
    data[`${node.id}`] = node;
  });
  return data;
};
