import { Admin } from 'flyteidl';
import { createPaginationTransformer } from 'models/AdminEntity/utils';
import { endpointPrefixes } from 'models/Common/constants';
import { NameIdentifierScope } from 'models/Common/types';
import { makeIdentifierPath } from 'models/Common/utils';
import {
  Execution,
  NodeExecution,
  NodeExecutionIdentifier,
  TaskExecution,
  TaskExecutionIdentifier,
  WorkflowExecutionIdentifier,
} from './types';

/** Generates the API endpoint for a given `WorkflowExecutionIdentifier` */
export const makeExecutionPath = ({ project, domain, name }: WorkflowExecutionIdentifier) =>
  [endpointPrefixes.execution, project, domain, name].join('/');

/** Generates the API endpoint for a given `NodeExecutionIdentifier` */
export const makeNodeExecutionPath = ({
  executionId: { project, domain, name },
  nodeId,
}: NodeExecutionIdentifier) =>
  [endpointPrefixes.nodeExecution, project, domain, name, nodeId].join('/');

export const makeNodeExecutionListPath = (scope: NameIdentifierScope) =>
  makeIdentifierPath(endpointPrefixes.nodeExecution, scope);

/** Generates the API endpoint for a list of `TaskExecution` children */
export const makeTaskExecutionChildrenPath = ({
  nodeExecutionId: {
    executionId: { domain, name, project },
    nodeId,
  },
  retryAttempt,
  taskId: { domain: taskDomain, name: taskName, project: taskProject, version: taskVersion },
}: TaskExecutionIdentifier) =>
  [
    endpointPrefixes.taskExecutionChildren,
    project,
    domain,
    name,
    nodeId,
    taskProject,
    taskDomain,
    taskName,
    taskVersion,
    retryAttempt,
  ].join('/');

/** Generates the API endpoint for the TaskExecution list belonging to a given
 * `NodeExecutionIdentifier`
 */
export const makeTaskExecutionListPath = ({
  executionId: { project, domain, name },
  nodeId,
}: NodeExecutionIdentifier) =>
  [endpointPrefixes.taskExecution, project, domain, name, nodeId].join('/');

/** Generates the API endpoint for a given `TaskExecutionIdentifier` */
export const makeTaskExecutionPath = ({
  nodeExecutionId: {
    executionId: { project, domain, name },
    nodeId,
  },
  taskId,
  retryAttempt,
}: TaskExecutionIdentifier) =>
  [
    endpointPrefixes.taskExecution,
    project,
    domain,
    name,
    nodeId,
    taskId.project,
    taskId.domain,
    taskId.name,
    taskId.version,
    retryAttempt,
  ].join('/');

/** Transformer to coerce an `Admin.ExecutionList` into a standard shape */
export const executionListTransformer = createPaginationTransformer<Execution, Admin.ExecutionList>(
  'executions',
);

/** Transformer to coerce an `Admin.NodeExecutionList` into a standard shape */
export const nodeExecutionListTransformer = createPaginationTransformer<
  NodeExecution,
  Admin.NodeExecutionList
>('nodeExecutions');

/** Transformer to coerce an `Admin.TaskExecutionList` into a standard shape */
export const taskExecutionListTransformer = createPaginationTransformer<
  TaskExecution,
  Admin.TaskExecutionList
>('taskExecutions');
