import { ThemeProvider } from '@material-ui/styles';
import {
    act,
    fireEvent,
    getAllByRole,
    getByRole,
    queryAllByRole,
    render,
    wait,
    waitForElement
} from '@testing-library/react';
import { mockAPIContextValue } from 'components/data/__mocks__/apiContext';
import { APIContext } from 'components/data/apiContext';
import { muiTheme } from 'components/Theme';
import { Core } from 'flyteidl';
import { get, mapValues } from 'lodash';
import * as Long from 'long';
import {
    createWorkflowExecution,
    CreateWorkflowExecutionArguments,
    getLaunchPlan,
    getWorkflow,
    Identifier,
    LaunchPlan,
    listLaunchPlans,
    listWorkflows,
    Literal,
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

const booleanInputName = 'simpleBoolean';
const stringInputName = 'simpleString';
const stringNoLabelName = 'stringNoLabel';
const integerInputName = 'simpleInteger';

function createMockObjects(variables: Record<string, Variable>) {
    const mockWorkflow = createMockWorkflow('MyWorkflow');

    const mockWorkflowVersions = createMockWorkflowVersions(
        mockWorkflow.id.name,
        10
    );

    const mockLaunchPlans = [mockWorkflow.id.name, 'OtherLaunchPlan'].map(
        name => {
            const parameterMap = {
                parameters: mapValues(variables, v => ({ var: v }))
            };
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

        // We want the second launch plan to have inputs which differ, so we'll
        // remove one of the inputs
        delete mockLaunchPlans[1].closure!.expectedInputs.parameters[
            stringNoLabelName
        ];

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

    const getSubmitButton = (container: HTMLElement) => {
        const buttons = queryAllByRole(container, 'button').filter(
            el => el.getAttribute('type') === 'submit'
        );
        expect(buttons.length).toBe(1);
        return buttons[0];
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
            await waitForElement(() =>
                getByLabelText(formStrings.workflowVersion)
            );
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
            const { getByLabelText, getByTitle } = renderForm();
            await wait();

            // Find the launch plan selector, verify it has no value selected
            const launchPlanInput = getByLabelText(formStrings.launchPlan);
            expect(launchPlanInput).toBeInTheDocument();
            expect(launchPlanInput).toHaveValue('');

            // Click the expander for the launch plan, select the first/only item
            const launchPlanDiv = getByTitle(formStrings.launchPlan);
            const expander = getByRole(launchPlanDiv, 'button');
            fireEvent.click(expander);
            const item = await waitForElement(() =>
                getByRole(launchPlanDiv, 'menuitem')
            );
            fireEvent.click(item);

            await wait();
            expect(
                getByLabelText(stringInputName, {
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
            const { container } = renderForm();

            await wait();

            const submitButton = getSubmitButton(container);

            expect(submitButton).toBeDisabled();
            resolve(createMockWorkflowWithInputs(identifier));

            await wait();
            expect(submitButton).not.toBeDisabled();
        });

        it('should not show validation errors until first submit', async () => {
            jest.useFakeTimers();
            const { container, getByLabelText } = renderForm();
            await wait();

            const integerInput = getByLabelText(integerInputName, {
                exact: false
            });
            fireEvent.change(integerInput, { target: { value: 'abc' } });

            act(jest.runAllTimers);
            await wait();
            expect(integerInput).not.toBeInvalid();

            fireEvent.click(getSubmitButton(container));
            await wait();

            expect(integerInput).toBeInvalid();
        });

        it('should update validation errors while typing', async () => {
            jest.useFakeTimers();
            const { container, getByLabelText } = renderForm();
            await wait();

            const integerInput = getByLabelText(integerInputName, {
                exact: false
            });
            fireEvent.change(integerInput, { target: { value: 'abc' } });
            fireEvent.click(getSubmitButton(container));
            await wait();
            expect(integerInput).toBeInvalid();

            fireEvent.change(integerInput, { target: { value: '123' } });
            act(jest.runAllTimers);
            await wait();
            expect(integerInput).toBeValid();
        });

        it('should update launch plan when selecting a new workflow version', async () => {
            const { getByTitle } = renderForm();
            await wait();

            mockListLaunchPlans.mockClear();

            // Click the expander for the workflow, select the second item
            const workflowDiv = getByTitle(formStrings.workflowVersion);
            const expander = getByRole(workflowDiv, 'button');
            fireEvent.click(expander);
            const items = await waitForElement(() =>
                getAllByRole(workflowDiv, 'menuitem')
            );
            fireEvent.click(items[1]);

            await wait();
            expect(mockListLaunchPlans).toHaveBeenCalled();
        });

        it('should update inputs when selecting a new launch plan', async () => {
            const { queryByLabelText, getByTitle } = renderForm();
            await wait();

            // Delete the string input so that its corresponding input will
            // disappear after the new launch plan is loaded.
            delete mockLaunchPlans[1].closure!.expectedInputs.parameters[
                stringInputName
            ];
            mockGetLaunchPlan.mockResolvedValue(mockLaunchPlans[1]);

            // Click the expander for the launch plan, select the second item
            const launchPlanDiv = getByTitle(formStrings.launchPlan);
            const expander = getByRole(launchPlanDiv, 'button');
            fireEvent.click(expander);
            const items = await waitForElement(() =>
                getAllByRole(launchPlanDiv, 'menuitem')
            );
            fireEvent.click(items[1]);

            await wait();
            expect(
                queryByLabelText(stringInputName, {
                    // Don't use exact match because the label will be decorated with type info
                    exact: false
                })
            ).toBeNull();
        });

        it('should reset form error when inputs change', async () => {
            const errorString = 'Something went wrong';
            mockCreateWorkflowExecution.mockRejectedValue(
                new Error(errorString)
            );

            const {
                container,
                getByText,
                getByTitle,
                queryByText
            } = renderForm();
            await wait();

            fireEvent.click(getSubmitButton(container));
            await wait();

            expect(getByText(errorString)).toBeInTheDocument();

            mockGetLaunchPlan.mockResolvedValue(mockLaunchPlans[1]);
            // Click the expander for the launch plan, select the second item
            const launchPlanDiv = getByTitle(formStrings.launchPlan);
            const expander = getByRole(launchPlanDiv, 'button');
            fireEvent.click(expander);
            const items = await waitForElement(() =>
                getAllByRole(launchPlanDiv, 'menuitem')
            );
            fireEvent.click(items[1]);
            await wait();
            expect(queryByText(errorString)).not.toBeInTheDocument();
        });

        describe('Input Values', () => {
            it('Should send false for untouched toggles', async () => {
                let inputs: Core.ILiteralMap = {};
                mockCreateWorkflowExecution.mockImplementation(
                    ({
                        inputs: passedInputs
                    }: CreateWorkflowExecutionArguments) => {
                        inputs = passedInputs;
                        return pendingPromise();
                    }
                );

                const { container } = renderForm();
                await wait();

                fireEvent.click(getSubmitButton(container));
                await wait();

                expect(mockCreateWorkflowExecution).toHaveBeenCalled();
                expect(inputs.literals).toBeDefined();
                const value = get(
                    inputs.literals,
                    `${booleanInputName}.scalar.primitive.boolean`
                );
                expect(value).toBe(false);
            });

            it('should use default values when provided', async () => {
                // Add defaults for the string/integer inputs and check that they are
                // correctly populated
                const parameters = mockLaunchPlans[0].closure!.expectedInputs
                    .parameters;
                parameters[stringInputName].default = {
                    scalar: { primitive: { stringValue: 'abc' } }
                } as Literal;
                parameters[integerInputName].default = {
                    scalar: { primitive: { integer: Long.fromNumber(10000) } }
                } as Literal;
                mockGetLaunchPlan.mockResolvedValue(mockLaunchPlans[0]);

                const { getByLabelText } = renderForm();
                await wait();

                expect(
                    getByLabelText(stringInputName, { exact: false })
                ).toHaveValue('abc');
                expect(
                    getByLabelText(integerInputName, { exact: false })
                ).toHaveValue('10000');
            });

            it('should decorate labels for required inputs', async () => {
                // Add defaults for the string/integer inputs and check that they are
                // correctly populated
                const parameters = mockLaunchPlans[0].closure!.expectedInputs
                    .parameters;
                parameters[stringInputName].required = true;
                mockGetLaunchPlan.mockResolvedValue(mockLaunchPlans[0]);

                const { getByText } = renderForm();
                await wait();
                expect(
                    getByText(stringInputName, { exact: false }).textContent
                ).toContain('*');
            });
        });
    });
});
