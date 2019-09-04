import { useAPIContext } from 'components/data/apiContext';
import {
    useFetchableData,
    useWorkflow,
    useWorkflows,
    waitForAllFetchables
} from 'components/hooks';
import {
    FilterOperationName,
    LaunchPlan,
    LiteralType,
    SortDirection,
    Workflow,
    WorkflowExecutionIdentifier,
    WorkflowId,
    workflowSortFields
} from 'models';
import { useEffect, useMemo, useState } from 'react';
import { history, Routes } from 'routes';
import { simpleTypeToInputType } from './constants';
import { SearchableSelectorOption } from './SearchableSelector';
import {
    InputProps,
    InputType,
    InputTypeDefinition,
    InputValue,
    LaunchWorkflowFormProps,
    LaunchWorkflowFormState
} from './types';
import {
    convertFormInputsToLiteralMap,
    formatLabelWithType,
    getWorkflowInputs,
    launchPlansToSearchableSelectorOptions,
    workflowsToSearchableSelectorOptions
} from './utils';

// We use a non-empty string for the description to allow display components
// to depend on the existence of a value
const emptyDescription = ' ';

function getInputDefintionForLiteralType(
    literalType: LiteralType
): InputTypeDefinition {
    if (literalType.blob) {
        return { type: InputType.Blob };
    }

    if (literalType.collectionType) {
        return {
            type: InputType.Collection,
            subtype: getInputDefintionForLiteralType(literalType.collectionType)
        };
    }

    if (literalType.mapValueType) {
        return {
            type: InputType.Map,
            subtype: getInputDefintionForLiteralType(literalType.mapValueType)
        };
    }

    if (literalType.schema) {
        return { type: InputType.Schema };
    }

    if (literalType.simple) {
        return { type: simpleTypeToInputType[literalType.simple] };
    }

    return { type: InputType.Unknown };
}

type ParsedInput = Pick<
    InputProps,
    'description' | 'label' | 'name' | 'required' | 'typeDefinition'
>;

function getInputs(workflow: Workflow, launchPlan: LaunchPlan): ParsedInput[] {
    if (!launchPlan.closure || !workflow) {
        // TODO: is this an error?
        return [];
    }

    const workflowInputs = getWorkflowInputs(workflow);
    const launchPlanInputs = launchPlan.closure.expectedInputs.parameters;
    return Object.entries(launchPlanInputs).map(value => {
        const [name, parameter] = value;
        const required = !!(parameter.default || parameter.required);
        const workflowInput = workflowInputs[name];
        const description =
            workflowInput && workflowInput.description
                ? workflowInput.description
                : emptyDescription;

        const typeDefinition = getInputDefintionForLiteralType(
            parameter.var.type
        );
        const label = formatLabelWithType(name, typeDefinition);

        // TODO:
        // Extract default value for more specific type (maybe just for simple)
        return {
            description,
            label,
            name,
            required,
            typeDefinition
        };
    });
}

interface FormInputsState {
    inputs: InputProps[];
}

// TODO: This could be made generic and composed with ParsedInput
function useFormInputsState(parsedInputs: ParsedInput[]): FormInputsState {
    const [values, setValues] = useState<Record<string, InputValue>>({});
    const [errors, setErrors] = useState<Record<string, string>>({});

    const inputs = parsedInputs.map<InputProps>(parsed => ({
        ...parsed,
        value: values[parsed.name],
        helperText: errors[parsed.name]
            ? errors[parsed.name]
            : parsed.description,
        onChange: (value: InputValue) =>
            setValues({ ...values, [parsed.name]: value })
    }));

    useEffect(
        () => {
            // TODO: Use default values from inputs
            setValues({});
        },
        [parsedInputs]
    );

    return {
        inputs
    };
}

export function useWorkflowSelectorOptions(workflows: Workflow[]) {
    return useMemo(
        () => {
            const options = workflowsToSearchableSelectorOptions(workflows);
            if (options.length > 0) {
                options[0].description = 'latest';
            }
            return options;
        },
        [workflows]
    );
}

function useLaunchPlanSelectorOptions(launchPlans: LaunchPlan[]) {
    return useMemo(() => launchPlansToSearchableSelectorOptions(launchPlans), [
        launchPlans
    ]);
}

function useLaunchPlansForWorkflow(workflowId: WorkflowId | null = null) {
    const { listLaunchPlans } = useAPIContext();
    return useFetchableData<LaunchPlan[], WorkflowId | null>(
        {
            autoFetch: workflowId !== null,
            debugName: 'useLaunchPlansForWorkflow',
            defaultValue: [],
            doFetch: async workflowId => {
                if (workflowId === null) {
                    return Promise.reject('No workflowId specified');
                }
                const { project, domain, name, version } = workflowId;
                const { entities } = await listLaunchPlans(
                    { project, domain },
                    // TODO: Only active?
                    {
                        filter: [
                            {
                                key: 'workflow.name',
                                operation: FilterOperationName.EQ,
                                value: name
                            },
                            {
                                key: 'workflow.version',
                                operation: FilterOperationName.EQ,
                                value: version
                            }
                        ],
                        limit: 10
                    }
                );
                return entities;
            }
        },
        workflowId
    );
}

/** Contains all of the form state for a LaunchWorkflowForm, including input
 * definitions, current input values, and errors.
 */
export function useLaunchWorkflowFormState({
    onClose,
    workflowId
}: LaunchWorkflowFormProps): LaunchWorkflowFormState {
    const { createWorkflowExecution } = useAPIContext();
    const workflows = useWorkflows(workflowId, {
        limit: 10,
        sort: {
            key: workflowSortFields.createdAt,
            direction: SortDirection.DESCENDING
        }
    });
    const workflowSelectorOptions = useWorkflowSelectorOptions(workflows.value);
    const [selectedWorkflow, setWorkflow] = useState<
        SearchableSelectorOption<WorkflowId>
    >();
    const selectedWorkflowId = selectedWorkflow ? selectedWorkflow.data : null;

    // We have to do a single item get once a workflow is selected so that we
    // receive the full workflow spec
    const workflow = useWorkflow(selectedWorkflowId);
    const launchPlans = useLaunchPlansForWorkflow(selectedWorkflowId);
    const launchPlanSelectorOptions = useLaunchPlanSelectorOptions(
        launchPlans.value
    );

    const [selectedLaunchPlan, setLaunchPlan] = useState<
        SearchableSelectorOption<LaunchPlan>
    >();
    const launchPlanData = selectedLaunchPlan
        ? selectedLaunchPlan.data
        : undefined;

    const workflowOptionsLoadingState = waitForAllFetchables([workflows]);
    const launchPlanOptionsLoadingState = waitForAllFetchables([launchPlans]);

    const inputLoadingState = waitForAllFetchables([workflow, launchPlans]);

    const [parsedInputs, setParsedInputs] = useState<ParsedInput[]>([]);
    const { inputs } = useFormInputsState(parsedInputs);
    const workflowName = workflowId.name;

    const onSelectWorkflow = (
        newWorkflow: SearchableSelectorOption<WorkflowId>
    ) => {
        setLaunchPlan(undefined);
        setWorkflow(newWorkflow);
    };

    const launchWorkflow = async () => {
        if (!launchPlanData) {
            throw new Error('Attempting to launch with no LaunchPlan');
        }
        const launchPlanId = launchPlanData.id;
        const { domain, project } = workflowId;
        const response = await createWorkflowExecution({
            domain,
            launchPlanId,
            project,
            inputs: convertFormInputsToLiteralMap(inputs)
        });
        const newExecutionId = response.id as WorkflowExecutionIdentifier;
        if (!newExecutionId) {
            throw new Error('API Response did not include new execution id');
        }
        history.push(Routes.ExecutionDetails.makeUrl(newExecutionId));
        return newExecutionId;
    };

    const submissionState = useFetchableData<WorkflowExecutionIdentifier>({
        autoFetch: false,
        debugName: 'LaunchWorkflowForm',
        defaultValue: {} as WorkflowExecutionIdentifier,
        doFetch: launchWorkflow
    });

    const onSubmit = submissionState.fetch;
    const onCancel = onClose;

    useEffect(
        () => {
            const parsedInputs =
                launchPlanData && workflow.hasLoaded
                    ? getInputs(workflow.value, launchPlanData)
                    : [];
            setParsedInputs(parsedInputs);
        },
        [workflow.hasLoaded, workflow.value, launchPlanData]
    );

    // Once workflows have loaded, attempt to select the first option
    useEffect(
        () => {
            if (workflowSelectorOptions.length > 0 && !selectedWorkflow) {
                setWorkflow(workflowSelectorOptions[0]);
            }
        },
        [workflows.value]
    );

    // Once launch plans have been loaded, attempt to select the default
    // launch plan
    useEffect(
        () => {
            if (!launchPlanSelectorOptions.length) {
                return;
            }
            const defaultLaunchPlan = launchPlanSelectorOptions.find(
                ({ id }) => id === workflowId.name
            );
            setLaunchPlan(defaultLaunchPlan);
        },
        [launchPlanSelectorOptions]
    );

    return {
        inputLoadingState,
        inputs,
        launchPlanOptionsLoadingState,
        launchPlanSelectorOptions,
        onCancel,
        onSelectWorkflow,
        onSubmit,
        selectedLaunchPlan,
        selectedWorkflow,
        submissionState,
        workflowName,
        workflowOptionsLoadingState,
        workflowSelectorOptions,
        onSelectLaunchPlan: setLaunchPlan
    };
}
