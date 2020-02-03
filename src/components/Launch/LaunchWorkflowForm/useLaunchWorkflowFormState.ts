import { log } from 'common/log';
import { getCacheKey } from 'components/Cache';
import { useAPIContext } from 'components/data/apiContext';
import {
    useFetchableData,
    useWorkflow,
    useWorkflows,
    waitForAllFetchables
} from 'components/hooks';
import { uniqBy } from 'lodash';
import {
    FilterOperationName,
    Identifier,
    LaunchPlan,
    SortDirection,
    Workflow,
    WorkflowExecutionIdentifier,
    WorkflowId,
    workflowSortFields
} from 'models';
import { useEffect, useMemo, useRef, useState } from 'react';
import { history, Routes } from 'routes';
import { getInputs } from './getInputs';
import { createInputValueCache } from './inputValueCache';
import { SearchableSelectorOption } from './SearchableSelector';
import {
    LaunchWorkflowFormInputsRef,
    LaunchWorkflowFormProps,
    LaunchWorkflowFormState,
    ParsedInput
} from './types';
import {
    launchPlansToSearchableSelectorOptions,
    workflowsToSearchableSelectorOptions
} from './utils';

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

interface UseLaunchPlansForWorkflowArgs {
    workflowId?: WorkflowId | null;
    preferredLaunchPlanId?: Identifier;
}
function useLaunchPlansForWorkflow({
    workflowId = null,
    preferredLaunchPlanId
}: UseLaunchPlansForWorkflowArgs) {
    const { getLaunchPlan, listLaunchPlans } = useAPIContext();
    return useFetchableData<LaunchPlan[], WorkflowId | null>(
        {
            autoFetch: workflowId !== null,
            debugName: 'useLaunchPlansForWorkflow',
            defaultValue: [],
            doFetch: async workflowId => {
                if (workflowId === null) {
                    return Promise.reject('No workflowId specified');
                }

                // If we fail to retrieve the preferred launch plan,
                // we still want to return the available options. So
                // we will catch and return undefined.
                const preferredLaunchPlanPromise = preferredLaunchPlanId
                    ? getLaunchPlan(preferredLaunchPlanId).catch(error =>
                          log.error(error)
                      )
                    : Promise.resolve(undefined);

                const { project, domain, name, version } = workflowId;
                const launchPlansPromise = listLaunchPlans(
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

                const [{ entities }, preferredLaunchPlan] = await Promise.all([
                    launchPlansPromise,
                    preferredLaunchPlanPromise
                ]);
                if (preferredLaunchPlan) {
                    entities.unshift(preferredLaunchPlan);
                }
                // The preferred launch plan may have existed in
                // the list we fetched. Make sure it isn't listed twice.
                return uniqBy(entities, ({ id: name }) => name);
            }
        },
        workflowId
    );
}

/** If preferred workflow id exists, will ensure the passed workflows array contains
 * it. Returns a new array if merging the id is required.
 */
function mergePreferredWorkflowId(
    workflows: Workflow[],
    preferredWorkflowId?: WorkflowId
): Workflow[] {
    if (!preferredWorkflowId) {
        return workflows;
    }
    if (
        workflows.find(
            ({ id: { version } }) => version === preferredWorkflowId.version
        ) === undefined
    ) {
        return [...workflows, { id: preferredWorkflowId }];
    }
    return workflows;
}

/** Contains all of the form state for a LaunchWorkflowForm, including input
 * definitions, current input values, and errors.
 */
export function useLaunchWorkflowFormState({
    initialParameters = {},
    onClose,
    workflowId
}: LaunchWorkflowFormProps): LaunchWorkflowFormState {
    // These values will be used to auto-select items from the workflow
    // version/launch plan drop downs.
    const {
        workflow: preferredWorkflowId,
        launchPlan: preferredLaunchPlanId
    } = initialParameters;

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
    // If we have a preferred workflow version, we want to merge
    // that with the options fetched for our workflow name
    const workflowSelectorOptions = useWorkflowSelectorOptions(
        mergePreferredWorkflowId(workflows.value, preferredWorkflowId)
    );
    const [selectedWorkflow, setWorkflow] = useState<
        SearchableSelectorOption<WorkflowId>
    >();
    const selectedWorkflowId = selectedWorkflow ? selectedWorkflow.data : null;

    const inputValueCache = useMemo(
        () => createInputValueCache(initialParameters.values),
        [initialParameters.values]
    );

    // We have to do a single item get once a workflow is selected so that we
    // receive the full workflow spec
    const workflow = useWorkflow(selectedWorkflowId);
    const launchPlans = useLaunchPlansForWorkflow({
        preferredLaunchPlanId,
        workflowId: selectedWorkflowId
    });
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

    const launchWorkflow = async () => {
        if (!launchPlanData) {
            throw new Error('Attempting to launch with no LaunchPlan');
        }
        if (formInputsRef.current === null) {
            throw new Error('Unexpected empty form inputs ref');
        }
        const literals = formInputsRef.current.getValues();
        const launchPlanId = launchPlanData.id;
        const { domain, project } = workflowId;

        const response = await createWorkflowExecution({
            domain,
            launchPlanId,
            project,
            inputs: { literals }
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
            doFetch: launchWorkflow
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

    // Once workflows have loaded, attempt to select the preferred workflow
    // plan, or fall back to selecting the first option
    useEffect(() => {
        if (workflowSelectorOptions.length > 0 && !selectedWorkflow) {
            setWorkflow(workflowSelectorOptions[0]);
        }
    }, [workflows.value]);

    // Once launch plans have been loaded, attempt to select the preferred
    // launch plan, the one matching the workflow name, or just the first
    // option.
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
        inputValueCache,
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
