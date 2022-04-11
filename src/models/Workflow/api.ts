import { Admin, Core } from 'flyteidl';
import { getAdminEntity, postAdminEntity } from 'models/AdminEntity/AdminEntity';
import { defaultPaginationConfig } from 'models/AdminEntity/constants';
import { RequestConfig } from 'models/AdminEntity/types';
import { Identifier, IdentifierScope } from 'models/Common/types';
import { makeNamedEntityPath } from 'models/Common/utils';
import { WorkflowExecutionState } from './enums';
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

/** Updates `Workflow` archive state */
export const updateWorkflowState = (
  id: Admin.NamedEntityIdentifier,
  newState: WorkflowExecutionState,
  config?: RequestConfig,
) => {
  const path = makeNamedEntityPath({ resourceType: Core.ResourceType.WORKFLOW, ...id });
  return postAdminEntity<Admin.INamedEntityUpdateRequest, Admin.NamedEntityUpdateResponse>(
    {
      data: {
        resourceType: Core.ResourceType.WORKFLOW,
        id,
        metadata: {
          state: newState,
        },
      },
      path,
      requestMessageType: Admin.NamedEntityUpdateRequest,
      responseMessageType: Admin.NamedEntityUpdateResponse,
      method: 'put',
    },
    config,
  );
};
