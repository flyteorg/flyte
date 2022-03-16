import { Admin } from 'flyteidl';
import { createPaginationTransformer } from 'models/AdminEntity/utils';
import { endpointPrefixes } from 'models/Common/constants';
import { IdentifierScope } from 'models/Common/types';
import { makeIdentifierPath } from 'models/Common/utils';
import { Workflow } from './types';

export function makeWorkflowPath(scope: IdentifierScope) {
  return makeIdentifierPath(endpointPrefixes.workflow, scope);
}

/** Transformer to coerce an `Admin.WorkflowList` into a standard shape */
export const workflowListTransformer = createPaginationTransformer<Workflow, Admin.WorkflowList>(
  'workflows',
);
