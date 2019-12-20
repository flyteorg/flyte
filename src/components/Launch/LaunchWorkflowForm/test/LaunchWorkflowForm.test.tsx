import { ThemeProvider } from '@material-ui/styles';
import { render, wait } from '@testing-library/react';
import { mockAPIContextValue } from 'components/data/__mocks__/apiContext';
import { APIContext, APIContextValue } from 'components/data/apiContext';
import { muiTheme } from 'components/Theme';
import { mapValues } from 'lodash';
import {
    createWorkflowExecution,
    getLaunchPlan,
    getWorkflow,
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
import { pendingPromise } from 'test/utils';
import {
    createMockWorkflowInputsInterface,
    mockSimpleVariables
} from '../__mocks__/mockInputs';
import { formStrings } from '../constants';
import { LaunchWorkflowForm } from '../LaunchWorkflowForm';

function createMockObjects(variables: Record<string, Variable>) {
    const mockWorkflow = createMockWorkflow('MyWorkflow');
    const mockLaunchPlan = createMockLaunchPlan(
        mockWorkflow.id.name,
        mockWorkflow.id.version
    );

    const mockWorkflowVersions = createMockWorkflowVersions(
        mockWorkflow.id.name,
        10
    );

    const parameterMap = {
        parameters: mapValues(variables, v => ({ var: v }))
    };

    mockLaunchPlan.closure!.expectedInputs = parameterMap;
    return { mockWorkflow, mockLaunchPlan, mockWorkflowVersions };
}

describe('LaunchWorkflowForm', () => {
    let onClose: jest.Mock;
    let mockLaunchPlan: LaunchPlan;
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

    const createMocks = () => {
        const mockObjects = createMockObjects(variables);
        mockWorkflow = mockObjects.mockWorkflow;
        mockLaunchPlan = mockObjects.mockLaunchPlan;
        mockWorkflowVersions = mockObjects.mockWorkflowVersions;

        workflowId = mockWorkflow.id;
        mockCreateWorkflowExecution = jest.fn();
        mockGetLaunchPlan = jest.fn().mockResolvedValue(mockLaunchPlan);
        // Return our mock inputs for any version requested
        mockGetWorkflow = jest.fn().mockImplementation(id => {
            const workflow: Workflow = {
                id
            };
            workflow.closure = createMockWorkflowClosure();
            workflow.closure!.compiledWorkflow!.primary.template.interface = createMockWorkflowInputsInterface(
                variables
            );
            return Promise.resolve(workflow);
        });
        mockListLaunchPlans = jest
            .fn()
            .mockResolvedValue({ entities: [mockLaunchPlan] });
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
            const { getByText, queryByText } = renderForm();
            await wait(() => getByText(formStrings.workflowVersion));
            expect(queryByText(formStrings.launchPlan)).not.toBeInTheDocument();
        });

        it('should select the most recent workflow version by default', async () => {});

        it('should select the first launch plan by default', async () => {});
    });
});
