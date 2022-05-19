import { Admin } from 'flyteidl';
import { createPaginationTransformer } from 'models/AdminEntity/utils';
import { endpointPrefixes } from 'models/Common/constants';
import { IdentifierScope } from 'models/Common/types';
import { makeIdentifierPath } from 'models/Common/utils';
import { TaskType } from './constants';
import { Task } from './types';

/** Generate the correct path for retrieving a task or list of tasks based on the
 * given scope.
 */
export function makeTaskPath(scope: IdentifierScope) {
  return makeIdentifierPath(endpointPrefixes.task, scope);
}

/** Transformer to coerce an `Admin.TaskList` into a standard shape */
export const taskListTransformer = createPaginationTransformer<Task, Admin.TaskList>('tasks');

/** Returns true if tasks schema is treated as a map task */
export function isMapTaskType(taskType?: string): boolean {
  return (
    taskType === TaskType.ARRAY ||
    taskType === TaskType.ARRAY_AWS ||
    taskType === TaskType.ARRAY_K8S
  );
}

/** Returns true if tasks schema is treated as a map task AND eventVersion >= 1 AND externalResources array length > 0 */
export function isMapTaskV1(
  eventVersion: number,
  externalResourcesLength: number,
  taskType?: string,
): boolean {
  return isMapTaskType(taskType) && eventVersion >= 1 && externalResourcesLength > 0;
}
