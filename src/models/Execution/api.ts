import { Admin, Core } from 'flyteidl';
import { getAdminEntity, postAdminEntity } from 'models/AdminEntity/AdminEntity';
import {
  defaultListExecutionChildrenConfig,
  defaultPaginationConfig,
} from 'models/AdminEntity/constants';
import { RequestConfig } from 'models/AdminEntity/types';
import { endpointPrefixes } from 'models/Common/constants';
import { Identifier, IdentifierScope, NameIdentifierScope } from 'models/Common/types';
import { makeIdentifierPath } from 'models/Common/utils';
import { defaultExecutionPrincipal } from './constants';
import { ExecutionState } from './enums';
import {
  Execution,
  ExecutionData,
  ExecutionMetadata,
  NodeExecution,
  NodeExecutionIdentifier,
  TaskExecution,
  TaskExecutionIdentifier,
  WorkflowExecutionIdentifier,
} from './types';
import {
  executionListTransformer,
  makeExecutionPath,
  makeNodeExecutionListPath,
  makeNodeExecutionPath,
  makeTaskExecutionChildrenPath,
  makeTaskExecutionListPath,
  makeTaskExecutionPath,
  nodeExecutionListTransformer,
  taskExecutionListTransformer,
} from './utils';

/** Fetches a list of `Execution` records matching the provided `scope` */
export const listExecutions = (scope: IdentifierScope, config?: RequestConfig) =>
  getAdminEntity(
    {
      path: makeIdentifierPath(endpointPrefixes.execution, scope),
      messageType: Admin.ExecutionList,
      transform: executionListTransformer,
    },
    {
      ...defaultPaginationConfig,
      ...config,
    },
  );

/** Fetches a single `Execution` record */
export const getExecution = (id: WorkflowExecutionIdentifier, config?: RequestConfig) =>
  getAdminEntity<Admin.Execution, Execution>(
    {
      path: makeExecutionPath(id),
      messageType: Admin.Execution,
    },
    config,
  );

const emptyExecutionData: ExecutionData = {
  inputs: {},
  outputs: {},
  fullInputs: null,
  fullOutputs: null,
};

/** Fetches data URLs for an `Execution` record */
export const getExecutionData = (id: WorkflowExecutionIdentifier, config?: RequestConfig) =>
  getAdminEntity<Admin.WorkflowExecutionGetDataResponse, ExecutionData>(
    {
      path: `/data${makeExecutionPath(id)}`,
      messageType: Admin.WorkflowExecutionGetDataResponse,
      // Admin isn't guaranteed to populate both inputs and outputs,
      // so ensure that a safe access is possible for each
      transform: (result) => ({ ...emptyExecutionData, ...result } as ExecutionData),
    },
    config,
  );

export interface CreateWorkflowExecutionArguments {
  annotations?: Admin.IAnnotations | null;
  authRole?: Admin.IAuthRole;
  securityContext?: Core.ISecurityContext;
  domain: string;
  disableAll?: boolean | null;
  labels?: Admin.ILabels | null;
  maxParallelism?: number | null;
  inputs: Core.ILiteralMap;
  launchPlanId: Identifier;
  project: string;
  referenceExecutionId?: WorkflowExecutionIdentifier;
}

/** Submits a request to create a new `WorkflowExecution` using the provided
 * LaunchPlan and input values.
 */
export const createWorkflowExecution = (
  {
    annotations,
    authRole,
    securityContext,
    domain,
    disableAll,
    labels,
    maxParallelism,
    inputs,
    launchPlanId: launchPlan,
    project,
    referenceExecutionId: referenceExecution,
  }: CreateWorkflowExecutionArguments,
  config?: RequestConfig,
) => {
  const spec: Admin.IExecutionSpec = {
    inputs,
    launchPlan,
    metadata: {
      referenceExecution,
      principal: defaultExecutionPrincipal,
    },
    labels,
    annotations,
  };

  if (authRole?.assumableIamRole || authRole?.kubernetesServiceAccount) {
    spec.authRole = authRole;
  }

  if (securityContext?.runAs?.iamRole || securityContext?.runAs?.k8sServiceAccount) {
    spec.securityContext = securityContext;
  }

  if (disableAll) {
    spec.disableAll = disableAll;
  }

  if (maxParallelism !== undefined) {
    spec.maxParallelism = maxParallelism;
  }

  return postAdminEntity<Admin.IExecutionCreateRequest, Admin.ExecutionCreateResponse>(
    {
      data: {
        project,
        domain,
        spec,
      },
      path: endpointPrefixes.execution,
      requestMessageType: Admin.ExecutionCreateRequest,
      responseMessageType: Admin.ExecutionCreateResponse,
    },
    config,
  );
};

/** Submits a request to terminate a WorkflowExecution by id */
export const terminateWorkflowExecution = (
  id: WorkflowExecutionIdentifier,
  cause: string,
  config?: RequestConfig,
) =>
  postAdminEntity<Admin.IExecutionTerminateRequest, Admin.ExecutionTerminateResponse>(
    {
      data: { cause },
      path: makeExecutionPath(id),
      requestMessageType: Admin.ExecutionTerminateRequest,
      responseMessageType: Admin.ExecutionTerminateResponse,
      method: 'delete',
    },
    config,
  );

interface RelaunchParams {
  id: WorkflowExecutionIdentifier;
  name?: string;
}

/** Submits a request to relaunch a WorkflowExecution by id */
export const relaunchWorkflowExecution = ({ id, name }: RelaunchParams, config?: RequestConfig) =>
  postAdminEntity<Admin.IExecutionRelaunchRequest, Admin.ExecutionCreateResponse>(
    {
      data: { id, name },
      path: endpointPrefixes.relaunchExecution,
      requestMessageType: Admin.ExecutionRelaunchRequest,
      responseMessageType: Admin.ExecutionCreateResponse,
    },
    config,
  );

interface RecoverParams {
  id: WorkflowExecutionIdentifier;
  name?: string;
  metadata?: ExecutionMetadata;
}

/**
 * Submits a request to recover a WorkflowExecution
 */
export const recoverWorkflowExecution = (
  { id, name, metadata }: RecoverParams,
  config?: RequestConfig,
) =>
  postAdminEntity<Admin.IExecutionRecoverRequest, Admin.ExecutionCreateResponse>(
    {
      data: { id, name, metadata },
      path: endpointPrefixes.recoverExecution,
      requestMessageType: Admin.ExecutionRecoverRequest,
      responseMessageType: Admin.ExecutionCreateResponse,
    },
    config,
  );

/** Retrieves a single `NodeExecution` record */
export const getNodeExecution = (id: NodeExecutionIdentifier, config?: RequestConfig) =>
  getAdminEntity<Admin.NodeExecution, NodeExecution>(
    {
      path: makeNodeExecutionPath(id),
      messageType: Admin.NodeExecution,
    },
    config,
  );

/** Fetches data URLs for a NodeExecution (used when fetching Dynamicworkflows
 *  from nodeExeecutions at runtime) */
export const getNodeExecutionData = (id: NodeExecutionIdentifier, config?: RequestConfig) =>
  getAdminEntity<Admin.NodeExecutionGetDataResponse, ExecutionData>(
    {
      path: `/data${makeNodeExecutionPath(id)}`,
      messageType: Admin.NodeExecutionGetDataResponse,
    },
    config,
  );

/** Fetches a list of `NodeExecution` records matching the provided `scope` */
export const listNodeExecutions = (scope: NameIdentifierScope, config?: RequestConfig) =>
  getAdminEntity(
    {
      path: makeNodeExecutionListPath(scope),
      messageType: Admin.NodeExecutionList,
      transform: nodeExecutionListTransformer,
    },
    {
      ...defaultPaginationConfig,
      ...defaultListExecutionChildrenConfig,
      ...config,
    },
  );

/** Fetches a list of `NodeExecution` records which are children of a given
 * `TaskExecution`.
 */
export const listTaskExecutionChildren = (
  taskExecutionId: TaskExecutionIdentifier,
  config?: RequestConfig,
) =>
  getAdminEntity(
    {
      path: makeTaskExecutionChildrenPath(taskExecutionId),
      messageType: Admin.NodeExecutionList,
      transform: nodeExecutionListTransformer,
    },
    {
      ...defaultPaginationConfig,
      ...defaultListExecutionChildrenConfig,
      ...config,
    },
  );

/** Fetches a single `TaskExecution` record */
export const getTaskExecution = (id: TaskExecutionIdentifier, config?: RequestConfig) =>
  getAdminEntity<Admin.TaskExecution, TaskExecution>(
    {
      path: makeTaskExecutionPath(id),
      messageType: Admin.TaskExecution,
    },
    config,
  );

/** Fetches the data URLs for a TaskExecution */
export const getTaskExecutionData = (id: TaskExecutionIdentifier, config?: RequestConfig) =>
  getAdminEntity<Admin.TaskExecutionGetDataResponse, ExecutionData>(
    {
      path: `/data${makeTaskExecutionPath(id)}`,
      messageType: Admin.TaskExecutionGetDataResponse,
    },
    config,
  );

/** Fetches a list of TaskExecutions for a given NodeExecution */
export const listTaskExecutions = (id: NodeExecutionIdentifier, config?: RequestConfig) =>
  getAdminEntity(
    {
      path: makeTaskExecutionListPath(id),
      messageType: Admin.TaskExecutionList,
      transform: taskExecutionListTransformer,
    },
    {
      ...defaultPaginationConfig,
      ...defaultListExecutionChildrenConfig,
      ...config,
    },
  );

/** Updates Execution archive state */
export const updateExecution = (
  id: WorkflowExecutionIdentifier,
  newState: ExecutionState,
  config?: RequestConfig,
): Promise<Admin.ExecutionUpdateResponse> => {
  return postAdminEntity<Admin.IExecutionUpdateRequest, Admin.ExecutionUpdateResponse>(
    {
      data: { id, state: newState },
      path: makeExecutionPath(id),
      requestMessageType: Admin.ExecutionUpdateRequest,
      responseMessageType: Admin.ExecutionUpdateResponse,
      method: 'put',
    },
    config,
  );
};
