import { Admin } from 'flyteidl';
import { defaultPaginationConfig, RequestConfig } from 'models/AdminEntity';
import { getAdminEntity } from 'models/AdminEntity/AdminEntity';
import { Identifier, IdentifierScope } from 'models/Common';
import { Task } from './types';
import { makeTaskPath, taskListTransformer } from './utils';

/** Fetches a list of `Task` records matching the provided `scope` */
export const listTasks = (scope: IdentifierScope, config?: RequestConfig) =>
    getAdminEntity(
        {
            path: makeTaskPath(scope),
            messageType: Admin.TaskList,
            transform: taskListTransformer
        },
        { ...defaultPaginationConfig, ...config }
    );

/** Fetches an individual `Task` record */
export const getTask = (id: Identifier, config?: RequestConfig) =>
    getAdminEntity<Admin.Task, Task>(
        {
            path: makeTaskPath(id),
            messageType: Admin.Task
        },
        config
    );
