import { dateToTimestamp, millisecondsToDuration } from 'common/utils';
import { Admin } from 'flyteidl';
import { cloneDeep, random, sample } from 'lodash';
import { CompiledNode } from 'models/Node/types';
import { mockNodes } from 'models/Node/__mocks__/mockNodeData';
import { NodeExecutionPhase } from '../enums';
import { NodeExecution } from '../types';
import { mockWorkflowExecutionId } from './constants';
import { sampleError } from './sampleExecutionError';

export const mockNodeExecutionResponse: Admin.INodeExecution = {
  id: {
    executionId: mockWorkflowExecutionId,
    nodeId: 'DefaultNodeId',
  },
  inputUri: 's3://path/to/my/inputs.pb',
  closure: {
    phase: NodeExecutionPhase.SUCCEEDED,
    startedAt: dateToTimestamp(new Date(Date.now() - 1000 * 60 * 10)),
    duration: millisecondsToDuration(1000 * 60 * 60 * 1.251),
    outputUri: 's3://path/to/my/outputs.pb',
  },
};

export const mockExecution = mockNodeExecutionResponse as NodeExecution;

/** Generates a set of mock NodeExecutions based on the `mockNodes` list. Since
 * the nodeIds will be generated as well, also returns a list of the generated/
 * updated nodes to be appended to a compiled workflow closure.
 */
export const createMockNodeExecutions = (length: number) => {
  const indexes: Record<string, number> = {};
  const generateNode = () => {
    const node = sample(mockNodes)!;
    const index = indexes[node.id] || 1;
    indexes[node.id] = index + 1;
    return { ...node, id: `${node.id}${index}` };
  };

  const nodes: CompiledNode[] = [];

  const executions = Array.from({ length }, (_, idx) => {
    const execution = cloneDeep(mockExecution);

    const node = generateNode();
    nodes.push(node);
    execution.id.nodeId = node.id;
    const startedAtDate = new Date(Date.now() - 1000 * 60 * (idx + 1));

    const startedAt = dateToTimestamp(startedAtDate);
    const createdAtDate = new Date(startedAtDate.getTime() - 1000 * 30 * 5 * (idx + 1));
    const createdAt = dateToTimestamp(createdAtDate);
    const phase = random(Object.keys(NodeExecutionPhase).length - 1);

    // random duration between 0-90 minutes
    const duration = millisecondsToDuration(Math.random() * 1000 * 60 * 90);

    const error =
      phase === NodeExecutionPhase.FAILED
        ? {
            code: 'user_error',
            errorUri: '',
            message: sampleError,
          }
        : undefined;

    Object.assign(execution.closure, {
      createdAt,
      error,
      duration,
      phase,
      startedAt,
    });

    return execution;
  });

  return { executions, nodes };
};
