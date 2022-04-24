import { cloneDeep } from 'lodash';
import { endNodeId, startNodeId } from 'models/Node/constants';
import {
  generateExecutionForWorkflow,
  generateNodeExecution,
  generateTask,
  generateTaskExecution,
  generateWorkflow,
} from '../generators';
import { makeDefaultLaunchPlan, taskNodeIds } from '../utils';

const topWorkflowName = 'LaunchExternalSubWorkflow';
const subWorkflowName = `${topWorkflowName}.SubWorkflow`;
const nodeIds = {
  subWorkflow: 'subWorkflowNode',
  python: 'pythonNode',
};
const subWorkflowTaskName = `${topWorkflowName}.LaunchSubworkflowTask`;

function generate() {
  const launchSubWorkflowTask = generateTask(
    { name: subWorkflowTaskName },
    { template: { type: 'dynamic-task' } },
  );
  const topWorkflow = generateWorkflow(
    { name: topWorkflowName },
    {
      closure: {
        compiledWorkflow: {
          primary: {
            connections: {
              downstream: {
                [startNodeId]: {
                  ids: [nodeIds.subWorkflow],
                },
                [nodeIds.subWorkflow]: { ids: [endNodeId] },
              },
              upstream: {
                [nodeIds.subWorkflow]: {
                  ids: [startNodeId],
                },
                [endNodeId]: { ids: [nodeIds.subWorkflow] },
              },
            },
            template: {
              nodes: [
                {
                  ...taskNodeIds(nodeIds.subWorkflow, launchSubWorkflowTask),
                  inputs: [],
                },
              ],
            },
          },
          tasks: [cloneDeep(launchSubWorkflowTask.closure.compiledTask)],
        },
      },
    },
  );

  const topWorkflowLaunchPlan = makeDefaultLaunchPlan(topWorkflow);
  const topExecution = generateExecutionForWorkflow(topWorkflow, topWorkflowLaunchPlan);

  const topNodeExecution = generateNodeExecution(topExecution, 'dynamicWorkflowChild', {
    metadata: {
      specNodeId: nodeIds.subWorkflow,
    },
  });

  const topTaskExecution = generateTaskExecution(topNodeExecution, launchSubWorkflowTask, {
    isParent: true,
  });

  const subWorkflowPythonTask = generateTask(
    { name: `${topWorkflowName}.PythonTask` },
    { template: { type: 'python-task' } },
  );

  const subWorkflow = generateWorkflow(
    { name: subWorkflowName },
    {
      closure: {
        compiledWorkflow: {
          primary: {
            connections: {
              downstream: {
                [startNodeId]: { ids: [nodeIds.python] },
                [nodeIds.python]: { ids: [endNodeId] },
              },
              upstream: {
                [nodeIds.python]: { ids: [startNodeId] },
                [endNodeId]: { ids: [nodeIds.python] },
              },
            },
            template: {
              nodes: [
                {
                  ...taskNodeIds(nodeIds.python, subWorkflowPythonTask),
                  inputs: [],
                },
              ],
            },
          },
          tasks: [subWorkflowPythonTask.closure.compiledTask],
        },
      },
    },
  );
  const subWorkflowLaunchPlan = makeDefaultLaunchPlan(subWorkflow);

  const subWorkflowExecution = generateExecutionForWorkflow(subWorkflow, subWorkflowLaunchPlan, {
    spec: { metadata: { nesting: 1 } },
  });
  const pythonNodeExecution = generateNodeExecution(subWorkflowExecution, 'pythonNode', {
    metadata: {
      specNodeId: nodeIds.python,
    },
  });
  const pythonTaskExecution = generateTaskExecution(pythonNodeExecution, subWorkflowPythonTask);

  const launchSubWorkflowNodeExecution = generateNodeExecution(topExecution, 'launchSubWorkflow', {
    closure: {
      workflowNodeMetadata: {
        executionId: subWorkflowExecution.id,
      },
    },
  });

  return {
    launchPlans: {
      top: topWorkflowLaunchPlan,
      subWorkflow: subWorkflowLaunchPlan,
    },
    tasks: {
      generateSubWorkflow: launchSubWorkflowTask,
      pythonTask: subWorkflowPythonTask,
    },
    workflows: {
      top: topWorkflow,
      sub: subWorkflow,
    },
    workflowExecutions: {
      top: {
        data: topExecution,
        nodeExecutions: {
          dynamicWorkflowGenerator: {
            data: topNodeExecution,
            taskExecutions: {
              firstAttempt: {
                data: topTaskExecution,
                nodeExecutions: {
                  launchSubWorkflow: {
                    data: launchSubWorkflowNodeExecution,
                  },
                },
              },
            },
          },
        },
      },
      sub: {
        data: subWorkflowExecution,
        nodeExecutions: {
          pythonNode: {
            data: pythonNodeExecution,
            taskExecutions: {
              firstAttempt: { data: pythonTaskExecution },
            },
          },
        },
      },
    },
  };
}

/**
 * A workflow with one dynamic task node which will yield an additional node at
 * runtime. The child node will launch a separate workflow execution referencing
 * our basic python workflow.
 */
export const dynamicExternalSubWorkflow = { generate };
