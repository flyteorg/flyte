import { cloneDeep } from 'lodash';
import { SimpleType } from 'models/Common/types';
import { endNodeId, startNodeId } from 'models/Node/constants';
import { variableNames } from '../constants';
import {
  generateExecutionForWorkflow,
  generateNodeExecution,
  generateTask,
  generateTaskExecution,
  generateWorkflow,
} from '../generators';
import { bindingFromNode, makeDefaultLaunchPlan, taskNodeIds } from '../utils';

const workflowName = 'BasicPythonWorkflow';
const taskName = `${workflowName}.PythonTask`;
const pythonNodeId = 'pythonNode';

function generate() {
  const pythonTask = generateTask(
    { name: taskName },
    {
      template: {
        type: 'python-task',
        interface: {
          inputs: {
            variables: {
              [variableNames.basicString]: {
                description: 'A string which will be echoed to output',
                type: { simple: SimpleType.STRING },
              },
            },
          },
          outputs: {
            variables: {
              [variableNames.basicString]: {
                description: 'A copy of the string provided to this task',
                type: { simple: SimpleType.STRING },
              },
            },
          },
        },
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
                [startNodeId]: { ids: [pythonNodeId] },
                [pythonNodeId]: { ids: [endNodeId] },
              },
              upstream: {
                [pythonNodeId]: { ids: [startNodeId] },
                [endNodeId]: { ids: [pythonNodeId] },
              },
            },
            template: {
              // This workflow has just one task, so the i/o will be those from
              // the task
              interface: cloneDeep(pythonTask.closure.compiledTask.template.interface),
              nodes: [
                {
                  ...taskNodeIds(pythonNodeId, pythonTask),
                  inputs: [
                    bindingFromNode(
                      variableNames.basicString,
                      startNodeId,
                      variableNames.basicString,
                    ),
                  ],
                },
              ],
              outputs: [
                bindingFromNode(variableNames.basicString, pythonNodeId, variableNames.basicString),
              ],
            },
          },
          tasks: [pythonTask.closure.compiledTask],
        },
      },
    },
  );

  const launchPlan = makeDefaultLaunchPlan(workflow);
  const execution = generateExecutionForWorkflow(workflow, launchPlan);
  const pythonNodeExecution = generateNodeExecution(execution, pythonNodeId);
  const pythonTaskExecution = generateTaskExecution(pythonNodeExecution, pythonTask);
  return {
    launchPlans: { top: launchPlan },
    tasks: { python: pythonTask },
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
        },
      },
    },
  };
}

/** This workflow has a single python node which takes a string as input
 * and copies it to the output.
 */
export const basicPythonWorkflow = { generate };
