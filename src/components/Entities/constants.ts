import { ResourceType } from 'models/Common/types';

type EntityStringMap = { [k in ResourceType]: string };

export const entityStrings: EntityStringMap = {
  [ResourceType.DATASET]: 'dataset',
  [ResourceType.LAUNCH_PLAN]: 'launch plan',
  [ResourceType.TASK]: 'task',
  [ResourceType.UNSPECIFIED]: 'item',
  [ResourceType.WORKFLOW]: 'workflow',
};

interface EntitySectionsFlags {
  description?: boolean;
  executions?: boolean;
  launch?: boolean;
  schedules?: boolean;
  versions?: boolean;
}

export const entitySections: { [k in ResourceType]: EntitySectionsFlags } = {
  [ResourceType.DATASET]: { description: true },
  [ResourceType.LAUNCH_PLAN]: {
    description: true,
    executions: true,
    launch: true,
    schedules: true,
  },
  [ResourceType.TASK]: { description: true, executions: true, launch: true },
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
