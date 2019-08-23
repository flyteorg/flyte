import { LiteralMapBlob, ResourceType } from './types';

export const endpointPrefixes = {
    execution: '/executions',
    launchPlan: '/launch_plans',
    nodeExecution: '/node_executions',
    project: '/projects',
    relaunchExecution: '/executions/relaunch',
    task: '/tasks',
    taskExecution: '/task_executions',
    taskExecutionChildren: '/children/task_executions',
    workflow: '/workflows'
};

export const identifierPrefixes: { [k in ResourceType]: string } = {
    [ResourceType.LAUNCH_PLAN]: '/launch_plan_ids',
    [ResourceType.TASK]: '/task_ids',
    [ResourceType.UNSPECIFIED]: '',
    [ResourceType.WORKFLOW]: '/workflow_ids'
};

export const emptyLiteralMapBlob: LiteralMapBlob = {
    values: { literals: {} }
};
