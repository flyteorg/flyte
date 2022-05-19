import { createLocalizedString } from '@flyteconsole/locale';

const str = {
  viewAll: 'View All',
  schedulesHeader: 'Schedules',
  collapseButton: (showAll: boolean) => (showAll ? 'Collapse' : 'Expand'),
  launchStrings_workflow: 'Launch Workflow',
  launchStrings_task: 'Launch Task',
  noDescription_workflow: 'This workflow has no description.',
  noDescription_task: 'This task has no description.',
  noSchedules_workflow: 'This workflow has no schedules.',
  noSchedules_task: 'This task has no schedules.',
  allExecutionsChartTitle_workflow: 'All Executions in the Workflow',
  allExecutionsChartTitle_task: 'All Execuations in the Task',
  versionsTitle_workflow: 'Recent Workflow Versions',
  versionsTitle_task: 'Recent Task Versions',
  details_task: 'Task Details',
  inputsFieldName: 'Inputs',
  outputsFieldName: 'Outputs',
  imageFieldName: 'Image',
  envVarsFieldName: 'Env Vars',
  commandsFieldName: 'Commands',
  empty: 'Empty',
  key: 'Key',
  value: 'Value',
  basicInformation: 'Basic Information',
  description: 'Description',
};

export { patternKey } from '@flyteconsole/locale';
export default createLocalizedString(str);
