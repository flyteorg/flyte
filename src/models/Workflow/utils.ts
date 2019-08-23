import { Admin } from 'flyteidl';
import { createPaginationTransformer } from 'models/AdminEntity';

import { Workflow } from './types';

/** Transformer to coerce an `Admin.WorkflowList` into a standard shape */
export const workflowListTransformer = createPaginationTransformer<
    Workflow,
    Admin.WorkflowList
>('workflows');
