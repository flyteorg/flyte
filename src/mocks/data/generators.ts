import { DeepPartial } from 'common/types';
import { dateToTimestamp, millisecondsToDuration } from 'common/utils';
import { Admin, Core } from 'flyteidl';
import { merge } from 'lodash';
import { timeStampOffset } from 'mocks/utils';
import { Identifier, ResourceType } from 'models/Common/types';
import {
    ExecutionState,
    NodeExecutionPhase,
    TaskExecutionPhase,
    WorkflowExecutionPhase
} from 'models/Execution/enums';
import {
    Execution,
    NodeExecution,
    TaskExecution
} from 'models/Execution/types';
import { LaunchPlan } from 'models/Launch/types';
import { endNodeId, startNodeId } from 'models/Node/constants';
import { CompiledTask, Task } from 'models/Task/types';
import { Workflow } from 'models/Workflow/types';
import {
    defaultExecutionDuration,
    emptyInputUri,
    emptyOutputUri,
    entityCreationDate,
    mockStartDate,
    testDomain,
    testProject,
    testVersions
} from './constants';
import { nodeExecutionId, sampleLogs, taskExecutionId } from './utils';

/** Wraps a `CompiledTask` in the necessary fields to create a `Task`. */
export function taskFromCompiledTask(compiledTask: CompiledTask): Task {
    return {
        closure: { createdAt: { ...entityCreationDate }, compiledTask },
        id: compiledTask.template.id
    };
}

/** Generate a new `Task` object based on a set of defaults. The base object
 * returned when `compiledTaskOverrides` is omitted will be a valid `Task`
 */
export function generateTask(
    idOverrides: Partial<Identifier>,
    compiledTaskOverrides?: DeepPartial<CompiledTask>
): Task {
    const id = {
        resourceType: ResourceType.TASK,
        project: testProject,
        domain: testDomain,
        name: '_base',
        version: testVersions.v1,
        ...idOverrides
    };
    const base: CompiledTask = {
        template: {
            custom: {},
            container: {},
            metadata: {},
            type: 'unknown-type',
            id,
            interface: {
                inputs: {
                    variables: {}
                },
                outputs: {
                    variables: {}
                }
            }
        }
    };
    return taskFromCompiledTask(merge(base, compiledTaskOverrides));
}

/** Generate a new `Workflow` object based on a set of defaults. The base object
 * returned when `overrides` is omitted will be a valid `Workflow`.
 */
export function generateWorkflow(
    idOverrides: Partial<Identifier>,
    overrides: DeepPartial<Workflow>
): Workflow {
    const id = {
        resourceType: Core.ResourceType.WORKFLOW,
        project: testProject,
        domain: testDomain,
        name: '_base',
        version: testVersions.v1,
        ...idOverrides
    };
    const base: Workflow = {
        id,
        closure: {
            createdAt: { ...entityCreationDate },
            compiledWorkflow: {
                primary: {
                    connections: {
                        downstream: {},
                        upstream: {}
                    },
                    template: {
                        metadata: {},
                        metadataDefaults: {},
                        id,
                        interface: {},
                        nodes: [{ id: startNodeId }, { id: endNodeId }],
                        outputs: []
                    }
                },
                tasks: []
            }
        }
    };
    return merge(base, overrides);
}

/** Generate an `Execution` for a given `Workflow` and `LaunchPlan`. The base object
 * returned when `overrides` is omitted will be a valid `Execution`.
 */
export function generateExecutionForWorkflow(
    workflow: Workflow,
    launchPlan: LaunchPlan,
    overrides?: DeepPartial<Execution>
): Execution {
    const executionStart = dateToTimestamp(mockStartDate);
    const { id: workflowId } = workflow;
    const id = {
        project: testProject,
        domain: testDomain,
        name: `${workflowId.name}Execution`
    };
    const base: Execution = {
        id,
        spec: {
            launchPlan: { ...launchPlan.id },
            inputs: { literals: {} },
            metadata: {
                mode: Admin.ExecutionMetadata.ExecutionMode.MANUAL,
                principal: 'sdk',
                nesting: 0
            },
            notifications: {
                notifications: []
            }
        },
        closure: {
            workflowId,
            computedInputs: { literals: {} },
            createdAt: executionStart,
            duration: millisecondsToDuration(defaultExecutionDuration),
            phase: WorkflowExecutionPhase.SUCCEEDED,
            startedAt: executionStart,
            stateChangeDetails: { state: ExecutionState.EXECUTION_ACTIVE }
        }
    };
    return merge(base, overrides);
}

/** Generate a `NodeExecution` for a nodeId that will be a child of the given `Execution`.
 * The base object returned when `overrides` is omitted will be a valid `NodeExecution`.
 */
export function generateNodeExecution(
    parentExecution: Execution,
    nodeId: string,
    overrides?: DeepPartial<NodeExecution>
): NodeExecution {
    const base: NodeExecution = {
        id: nodeExecutionId(parentExecution.id, nodeId),
        metadata: { specNodeId: nodeId },
        closure: {
            createdAt: timeStampOffset(parentExecution.closure.createdAt, 0),
            startedAt: timeStampOffset(parentExecution.closure.createdAt, 0),
            outputUri: emptyOutputUri,
            phase: NodeExecutionPhase.SUCCEEDED,
            duration: millisecondsToDuration(defaultExecutionDuration)
        },
        inputUri: emptyInputUri
    };
    return merge(base, overrides);
}

/** Generate a `TaskExecution` for a `Task` that will be a child of the given `NodeExecution`.
 * The base object returned when `overrides` is omitted will be a valid `TaskExecution`.
 */
export function generateTaskExecution(
    nodeExecution: NodeExecution,
    task: Task,
    overrides?: DeepPartial<TaskExecution>
): TaskExecution {
    const base: TaskExecution = {
        id: taskExecutionId(nodeExecution, task, 0),
        inputUri: emptyInputUri,
        isParent: false,
        closure: {
            customInfo: {},
            phase: TaskExecutionPhase.SUCCEEDED,
            duration: millisecondsToDuration(defaultExecutionDuration),
            createdAt: timeStampOffset(nodeExecution.closure.createdAt, 0),
            startedAt: timeStampOffset(nodeExecution.closure.createdAt, 0),
            outputUri: emptyOutputUri,
            logs: sampleLogs()
        }
    };
    return merge(base, overrides);
}
