import { dateToTimestamp, millisecondsToDuration } from 'common/utils';
import { Admin, Core } from 'flyteidl';
import { cloneDeep } from 'lodash';
import { NodeExecutionPhase, TaskExecutionPhase } from '../enums';
import { TaskExecution } from '../types';
import { sampleError } from './sampleExecutionError';

const sampleLogs: Core.ITaskLog[] = [
    { name: 'Kubernetes Logs', uri: 'http://localhost/k8stasklog' },
    { name: 'User Logs', uri: 'http://localhost/containerlog' },
    { name: 'AWS Batch Logs', uri: 'http://localhost/awsbatchlog' },
    { name: 'Other Custom Logs', uri: 'http://localhost/customlog' }
];

export const mockTaskExecutionResponse: Admin.ITaskExecution = {
    id: {
        nodeExecutionId: {
            executionId: {
                project: 'flytekit',
                domain: 'development',
                name: '4a580545ce6344fc9950'
            },
            nodeId: 'DefaultNodeId'
        },
        retryAttempt: 0,
        taskId: {
            project: 'flytekit',
            domain: 'development',
            name: 'MyTask',
            version: 'abcdef'
        }
    },
    inputUri: 's3://path/to/my/inputs.pb',
    closure: {
        phase: TaskExecutionPhase.SUCCEEDED,
        startedAt: dateToTimestamp(new Date(Date.now() - 1000 * 60 * 10)),
        createdAt: dateToTimestamp(new Date(Date.now() - 1000 * 60 * 10)),
        duration: millisecondsToDuration(1000 * 60 * 60 * 1.251),
        outputUri: 's3://path/to/my/outputs.pb',
        logs: [...sampleLogs]
    }
};

export const mockExecution = mockTaskExecutionResponse as TaskExecution;

export const createMockTaskExecutionsListResponse = (length: number) => {
    return {
        taskExecutions: Array.from({ length }, (_, idx) => {
            const execution = cloneDeep(mockExecution);
            execution.id.retryAttempt = idx;
            const startedAt = dateToTimestamp(
                new Date(Date.now() - 1000 * 60 * (idx + 1))
            );
            const phase =
                idx < length - 1
                    ? TaskExecutionPhase.FAILED
                    : TaskExecutionPhase.SUCCEEDED;

            // random duration between 0-90 minutes
            const duration = millisecondsToDuration(
                Math.random() * 1000 * 60 * 90
            );

            const error =
                phase === TaskExecutionPhase.FAILED
                    ? {
                          code: 'user_error',
                          errorUri: '',
                          message: sampleError
                      }
                    : undefined;

            Object.assign(execution.closure, {
                error,
                duration,
                phase,
                startedAt,
                createdAt: startedAt
            });

            return execution;
        })
    };
};
