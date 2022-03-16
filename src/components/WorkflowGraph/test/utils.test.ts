import {
  mockCompiledWorkflow,
  mockCompiledEndNode,
  mockCompiledStartNode,
  mockCompiledTaskNode,
} from 'models/__mocks__/graphWorkflowData';
import { dTypes } from 'models/Graph/types';
import {
  DISPLAY_NAME_START,
  DISPLAY_NAME_END,
  checkIfObjectsAreSame,
  getDisplayName,
  getNodeTypeFromCompiledNode,
  isStartNode,
  isEndNode,
} from '../utils';

describe('getDisplayName', () => {
  it('should return correct name', () => {
    expect(getDisplayName(mockCompiledStartNode)).toBe(DISPLAY_NAME_START);
    expect(getDisplayName(mockCompiledEndNode)).toBe(DISPLAY_NAME_END);
    expect(getDisplayName(mockCompiledTaskNode)).toBe('DEADBEEF');
    expect(getDisplayName(mockCompiledWorkflow)).toBe('myWorkflowName');
  });
});

describe('checkIfObjectsAreSame', () => {
  const a = {
    red: 'red',
    blue: 'blue',
    green: 'green',
  };
  const b = {
    red: 'red',
    blue: 'blue',
    green: 'green',
  };
  const fail_a = {
    red: 'red',
    blue: 'not blue',
    green: 'green',
  };
  const fail_b = {
    red: 'red',
    green: 'green',
    orange: 'orange',
  };
  it('should return true when a-keys match b-values', () => {
    expect(checkIfObjectsAreSame(a, b)).toEqual(true);
  });
  it("should return false when a-keys don't match b-values", () => {
    expect(checkIfObjectsAreSame(fail_a, b)).toEqual(false);
    expect(checkIfObjectsAreSame(a, fail_b)).toEqual(false);
  });
});

describe('getNodeTypeFromCompiledNode', () => {
  const branchNode = {
    branchNode: {},
  };
  const workflowNode = {
    workflowNode: {},
  };
  const mockBranchNode = { ...mockCompiledTaskNode, ...branchNode };
  const mockWorkflowNode = { ...mockCompiledTaskNode, ...workflowNode };

  it('should return dTypes.start when is start-node', () => {
    expect(getNodeTypeFromCompiledNode(mockCompiledStartNode)).toBe(dTypes.start);
  });
  it('should return dTypes.end when is end-node', () => {
    expect(getNodeTypeFromCompiledNode(mockCompiledEndNode)).toBe(dTypes.end);
  });
  it('should return *dTypes.subworkflow (branch is typed as subworkflow for graph) when is node has branchNodes', () => {
    expect(getNodeTypeFromCompiledNode(mockBranchNode)).toBe(dTypes.subworkflow);
  });
  it('should return dTypes.subworkflow when is node has workflowNode', () => {
    expect(getNodeTypeFromCompiledNode(mockWorkflowNode)).toBe(dTypes.subworkflow);
  });
  it('should return dTypes.task when is node is taskNode', () => {
    expect(getNodeTypeFromCompiledNode(mockCompiledTaskNode)).toBe(dTypes.task);
  });
});

describe('isStartNode', () => {
  it('should return true when start-node', () => {
    expect(isStartNode(mockCompiledStartNode)).toBe(true);
  });
  it('should return false when not start-node', () => {
    expect(isStartNode(mockCompiledTaskNode)).toBe(false);
  });
});

describe('isEndNode', () => {
  it('should return true when start-node', () => {
    expect(isEndNode(mockCompiledEndNode)).toBe(true);
  });
  it('should return false when not start-node', () => {
    expect(isEndNode(mockCompiledTaskNode)).toBe(false);
  });
});

describe('getSubWorkflowFromId', () => {
  it('should return subworkflow from id', () => {
    expect(isStartNode(mockCompiledStartNode)).toBe(true);
  });
  it('should return false when not start-node', () => {
    expect(isStartNode(mockCompiledTaskNode)).toBe(false);
  });
});
