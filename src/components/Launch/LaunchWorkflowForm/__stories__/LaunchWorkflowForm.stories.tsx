import { storiesOf } from '@storybook/react';
import { resolveAfter } from 'common/promiseUtils';
import { mockAPIContextValue } from 'components/data/__mocks__/apiContext';
import { APIContext } from 'components/data/apiContext';
import { Workflow } from 'models';
import { createMockLaunchPlan } from 'models/__mocks__/launchPlanData';
import {
    createMockWorkflow,
    createMockWorkflowClosure,
    createMockWorkflowVersions
} from 'models/__mocks__/workflowData';
import * as React from 'react';
import { LaunchWorkflowForm } from '../LaunchWorkflowForm';
import { mockParameterMap, mockWorkflowInputsInterface } from './mockInputs';

const mockWorkflow = createMockWorkflow('MyWorkflow');
const mockLaunchPlan = createMockLaunchPlan(
    mockWorkflow.id.name,
    mockWorkflow.id.version
);

const mockWorkflowVersions = createMockWorkflowVersions(
    mockWorkflow.id.name,
    10
);

mockLaunchPlan.closure!.expectedInputs = mockParameterMap;

const mockApi = mockAPIContextValue({
    getLaunchPlan: () => resolveAfter(500, mockLaunchPlan),
    getWorkflow: id => {
        const workflow: Workflow = {
            id
        };
        workflow.closure = createMockWorkflowClosure();
        workflow.closure!.compiledWorkflow!.primary.template.interface = mockWorkflowInputsInterface;

        return resolveAfter(500, workflow);
    },
    listWorkflows: () => resolveAfter(500, { entities: mockWorkflowVersions }),
    listLaunchPlans: () => resolveAfter(500, { entities: [mockLaunchPlan] })
});

const onClose = () => console.log('Close');

const stories = storiesOf('Launch/LaunchWorkflowForm', module);
stories.addDecorator(story => (
    <APIContext.Provider value={mockApi}>
        <div style={{ width: 600, height: '95vh' }}>{story()}</div>
    </APIContext.Provider>
));

stories.add('Basic', () => (
    <LaunchWorkflowForm onClose={onClose} workflowId={mockWorkflow.id} />
));
