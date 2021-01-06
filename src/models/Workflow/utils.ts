import { Admin } from 'flyteidl';
import { createPaginationTransformer } from 'models/AdminEntity';
import {
    endpointPrefixes,
    IdentifierScope,
    makeIdentifierPath
} from 'models/Common';
import { Workflow } from './types';

export function makeWorkflowPath(scope: IdentifierScope) {
    return makeIdentifierPath(endpointPrefixes.workflow, scope);
}

/** Transformer to coerce an `Admin.WorkflowList` into a standard shape */
export const workflowListTransformer = createPaginationTransformer<
    Workflow,
    Admin.WorkflowList
>('workflows');
