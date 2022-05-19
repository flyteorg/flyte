import { ResourceType } from 'models/Common/types';
import { listTasks } from 'models/Task/api';
import { listWorkflows } from 'models/Workflow/api';
import { listLaunchPlans } from 'models/Launch/api';
import { Workflow } from 'models/Workflow/types';
import { Task } from 'models/Task/types';
import { LaunchPlan } from 'models/Launch/types';

interface EntityFunctions {
  description?: boolean;
  executions?: boolean;
  launch?: boolean;
  listEntity?: any;
}

export type EntityType = Workflow | Task | LaunchPlan;

const unspecifiedFn = () => {
  throw new Error('Unspecified Resourcetype.');
};
const unimplementedFn = () => {
  throw new Error('Method not implemented.');
};

export const entityFunctions: { [k in ResourceType]: EntityFunctions } = {
  [ResourceType.DATASET]: {
    listEntity: unimplementedFn,
  },
  [ResourceType.LAUNCH_PLAN]: {
    listEntity: listLaunchPlans,
  },
  [ResourceType.TASK]: {
    listEntity: listTasks,
  },
  [ResourceType.UNSPECIFIED]: {
    listEntity: unspecifiedFn,
  },
  [ResourceType.WORKFLOW]: {
    listEntity: listWorkflows,
  },
};
