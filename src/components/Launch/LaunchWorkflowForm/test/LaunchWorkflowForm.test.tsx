import { ThemeProvider } from '@material-ui/styles';
import {
    fireEvent,
    getByRole,
    getByText,
    render,
    wait
} from '@testing-library/react';
import { mockAPIContextValue } from 'components/data/__mocks__/apiContext';
import { APIContext } from 'components/data/apiContext';
import { muiTheme } from 'components/Theme';
import { mapValues } from 'lodash';
import {
    createWorkflowExecution,
    getLaunchPlan,
    getWorkflow,
    Identifier,
    LaunchPlan,
    listLaunchPlans,
    listWorkflows,
    NamedEntityIdentifier,
    Variable,
    Workflow
} from 'models';
import { createMockLaunchPlan } from 'models/__mocks__/launchPlanData';
import {
    createMockWorkflow,
    createMockWorkflowClosure,
    createMockWorkflowVersions
} from 'models/__mocks__/workflowData';
import * as React from 'react';
import { delayedPromise, pendingPromise } from 'test/utils';
import {
    createMockWorkflowInputsInterface,
    mockSimpleVariables
} from '../__mocks__/mockInputs';
import { formStrings } from '../constants';
import { LaunchWorkflowForm } from '../LaunchWorkflowForm';

function createMockObjects(variables: Record<string, Variable>) {
    const mockWorkflow = createMockWorkflow('MyWorkflow');

    const mockWorkflowVersions = createMockWorkflowVersions(
        mockWorkflow.id.name,
        10
    );

    const parameterMap = {
        parameters: mapValues(variables, v => ({ var: v }))
    };

    const mockLaunchPlans = [mockWorkflow.id.name, 'OtherLaunchPlan'].map(
        name => {
            const launchPlan = createMockLaunchPlan(
                name,
                mockWorkflow.id.version
            );
            launchPlan.closure!.expectedInputs = parameterMap;
            return launchPlan;
        }
    );
    return { mockWorkflow, mockLaunchPlans, mockWorkflowVersions };
}

describe('LaunchWorkflowForm', () => {
    let onClose: jest.Mock;
    let mockLaunchPlans: LaunchPlan[];
    let mockWorkflow: Workflow;
    let mockWorkflowVersions: Workflow[];
    let workflowId: NamedEntityIdentifier;
    let variables: Record<string, Variable>;

    let mockListLaunchPlans: jest.Mock<ReturnType<typeof listLaunchPlans>>;
    let mockListWorkflows: jest.Mock<ReturnType<typeof listWorkflows>>;
    let mockGetLaunchPlan: jest.Mock<ReturnType<typeof getLaunchPlan>>;
    let mockGetWorkflow: jest.Mock<ReturnType<typeof getWorkflow>>;
    let mockCreateWorkflowExecution: jest.Mock<ReturnType<
        typeof createWorkflowExecution
    >>;

    beforeEach(() => {
        onClose = jest.fn();
    });

    const createMockWorkflowWithInputs = (id: Identifier) => {
        const workflow: Workflow = {
            id
        };
        workflow.closure = createMockWorkflowClosure();
        workflow.closure!.compiledWorkflow!.primary.template.interface = createMockWorkflowInputsInterface(
            variables
        );
        return workflow;
    };

    const createMocks = () => {
        const mockObjects = createMockObjects(variables);
        mockWorkflow = mockObjects.mockWorkflow;
        mockLaunchPlans = mockObjects.mockLaunchPlans;
        mockWorkflowVersions = mockObjects.mockWorkflowVersions;

        workflowId = mockWorkflow.id;
        mockCreateWorkflowExecution = jest.fn();
        mockGetLaunchPlan = jest.fn().mockResolvedValue(mockLaunchPlans[0]);
        // Return our mock inputs for any version requested
        mockGetWorkflow = jest
            .fn()
            .mockImplementation(id =>
                Promise.resolve(createMockWorkflowWithInputs(id))
            );
        mockListLaunchPlans = jest
            .fn()
            .mockResolvedValue({ entities: mockLaunchPlans });
        mockListWorkflows = jest
            .fn()
            .mockResolvedValue({ entities: mockWorkflowVersions });
    };

    const renderForm = () => {
        return render(
            <ThemeProvider theme={muiTheme}>
                <APIContext.Provider
                    value={mockAPIContextValue({
                        createWorkflowExecution: mockCreateWorkflowExecution,
                        getLaunchPlan: mockGetLaunchPlan,
                        getWorkflow: mockGetWorkflow,
                        listLaunchPlans: mockListLaunchPlans,
                        listWorkflows: mockListWorkflows
                    })}
                >
                    <LaunchWorkflowForm
                        onClose={onClose}
                        workflowId={workflowId}
                    />
                </APIContext.Provider>
            </ThemeProvider>
        );
    };

    describe('With Simple Inputs', () => {
        beforeEach(() => {
            variables = mockSimpleVariables;
            createMocks();
        });

        it('should not show workflow selector until options have loaded', async () => {
            mockListWorkflows.mockReturnValue(pendingPromise());
            const { queryByText } = renderForm();
            await wait();
            expect(
                queryByText(formStrings.workflowVersion)
            ).not.toBeInTheDocument();
        });

        it('should not show launch plan selector until list has loaded', async () => {
            mockListLaunchPlans.mockReturnValue(pendingPromise());
            const { getByLabelText, queryByText } = renderForm();
            await wait(() => getByLabelText(formStrings.workflowVersion));
            expect(queryByText(formStrings.launchPlan)).not.toBeInTheDocument();
        });

        it('should select the most recent workflow version by default', async () => {
            const { getByLabelText } = renderForm();
            await wait();
            expect(getByLabelText(formStrings.workflowVersion)).toHaveValue(
                mockWorkflowVersions[0].id.version
            );
        });

        it('should select the launch plan matching the workflow name by default', async () => {
            const { getByLabelText } = renderForm();
            await wait();
            expect(getByLabelText(formStrings.launchPlan)).toHaveValue(
                mockWorkflow.id.name
            );
        });

        it('should not render inputs until workflow and launch plan are selected', async () => {
            // Remove default launch plan so it is not auto-selected
            const launchPlans = mockLaunchPlans.filter(
                lp => lp.id.name !== workflowId.name
            );
            mockListLaunchPlans.mockResolvedValue({
                entities: launchPlans
            });
            const { getByLabelText, getByTitle, container } = renderForm();
            await wait();

            // Find the launch plan selector, verify it has no value selected
            const launchPlanInput = getByLabelText(formStrings.launchPlan);
            expect(launchPlanInput).toBeInTheDocument();
            expect(launchPlanInput).toHaveValue('');

            // Click the expander for the launch plan, select the first/only item
            const launchPlanDiv = getByTitle(formStrings.launchPlan);
            const expander = getByRole(launchPlanDiv, 'button');
            fireEvent.click(expander);
            await wait(() => getByRole(launchPlanDiv, 'menuitem'));
            fireEvent.click(getByRole(launchPlanDiv, 'menuitem'));

            await wait();
            const simpleInputName = Object.keys(variables)[0];
            expect(
                getByLabelText(simpleInputName, {
                    // Don't use exact match because the label will be decorated with type info
                    exact: false
                })
            ).toBeInTheDocument();
        });

        it('should disable submit button until inputs have loaded', async () => {
            let identifier: Identifier = {} as Identifier;
            const { promise, resolve } = delayedPromise<Workflow>();
            mockGetWorkflow.mockImplementation(id => {
                identifier = id;
                return promise;
            });
            const { queryAllByRole } = renderForm();

            await wait();

            const buttons = queryAllByRole('button').filter(
                el => el.getAttribute('type') === 'submit'
            );
            expect(buttons.length).toBe(1);
            const submitButton = buttons[0];

            expect(submitButton).toBeDisabled();
            resolve(createMockWorkflowWithInputs(identifier));

            await wait();
            expect(submitButton).not.toBeDisabled();
        });

        it('should not show validation errors until first submit', async () => {});

        it('should update validation errors while typing', async () => {});

        it('should update launch plan when selecting a new workflow version', async () => {});

        it('should update inputs when selecting a new launch plan', () => {});

        it('should reset form error when selecting a new launch plan', async () => {});

        describe('Input Values', () => {
            it('Should send false for untouched toggles', async () => {});
        });
    });
});
