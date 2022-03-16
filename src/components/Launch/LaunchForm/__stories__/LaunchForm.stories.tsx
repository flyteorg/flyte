import { action } from '@storybook/addon-actions';
import { storiesOf } from '@storybook/react';
import { resolveAfter } from 'common/promiseUtils';
import { WaitForData } from 'components/common/WaitForData';
import { APIContext } from 'components/data/apiContext';
import { mockAPIContextValue } from 'components/data/__mocks__/apiContext';
import { mapValues } from 'lodash';
import * as Long from 'long';
import { Literal, LiteralMap, Variable } from 'models/Common/types';
import { Execution, ExecutionData } from 'models/Execution/types';
import { mockWorkflowExecutionResponse } from 'models/Execution/__mocks__/mockWorkflowExecutionsData';
import { Workflow } from 'models/Workflow/types';
import { createMockLaunchPlan } from 'models/__mocks__/launchPlanData';
import {
  createMockWorkflow,
  createMockWorkflowClosure,
  createMockWorkflowVersions,
} from 'models/__mocks__/workflowData';
import * as React from 'react';
import { LaunchForm } from '../LaunchForm';
import { binaryInputName, errorInputName } from '../test/constants';
import { WorkflowInitialLaunchParameters } from '../types';
import { useMappedExecutionInputValues } from '../useMappedExecutionInputValues';
import { getWorkflowInputs } from '../utils';
import {
  createMockInputsInterface,
  mockCollectionVariables,
  mockNestedCollectionVariables,
  mockSimpleVariables,
  simpleVariableDefaults,
  SimpleVariableKey,
} from '../__mocks__/mockInputs';

const booleanInputName = 'simpleBoolean';
const blobInputName = 'simpleBlob';
const stringInputName = 'simpleString';
const integerInputName = 'simpleInteger';
const submitAction = action('createWorkflowExecution');

const generateMocks = (variables: Record<string, Variable>) => {
  const mockWorkflow = createMockWorkflow('MyWorkflow');
  mockWorkflow.closure = createMockWorkflowClosure();
  mockWorkflow.closure!.compiledWorkflow!.primary.template.interface =
    createMockInputsInterface(variables);

  const mockLaunchPlan = createMockLaunchPlan(mockWorkflow.id.name, mockWorkflow.id.version);

  const mockWorkflowVersions = createMockWorkflowVersions(mockWorkflow.id.name, 10);

  const parameterMap = {
    parameters: mapValues(variables, (v) => ({ var: v })),
  };

  mockLaunchPlan.closure!.expectedInputs = parameterMap;

  const mockExecutionData: ExecutionData = {
    inputs: { url: 'inputsUrl', bytes: Long.fromNumber(1000) },
    outputs: { url: 'outputsUrl', bytes: Long.fromNumber(1000) },
    fullInputs: null,
    fullOutputs: null,
  };

  const mockExecutionInputs: LiteralMap = Object.keys(parameterMap.parameters).reduce(
    (out, paramName) => {
      const defaultValue = simpleVariableDefaults[paramName as SimpleVariableKey];
      out.literals[paramName] = defaultValue as Literal;
      return out;
    },
    { literals: {} } as LiteralMap,
  );

  const mockApi = mockAPIContextValue({
    createWorkflowExecution: (input) => {
      console.log(input);
      submitAction('See console for data');
      return Promise.reject('Not implemented');
    },
    getExecutionData: () => resolveAfter(100, mockExecutionData),
    getRemoteLiteralMap: () => resolveAfter(100, mockExecutionInputs),
    getLaunchPlan: () => resolveAfter(500, mockLaunchPlan),
    getWorkflow: (id) => {
      const workflow: Workflow = {
        id,
      };
      workflow.closure = createMockWorkflowClosure();
      workflow.closure!.compiledWorkflow!.primary.template.interface =
        createMockInputsInterface(variables);

      return resolveAfter(500, workflow);
    },
    listWorkflows: () => resolveAfter(500, { entities: mockWorkflowVersions }),
    listLaunchPlans: () => resolveAfter(500, { entities: [mockLaunchPlan] }),
  });

  return { mockWorkflow, mockLaunchPlan, mockWorkflowVersions, mockApi };
};

interface RenderFormArgs {
  execution?: Execution;
  mocks: ReturnType<typeof generateMocks>;
}

const LaunchFormWithExecution: React.FC<
  RenderFormArgs & {
    execution: Execution;
  }
> = ({ execution, mocks: { mockWorkflow } }) => {
  const {
    closure: { workflowId },
    spec: { launchPlan },
  } = execution;
  const executionInputs = useMappedExecutionInputValues({
    execution,
    inputDefinitions: getWorkflowInputs(mockWorkflow),
  });
  const initialParameters: WorkflowInitialLaunchParameters = {
    workflowId,
    launchPlan,
    values: executionInputs.value,
  };
  const onClose = () => console.log('Close');
  return (
    <WaitForData {...executionInputs}>
      <LaunchForm
        onClose={onClose}
        workflowId={mockWorkflow.id}
        initialParameters={initialParameters}
      />
    </WaitForData>
  );
};

const renderForm = (args: RenderFormArgs) => {
  const {
    mocks: { mockApi, mockWorkflow },
  } = args;
  const onClose = () => console.log('Close');

  return (
    <APIContext.Provider value={mockApi}>
      <div style={{ width: 600, height: '95vh' }}>
        {args.execution ? (
          <LaunchFormWithExecution {...args} execution={args.execution} />
        ) : (
          <LaunchForm onClose={onClose} workflowId={mockWorkflow.id} />
        )}
      </div>
    </APIContext.Provider>
  );
};

const stories = storiesOf('Launch/LaunchForm/Workflow', module);

stories.add('Simple', () => renderForm({ mocks: generateMocks(mockSimpleVariables) }));
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
  Object.keys(parameters).forEach((paramName) => {
    const defaultValue = simpleVariableDefaults[paramName as SimpleVariableKey];
    parameters[paramName].default = defaultValue as Literal;
  });
  return renderForm({ mocks });
});
stories.add('Collections', () => renderForm({ mocks: generateMocks(mockCollectionVariables) }));
stories.add('Nested Collections', () =>
  renderForm({ mocks: generateMocks(mockNestedCollectionVariables) }),
);
stories.add('Launched from Execution', () => {
  const mocks = generateMocks(mockSimpleVariables);
  // Only providing values for the properties in the execution which
  // are actually used by the component
  const execution = {
    closure: {
      workflowId: mocks.mockWorkflow.id,
    },
    spec: {
      launchPlan: mocks.mockLaunchPlan.id,
    },
    id: mockWorkflowExecutionResponse.id,
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
