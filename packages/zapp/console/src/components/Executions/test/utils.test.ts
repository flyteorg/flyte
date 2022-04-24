import { Protobuf } from 'flyteidl';
import {
  NodeExecutionPhase,
  TaskExecutionPhase,
  WorkflowExecutionPhase,
} from 'models/Execution/enums';
import { Execution, NodeExecution, TaskExecution } from 'models/Execution/types';
import { createMockNodeExecutions } from 'models/Execution/__mocks__/mockNodeExecutionsData';
import { createMockTaskExecutionsListResponse } from 'models/Execution/__mocks__/mockTaskExecutionsData';
import { createMockWorkflowExecutionsListResponse } from 'models/Execution/__mocks__/mockWorkflowExecutionsData';
import { long, waitFor } from 'test/utils';
import {
  getNodeExecutionTimingMS,
  getTaskExecutionTimingMS,
  getWorkflowExecutionTimingMS,
} from '../utils';

const getMockWorkflowExecution = () => createMockWorkflowExecutionsListResponse(1).executions[0];
const getMockNodeExecution = () => createMockNodeExecutions(1).executions[0];
const getMockTaskExecution = () => createMockTaskExecutionsListResponse(1).taskExecutions[0];

const sampleDuration: Protobuf.Duration = {
  seconds: long(65),
  nanos: 0,
};
const sampleDurationMS = 65000;

const sampleCreatedAt: Protobuf.ITimestamp = {
  seconds: long(1000),
  nanos: 0,
};
const sampleStartedAt: Protobuf.ITimestamp = {
  seconds: long(1100),
  nanos: 0,
};
const sampleQueuedMS = 100000;

describe('getWorkflowExecutionTimingMS', () => {
  let execution: Execution;
  beforeEach(() => {
    execution = getMockWorkflowExecution();
    execution.closure.phase = WorkflowExecutionPhase.SUCCEEDED;
    execution.closure.createdAt = sampleCreatedAt;
    execution.closure.duration = sampleDuration;
    execution.closure.startedAt = sampleStartedAt;
  });

  it('should return null when terminal with no duration', () => {
    delete execution.closure.duration;
    expect(getWorkflowExecutionTimingMS(execution)).toBeNull();

    execution.closure.phase = WorkflowExecutionPhase.FAILED;
    expect(getWorkflowExecutionTimingMS(execution)).toBeNull();
  });

  it('should return null when no startedAt field is present', () => {
    delete execution.closure.startedAt;
    expect(getWorkflowExecutionTimingMS(execution)).toBeNull();
  });

  it('should return calculated MS values for valid input', () => {
    expect(getWorkflowExecutionTimingMS(execution)).toEqual({
      duration: sampleDurationMS,
      queued: sampleQueuedMS,
    });
  });

  it('should calculate duration using current time if execution is non-terminal', async () => {
    execution.closure.phase = WorkflowExecutionPhase.RUNNING;
    const firstResult = getWorkflowExecutionTimingMS(execution);
    expect(firstResult).not.toBeNull();

    // Force a tick so new Date() will return a different value
    await waitFor(10);
    const secondResult = getWorkflowExecutionTimingMS(execution);
    expect(secondResult).not.toBeNull();
    expect(firstResult!.duration).toBeLessThan(secondResult!.duration);
  });
});

describe('getNodeExecutionTimingMS', () => {
  let execution: NodeExecution;
  beforeEach(() => {
    execution = getMockNodeExecution();
    execution.closure.phase = NodeExecutionPhase.SUCCEEDED;
    execution.closure.createdAt = sampleCreatedAt;
    execution.closure.duration = sampleDuration;
    execution.closure.startedAt = sampleStartedAt;
  });

  it('should return null when terminal with no duration', () => {
    delete execution.closure.duration;
    expect(getNodeExecutionTimingMS(execution)).toBeNull();

    execution.closure.phase = NodeExecutionPhase.FAILED;
    expect(getNodeExecutionTimingMS(execution)).toBeNull();
  });

  it('should return null when no startedAt field is present', () => {
    delete execution.closure.startedAt;
    expect(getNodeExecutionTimingMS(execution)).toBeNull();
  });

  it('should return calculated MS values for valid input', () => {
    expect(getNodeExecutionTimingMS(execution)).toEqual({
      duration: sampleDurationMS,
      queued: sampleQueuedMS,
    });
  });

  it('should calculate duration using current time if execution is non-terminal', async () => {
    execution.closure.phase = NodeExecutionPhase.RUNNING;
    const firstResult = getNodeExecutionTimingMS(execution);
    expect(firstResult).not.toBeNull();

    // Force a tick so new Date() will return a different value
    await waitFor(10);
    const secondResult = getNodeExecutionTimingMS(execution);
    expect(secondResult).not.toBeNull();
    expect(firstResult!.duration).toBeLessThan(secondResult!.duration);
  });
});

describe('getTaskExecutionTimingMS', () => {
  let execution: TaskExecution;
  beforeEach(() => {
    execution = getMockTaskExecution();
    execution.closure.phase = TaskExecutionPhase.SUCCEEDED;
    execution.closure.createdAt = sampleCreatedAt;
    execution.closure.duration = sampleDuration;
    execution.closure.startedAt = sampleStartedAt;
  });

  it('should return null when terminal with no duration', () => {
    delete execution.closure.duration;
    expect(getTaskExecutionTimingMS(execution)).toBeNull();

    execution.closure.phase = TaskExecutionPhase.FAILED;
    expect(getTaskExecutionTimingMS(execution)).toBeNull();
  });

  it('should return null when no startedAt field is present', () => {
    delete execution.closure.startedAt;
    expect(getTaskExecutionTimingMS(execution)).toBeNull();
  });

  it('should return calculated MS values for valid input', () => {
    expect(getTaskExecutionTimingMS(execution)).toEqual({
      duration: sampleDurationMS,
      queued: sampleQueuedMS,
    });
  });

  it('should calculate duration using current time if execution is non-terminal', async () => {
    execution.closure.phase = TaskExecutionPhase.RUNNING;
    const firstResult = getTaskExecutionTimingMS(execution);
    expect(firstResult).not.toBeNull();

    // Force a tick so new Date() will return a different value
    await waitFor(10);
    const secondResult = getTaskExecutionTimingMS(execution);
    expect(secondResult).not.toBeNull();
    expect(firstResult!.duration).toBeLessThan(secondResult!.duration);
  });
});
