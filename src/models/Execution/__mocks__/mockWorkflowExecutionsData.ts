import { dateToTimestamp, millisecondsToDuration } from 'common/utils';
import { Admin, Core } from 'flyteidl';
import { cloneDeep, random } from 'lodash';
import * as Long from 'long';
import { WorkflowExecutionPhase } from '../enums';
import { Execution } from '../types';
import { mockWorkflowExecutionId } from './constants';
import { sampleError } from './sampleExecutionError';

export const mockWorkflowExecutionResponse: Admin.IExecution = {
    id: mockWorkflowExecutionId,
    spec: {
        launchPlan: {
            resourceType: Core.ResourceType.LAUNCH_PLAN,
            project: 'flytekit',
            domain: 'development',
            name: 'MyExampleWorkflow',
            version: 'ABC123'
        },
        inputs: {},
        metadata: {
            principal: 'sdk'
        }
    },
    closure: {
        phase: WorkflowExecutionPhase.SUCCEEDED,
        startedAt: dateToTimestamp(new Date(Date.now() - 1000 * 60 * 10)),
        duration: millisecondsToDuration(1000 * 60 * 60 * 1.251),
        workflowId: {
            project: 'flytekit',
            domain: 'development',
            name: 'MyExampleWorkflow',
            version: 'ABC123'
        },
        computedInputs: {
            literals: {
                input_1: {
                    scalar: {
                        primitive: {
                            integer: Long.fromString('3')
                        }
                    }
                },
                input_2: {
                    scalar: {
                        primitive: {
                            integer: Long.fromString('5')
                        }
                    }
                }
            }
        }
    }
};

export const mockExecution = mockWorkflowExecutionResponse as Execution;

export const createMockWorkflowExecutionsListResponse = (length: number) => ({
    executions: Array.from({ length }, (_, idx) => {
        const execution = cloneDeep(mockExecution);
        const startedAtDate = new Date(Date.now() - 1000 * 60 * (idx + 1));
        const createdAtDate = new Date(
            startedAtDate.getTime() - 1000 * 30 * 5 * (idx + 1)
        );
        const phase = random(Object.keys(WorkflowExecutionPhase).length - 1);
        // random duration between 0-90 minutes
        const durationMS = Math.random() * 1000 * 60 * 90;
        execution.closure.createdAt = dateToTimestamp(createdAtDate);
        execution.closure.startedAt = dateToTimestamp(startedAtDate);
        execution.closure.duration = millisecondsToDuration(durationMS);
        execution.closure.phase = phase;
        if (phase === WorkflowExecutionPhase.FAILED) {
            execution.closure.error = {
                code: 'user_error',
                errorUri: '',
                kind: Core.ExecutionError.ErrorKind.USER,
                message: sampleError
            };
        }
        return execution;
    })
});
