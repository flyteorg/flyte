import { Admin } from 'flyteidl';
import { createPaginationTransformer } from 'models/AdminEntity';

import { Task } from './types';

/** Transformer to coerce an `Admin.TaskList` into a standard shape */
export const taskListTransformer = createPaginationTransformer<
    Task,
    Admin.TaskList
>('tasks');
