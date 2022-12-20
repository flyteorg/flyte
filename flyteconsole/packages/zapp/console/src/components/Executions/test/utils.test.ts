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
import { mockNodes, mockNodesWithGateNode } from 'models/Node/__mocks__/mockNodeData';
import { long, waitFor } from 'test/utils';
import {
  getNodeExecutionTimingMS,
  getNodeFrontendPhase,
  getTaskExecutionTimingMS,
  getWorkflowExecutionTimingMS,
  isNodeGateNode,
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

describe('isNodeGateNode', () => {
  const executionId = { project: 'project', domain: 'domain', name: 'name' };

  it('should return true if nodeId is in the list and has a gateNode field', () => {
    expect(isNodeGateNode(mockNodesWithGateNode, { nodeId: 'GateNode', executionId })).toBeTruthy();
  });

  it('should return false if nodeId is in the list, but a gateNode field is missing', () => {
    expect(isNodeGateNode(mockNodes, { nodeId: 'BasicNode', executionId })).toBeFalsy();
  });

  it('should return false if nodeId is not in the list, but has a gateNode field', () => {
    expect(isNodeGateNode(mockNodes, { nodeId: 'GateNode', executionId })).toBeFalsy();
  });

  it('should return false if nodeId is a gateNode, but the list is empty', () => {
    expect(isNodeGateNode([], { nodeId: 'GateNode', executionId })).toBeFalsy();
  });

  it('should return false if nodeId is not a gateNode and the list is empty', () => {
    expect(isNodeGateNode([], { nodeId: 'BasicNode', executionId })).toBeFalsy();
  });
});

describe('getNodeFrontendPhase', () => {
  it('should return PAUSED if node is a gateNode in the RUNNING phase', () => {
    expect(getNodeFrontendPhase(NodeExecutionPhase.RUNNING, true)).toEqual(
      NodeExecutionPhase.PAUSED,
    );
  });

  it('should return phase if node is a gateNode not in the RUNNING phase', () => {
    expect(getNodeFrontendPhase(NodeExecutionPhase.FAILED, true)).toEqual(
      NodeExecutionPhase.FAILED,
    );
  });

  it('should return RUNNING if node is not a gateNode in the RUNNING phase', () => {
    expect(getNodeFrontendPhase(NodeExecutionPhase.RUNNING, false)).toEqual(
      NodeExecutionPhase.RUNNING,
    );
  });

  it('should return phase if node is not a gateNode not in the RUNNING phase', () => {
    expect(getNodeFrontendPhase(NodeExecutionPhase.SUCCEEDED, false)).toEqual(
      NodeExecutionPhase.SUCCEEDED,
    );
  });
});
