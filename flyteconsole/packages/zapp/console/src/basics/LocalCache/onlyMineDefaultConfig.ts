export enum OnlyMyFilter {
  SelectAll = 'selectAll',
  OnlyMyWorkflow = 'onlyMyWorkflow',
  OnlyMyExecutions = 'onlyMyExecutions',
  OnlyMyTasks = 'onlyMyTasks',
  OnlyMyWorkflowVersions = 'onlyMyWorkflowVersions',
  OnlyMyLaunchForm = 'onlyMyLaunchForm',
}

export const filterByDefault = [
  {
    value: 'selectAll',
    label: 'Select All',
    data: true,
  },
  {
    value: 'onlyMyWorkflow',
    label: 'Only My Workflow',
    data: true,
  },
  {
    value: 'onlyMyExecutions',
    label: 'Only My Executions',
    data: true,
  },
  {
    value: 'onlyMyTasks',
    label: 'Only My Tasks',
    data: true,
  },
  {
    value: 'onlyMyWorkflowVersions',
    label: 'Only My Workflow Versions',
    data: true,
  },
  {
    value: 'onlyMyLaunchForm',
    label: 'Only My Launch Form',
    data: true,
  },
];

export const defaultSelectedValues = Object.assign(
  {},
  ...filterByDefault.map((x) => ({ [x.value]: x.data })),
);
