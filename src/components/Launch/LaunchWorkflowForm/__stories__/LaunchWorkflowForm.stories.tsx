import { action } from '@storybook/addon-actions';
import { storiesOf } from '@storybook/react';
import { resolveAfter } from 'common/promiseUtils';
import { WaitForData } from 'components/common';
import { mockAPIContextValue } from 'components/data/__mocks__/apiContext';
import { APIContext } from 'components/data/apiContext';
import { mapValues } from 'lodash';
import * as Long from 'long';
import {
    Execution,
    ExecutionData,
    Literal,
    LiteralMap,
    Variable,
    Workflow
} from 'models';
import { createMockLaunchPlan } from 'models/__mocks__/launchPlanData';
import {
    createMockWorkflow,
    createMockWorkflowClosure,
    createMockWorkflowVersions
} from 'models/__mocks__/workflowData';
import { mockWorkflowExecutionResponse } from 'models/Execution/__mocks__/mockWorkflowExecutionsData';
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
import { binaryInputName, errorInputName } from '../test/constants';
import { useExecutionLaunchConfiguration } from '../useExecutionLaunchConfiguration';
import { getWorkflowInputs } from '../utils';

const booleanInputName = 'simpleBoolean';
const blobInputName = 'simpleBlob';
const stringInputName = 'simpleString';
const integerInputName = 'simpleInteger';
const submitAction = action('createWorkflowExecution');

const generateMocks = (variables: Record<string, Variable>) => {
    const mockWorkflow = createMockWorkflow('MyWorkflow');
    mockWorkflow.closure = createMockWorkflowClosure();
    mockWorkflow.closure!.compiledWorkflow!.primary.template.interface = createMockWorkflowInputsInterface(
        variables
    );

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

    const mockExecutionData: ExecutionData = {
        inputs: { url: 'inputsUrl', bytes: Long.fromNumber(1000) },
        outputs: { url: 'outputsUrl', bytes: Long.fromNumber(1000) }
    };

    const mockExecutionInputs: LiteralMap = Object.keys(
        parameterMap.parameters
    ).reduce(
        (out, paramName) => {
            const defaultValue =
                simpleVariableDefaults[paramName as SimpleVariableKey];
            out.literals[paramName] = defaultValue as Literal;
            return out;
        },
        { literals: {} } as LiteralMap
    );

    const mockApi = mockAPIContextValue({
        createWorkflowExecution: input => {
            console.log(input);
            submitAction('See console for data');
            return Promise.reject('Not implemented');
        },
        getExecutionData: () => resolveAfter(100, mockExecutionData),
        getRemoteLiteralMap: () => resolveAfter(100, mockExecutionInputs),
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

interface RenderFormArgs {
    execution?: Execution;
    mocks: ReturnType<typeof generateMocks>;
}

const LaunchFormWithExecution: React.FC<RenderFormArgs & {
    execution: Execution;
}> = ({ execution, mocks: { mockWorkflow } }) => {
    const launchConfig = useExecutionLaunchConfiguration({
        execution,
        workflowInputs: getWorkflowInputs(mockWorkflow)
    });
    const onClose = () => console.log('Close');
    return (
        <WaitForData {...launchConfig}>
            <LaunchWorkflowForm
                onClose={onClose}
                workflowId={mockWorkflow.id}
                initialParameters={launchConfig.value}
            />
        </WaitForData>
    );
};

const renderForm = (args: RenderFormArgs) => {
    const {
        mocks: { mockApi, mockWorkflow }
    } = args;
    const onClose = () => console.log('Close');

    return (
        <APIContext.Provider value={mockApi}>
            <div style={{ width: 600, height: '95vh' }}>
                {args.execution ? (
                    <LaunchFormWithExecution
                        {...args}
                        execution={args.execution}
                    />
                ) : (
                    <LaunchWorkflowForm
                        onClose={onClose}
                        workflowId={mockWorkflow.id}
                    />
                )}
            </div>
        </APIContext.Provider>
    );
};

const stories = storiesOf('Launch/LaunchWorkflowForm', module);

stories.add('Simple', () =>
    renderForm({ mocks: generateMocks(mockSimpleVariables) })
);
stories.add('Required Inputs', () => {
    const mocks = generateMocks(mockSimpleVariables);
    const parameters = mocks.mockLaunchPlan.closure!.expectedInputs.parameters;
    parameters[stringInputName].required = true;
    parameters[integerInputName].required = true;
    parameters[booleanInputName].required = true;
    parameters[blobInputName].required = true;
    return renderForm({ mocks });
});
stories.add('Default Values', () => {
    const mocks = generateMocks(mockSimpleVariables);
    const parameters = mocks.mockLaunchPlan.closure!.expectedInputs.parameters;
    Object.keys(parameters).forEach(paramName => {
        const defaultValue =
            simpleVariableDefaults[paramName as SimpleVariableKey];
        parameters[paramName].default = defaultValue as Literal;
    });
    return renderForm({ mocks });
});
stories.add('Collections', () =>
    renderForm({ mocks: generateMocks(mockCollectionVariables) })
);
stories.add('Nested Collections', () =>
    renderForm({ mocks: generateMocks(mockNestedCollectionVariables) })
);
stories.add('Launched from Execution', () => {
    const mocks = generateMocks(mockSimpleVariables);
    // Only providing values for the properties in the execution which
    // are actually used by the component
    const execution = {
        closure: {
            workflowId: mocks.mockWorkflow.id
        },
        spec: {
            launchPlan: mocks.mockLaunchPlan.id
        },
        id: mockWorkflowExecutionResponse.id
    } as Execution;
    return renderForm({ mocks, execution });
});
stories.add('Unsupported Required Values', () => {
    const mocks = generateMocks(mockSimpleVariables);
    const parameters = mocks.mockLaunchPlan.closure!.expectedInputs.parameters;
    parameters[binaryInputName].required = true;
    parameters[errorInputName].required = true;
    return renderForm({ mocks });
});
