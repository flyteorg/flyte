import { LiteralMapBlob, ResourceType, SystemStatus } from './types';

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

export const defaultSystemStatus: SystemStatus = { status: 'normal' };
