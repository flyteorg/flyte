import { mockSimpleVariables } from '../__mocks__/mockInputs';
import { getInputs } from '../getInputs';
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

        expect(() => getInputs(mockWorkflow, launchPlan)).not.toThrowError();

        delete parameters[stringInputName].default;
        expect(() => getInputs(mockWorkflow, launchPlan)).not.toThrowError();
    });
});
