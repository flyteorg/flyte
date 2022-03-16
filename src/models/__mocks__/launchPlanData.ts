import { cloneDeep } from 'lodash';
import { FixedRateUnit, Identifier } from '../Common/types';
import { LaunchPlan, LaunchPlanClosure, LaunchPlanSpec, LaunchPlanState } from '../Launch/types';

const basicLaunchPlanClosure: LaunchPlanClosure = {
  state: LaunchPlanState.ACTIVE,
  expectedInputs: {
    parameters: {},
  },
  expectedOutputs: {
    variables: {},
  },
};

export const createMockLaunchPlanSpec: () => LaunchPlanSpec = () => ({
  defaultInputs: {
    parameters: {},
  },
  entityMetadata: {
    notifications: [],
    schedule: {},
  },
  fixedInputs: {
    literals: {},
  },
  role: '',
  workflowId: workflowId('name', 'version'),
});

const workflowId: (name: string, version: string) => Identifier = (name, version) => ({
  name,
  version,
  project: 'flyte',
  domain: 'development',
});

export const createMockLaunchPlan: (name: string, version?: string) => LaunchPlan = (
  name: string,
  version = 'abcdefg',
) => ({
  id: workflowId(name, version),
  spec: createMockLaunchPlanSpec(),
  closure: cloneDeep(basicLaunchPlanClosure),
});

export const mockLaunchPlanSchedules = {
  everyTenMinutes: {
    rate: {
      unit: FixedRateUnit.MINUTE,
      value: 10,
    },
  },
  everyDay6AM: {
    cronExpression: '0 6 * * ? *',
  },
};
