import { BlobDimensionality, SimpleType } from 'models/Common/types';
import { BlobValue, InputType, AuthRoleTypes, AuthRoleMeta } from './types';

export const AuthRoleStrings: { [k in AuthRoleTypes]: AuthRoleMeta } = {
  [AuthRoleTypes.k8]: {
    helperText: 'example: default-service-account',
    inputLabel: 'service account name',
    label: 'Kubernetes Service Account',
    value: 'kubernetesServiceAccount',
  },
  [AuthRoleTypes.IAM]: {
    helperText: 'example: arn:aws:iam::12345678:role/defaultrole',
    inputLabel: 'role urn',
    label: 'IAM Role',
    value: 'assumableIamRole',
  },
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
  [InputType.Union]: 'union',
  [InputType.Unknown]: 'unknown',
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
  [SimpleType.STRUCT]: InputType.Struct,
};

export const defaultBlobValue: BlobValue = {
  uri: '',
  dimensionality: BlobDimensionality.SINGLE,
};

export const launchInputDebouncDelay = 500;

export const qualityOfServiceTier = {
  UNDEFINED: 0,
  HIGH: 1,
  MEDIUM: 2,
  LOW: 3,
};

export const qualityOfServiceTierLabels = {
  [qualityOfServiceTier.UNDEFINED]: 'Undefined',
  [qualityOfServiceTier.HIGH]: 'High',
  [qualityOfServiceTier.MEDIUM]: 'Medium',
  [qualityOfServiceTier.LOW]: 'Low',
};

export const workflowNoInputsString = 'This workflow does not accept any inputs.';
