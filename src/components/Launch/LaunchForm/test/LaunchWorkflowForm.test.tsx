import { ThemeProvider } from '@material-ui/styles';
import {
  act,
  fireEvent,
  getAllByRole,
  getByRole,
  queryAllByRole,
  render,
  waitFor,
} from '@testing-library/react';
import { APIContext } from 'components/data/apiContext';
import { mockAPIContextValue } from 'components/data/__mocks__/apiContext';
import { muiTheme } from 'components/Theme/muiTheme';
import { Core, Protobuf } from 'flyteidl';
import { cloneDeep, get } from 'lodash';
import * as Long from 'long';
import { RequestConfig } from 'models/AdminEntity/types';
import { Identifier, Literal, NamedEntityIdentifier, Variable } from 'models/Common/types';
import { createWorkflowExecution, CreateWorkflowExecutionArguments } from 'models/Execution/api';
import { listLaunchPlans } from 'models/Launch/api';
import { LaunchPlan } from 'models/Launch/types';
import { getWorkflow, listWorkflows } from 'models/Workflow/api';
import { Workflow } from 'models/Workflow/types';
import { createMockWorkflowClosure } from 'models/__mocks__/workflowData';
import * as React from 'react';
import { delayedPromise, pendingPromise } from 'test/utils';
import {
  cannotLaunchWorkflowString,
  formStrings,
  inputsDescription,
  requiredInputSuffix,
  workflowNoInputsString,
} from '../constants';
import { LaunchForm } from '../LaunchForm';
import { LaunchFormProps, WorkflowInitialLaunchParameters } from '../types';
import { createInputCacheKey, getInputDefintionForLiteralType } from '../utils';
import {
  createMockInputsInterface,
  mockSimpleVariables,
  simpleVariableDefaults,
} from '../__mocks__/mockInputs';
import {
  binaryInputName,
  booleanInputName,
  integerInputName,
  stringInputName,
  stringNoLabelName,
} from './constants';
import { createMockObjects } from './utils';

describe('LaunchForm: Workflow', () => {
  let onClose: jest.Mock;
  let mockLaunchPlans: LaunchPlan[];
  let mockSingleLaunchPlan: LaunchPlan;
  let mockWorkflow: Workflow;
  let mockWorkflowVersions: Workflow[];
  let workflowId: NamedEntityIdentifier;
  let variables: Record<string, Variable>;

  let mockListLaunchPlans: jest.Mock<ReturnType<typeof listLaunchPlans>>;
  let mockListWorkflows: jest.Mock<ReturnType<typeof listWorkflows>>;
  let mockGetWorkflow: jest.Mock<ReturnType<typeof getWorkflow>>;
  let mockCreateWorkflowExecution: jest.Mock<ReturnType<typeof createWorkflowExecution>>;

  beforeEach(() => {
    onClose = jest.fn();
    jest.useFakeTimers();
  });

  afterEach(() => {
    jest.runOnlyPendingTimers();
    jest.useRealTimers();
  });

  const createMockWorkflowWithInputs = (id: Identifier) => {
    const workflow: Workflow = {
      id,
    };
    workflow.closure = createMockWorkflowClosure();
    workflow.closure!.compiledWorkflow!.primary.template.interface =
      createMockInputsInterface(variables);
    return workflow;
  };

  const createMocks = () => {
    const mockObjects = createMockObjects(variables);
    mockWorkflow = mockObjects.mockWorkflow;
    mockLaunchPlans = mockObjects.mockLaunchPlans;
    mockSingleLaunchPlan = mockLaunchPlans[0];

    // We want the second launch plan to have inputs which differ, so we'll
    // remove one of the inputs
    delete mockLaunchPlans[1].closure!.expectedInputs.parameters[stringNoLabelName];

    mockWorkflowVersions = mockObjects.mockWorkflowVersions;

    workflowId = mockWorkflow.id;
    mockCreateWorkflowExecution = jest.fn();
    // Return our mock inputs for any version requested
    mockGetWorkflow = jest
      .fn()
      .mockImplementation((id) => Promise.resolve(createMockWorkflowWithInputs(id)));
    mockListLaunchPlans = jest
      .fn()
      .mockImplementation((scope: Partial<Identifier>, { filter }: RequestConfig) => {
        // If the scope has a filter, the calling
        // code is searching for a specific item. So we'll
        // return a single-item list containing it.
        if (filter && filter[0].key === 'version') {
          const launchPlan = { ...mockSingleLaunchPlan };
          launchPlan.id = {
            ...scope,
            version: filter[0].value,
          } as Identifier;
          return Promise.resolve({
            entities: [launchPlan],
          });
        }
        return Promise.resolve({ entities: mockLaunchPlans });
      });

    // For workflow/task list endpoints: If the scope has a filter, the calling
    // code is searching for a specific item. So we'll return a single-item
    // list containing it.
    mockListWorkflows = jest
      .fn()
      .mockImplementation((scope: Partial<Identifier>, { filter }: RequestConfig) => {
        if (filter && filter[0].key === 'version') {
          const workflow = { ...mockWorkflowVersions[0] };
          workflow.id = {
            ...scope,
            version: filter[0].value,
          } as Identifier;
          return Promise.resolve({
            entities: [workflow],
          });
        }
        return Promise.resolve({ entities: mockWorkflowVersions });
      });
  };

  const renderForm = (props?: Partial<LaunchFormProps>) => {
    return render(
      <ThemeProvider theme={muiTheme}>
        <APIContext.Provider
          value={mockAPIContextValue({
            createWorkflowExecution: mockCreateWorkflowExecution,
            getWorkflow: mockGetWorkflow,
            listLaunchPlans: mockListLaunchPlans,
            listWorkflows: mockListWorkflows,
          })}
        >
          <LaunchForm onClose={onClose} workflowId={workflowId} {...props} />
        </APIContext.Provider>
      </ThemeProvider>,
    );
  };

  const getSubmitButton = (container: HTMLElement) => {
    const buttons = queryAllByRole(container, 'button').filter(
      (el) => el.getAttribute('type') === 'submit',
    );
    expect(buttons.length).toBe(1);
    return buttons[0];
  };

  describe('With No Inputs', () => {
    beforeEach(() => {
      variables = {};
      createMocks();
    });

    it('should render info message', async () => {
      const { container, getByText } = renderForm();
      const submitButton = await waitFor(() => getSubmitButton(container));
      await waitFor(() => expect(submitButton).toBeEnabled());

      expect(getByText(workflowNoInputsString)).toBeInTheDocument();
    });

    it('should not render inputs header/description', async () => {
      const { container, queryByText } = renderForm();
      const submitButton = await waitFor(() => getSubmitButton(container));
      await waitFor(() => expect(submitButton).toBeEnabled());

      expect(queryByText(formStrings.inputs)).toBeNull();
      expect(queryByText(inputsDescription)).toBeNull();
    });
  });

  describe('With Simple Inputs', () => {
    beforeEach(() => {
      variables = cloneDeep(mockSimpleVariables);
      createMocks();
    });

    it('should not show workflow selector until options have loaded', async () => {
      mockListWorkflows.mockReturnValue(pendingPromise());
      const { queryByText } = renderForm();
      await waitFor(() => {});
      expect(queryByText(formStrings.workflowVersion)).not.toBeInTheDocument();
    });

    it('should not show launch plan selector until list has loaded', async () => {
      mockListLaunchPlans.mockReturnValue(pendingPromise());
      const { getByLabelText, queryByText } = renderForm();
      await waitFor(() => getByLabelText(formStrings.workflowVersion));
      expect(queryByText(formStrings.launchPlan)).not.toBeInTheDocument();
    });

    it('should select the most recent workflow version by default', async () => {
      const { getByLabelText } = renderForm();
      await waitFor(() => {});
      expect(getByLabelText(formStrings.workflowVersion)).toHaveValue(
        mockWorkflowVersions[0].id.version,
      );
    });

    it('should select the launch plan matching the workflow name by default', async () => {
      const { getByLabelText } = renderForm();
      const launchPlanEl = await waitFor(() => getByLabelText(formStrings.launchPlan));
      expect(launchPlanEl).toHaveValue(mockWorkflow.id.name);
    });

    it('should not render inputs if no launch plan is selected', async () => {
      mockListLaunchPlans.mockResolvedValue({
        entities: [],
      });
      const { getByLabelText, queryByLabelText } = renderForm();
      await waitFor(() => {});

      // Find the launch plan selector, verify it has no value selected
      const launchPlanInput = getByLabelText(formStrings.launchPlan);
      expect(launchPlanInput).toBeInTheDocument();
      expect(launchPlanInput).toHaveValue('');
      expect(
        queryByLabelText(stringInputName, {
          // Don't use exact match because the label will be decorated with type info
          exact: false,
        }),
      ).toBeNull();
    });

    it('should disable submit button until inputs have loaded', async () => {
      let identifier: Identifier = {} as Identifier;
      const { promise, resolve } = delayedPromise<Workflow>();
      mockGetWorkflow.mockImplementation((id) => {
        identifier = id;
        return promise;
      });
      const { container } = renderForm();

      await waitFor(() => {});

      const submitButton = getSubmitButton(container);

      expect(submitButton).toBeDisabled();
      resolve(createMockWorkflowWithInputs(identifier));

      await waitFor(() => {});
      expect(submitButton).not.toBeDisabled();
    });

    it('should not show validation errors until first submit', async () => {
      const { container, getByLabelText } = renderForm();
      await waitFor(() => {});

      const integerInput = getByLabelText(integerInputName, {
        exact: false,
      });
      fireEvent.change(integerInput, { target: { value: 'abc' } });

      act(() => {
        jest.runAllTimers();
      });
      await waitFor(() => {});
      expect(integerInput).not.toBeInvalid();

      fireEvent.click(getSubmitButton(container));
      await waitFor(() => {});

      expect(integerInput).toBeInvalid();
    });

    it('should update validation errors while typing', async () => {
      const { container, getByLabelText } = renderForm();
      await waitFor(() => {});

      const integerInput = getByLabelText(integerInputName, {
        exact: false,
      });
      fireEvent.change(integerInput, { target: { value: 'abc' } });
      fireEvent.click(getSubmitButton(container));
      await waitFor(() => {});
      expect(integerInput).toBeInvalid();

      fireEvent.change(integerInput, { target: { value: '123' } });
      act(() => {
        jest.runAllTimers();
      });
      await waitFor(() => {});
      expect(integerInput).toBeValid();
    });

    it('should allow submission after fixing validation errors', async () => {
      const { container, getByLabelText } = renderForm();
      await waitFor(() => {});

      const integerInput = await waitFor(() =>
        getByLabelText(integerInputName, {
          exact: false,
        }),
      );
      const submitButton = getSubmitButton(container);
      fireEvent.change(integerInput, { target: { value: 'abc' } });
      fireEvent.click(submitButton);
      await waitFor(() => expect(integerInput).toBeInvalid());
      expect(mockCreateWorkflowExecution).not.toHaveBeenCalled();

      fireEvent.change(integerInput, { target: { value: '123' } });
      await waitFor(() => expect(integerInput).toBeValid());
      fireEvent.click(submitButton);
      await waitFor(() => expect(mockCreateWorkflowExecution).toHaveBeenCalled());
    });

    it('should update launch plan when selecting a new workflow version', async () => {
      const { getByTitle } = renderForm();
      await waitFor(() => getByTitle(formStrings.launchPlan));

      mockListLaunchPlans.mockClear();

      // Click the expander for the workflow, select the second item
      const workflowDiv = getByTitle(formStrings.workflowVersion);
      const expander = getByRole(workflowDiv, 'button');
      fireEvent.click(expander);
      const items = await waitFor(() => getAllByRole(workflowDiv, 'menuitem'));
      fireEvent.click(items[1]);

      await waitFor(() => getByTitle(formStrings.launchPlan));
      expect(mockListLaunchPlans).toHaveBeenCalled();
    });

    it('should not clear launch plan when selecting the already selected workflow version', async () => {
      const { getByLabelText, getByTitle } = renderForm();
      await waitFor(() => {});

      mockListLaunchPlans.mockClear();

      // Click the expander for the workflow, select the second item
      const workflowDiv = getByTitle(formStrings.workflowVersion);
      const expander = getByRole(workflowDiv, 'button');
      fireEvent.click(expander);
      const items = await waitFor(() => getAllByRole(workflowDiv, 'menuitem'));
      fireEvent.click(items[0]);

      await waitFor(() => {});
      expect(mockListLaunchPlans).not.toHaveBeenCalled();
      expect(getByLabelText(formStrings.launchPlan)).toHaveValue(mockWorkflow.id.name);
    });

    it('should update inputs when selecting a new launch plan', async () => {
      const { queryByLabelText, getByTitle } = renderForm();
      const launchPlanDiv = await waitFor(() => getByTitle(formStrings.launchPlan));

      // Delete the string input so that its corresponding input will
      // disappear after the new launch plan is loaded.
      delete mockLaunchPlans[1].closure!.expectedInputs.parameters[stringInputName];

      // Click the expander for the launch plan, select the second item
      const expander = getByRole(launchPlanDiv, 'button');
      fireEvent.click(expander);
      const items = await waitFor(() => getAllByRole(launchPlanDiv, 'menuitem'));
      fireEvent.click(items[1]);

      await waitFor(() => getByTitle(formStrings.inputs));
      expect(
        queryByLabelText(stringInputName, {
          // Don't use exact match because the label will be decorated with type info
          exact: false,
        }),
      ).toBeNull();
    });

    it('should preserve input values when changing launch plan', async () => {
      const { getByLabelText, getByTitle } = renderForm();
      await waitFor(() => {});

      const integerInput = getByLabelText(integerInputName, {
        exact: false,
      });
      fireEvent.change(integerInput, { target: { value: '10' } });
      await waitFor(() => {});

      // Click the expander for the launch plan, select the second item
      const launchPlanDiv = getByTitle(formStrings.launchPlan);
      const expander = getByRole(launchPlanDiv, 'button');
      fireEvent.click(expander);
      const items = await waitFor(() => getAllByRole(launchPlanDiv, 'menuitem'));
      fireEvent.click(items[1]);
      await waitFor(() => {});

      expect(
        getByLabelText(integerInputName, {
          exact: false,
        }),
      ).toHaveValue('10');
    });

    it('should reset form error when inputs change', async () => {
      const errorString = 'Something went wrong';
      mockCreateWorkflowExecution.mockRejectedValue(new Error(errorString));

      const { container, getByText, getByTitle, queryByText } = renderForm();
      await waitFor(() => {});

      fireEvent.click(getSubmitButton(container));
      await waitFor(() => {});

      expect(getByText(errorString)).toBeInTheDocument();

      mockSingleLaunchPlan = mockLaunchPlans[1];
      // Click the expander for the launch plan, select the second item
      const launchPlanDiv = getByTitle(formStrings.launchPlan);
      const expander = getByRole(launchPlanDiv, 'button');
      fireEvent.click(expander);
      const items = await waitFor(() => getAllByRole(launchPlanDiv, 'menuitem'));
      fireEvent.click(items[1]);
      await waitFor(() => {});
      expect(queryByText(errorString)).not.toBeInTheDocument();
    });

    describe('Input Values', () => {
      it('Should send false for untouched toggles', async () => {
        let inputs: Core.ILiteralMap = {};
        mockCreateWorkflowExecution.mockImplementation(
          ({ inputs: passedInputs }: CreateWorkflowExecutionArguments) => {
            inputs = passedInputs;
            return pendingPromise();
          },
        );

        const { container } = renderForm();
        await waitFor(() => {});

        fireEvent.click(getSubmitButton(container));
        await waitFor(() => {});

        expect(mockCreateWorkflowExecution).toHaveBeenCalled();
        expect(inputs.literals).toBeDefined();
        const value = get(inputs.literals, `${booleanInputName}.scalar.primitive.boolean`);
        expect(value).toBe(false);
      });

      it('should use default values when provided', async () => {
        // Add defaults for the string/integer inputs and check that they are
        // correctly populated
        const parameters = mockLaunchPlans[0].closure!.expectedInputs.parameters;
        parameters[stringInputName].default = {
          scalar: { primitive: { stringValue: 'abc' } },
        } as Literal;
        parameters[integerInputName].default = {
          scalar: { primitive: { integer: Long.fromNumber(10000) } },
        } as Literal;

        const { getByLabelText } = renderForm();
        await waitFor(() => {});

        expect(getByLabelText(stringInputName, { exact: false })).toHaveValue('abc');
        expect(getByLabelText(integerInputName, { exact: false })).toHaveValue('10000');
      });

      it('should decorate labels for required inputs', async () => {
        // Add defaults for the string/integer inputs and check that they are
        // correctly populated
        const parameters = mockLaunchPlans[0].closure!.expectedInputs.parameters;
        parameters[stringInputName].required = true;

        const { getByText } = renderForm();
        await waitFor(() => {});
        expect(
          getByText(stringInputName, {
            exact: false,
            selector: 'label',
          }).textContent,
        ).toContain('*');
      });
    });

    describe('When using initial parameters', () => {
      it('should prefer the provided workflow version', async () => {
        const initialParameters: WorkflowInitialLaunchParameters = {
          workflowId: mockWorkflowVersions[2].id,
        };
        const { getByLabelText } = renderForm({ initialParameters });
        await waitFor(() => {});
        expect(getByLabelText(formStrings.workflowVersion)).toHaveValue(
          mockWorkflowVersions[2].id.version,
        );
      });

      it('should only include one instance of the preferred version in the selector', async () => {
        const initialParameters: WorkflowInitialLaunchParameters = {
          workflowId: mockWorkflowVersions[2].id,
        };
        const { getByTitle } = renderForm({ initialParameters });
        await waitFor(() => {});
        // Click the expander for the workflow, select the second item
        const versionDiv = getByTitle(formStrings.workflowVersion);
        const expander = getByRole(versionDiv, 'button');
        fireEvent.click(expander);
        const items = await waitFor(() => getAllByRole(versionDiv, 'menuitem'));

        const expectedVersion = mockWorkflowVersions[2].id.version;
        expect(
          items.filter((item) => item.textContent && item.textContent.includes(expectedVersion)),
        ).toHaveLength(1);
      });

      it('should fall back to the first item in the list if preferred workflow is not found', async () => {
        mockListWorkflows.mockImplementation((scope: Partial<Identifier>) => {
          // If we get a request for a specific item,
          // simulate not found
          if (scope.version) {
            return Promise.resolve({ entities: [] });
          }
          return Promise.resolve({
            entities: mockWorkflowVersions,
          });
        });
        const baseId = mockWorkflowVersions[2].id;
        const initialParameters: WorkflowInitialLaunchParameters = {
          workflowId: { ...baseId, version: 'nonexistentValue' },
        };
        const { getByLabelText } = renderForm({ initialParameters });
        await waitFor(() => {});
        expect(getByLabelText(formStrings.workflowVersion)).toHaveValue(
          mockWorkflowVersions[0].id.version,
        );
      });

      it('should prefer the provided launch plan', async () => {
        const initialParameters: WorkflowInitialLaunchParameters = {
          launchPlan: mockLaunchPlans[1].id,
        };
        const { getByLabelText } = renderForm({ initialParameters });
        await waitFor(() => {});
        expect(getByLabelText(formStrings.launchPlan)).toHaveValue(mockLaunchPlans[1].id.name);
      });

      it('should only include one instance of the preferred launch plan in the selector', async () => {
        const initialParameters: WorkflowInitialLaunchParameters = {
          launchPlan: mockLaunchPlans[1].id,
        };
        const { getByTitle } = renderForm({ initialParameters });
        await waitFor(() => {});
        // Click the expander for the LaunchPlan, select the second item
        const launchPlanDiv = getByTitle(formStrings.launchPlan);
        const expander = getByRole(launchPlanDiv, 'button');
        fireEvent.click(expander);
        const items = await waitFor(() => getAllByRole(launchPlanDiv, 'menuitem'));

        const expectedName = mockLaunchPlans[1].id.name;
        expect(
          items.filter((item) => item.textContent && item.textContent.includes(expectedName)),
        ).toHaveLength(1);
      });

      it('should fall back to the default launch plan if the preferred is not found', async () => {
        mockListLaunchPlans.mockImplementation((scope: Partial<Identifier>) => {
          // If we get a request for a specific item,
          // simulate not found
          if (scope.version) {
            return Promise.resolve({ entities: [] });
          }
          return Promise.resolve({ entities: mockLaunchPlans });
        });
        const launchPlanId = { ...mockLaunchPlans[1].id };
        launchPlanId.name = 'InvalidLauchPlan';
        const initialParameters: WorkflowInitialLaunchParameters = {
          launchPlan: launchPlanId,
        };
        const { getByLabelText } = renderForm({ initialParameters });
        await waitFor(() => {});
        expect(getByLabelText(formStrings.launchPlan)).toHaveValue(mockLaunchPlans[0].id.name);
      });

      it('should maintain selected launch plan by name after switching workflow versions', async () => {
        const { getByLabelText, getByTitle } = renderForm();
        await waitFor(() => {});

        // Click the expander for the launch plan, select the second item
        const launchPlanDiv = getByTitle(formStrings.launchPlan);
        const launchPlanExpander = getByRole(launchPlanDiv, 'button');
        fireEvent.click(launchPlanExpander);
        const launchPlanItems = await waitFor(() => getAllByRole(launchPlanDiv, 'menuitem'));
        fireEvent.click(launchPlanItems[1]);
        await waitFor(() => {});

        // Click the expander for the workflow, select the second item
        const workflowDiv = getByTitle(formStrings.workflowVersion);
        const expander = getByRole(workflowDiv, 'button');
        fireEvent.click(expander);
        const items = await waitFor(() => getAllByRole(workflowDiv, 'menuitem'));
        fireEvent.click(items[1]);

        await waitFor(() => {});
        expect(getByLabelText(formStrings.launchPlan)).toHaveValue(mockLaunchPlans[1].id.name);
      });

      it('should prepopulate inputs with provided initial values', async () => {
        const stringValue = 'initialStringValue';
        const initialStringValue: Core.ILiteral = {
          scalar: { primitive: { stringValue } },
        };
        const parameters = mockLaunchPlans[0].closure!.expectedInputs.parameters;
        const values = new Map();
        const stringCacheKey = createInputCacheKey(
          stringInputName,
          getInputDefintionForLiteralType(parameters[stringInputName].var.type),
        );
        values.set(stringCacheKey, initialStringValue);
        const { getByLabelText } = renderForm({
          initialParameters: { values },
        });
        await waitFor(() => {});

        expect(getByLabelText(stringInputName, { exact: false })).toHaveValue(stringValue);
      });

      it('loads preferred workflow version when it does not exist in the list of suggestions', async () => {
        const missingWorkflow = mockWorkflowVersions[0];
        missingWorkflow.id.version = 'missingVersionString';
        const initialParameters: WorkflowInitialLaunchParameters = {
          workflowId: missingWorkflow.id,
        };
        const { getByLabelText } = renderForm({ initialParameters });
        await waitFor(() => {});
        expect(getByLabelText(formStrings.workflowVersion)).toHaveValue(missingWorkflow.id.version);
      });

      it('loads the preferred launch plan when it does not exist in the list of suggestions', async () => {
        const missingLaunchPlan = mockLaunchPlans[0];
        missingLaunchPlan.id.name = 'missingLaunchPlanName';
        const initialParameters: WorkflowInitialLaunchParameters = {
          launchPlan: missingLaunchPlan.id,
        };
        const { getByLabelText } = renderForm({ initialParameters });
        await waitFor(() => {});
        expect(getByLabelText(formStrings.launchPlan)).toHaveValue(missingLaunchPlan.id.name);
      });

      it('should select contents of workflow version input on focus', async () => {
        const { getByLabelText } = renderForm();
        await waitFor(() => {});

        // Focus the workflow version input
        const workflowInput = getByLabelText(formStrings.workflowVersion);
        fireEvent.focus(workflowInput);

        act(() => {
          jest.runAllTimers();
        });

        const expectedValue = mockWorkflowVersions[0].id.version;

        // The value should remain, but selection should be the entire string
        expect(workflowInput).toHaveValue(expectedValue);
        expect((workflowInput as HTMLInputElement).selectionEnd).toBe(expectedValue.length);
      });

      it('should correctly render workflow version search results', async () => {
        const initialParameters: WorkflowInitialLaunchParameters = {
          workflowId: mockWorkflowVersions[2].id,
        };
        const inputString = mockWorkflowVersions[1].id.version.substring(0, 4);
        const { getByLabelText } = renderForm({ initialParameters });
        await waitFor(() => {});

        mockListWorkflows.mockClear();

        const versionInput = getByLabelText(formStrings.workflowVersion);
        fireEvent.change(versionInput, {
          target: { value: inputString },
        });

        act(() => {
          jest.runAllTimers();
        });
        await waitFor(() => {});
        const { project, domain, name } = mockWorkflowVersions[2].id;
        expect(mockListWorkflows).toHaveBeenCalledWith(
          { project, domain, name },
          expect.anything(),
        );
      });
    });

    describe('With Unsupported Required Inputs', () => {
      beforeEach(() => {
        // Binary is currently unsupported, setting the binary input to
        // required and removing the default value will trigger our use case
        const parameters = mockLaunchPlans[0].closure!.expectedInputs.parameters;
        parameters[binaryInputName].required = true;
        delete parameters[binaryInputName].default;
      });

      it('should render error message', async () => {
        const { getByText } = renderForm();
        const errorElement = await waitFor(() => getByText(cannotLaunchWorkflowString));
        expect(errorElement).toBeInTheDocument();
      });

      it('should show unsupported inputs', async () => {
        const { getByText } = renderForm();
        const inputElement = await waitFor(() => getByText(binaryInputName, { exact: false }));
        expect(inputElement).toBeInTheDocument();
      });

      it('should print input labels without decoration', async () => {
        const { getByText } = renderForm();
        const inputElement = await waitFor(() => getByText(binaryInputName, { exact: false }));
        expect(inputElement.textContent).not.toContain(requiredInputSuffix);
      });

      it('should disable submission', async () => {
        const { getByRole } = renderForm();

        const submitButton = await waitFor(() => getByRole('button', { name: formStrings.submit }));

        expect(submitButton).toBeDisabled();
      });

      it('should not show error if launch plan has default value', async () => {
        mockLaunchPlans[0].closure!.expectedInputs.parameters[binaryInputName].default =
          simpleVariableDefaults.simpleBinary as Literal;
        const { queryByText } = renderForm();
        await waitFor(() => queryByText(binaryInputName, { exact: false }));
        expect(queryByText(cannotLaunchWorkflowString)).toBeNull();
      });

      it('should not show error if initial value is provided', async () => {
        const parameters = mockLaunchPlans[0].closure!.expectedInputs.parameters;
        const values = new Map();
        const cacheKey = createInputCacheKey(
          binaryInputName,
          getInputDefintionForLiteralType(parameters[binaryInputName].var.type),
        );
        values.set(cacheKey, simpleVariableDefaults.simpleBinary);
        const { queryByText } = renderForm({
          initialParameters: { values },
        });

        await waitFor(() => queryByText(binaryInputName, { exact: false }));
        expect(queryByText(cannotLaunchWorkflowString)).toBeNull();
      });
    });

    describe('Interruptible', () => {
      it('should render checkbox', async () => {
        const { getByLabelText } = renderForm();
        const inputElement = await waitFor(() =>
          getByLabelText(formStrings.interruptible, { exact: false }),
        );
        expect(inputElement).toBeInTheDocument();
        expect(inputElement).not.toBeChecked();
      });

      it('should use initial values when provided', async () => {
        const initialParameters: WorkflowInitialLaunchParameters = {
          workflowId: mockWorkflowVersions[2].id,
          interruptible: Protobuf.BoolValue.create({ value: true }),
        };

        const { getByLabelText } = renderForm({
          initialParameters,
        });

        const inputElement = await waitFor(() =>
          getByLabelText(formStrings.interruptible, { exact: false }),
        );
        expect(inputElement).toBeInTheDocument();
        expect(inputElement).toBeChecked();
      });

      it('should cycle between states correctly when clicked', async () => {
        const { getByLabelText } = renderForm();

        let inputElement = await waitFor(() =>
          getByLabelText(`${formStrings.interruptible} (no override)`, { exact: true }),
        );
        expect(inputElement).toBeInTheDocument();
        expect(inputElement).not.toBeChecked();
        expect(inputElement).toHaveAttribute('data-indeterminate', 'true');

        fireEvent.click(inputElement);
        inputElement = await waitFor(() =>
          getByLabelText(`${formStrings.interruptible} (enabled)`, { exact: true }),
        );
        expect(inputElement).toBeInTheDocument();
        expect(inputElement).toBeChecked();
        expect(inputElement).toHaveAttribute('data-indeterminate', 'false');

        fireEvent.click(inputElement);
        inputElement = await waitFor(() =>
          getByLabelText(`${formStrings.interruptible} (disabled)`, { exact: true }),
        );
        expect(inputElement).toBeInTheDocument();
        expect(inputElement).not.toBeChecked();
        expect(inputElement).toHaveAttribute('data-indeterminate', 'false');

        fireEvent.click(inputElement);
        inputElement = await waitFor(() =>
          getByLabelText(`${formStrings.interruptible} (no override)`, { exact: true }),
        );
        expect(inputElement).toBeInTheDocument();
        expect(inputElement).not.toBeChecked();
        expect(inputElement).toHaveAttribute('data-indeterminate', 'true');
      });

      it('should submit without interruptible override set', async () => {
        const { container, getByLabelText } = renderForm();

        const inputElement = await waitFor(() =>
          getByLabelText(formStrings.interruptible, { exact: false }),
        );
        expect(inputElement).toBeInTheDocument();
        expect(inputElement).not.toBeChecked();
        expect(inputElement).toHaveAttribute('data-indeterminate', 'true');

        const integerInput = getByLabelText(integerInputName, {
          exact: false,
        });
        fireEvent.change(integerInput, { target: { value: '123' } });
        await waitFor(() => expect(integerInput).toBeValid());
        fireEvent.click(getSubmitButton(container));

        await waitFor(() =>
          expect(mockCreateWorkflowExecution).toHaveBeenCalledWith(
            expect.objectContaining({
              interruptible: null,
            }),
          ),
        );
      });

      it('should submit with interruptible override enabled', async () => {
        const initialParameters: WorkflowInitialLaunchParameters = {
          interruptible: Protobuf.BoolValue.create({ value: true }),
        };
        const { container, getByLabelText } = renderForm({ initialParameters });

        const inputElement = await waitFor(() =>
          getByLabelText(formStrings.interruptible, { exact: false }),
        );
        expect(inputElement).toBeInTheDocument();
        expect(inputElement).toBeChecked();
        expect(inputElement).toHaveAttribute('data-indeterminate', 'false');

        const integerInput = getByLabelText(integerInputName, {
          exact: false,
        });
        fireEvent.change(integerInput, { target: { value: '123' } });
        fireEvent.click(getSubmitButton(container));

        await waitFor(() =>
          expect(mockCreateWorkflowExecution).toHaveBeenCalledWith(
            expect.objectContaining({
              interruptible: Protobuf.BoolValue.create({ value: true }),
            }),
          ),
        );
      });

      it('should submit with interruptible override disabled', async () => {
        const initialParameters: WorkflowInitialLaunchParameters = {
          interruptible: Protobuf.BoolValue.create({ value: false }),
        };
        const { container, getByLabelText } = renderForm({ initialParameters });

        const inputElement = await waitFor(() =>
          getByLabelText(formStrings.interruptible, { exact: false }),
        );
        expect(inputElement).toBeInTheDocument();
        expect(inputElement).not.toBeChecked();
        expect(inputElement).toHaveAttribute('data-indeterminate', 'false');

        const integerInput = getByLabelText(integerInputName, {
          exact: false,
        });
        fireEvent.change(integerInput, { target: { value: '123' } });
        fireEvent.click(getSubmitButton(container));

        await waitFor(() =>
          expect(mockCreateWorkflowExecution).toHaveBeenCalledWith(
            expect.objectContaining({
              interruptible: Protobuf.BoolValue.create({ value: false }),
            }),
          ),
        );
      });
    });
  });
});
