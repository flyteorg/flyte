import { Admin } from 'flyteidl';
import { getAdminEntity } from 'models/AdminEntity/AdminEntity';
import { defaultPaginationConfig } from 'models/AdminEntity/constants';
import { RequestConfig } from 'models/AdminEntity/types';
import { Identifier, IdentifierScope } from 'models/Common/types';
import { Workflow } from './types';
import { makeWorkflowPath, workflowListTransformer } from './utils';

/** Fetches a list of `Workflow` records matching the provided `scope` */
export const listWorkflows = (scope: IdentifierScope, config?: RequestConfig) =>
  getAdminEntity(
    {
      path: makeWorkflowPath(scope),
      messageType: Admin.WorkflowList,
      transform: workflowListTransformer,
    },
    { ...defaultPaginationConfig, ...config },
  );

/** Retrieves a single `Workflow` record */
export const getWorkflow = (id: Identifier, config?: RequestConfig) =>
  getAdminEntity<Admin.Workflow, Workflow>(
    {
      path: makeWorkflowPath(id),
      messageType: Admin.Workflow,
    },
    config,
  );
