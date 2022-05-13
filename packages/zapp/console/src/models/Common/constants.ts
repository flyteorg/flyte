import axios, { AxiosRequestConfig, AxiosTransformer } from 'axios';
import * as camelcaseKeys from 'camelcase-keys';
import * as snakecaseKeys from 'snakecase-keys';
import { LiteralMapBlob, ResourceType, SystemStatus } from './types';

// recommended util.d.ts implementation
const isObject = (value: unknown): boolean => {
  return value !== null && typeof value === 'object';
};

export const endpointPrefixes = {
  execution: '/executions',
  launchPlan: '/launch_plans',
  namedEntity: '/named_entities',
  nodeExecution: '/node_executions',
  dynamicWorkflowExecution: '/data/node_executions',
  project: '/projects',
  projectDomainAtributes: '/project_domain_attributes',
  relaunchExecution: '/executions/relaunch',
  recoverExecution: '/executions/recover',
  task: '/tasks',
  taskExecution: '/task_executions',
  taskExecutionChildren: '/children/task_executions',
  workflow: '/workflows',
};

export const identifierPrefixes: { [k in ResourceType]: string } = {
  [ResourceType.DATASET]: '',
  [ResourceType.LAUNCH_PLAN]: '/launch_plan_ids',
  [ResourceType.TASK]: '/task_ids',
  [ResourceType.UNSPECIFIED]: '',
  [ResourceType.WORKFLOW]: '/workflow_ids',
};

export const emptyLiteralMapBlob: LiteralMapBlob = {
  values: { literals: {} },
};

/** Config object that can be used for requests that are not sent to
 * the Admin entity API (`/api/v1/...`), such as the `/me` endpoint. This config
 * ensures that requests/responses are correctly converted and that cookies are
 * included.
 */
export const defaultAxiosConfig: AxiosRequestConfig = {
  transformRequest: [
    (data: any) => (isObject(data) ? snakecaseKeys(data) : data),
    ...(axios.defaults.transformRequest as AxiosTransformer[]),
  ],
  transformResponse: [...(axios.defaults.transformResponse as AxiosTransformer[]), camelcaseKeys],
  withCredentials: true,
};

export const defaultSystemStatus: SystemStatus = { status: 'normal' };
