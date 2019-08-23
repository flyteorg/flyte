import { Admin } from 'flyteidl';
import {
    defaultPaginationConfig,
    getAdminEntity,
    RequestConfig
} from 'models/AdminEntity';
import {
    endpointPrefixes,
    Identifier,
    IdentifierScope,
    makeIdentifierPath
} from 'models/Common';

import { Workflow } from './types';
import { workflowListTransformer } from './utils';

/** Fetches a list of `Workflow` records matching the provided `scope` */
export const listWorkflows = (scope: IdentifierScope, config?: RequestConfig) =>
    getAdminEntity(
        {
            path: makeIdentifierPath(endpointPrefixes.workflow, scope),
            messageType: Admin.WorkflowList,
            transform: workflowListTransformer
        },
        { ...defaultPaginationConfig, ...config }
    );

/** Retrieves a single `Workflow` record */
export const getWorkflow = (id: Identifier, config?: RequestConfig) =>
    getAdminEntity<Admin.Workflow, Workflow>(
        {
            path: makeIdentifierPath(endpointPrefixes.workflow, id),
            messageType: Admin.Workflow
        },
        config
    );
