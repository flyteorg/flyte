import { FilterOperation, FilterOperationName } from 'models/AdminEntity/types';
import { ResourceIdentifier, ResourceType, Identifier } from 'models/Common/types';
import { Routes } from 'routes/routes';
import { entityStrings } from './constants';

const noFilters = () => [];

export const executionFilterGenerator: {
  [k in ResourceType]: (id: ResourceIdentifier, version?: string) => FilterOperation[];
} = {
  [ResourceType.DATASET]: noFilters,
  [ResourceType.LAUNCH_PLAN]: ({ name }, version) =>
    version
      ? [
          {
            key: 'launch_plan.name',
            operation: FilterOperationName.EQ,
            value: name,
          },
          {
            key: 'launch_plan.version',
            operation: FilterOperationName.EQ,
            value: version,
          },
        ]
      : [
          {
            key: 'launch_plan.name',
            operation: FilterOperationName.EQ,
            value: name,
          },
        ],
  [ResourceType.TASK]: ({ name }) => [
    {
      key: 'task.name',
      operation: FilterOperationName.EQ,
      value: name,
    },
  ],
  [ResourceType.UNSPECIFIED]: noFilters,
  [ResourceType.WORKFLOW]: ({ name }, version) =>
    version
      ? [
          {
            key: 'workflow.version',
            operation: FilterOperationName.EQ,
            value: version,
          },
        ]
      : [
          {
            key: 'workflow.name',
            operation: FilterOperationName.EQ,
            value: name,
          },
        ],
};

const workflowListGenerator = ({ project, domain }: ResourceIdentifier) =>
  Routes.ProjectDetails.sections.workflows.makeUrl(project, domain);
const launchPlanListGenerator = ({ project, domain }: ResourceIdentifier) =>
  Routes.ProjectDetails.sections.launchPlans.makeUrl(project, domain);
const taskListGenerator = ({ project, domain }: ResourceIdentifier) =>
  Routes.ProjectDetails.sections.tasks.makeUrl(project, domain);
const unspecifiedGenerator = ({ project, domain }: ResourceIdentifier | Identifier) => {
  throw new Error('Unspecified Resourcetype.');
};
const unimplementedGenerator = ({ project, domain }: ResourceIdentifier | Identifier) => {
  throw new Error('Method not implemented.');
};

export const backUrlGenerator: {
  [k in ResourceType]: (id: ResourceIdentifier) => string;
} = {
  [ResourceType.DATASET]: unimplementedGenerator,
  [ResourceType.LAUNCH_PLAN]: launchPlanListGenerator,
  [ResourceType.TASK]: taskListGenerator,
  [ResourceType.UNSPECIFIED]: unspecifiedGenerator,
  [ResourceType.WORKFLOW]: workflowListGenerator,
};

const workflowDetailGenerator = ({ project, domain, name }: ResourceIdentifier) =>
  Routes.WorkflowDetails.makeUrl(project, domain, name);
const launchPlanDetailGenerator = ({ project, domain, name }: ResourceIdentifier) =>
  Routes.LaunchPlanDetails.makeUrl(project, domain, name);
const taskDetailGenerator = ({ project, domain, name }: ResourceIdentifier) =>
  Routes.TaskDetails.makeUrl(project, domain, name);

export const backToDetailUrlGenerator: {
  [k in ResourceType]: (id: ResourceIdentifier) => string;
} = {
  [ResourceType.DATASET]: unimplementedGenerator,
  [ResourceType.LAUNCH_PLAN]: launchPlanDetailGenerator,
  [ResourceType.TASK]: taskDetailGenerator,
  [ResourceType.UNSPECIFIED]: unspecifiedGenerator,
  [ResourceType.WORKFLOW]: workflowDetailGenerator,
};

const workflowVersopmDetailsGenerator = ({ project, domain, name, version }: Identifier) =>
  Routes.EntityVersionDetails.makeUrl(
    project,
    domain,
    name,
    entityStrings[ResourceType.WORKFLOW],
    version,
  );
const taskVersionDetailsGenerator = ({ project, domain, name, version }: Identifier) =>
  Routes.EntityVersionDetails.makeUrl(
    project,
    domain,
    name,
    entityStrings[ResourceType.TASK],
    version,
  );
const launchPlanVersionDetailsGenerator = ({ project, domain, name, version }: Identifier) =>
  Routes.EntityVersionDetails.makeUrl(
    project,
    domain,
    name,
    entityStrings[ResourceType.LAUNCH_PLAN],
    version,
  );

const entityMapVersionDetailsUrl: {
  [k in ResourceType]: (id: Identifier) => string;
} = {
  [ResourceType.DATASET]: unimplementedGenerator,
  [ResourceType.LAUNCH_PLAN]: launchPlanVersionDetailsGenerator,
  [ResourceType.TASK]: taskVersionDetailsGenerator,
  [ResourceType.UNSPECIFIED]: unspecifiedGenerator,
  [ResourceType.WORKFLOW]: workflowVersopmDetailsGenerator,
};

export const versionDetailsUrlGenerator = (id: Identifier): string => {
  if (id?.resourceType) return entityMapVersionDetailsUrl[id?.resourceType](id);
  return '';
};
