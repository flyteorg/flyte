import {
    BaseExecutionClosure,
    Execution,
    NodeExecution,
    TaskExecution,
    terminalExecutionStates,
    terminalNodeExecutionStates,
    terminalTaskExecutionStates
} from 'models/Execution';

import {
    NodeExecutionPhase,
    TaskExecutionPhase,
    WorkflowExecutionPhase
} from 'models/Execution/enums';

import { log } from 'common/log';
import { durationToMilliseconds, timestampToDate } from 'common/utils';
import { getCacheKey } from 'components/Cache';
import { extractTaskTemplates } from 'components/hooks/utils';
import { keyBy } from 'lodash';
import {
    CompiledNode,
    endNodeId,
    Identifier,
    startNodeId,
    TaskTemplate,
    TaskType,
    Workflow
} from 'models';
import {
    nodeExecutionPhaseConstants,
    taskExecutionPhaseConstants,
    taskTypeToNodeExecutionDisplayType,
    workflowExecutionPhaseConstants
} from './constants';
import {
    DetailedNodeExecution,
    ExecutionPhaseConstants,
    NodeExecutionDisplayType
} from './types';

/** Given an execution phase, returns a set of constants (i.e. color, display
 * string) used to represent it in various UI components.
 */
export function getWorkflowExecutionPhaseConstants(
    phase: WorkflowExecutionPhase
): ExecutionPhaseConstants {
    return (
        workflowExecutionPhaseConstants[phase] ||
        workflowExecutionPhaseConstants[WorkflowExecutionPhase.UNDEFINED]
    );
}

/** Maps a `WorkflowExecutionPhase` value to a corresponding color string */
export function workflowExecutionPhaseToColor(
    phase: WorkflowExecutionPhase
): string {
    return getWorkflowExecutionPhaseConstants(phase).badgeColor;
}

/** Given an execution phase, returns a set of constants (i.e. color, display
 * string) used to represent it in various UI components.
 */
export function getNodeExecutionPhaseConstants(
    phase: NodeExecutionPhase
): ExecutionPhaseConstants {
    return (
        nodeExecutionPhaseConstants[phase] ||
        nodeExecutionPhaseConstants[NodeExecutionPhase.UNDEFINED]
    );
}

/** Maps a `NodeExecutionPhase` value to a corresponding color string */
export function nodeExecutionPhaseToColor(phase: NodeExecutionPhase): string {
    return getNodeExecutionPhaseConstants(phase).badgeColor;
}

/** Given an execution phase, returns a set of constants (i.e. color, display
 * string) used to represent it in various UI components.
 */
export function getTaskExecutionPhaseConstants(
    phase: TaskExecutionPhase
): ExecutionPhaseConstants {
    return (
        taskExecutionPhaseConstants[phase] ||
        taskExecutionPhaseConstants[TaskExecutionPhase.UNDEFINED]
    );
}

/** Maps a `TaskExecutionPhase` value to a corresponding color string */
export function taskExecutionPhaseToColor(phase: TaskExecutionPhase): string {
    return getTaskExecutionPhaseConstants(phase).badgeColor;
}

/** Determines if a workflow execution can be considered finalized and will not
 * change state again.
 */
export const executionIsTerminal = (execution: Execution) =>
    execution.closure &&
    terminalExecutionStates.includes(execution.closure.phase);

/** Determines if a node execution can be considered finalized and will not
 * change state again.
 */
export const nodeExecutionIsTerminal = (nodeExecution: NodeExecution) =>
    nodeExecution.closure &&
    terminalNodeExecutionStates.includes(nodeExecution.closure.phase);

/** Determines if a task execution can be considered finalized and will not
 * change state again.
 */
export const taskExecutionIsTerminal = (taskExecution: TaskExecution) =>
    taskExecution.closure &&
    terminalTaskExecutionStates.includes(taskExecution.closure.phase);

/** Assigns display information to NodeExecutions. Each NodeExecution has an
 * associated `nodeId`. If a `Workflow` is provided, node/task information will
 * be pulled from the workflow closure to determine the node type.
 */
export function mapNodeExecutionDetails(
    executions: NodeExecution[],
    workflow?: Workflow
) {
    let nodesById: Dictionary<CompiledNode> = {};
    let taskTemplates: Dictionary<TaskTemplate> = {};

    if (workflow) {
        if (!workflow.closure) {
            throw new Error('Workflow has no closure');
        }
        if (!workflow.closure.compiledWorkflow) {
            throw new Error('Workflow closure missing a compiled workflow');
        }

        taskTemplates = keyBy(extractTaskTemplates(workflow), t =>
            getCacheKey(t.id)
        );
        nodesById = keyBy(
            workflow.closure.compiledWorkflow.primary.template.nodes,
            'id'
        );
    }

    return executions
        .filter(execution => {
            // Exclude the start/end nodes from the renderered list
            const { nodeId } = execution.id;
            return !(nodeId === startNodeId || nodeId === endNodeId);
        })
        .map<DetailedNodeExecution>(execution => {
            const { nodeId } = execution.id;
            const node = nodesById[nodeId];
            const cacheKey = getCacheKey(execution.id);
            let displayId = nodeId;
            let displayType = NodeExecutionDisplayType.Unknown;
            let taskTemplate: TaskTemplate | undefined = undefined;

            if (!node) {
                return { ...execution, cacheKey, displayId, displayType };
            }

            if (node.branchNode) {
                displayId = nodeId;
                displayType = NodeExecutionDisplayType.BranchNode;
            } else if (node.taskNode) {
                displayType = NodeExecutionDisplayType.UnknownTask;
                taskTemplate =
                    taskTemplates[getCacheKey(node.taskNode.referenceId)];

                if (!taskTemplate) {
                    log.warn(
                        `Unexpected missing workflow task for node ${nodeId}`
                    );
                    displayType = NodeExecutionDisplayType.UnknownTask;
                } else {
                    displayId = taskTemplate.id.name;
                    displayType =
                        taskTypeToNodeExecutionDisplayType[
                            taskTemplate.type as TaskType
                        ];
                    if (!displayType) {
                        displayType = NodeExecutionDisplayType.UnknownTask;
                    }
                }
            } else if (node.workflowNode) {
                displayType = NodeExecutionDisplayType.Workflow;
                const { launchplanRef, subWorkflowRef } = node.workflowNode;
                const identifier = (launchplanRef
                    ? launchplanRef
                    : subWorkflowRef) as Identifier;
                if (!identifier) {
                    log.warn(`Unexpected workflow node with no ref: ${nodeId}`);
                } else {
                    displayId = identifier.name;
                }
            }

            return {
                ...execution,
                cacheKey,
                displayId,
                displayType,
                taskTemplate
            };
        });
}

interface GetExecutionDurationMSArgs {
    closure: BaseExecutionClosure;
    isTerminal: boolean;
}

interface GetExecutionTimingMSResult {
    duration: number;
    queued: number;
}

/** Computes timing information for an execution based on its create/start times and duration. */
function getExecutionTimingMS({
    closure: { duration, createdAt, startedAt },
    isTerminal
}: GetExecutionDurationMSArgs): GetExecutionTimingMSResult | null {
    if ((isTerminal && duration == null) || createdAt == null) {
        return null;
    }

    const createdAtDate = timestampToDate(createdAt);
    const durationMS =
        isTerminal && duration != null
            ? durationToMilliseconds(duration)
            : Date.now() - createdAtDate.getTime();
    const queuedEndDate = startedAt ? timestampToDate(startedAt) : new Date();
    const queuedMS = queuedEndDate.getTime() - createdAtDate.getTime();

    return { duration: durationMS, queued: queuedMS };
}

/** Returns timing information (duration, queue time, ...) for a WorkflowExecution */
export function getWorkflowExecutionTimingMS(execution: Execution) {
    const { closure } = execution;
    const isTerminal = executionIsTerminal(execution);
    return getExecutionTimingMS({ closure, isTerminal });
}

/** Returns timing information (duration, queue time, ...) for a NodeExecution */
export function getNodeExecutionTimingMS(execution: NodeExecution) {
    const { closure } = execution;
    const isTerminal = nodeExecutionIsTerminal(execution);
    return getExecutionTimingMS({ closure, isTerminal });
}

/** Returns timing information (duration, queue time, ...) for a TaskExecution */
export function getTaskExecutionTimingMS(execution: TaskExecution) {
    const { closure } = execution;
    const isTerminal = taskExecutionIsTerminal(execution);
    return getExecutionTimingMS({ closure, isTerminal });
}
