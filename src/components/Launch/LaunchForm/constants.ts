import { BlobDimensionality, SimpleType } from 'models/Common/types';
import { BlobValue, InputType, RoleType } from './types';

export const formStrings = {
    cancel: 'Cancel',
    inputs: 'Inputs',
    role: 'Role',
    submit: 'Launch',
    taskVersion: 'Task Version',
    title: 'Create New Execution',
    workflowVersion: 'Workflow Version',
    launchPlan: 'Launch Plan'
};

type RoleTypesKey = 'iamRole' | 'k8sServiceAccount';
export const roleTypes: { [k in RoleTypesKey]: RoleType } = {
    iamRole: {
        helperText: 'example: arn:aws:iam::12345678:role/defaultrole',
        inputLabel: 'role urn',
        label: 'IAM Role',
        value: 'assumableIamRole'
    },
    k8sServiceAccount: {
        helperText: 'example: default-service-account',
        inputLabel: 'service account name',
        label: 'Kubernetes Service Account',
        value: 'kubernetesServiceAccount'
    }
};

/** Maps any valid InputType enum to a display string */
export const typeLabels: { [k in InputType]: string } = {
    [InputType.Binary]: 'binary',
    [InputType.Blob]: 'file/blob',
    [InputType.Boolean]: 'boolean',
    [InputType.Collection]: '',
    [InputType.Datetime]: 'datetime - UTC',
    [InputType.Duration]: 'duration - ms',
    [InputType.Error]: 'error',
    [InputType.Enum]: 'enum',
    [InputType.Float]: 'float',
    [InputType.Integer]: 'integer',
    [InputType.Map]: '',
    [InputType.None]: 'none',
    [InputType.Schema]: 'schema - uri',
    [InputType.String]: 'string',
    [InputType.Struct]: 'struct',
    [InputType.Unknown]: 'unknown'
};

/** Maps nested `SimpleType`s to our flattened `InputType` enum. */
export const simpleTypeToInputType: { [k in SimpleType]: InputType } = {
    [SimpleType.BINARY]: InputType.Binary,
    [SimpleType.BOOLEAN]: InputType.Boolean,
    [SimpleType.DATETIME]: InputType.Datetime,
    [SimpleType.DURATION]: InputType.Duration,
    [SimpleType.ERROR]: InputType.Error,
    [SimpleType.FLOAT]: InputType.Float,
    [SimpleType.INTEGER]: InputType.Integer,
    [SimpleType.NONE]: InputType.None,
    [SimpleType.STRING]: InputType.String,
    [SimpleType.STRUCT]: InputType.Struct
};

export const defaultBlobValue: BlobValue = {
    uri: '',
    dimensionality: BlobDimensionality.SINGLE
};

export const launchInputDebouncDelay = 500;

export const requiredInputSuffix = '*';
export const cannotLaunchWorkflowString = 'Workflow cannot be launched';
export const cannotLaunchTaskString = 'Task cannot be launched';
export const inputsDescription =
    'Enter input values below. Items marked with an asterisk(*) are required.';
export const workflowNoInputsString =
    'This workflow does not accept any inputs.';
export const taskNoInputsString = 'This task does not accept any inputs.';
export const workflowUnsupportedRequiredInputsString = `This Workflow version contains one or more required inputs which are not supported by Flyte Console and do not have default values specified in the Workflow definition or the selected Launch Plan.\n\nYou can launch this Workflow version with the Flyte CLI or by selecting a Launch Plan which provides values for the unsupported inputs.\n\nThe required inputs are :`;
export const taskUnsupportedRequiredInputsString = `This Task version contains one or more required inputs which are not supported by Flyte Console.\n\nYou can launch this Task version with the Flyte CLI instead.\n\nThe required inputs are :`;
export const blobUriHelperText = '(required) location of the data';
export const blobFormatHelperText = '(optional) csv, parquet, etc...';
export const correctInputErrors =
    'Some inputs have errors. Please correct them before submitting.';

export const qualityOfServiceTier = {
    UNDEFINED: 0,
    HIGH: 1,
    MEDIUM: 2,
    LOW: 3
};

export const qualityOfServiceTierLabels = {
    [qualityOfServiceTier.UNDEFINED]: 'Undefined',
    [qualityOfServiceTier.HIGH]: 'High',
    [qualityOfServiceTier.MEDIUM]: 'Medium',
    [qualityOfServiceTier.LOW]: 'Low'
};
