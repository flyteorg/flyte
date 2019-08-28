import { storiesOf } from '@storybook/react';
import { createMockLaunchPlan } from 'models/__mocks__/launchPlanData';
import {
    createMockWorkflow,
    createMockWorkflowClosure
} from 'models/__mocks__/workflowData';
import * as React from 'react';
import { LaunchWorkflowForm } from '../LaunchWorkflowForm';
import { mockParameterMap, mockWorkflowInputsInterface } from './mockInputs';

const mockWorkflow = createMockWorkflow('MyWorkflow');
const mockLaunchPlan = createMockLaunchPlan(
    mockWorkflow.id.name,
    mockWorkflow.id.version
);

mockLaunchPlan.closure!.expectedInputs = mockParameterMap;
mockWorkflow.closure = createMockWorkflowClosure();
mockWorkflow.closure!.compiledWorkflow!.primary.template.interface = mockWorkflowInputsInterface;

const stories = storiesOf('Launch/LaunchWorkflowForm', module);
stories.addDecorator(story => (
    <div style={{ width: 600, height: '95vh' }}>{story()}</div>
));

stories.add('Basic', () => (
    <LaunchWorkflowForm
        workflow={mockWorkflow}
        workflowId={mockWorkflow.id}
        launchPlan={mockLaunchPlan}
    />
));
