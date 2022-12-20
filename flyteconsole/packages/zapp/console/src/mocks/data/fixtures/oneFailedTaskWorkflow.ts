import { NodeExecutionPhase, TaskExecutionPhase } from 'models/Execution/enums';
import { endNodeId, startNodeId } from 'models/Node/constants';
import {
  generateExecutionForWorkflow,
  generateNodeExecution,
  generateTask,
  generateTaskExecution,
  generateWorkflow,
} from '../generators';
import { makeDefaultLaunchPlan, taskNodeIds } from '../utils';

const workflowName = 'OneFailedTaskWorkflow';
const pythonTaskName = `${workflowName}.PythonTask`;
const failedTaskName = `${workflowName}.FailedTask`;
const pythonNodeId = 'pythonNode';
const failedNodeId = 'failedNode';

function generate() {
  const pythonTask = generateTask(
    { name: pythonTaskName },
    {
      template: {
        type: 'python-task',
      },
    },
  );
  const failedTask = generateTask({ name: failedTaskName }, { template: { type: 'python-task' } });

  const workflow = generateWorkflow(
    { name: workflowName },
    {
      closure: {
        compiledWorkflow: {
          primary: {
            connections: {
              downstream: {
                [startNodeId]: {
                  ids: [pythonNodeId, failedNodeId],
                },
                [pythonNodeId]: { ids: [endNodeId] },
                [failedNodeId]: { ids: [endNodeId] },
              },
              upstream: {
                [failedNodeId]: { ids: [startNodeId] },
                [pythonNodeId]: { ids: [startNodeId] },
                [endNodeId]: {
                  ids: [pythonNodeId, failedNodeId],
                },
              },
            },
            template: {
              nodes: [
                {
                  ...taskNodeIds(pythonNodeId, pythonTask),
                  inputs: [],
                },
                {
                  ...taskNodeIds(failedNodeId, pythonTask),
                  inputs: [],
                },
              ],
              outputs: [],
            },
          },
          tasks: [pythonTask.closure.compiledTask, failedTask.closure.compiledTask],
        },
      },
    },
  );

  const launchPlan = makeDefaultLaunchPlan(workflow);
  const execution = generateExecutionForWorkflow(workflow, launchPlan);
  const pythonNodeExecution = generateNodeExecution(execution, pythonNodeId);
  const pythonTaskExecution = generateTaskExecution(pythonNodeExecution, pythonTask);
  const errorMessage = 'This task failed by design.';
  const failedNodeExecution = generateNodeExecution(execution, failedNodeId, {
    closure: {
      phase: NodeExecutionPhase.FAILED,
      error: { message: errorMessage },
    },
  });
  const failedTaskExecution = generateTaskExecution(failedNodeExecution, failedTask, {
    closure: {
      phase: TaskExecutionPhase.FAILED,
      error: { message: errorMessage },
    },
  });
  return {
    launchPlans: { top: launchPlan },
    tasks: { python: pythonTask, failed: failedTask },
    workflows: { top: workflow },
    workflowExecutions: {
      top: {
        data: execution,
        nodeExecutions: {
          pythonNode: {
            data: pythonNodeExecution,
            taskExecutions: {
              firstAttempt: { data: pythonTaskExecution },
            },
          },
          failedNode: {
            data: failedNodeExecution,
            taskExecutions: {
              firstAttempt: { data: failedTaskExecution },
            },
          },
        },
      },
    },
  };
}

/** This workflow has two python nodes with no inputs/outputs. One of them
 * will always fail.
 */
export const oneFailedTaskWorkflow = { generate };
