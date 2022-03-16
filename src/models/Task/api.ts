import { Admin } from 'flyteidl';
import { getAdminEntity } from 'models/AdminEntity/AdminEntity';
import { defaultPaginationConfig } from 'models/AdminEntity/constants';
import { RequestConfig } from 'models/AdminEntity/types';
import { Identifier, IdentifierScope } from 'models/Common/types';
import { Task } from './types';
import { makeTaskPath, taskListTransformer } from './utils';

/** Fetches a list of `Task` records matching the provided `scope` */
export const listTasks = (scope: IdentifierScope, config?: RequestConfig) =>
  getAdminEntity(
    {
      path: makeTaskPath(scope),
      messageType: Admin.TaskList,
      transform: taskListTransformer,
    },
    { ...defaultPaginationConfig, ...config },
  );

/** Fetches an individual `Task` record */
export const getTask = (id: Identifier, config?: RequestConfig) =>
  getAdminEntity<Admin.Task, Task>(
    {
      path: makeTaskPath(id),
      messageType: Admin.Task,
    },
    config,
  );
