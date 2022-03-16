import { mockSimpleVariables } from '../__mocks__/mockInputs';
import { getInputsForWorkflow } from '../getInputs';
import { stringInputName } from './constants';
import { createMockObjects } from './utils';

describe('getInputs', () => {
  let mocks: ReturnType<typeof createMockObjects>;
  beforeEach(() => {
    mocks = createMockObjects(mockSimpleVariables);
  });

  it('should not throw when parameter default is null/undefined', () => {
    const { mockLaunchPlans, mockWorkflow } = mocks;
    const launchPlan = mockLaunchPlans[0];
    const parameters = launchPlan.closure!.expectedInputs.parameters;
    parameters[stringInputName].default = null;

    expect(() => getInputsForWorkflow(mockWorkflow, launchPlan)).not.toThrowError();

    delete parameters[stringInputName].default;
    expect(() => getInputsForWorkflow(mockWorkflow, launchPlan)).not.toThrowError();
  });
});
