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
  allExecutionsChartTitle_task: 'All Executions in the Task',
  allExecutionsChartTitle_launch_plan: 'All Executions Using Launch Plan',
  versionsTitle_workflow: 'Recent Workflow Versions',
  versionsTitle_task: 'Recent Task Versions',
  versionsTitle_launch_plan: 'Launch Plan Versions',
  searchName_launch_plan: 'Search Launch Plan Name',
  searchName_task: 'Search Task Name',
  searchName_workflow: 'Search Workflow Name',
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
  launchPlanLatest: 'Launch Plan Detail  (Latest Version)',
  expectedInputs: 'Expected Inputs',
  fixedInputs: 'Fixed Inputs',
  inputsName: 'Name',
  inputsType: 'Type',
  inputsRequired: 'Required',
  inputsDefault: 'Default Value',
  configuration: 'Configuration',
  configType: 'type',
  configUrl: 'url',
  configSeed: 'seed',
  configTestSplitRatio: 'test_split_ratio',
  noExpectedInputs: 'This launch plan has no expected inputs.',
  noFixedInputs: 'This launch plan has no fixed inputs.',
};

export { patternKey } from '@flyteconsole/locale';
export default createLocalizedString(str);
