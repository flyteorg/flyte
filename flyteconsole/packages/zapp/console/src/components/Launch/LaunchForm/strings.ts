import { createLocalizedString } from '@flyteconsole/locale';

const str = {
  requiredInputSuffix: '*',
  cannotLaunchWorkflowString: 'Workflow cannot be launched',
  cannotLaunchTaskString: 'Task cannot be launched',
  inputsDescription: 'Enter input values below. Items marked with an asterisk(*) are required.',
  gateInputDescription: 'Enter input values below.',
  taskNoInputsString: 'This task does not accept any inputs.',
  workflowUnsupportedRequiredInputsString: `This Workflow version contains one or more required inputs which are not supported by Flyte Console and do not have default values specified in the Workflow definition or the selected Launch Plan.\n\nYou can launch this Workflow version with the Flyte CLI or by selecting a Launch Plan which provides values for the unsupported inputs.\n\nThe required inputs are :`,
  taskUnsupportedRequiredInputsString: `This Task version contains one or more required inputs which are not supported by Flyte Console.\n\nYou can launch this Task version with the Flyte CLI instead.\n\nThe required inputs are :`,
  blobUriHelperText: '(required) location of the data',
  blobFormatHelperText: '(optional) csv, parquet, etc...',
  correctInputErrors: 'Some inputs have errors. Please correct them before submitting.',
  noneInputTypeDescription: 'The value of none type is empty',
  cancel: 'Cancel',
  inputs: 'Inputs',
  gateInput: 'Gate input',
  role: 'Role',
  submit: 'Launch',
  resume: 'Resume',
  taskVersion: 'Task Version',
  title: 'Create New Execution',
  resumeTitle: 'Resume Paused Execution',
  workflowVersion: 'Workflow Version',
  launchPlan: 'Launch Plan',
  interruptible: 'Interruptible',
  viewNodeInputs: 'View node inputs',
};

export { patternKey } from '@flyteconsole/locale';
export default createLocalizedString(str);
