import { Core } from 'flyteidl';
import {
    NodeExecutionDetails,
    NodeExecutionDisplayType
} from 'components/Executions/types';
import { flattenBranchNodes } from 'components/Executions/utils';
import { Workflow } from 'models/Workflow/types';
import { Identifier } from 'models/Common/types';
import { CompiledNode } from 'models/Node/types';
import { CompiledTask } from 'models/Task/types';
import { endNodeId, startNodeId } from 'models/Node/constants';
import { isIdEqual, UNKNOWN_DETAILS } from './types';

interface NodeExecutionInfo extends NodeExecutionDetails {
    parentTemplate: Identifier;
}

export interface CurrentExecutionDetails {
    executionId: Identifier;
    nodes: NodeExecutionInfo[];
}

const isParentNode = (type: string) =>
    type === NodeExecutionDisplayType.Workflow;

const getNodeDetails = (
    node: CompiledNode,
    tasks: CompiledTask[]
): NodeExecutionDetails => {
    if (node.taskNode) {
        const templateName = node.taskNode.referenceId.name;
        const task = tasks.find(t => t.template.id.name === templateName);
        return {
            displayId: node.id,
            displayName: templateName,
            displayType:
                task?.template.type ?? NodeExecutionDisplayType.UnknownTask,
            taskTemplate: task?.template
        };
    }

    if (node.workflowNode) {
        const info =
            node.workflowNode.launchplanRef ?? node.workflowNode.subWorkflowRef;
        return {
            displayId: node.id,
            displayName: info?.name ?? 'N/A',
            displayType: NodeExecutionDisplayType.Workflow
        };
    }

    // TODO: https://github.com/lyft/flyte/issues/655
    if (node.branchNode) {
        return {
            displayId: node.id,
            displayName: 'branchNode',
            displayType: NodeExecutionDisplayType.BranchNode
        };
    }

    return UNKNOWN_DETAILS;
};

export function createExecutionDetails(
    workflow: Workflow
): { nodes: CurrentExecutionDetails; map: Map<string, Core.IIdentifier> } {
    const mapNodeIdToTemplate = new Map<string, Core.IIdentifier>();
    const result: CurrentExecutionDetails = {
        executionId: workflow.id,
        nodes: []
    };

    if (!workflow.closure?.compiledWorkflow) {
        return { nodes: result, map: mapNodeIdToTemplate };
    }

    const {
        primary,
        subWorkflows = [],
        tasks = []
    } = workflow.closure?.compiledWorkflow;

    if (!isIdEqual(primary.template.id, workflow.id)) {
        console.log('WRONG');
    }

    const allWorkflows = [primary, ...subWorkflows];
    allWorkflows.forEach(w => {
        const nodes = w.template.nodes;
        nodes
            .map(flattenBranchNodes)
            .flat()
            .forEach(n => {
                if (n.id === startNodeId || n.id === endNodeId) {
                    // skip start and end nodes
                    return;
                }

                const details = getNodeDetails(n, tasks);
                if (
                    details?.displayName &&
                    isParentNode(details?.displayType)
                ) {
                    const info =
                        n.workflowNode?.launchplanRef ??
                        n.workflowNode?.subWorkflowRef;
                    if (info) {
                        mapNodeIdToTemplate.set(n.id, info);
                    }
                }
                result.nodes.push({
                    parentTemplate: w.template.id,
                    ...details
                });
            });
    });

    return { nodes: result, map: mapNodeIdToTemplate };
}
