import { ResourceType } from 'models/Common/types';

type EntityStringMap = { [k in ResourceType]: string };

export const entityStrings: EntityStringMap = {
  [ResourceType.DATASET]: 'dataset',
  [ResourceType.LAUNCH_PLAN]: 'launch plan',
  [ResourceType.TASK]: 'task',
  [ResourceType.UNSPECIFIED]: 'item',
  [ResourceType.WORKFLOW]: 'workflow',
};

type TypeNameToEntityResourceType = { [key: string]: ResourceType };

export const typeNameToEntityResource: TypeNameToEntityResourceType = {
  ['dataset']: ResourceType.DATASET,
  ['launch plan']: ResourceType.LAUNCH_PLAN,
  ['task']: ResourceType.TASK,
  ['item']: ResourceType.UNSPECIFIED,
  ['workflow']: ResourceType.WORKFLOW,
};

interface EntitySectionsFlags {
  description?: boolean;
  executions?: boolean;
  launch?: boolean;
  schedules?: boolean;
  versions?: boolean;
  descriptionInputsAndOutputs?: boolean;
}

export const entitySections: { [k in ResourceType]: EntitySectionsFlags } = {
  [ResourceType.DATASET]: { description: true },
  [ResourceType.LAUNCH_PLAN]: {
    description: true,
    executions: true,
    launch: true,
    schedules: true,
  },
  [ResourceType.TASK]: {
    description: true,
    executions: true,
    launch: true,
    versions: true,
    descriptionInputsAndOutputs: true,
  },
  [ResourceType.UNSPECIFIED]: { description: true },
  [ResourceType.WORKFLOW]: {
    description: true,
    executions: true,
    launch: true,
    schedules: true,
    versions: true,
  },
};

export const WorkflowVersionsTablePageSize = 5;
