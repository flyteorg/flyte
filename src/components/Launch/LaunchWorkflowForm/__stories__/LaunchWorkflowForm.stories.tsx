import { action } from '@storybook/addon-actions';
import { storiesOf } from '@storybook/react';
import { resolveAfter } from 'common/promiseUtils';
import { mockAPIContextValue } from 'components/data/__mocks__/apiContext';
import { APIContext } from 'components/data/apiContext';
import { mapValues } from 'lodash';
import { Literal, Variable, Workflow } from 'models';
import { createMockLaunchPlan } from 'models/__mocks__/launchPlanData';
import {
    createMockWorkflow,
    createMockWorkflowClosure,
    createMockWorkflowVersions
} from 'models/__mocks__/workflowData';
import * as React from 'react';
import {
    createMockWorkflowInputsInterface,
    mockCollectionVariables,
    mockNestedCollectionVariables,
    mockSimpleVariables,
    simpleVariableDefaults,
    SimpleVariableKey
} from '../__mocks__/mockInputs';
import { LaunchWorkflowForm } from '../LaunchWorkflowForm';

const booleanInputName = 'simpleBoolean';
const stringInputName = 'simpleString';
const integerInputName = 'simpleInteger';
const submitAction = action('createWorkflowExecution');

const generateMocks = (variables: Record<string, Variable>) => {
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

    const mockApi = mockAPIContextValue({
        createWorkflowExecution: input => {
            console.log(input);
            submitAction('See console for data');
            return Promise.reject('Not implemented');
        },
        getLaunchPlan: () => resolveAfter(500, mockLaunchPlan),
        getWorkflow: id => {
            const workflow: Workflow = {
                id
            };
            workflow.closure = createMockWorkflowClosure();
            workflow.closure!.compiledWorkflow!.primary.template.interface = createMockWorkflowInputsInterface(
                variables
            );

            return resolveAfter(500, workflow);
        },
        listWorkflows: () =>
            resolveAfter(500, { entities: mockWorkflowVersions }),
        listLaunchPlans: () => resolveAfter(500, { entities: [mockLaunchPlan] })
    });

    return { mockWorkflow, mockLaunchPlan, mockWorkflowVersions, mockApi };
};

const renderForm = ({
    mockApi,
    mockWorkflow
}: ReturnType<typeof generateMocks>) => {
    const onClose = () => console.log('Close');

    return (
        <APIContext.Provider value={mockApi}>
            <div style={{ width: 600, height: '95vh' }}>
                <LaunchWorkflowForm
                    onClose={onClose}
                    workflowId={mockWorkflow.id}
                />
            </div>
        </APIContext.Provider>
    );
};

const stories = storiesOf('Launch/LaunchWorkflowForm', module);

stories.add('Simple', () => renderForm(generateMocks(mockSimpleVariables)));
stories.add('Required Inputs', () => {
    const mocks = generateMocks(mockSimpleVariables);
    const parameters = mocks.mockLaunchPlan.closure!.expectedInputs.parameters;
    parameters[stringInputName].required = true;
    parameters[integerInputName].required = true;
    parameters[booleanInputName].required = true;
    return renderForm(mocks);
});
stories.add('Default Values', () => {
    const mocks = generateMocks(mockSimpleVariables);
    const parameters = mocks.mockLaunchPlan.closure!.expectedInputs.parameters;
    Object.keys(parameters).forEach(paramName => {
        const defaultValue =
            simpleVariableDefaults[paramName as SimpleVariableKey];
        parameters[paramName].default = defaultValue as Literal;
    });
    return renderForm(mocks);
});
stories.add('Collections', () =>
    renderForm(generateMocks(mockCollectionVariables))
);
stories.add('Nested Collections', () =>
    renderForm(generateMocks(mockNestedCollectionVariables))
);
