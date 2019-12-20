import { sortedObjectEntries } from 'common/utils';
import { getCacheKey } from 'components/Cache';
import { useAPIContext } from 'components/data/apiContext';
import {
    useFetchableData,
    useWorkflow,
    useWorkflows,
    waitForAllFetchables
} from 'components/hooks';
import { Core } from 'flyteidl';
import {
    FilterOperationName,
    LaunchPlan,
    SortDirection,
    Workflow,
    WorkflowExecutionIdentifier,
    WorkflowId,
    workflowSortFields
} from 'models';
import { useEffect, useMemo, useRef, useState } from 'react';
import { history, Routes } from 'routes';
import { SearchableSelectorOption } from './SearchableSelector';
import {
    LaunchWorkflowFormInputsRef,
    LaunchWorkflowFormProps,
    LaunchWorkflowFormState,
    ParsedInput
} from './types';
import {
    formatLabelWithType,
    getInputDefintionForLiteralType,
    getWorkflowInputs,
    launchPlansToSearchableSelectorOptions,
    workflowsToSearchableSelectorOptions
} from './utils';

// We use a non-empty string for the description to allow display components
// to depend on the existence of a value
const emptyDescription = ' ';

function getInputs(workflow: Workflow, launchPlan: LaunchPlan): ParsedInput[] {
    if (!launchPlan.closure || !workflow) {
        // TODO: is this an error?
        return [];
    }

    const workflowInputs = getWorkflowInputs(workflow);
    const launchPlanInputs = launchPlan.closure.expectedInputs.parameters;
    return sortedObjectEntries(launchPlanInputs).map(value => {
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

export function useWorkflowSelectorOptions(workflows: Workflow[]) {
    return useMemo(() => {
        const options = workflowsToSearchableSelectorOptions(workflows);
        if (options.length > 0) {
            options[0].description = 'latest';
        }
        return options;
    }, [workflows]);
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
    const formInputsRef = useRef<LaunchWorkflowFormInputsRef>(null);
    const [showErrors, setShowErrors] = useState(false);
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

    // Any time the inputs change (even if it's just re-ordering), we must
    // change the form key so that the inputs component will re-mount.
    const formKey = useMemo<string>(() => {
        if (!selectedWorkflowId || !selectedLaunchPlan) {
            return '';
        }
        return getCacheKey(parsedInputs);
    }, [parsedInputs]);

    // Only show errors after first submission for a set of inputs.
    useEffect(() => setShowErrors(false), [formKey]);

    const workflowName = workflowId.name;

    const onSelectWorkflow = (
        newWorkflow: SearchableSelectorOption<WorkflowId>
    ) => {
        setLaunchPlan(undefined);
        setWorkflow(newWorkflow);
    };

    const launchWorkflow = async (
        inputValues: Record<string, Core.ILiteral>
    ) => {
        if (!launchPlanData) {
            throw new Error('Attempting to launch with no LaunchPlan');
        }
        const launchPlanId = launchPlanData.id;
        const { domain, project } = workflowId;

        const response = await createWorkflowExecution({
            domain,
            launchPlanId,
            project,
            inputs: { literals: inputValues }
        });
        const newExecutionId = response.id as WorkflowExecutionIdentifier;
        if (!newExecutionId) {
            throw new Error('API Response did not include new execution id');
        }
        history.push(Routes.ExecutionDetails.makeUrl(newExecutionId));
        return newExecutionId;
    };

    const submissionState = useFetchableData<
        WorkflowExecutionIdentifier,
        string
    >(
        {
            autoFetch: false,
            debugName: 'LaunchWorkflowForm',
            defaultValue: {} as WorkflowExecutionIdentifier,
            doFetch: () => {
                if (formInputsRef.current === null) {
                    throw new Error('Unexpected empty form inputs ref');
                }
                return launchWorkflow(formInputsRef.current.getValues());
            }
        },
        formKey
    );

    const onSubmit = () => {
        if (formInputsRef.current === null) {
            console.error('Unexpected empty form inputs ref');
            return;
        }

        // Show errors after the first submission
        setShowErrors(true);
        // We validate separately so that a request isn't triggered unless
        // the inputs are valid.
        if (!formInputsRef.current.validate()) {
            return;
        }
        submissionState.fetch();
    };
    const onCancel = onClose;

    useEffect(() => {
        const parsedInputs =
            launchPlanData && workflow.hasLoaded
                ? getInputs(workflow.value, launchPlanData)
                : [];
        setParsedInputs(parsedInputs);
    }, [workflow.hasLoaded, workflow.value, launchPlanData]);

    // Once workflows have loaded, attempt to select the first option
    useEffect(() => {
        if (workflowSelectorOptions.length > 0 && !selectedWorkflow) {
            setWorkflow(workflowSelectorOptions[0]);
        }
    }, [workflows.value]);

    // Once launch plans have been loaded, attempt to select the default
    // launch plan
    useEffect(() => {
        if (!launchPlanSelectorOptions.length) {
            return;
        }
        const defaultLaunchPlan = launchPlanSelectorOptions.find(
            ({ id }) => id === workflowId.name
        );
        setLaunchPlan(defaultLaunchPlan);
    }, [launchPlanSelectorOptions]);

    return {
        formInputsRef,
        formKey,
        inputLoadingState,
        launchPlanOptionsLoadingState,
        launchPlanSelectorOptions,
        onCancel,
        onSelectWorkflow,
        onSubmit,
        selectedLaunchPlan,
        selectedWorkflow,
        showErrors,
        submissionState,
        workflowName,
        workflowOptionsLoadingState,
        workflowSelectorOptions,
        inputs: parsedInputs,
        onSelectLaunchPlan: setLaunchPlan
    };
}
