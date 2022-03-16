import { Admin } from 'flyteidl';
import { createPaginationTransformer } from 'models/AdminEntity/utils';
import { endpointPrefixes } from 'models/Common/constants';
import { IdentifierScope } from 'models/Common/types';
import { makeIdentifierPath } from 'models/Common/utils';
import { Task } from './types';

/** Generate the correct path for retrieving a task or list of tasks based on the
 * given scope.
 */
export function makeTaskPath(scope: IdentifierScope) {
  return makeIdentifierPath(endpointPrefixes.task, scope);
}

/** Transformer to coerce an `Admin.TaskList` into a standard shape */
export const taskListTransformer = createPaginationTransformer<Task, Admin.TaskList>('tasks');
