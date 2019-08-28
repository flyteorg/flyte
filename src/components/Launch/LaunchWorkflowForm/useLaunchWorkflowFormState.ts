import { useAPIContext } from 'components/data/apiContext';
import { useFetchableData, useWorkflows } from 'components/hooks';
import {
    FilterOperationName,
    LaunchPlan,
    LiteralType,
    Workflow,
    WorkflowExecutionIdentifier
} from 'models';
import { useEffect, useMemo, useState } from 'react';
import { useDefaultLaunchPlan } from '../useDefaultLaunchPlan';
import { simpleTypeToInputType } from './constants';
import {
    InputProps,
    InputType,
    InputTypeDefinition,
    InputValue,
    LaunchWorkflowFormProps,
    LaunchWorkflowFormState
} from './types';
import {
    formatLabelWithType,
    getWorkflowInputs,
    workflowsToWorkflowSelectorOptions
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

function useWorkflowSelectorOptions(workflows: Workflow[]) {
    return useMemo(() => workflowsToWorkflowSelectorOptions(workflows), [
        workflows
    ]);
}

function useLaunchPlansForWorkflow(workflow?: Workflow) {
    const { listLaunchPlans } = useAPIContext();
    return useFetchableData<LaunchPlan[], Workflow | undefined>({
        autoFetch: !!workflow,
        debugName: 'useLaunchPlansForWorkflow',
        defaultValue: [],
        doFetch: async workflow => {
            if (!workflow) return Promise.reject('No workflow specified');
            const { project, domain, name, version } = workflow.id;
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
    });
}

/** Contains all of the form state for a LaunchWorkflowForm, including input
 * definitions, current input values, and errors.
 */
export function useLaunchWorkflowFormState({
    workflowId
}: LaunchWorkflowFormProps): LaunchWorkflowFormState {
    const workflows = useWorkflows(workflowId, { limit: 10 });
    const workflowSelectorOptions = useWorkflowSelectorOptions(workflows.value);
    const [selectedWorkflow, setWorkflow] = useState<Workflow>();
    const defaultLaunchPlan = useDefaultLaunchPlan(selectedWorkflow);

    const launchPlans = useLaunchPlansForWorkflow(selectedWorkflow);
    const [selectedLaunchPlan, setLaunchPlan] = useState<
        LaunchPlan | undefined
    >();

    const [parsedInputs, setParsedInputs] = useState<ParsedInput[]>([]);
    const { inputs } = useFormInputsState(parsedInputs);
    const workflowName = workflowId.name;

    const launchWorkflow = () => {
        console.log('launch', inputs);
        return new Promise<WorkflowExecutionIdentifier>((resolve, reject) => {
            setTimeout(() => reject('Launching is not implemented'), 1500);
        });
    };

    const submissionState = useFetchableData<WorkflowExecutionIdentifier>({
        autoFetch: false,
        debugName: 'LaunchWorkflowForm',
        defaultValue: {} as WorkflowExecutionIdentifier,
        doFetch: launchWorkflow
    });

    const onSubmit = submissionState.fetch;
    const onCancel = () => {
        console.log('cancel');
    };

    useEffect(
        () => {
            const parsedInputs =
                selectedLaunchPlan && selectedWorkflow
                    ? getInputs(selectedWorkflow, selectedLaunchPlan)
                    : [];
            setParsedInputs(parsedInputs);
        },
        [selectedWorkflow, selectedLaunchPlan]
    );

    return {
        defaultLaunchPlan,
        inputs,
        launchPlans,
        onCancel,
        onSubmit,
        selectedLaunchPlan,
        selectedWorkflow,
        setLaunchPlan,
        setWorkflow,
        submissionState,
        workflowName,
        workflowSelectorOptions
    };
}
