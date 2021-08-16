import { Identifier } from 'models/Common/types';
import { endNodeId, startNodeId } from 'models/Node/constants';
import { CompiledWorkflow } from 'models/Workflow/types';
import { CompiledNode, TaskNode } from 'models/Node/types';
import { CompiledTask, TaskTemplate } from 'models/Task/types';
import { dTypes } from 'models/Graph/types';

/**
 * @TODO these are dupes for testing, remove once tests fixed
 */
export const DISPLAY_NAME_START = 'start';
export const DISPLAY_NAME_END = 'end';

export function isStartNode(node: any) {
    return node.id === startNodeId;
}

export function isEndNode(node: any) {
    return node.id === endNodeId;
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
export const getDisplayName = (context: any): string => {
    let fullName;
    if (context.metadata) {
        // Compiled Node with Meta
        fullName = context.metadata.name;
    } else if (context.id) {
        // Compiled Node (start/end)
        fullName = context.id;
    } else {
        // CompiledWorkflow
        fullName = context.template.id.name;
    }

    if (fullName == startNodeId) {
        return DISPLAY_NAME_START;
    } else if (fullName == endNodeId) {
        return DISPLAY_NAME_END;
    } else if (fullName.indexOf('.') > 0) {
        return fullName.substr(
            fullName.lastIndexOf('.') + 1,
            fullName.length - 1
        );
    } else {
        return fullName;
    }
};

export const getThenNodeFromBranch = (node: CompiledNode) => {
    return node.branchNode?.ifElse?.case?.thenNode as CompiledNode;
};

/**
 * Returns the id for CompiledWorkflow
 * @param context   will find id for this entity
 * @returns id
 */
export const getWorkflowId = (workflow: CompiledWorkflow): string => {
    return workflow.template.id.name;
};

export const getNodeTypeFromCompiledNode = (node: CompiledNode): dTypes => {
    if (isStartNode(node)) {
        return dTypes.start;
    } else if (isEndNode(node)) {
        return dTypes.end;
    } else if (node.branchNode) {
        return dTypes.branch;
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

export const getTaskTypeFromCompiledNode = (
    taskNode: TaskNode,
    tasks: CompiledTask[]
) => {
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
