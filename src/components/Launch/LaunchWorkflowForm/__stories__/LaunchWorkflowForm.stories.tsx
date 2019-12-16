import { action } from '@storybook/addon-actions';
import { storiesOf } from '@storybook/react';
import { resolveAfter } from 'common/promiseUtils';
import { mockAPIContextValue } from 'components/data/__mocks__/apiContext';
import { APIContext } from 'components/data/apiContext';
import { Admin } from 'flyteidl';
import { mapValues } from 'lodash';
import { Variable, Workflow } from 'models';
import { createMockLaunchPlan } from 'models/__mocks__/launchPlanData';
import {
    createMockWorkflow,
    createMockWorkflowClosure,
    createMockWorkflowVersions
} from 'models/__mocks__/workflowData';
import { mockExecution } from 'models/Execution/__mocks__/mockWorkflowExecutionsData';
import * as React from 'react';
import { LaunchWorkflowForm } from '../LaunchWorkflowForm';
import {
    createMockWorkflowInputsInterface,
    mockCollectionVariables,
    mockSimpleVariables
} from './mockInputs';

const submitAction = action('createWorkflowExecution');

const renderForm = (variables: Record<string, Variable>) => {
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
            submitAction(input);
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

stories.add('Simple', () => renderForm(mockSimpleVariables));
stories.add('Collections', () => renderForm(mockCollectionVariables));
