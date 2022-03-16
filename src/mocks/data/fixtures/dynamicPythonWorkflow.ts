import { TaskExecutionPhase } from 'models/Execution/enums';
import { endNodeId, startNodeId } from 'models/Node/constants';
import { nodeIds } from '../constants';
import {
  generateExecutionForWorkflow,
  generateNodeExecution,
  generateTask,
  generateTaskExecution,
  generateWorkflow,
} from '../generators';
import { makeDefaultLaunchPlan, taskNodeIds } from '../utils';

const workflowName = 'DynamicPythonTaskWorkflow';
const pythonTaskName = `${workflowName}.PythonTask`;
const pythonNodeId = 'pythonNode';
const dynamicTaskName = `${workflowName}.DynamicTask`;
const dynamicNodeId = 'dynamicNode';

function getSharedEntities() {
  const pythonTask = generateTask(
    { name: pythonTaskName },
    {
      template: {
        type: 'python-task',
      },
    },
  );

  const dynamicTask = generateTask(
    { name: dynamicTaskName },
    {
      template: {
        type: 'dynamic-task',
      },
    },
  );

  const workflow = generateWorkflow(
    { name: workflowName },
    {
      closure: {
        compiledWorkflow: {
          primary: {
            connections: {
              downstream: {
                [startNodeId]: {
                  ids: [dynamicNodeId],
                },
                [nodeIds.dynamicTask]: { ids: [endNodeId] },
              },
              upstream: {
                [nodeIds.dynamicTask]: { ids: [startNodeId] },
                [endNodeId]: {
                  ids: [nodeIds.dynamicTask],
                },
              },
            },
            template: {
              nodes: [
                {
                  ...taskNodeIds(dynamicNodeId, dynamicTask),
                  inputs: [],
                },
              ],
              outputs: [],
            },
          },
          tasks: [dynamicTask.closure.compiledTask, pythonTask.closure.compiledTask],
        },
      },
    },
  );

  const launchPlan = makeDefaultLaunchPlan(workflow);
  const execution = generateExecutionForWorkflow(workflow, launchPlan);
  return { pythonTask, dynamicTask, workflow, launchPlan, execution };
}

function generateWithDynamicTaskChild() {
  const { dynamicTask, pythonTask, workflow, launchPlan, execution } = getSharedEntities();
  const pythonNodeExecution = generateNodeExecution(execution, pythonNodeId);
  const pythonTaskExecutions = [
    generateTaskExecution(pythonNodeExecution, pythonTask, {
      closure: {
        phase: TaskExecutionPhase.FAILED,
        error: { message: 'Something went wrong.' },
      },
    }),
    generateTaskExecution(pythonNodeExecution, pythonTask, {
      id: { retryAttempt: 1 },
    }),
  ];
  const dynamicNodeExecution = generateNodeExecution(execution, dynamicNodeId);
  const dynamicTaskExecution = generateTaskExecution(dynamicNodeExecution, dynamicTask, {
    isParent: true,
  });
  return {
    launchPlans: { top: launchPlan },
    tasks: { dynamic: dynamicTask, python: pythonTask },
    workflows: { top: workflow },
    workflowExecutions: {
      top: {
        data: execution,
        nodeExecutions: {
          dynamicNode: {
            data: dynamicNodeExecution,
            taskExecutions: {
              firstAttempt: {
                data: dynamicTaskExecution,
                nodeExecutions: {
                  pythonNode: {
                    data: pythonNodeExecution,
                    taskExecutions: {
                      firstAttempt: {
                        data: pythonTaskExecutions[0],
                      },
                      secondAttempt: {
                        data: pythonTaskExecutions[1],
                      },
                    },
                  },
                },
              },
            },
          },
        },
      },
    },
  };
}

function generateWithNodeExecutionChild() {
  const { dynamicTask, pythonTask, workflow, launchPlan, execution } = getSharedEntities();
  const dynamicNodeExecution = generateNodeExecution(execution, dynamicNodeId, {
    metadata: { isParentNode: true },
  });
  const dynamicTaskExecution = generateTaskExecution(dynamicNodeExecution, dynamicTask);
  const pythonNodeExecutions = [
    generateNodeExecution(execution, `${pythonNodeId}-1`, {
      metadata: { retryGroup: '0' },
    }),
    generateNodeExecution(execution, `${pythonNodeId}-2`, {
      metadata: { retryGroup: '1' },
    }),
  ];
  const pythonTaskExecutions = [
    generateTaskExecution(pythonNodeExecutions[0], pythonTask, {
      closure: {
        phase: TaskExecutionPhase.FAILED,
        error: { message: 'Something went wrong.' },
      },
    }),
    generateTaskExecution(pythonNodeExecutions[1], pythonTask),
  ];
  return {
    launchPlans: { top: launchPlan },
    tasks: {
      dynamic: dynamicTask,
      python: pythonTask,
    },
    workflows: { top: workflow },
    workflowExecutions: {
      top: {
        data: execution,
        nodeExecutions: {
          dynamicNode: {
            data: dynamicNodeExecution,
            nodeExecutions: {
              firstChild: {
                data: pythonNodeExecutions[0],
                taskExecutions: {
                  firstAttempt: {
                    data: pythonTaskExecutions[0],
                  },
                },
              },
              secondChild: {
                data: pythonNodeExecutions[1],
                taskExecutions: {
                  firstAttempt: {
                    data: pythonTaskExecutions[1],
                  },
                },
              },
            },
            taskExecutions: {
              firstAttempt: {
                data: dynamicTaskExecution,
              },
            },
          },
        },
      },
    },
  };
}

/** This workflow has a single dynamic task node which will yield an additional
 * python task node at runtime using the `TaskExecution.isParent` field.
 * The nested python task has two attempts.
 */
export const dynamicPythonTaskWorkflow = {
  generate: generateWithDynamicTaskChild,
};

/** This workflow has a single dynamic task node which will yield an additional
 * python task node at runtime using the `NodeExecution.metadata.isParentNode` field.
 * The nested python task has two attempts.
 */
export const dynamicPythonNodeExecutionWorkflow = {
  generate: generateWithNodeExecutionChild,
};
