import { getCacheKey } from 'components/Cache';
import { useAPIContext } from 'components/data/apiContext';
import { useFetchableData, useWorkflow } from 'components/hooks';
import { fetchStates } from 'components/hooks/types';
import { isEqual, uniqBy } from 'lodash';
import {
    FilterOperationName,
    Identifier,
    LaunchPlan,
    NamedEntityIdentifier,
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
    getUnsupportedRequiredInputs,
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
/** Lists launch plans for a given workflowId, optionally fetching a preferred
 * launch plan. The result is a merged, de-duplicated list.
 */
function useLaunchPlansForWorkflow({
    workflowId = null,
    preferredLaunchPlanId
}: UseLaunchPlansForWorkflowArgs) {
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

                let preferredLaunchPlanPromise = Promise.resolve({
                    entities: [] as LaunchPlan[]
                });
                if (preferredLaunchPlanId) {
                    const { version, ...scope } = preferredLaunchPlanId;
                    preferredLaunchPlanPromise = listLaunchPlans(scope, {
                        limit: 1,
                        filter: [
                            {
                                key: 'version',
                                operation: FilterOperationName.EQ,
                                value: version
                            }
                        ]
                    });
                }

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

                const [
                    launchPlansResult,
                    preferredLaunchPlanResult
                ] = await Promise.all([
                    launchPlansPromise,
                    preferredLaunchPlanPromise
                ]);
                const merged = [
                    ...launchPlansResult.entities,
                    ...preferredLaunchPlanResult.entities
                ];
                return uniqBy(merged, ({ id }) => id.name);
            }
        },
        workflowId
    );
}

/** Fetches workflow versions matching a specific scope, optionally also
 * fetching a preferred version. The result is a merged, de-duplicated list.
 */
function useWorkflowsWithPreferredVersion(
    workflowName: NamedEntityIdentifier,
    preferredVersion?: WorkflowId
) {
    const { listWorkflows } = useAPIContext();
    return useFetchableData(
        {
            debugName: 'UseWorkflowsWithPreferredVersion',
            defaultValue: [] as Workflow[],
            doFetch: async () => {
                const { project, domain, name } = workflowName;
                const workflowsPromise = listWorkflows(
                    { project, domain, name },
                    {
                        limit: 10,
                        sort: {
                            key: workflowSortFields.createdAt,
                            direction: SortDirection.DESCENDING
                        }
                    }
                );

                let preferredWorkflowPromise = Promise.resolve({
                    entities: [] as Workflow[]
                });
                if (preferredVersion) {
                    const { version, ...scope } = preferredVersion;
                    preferredWorkflowPromise = listWorkflows(scope, {
                        limit: 1,
                        filter: [
                            {
                                key: 'version',
                                operation: FilterOperationName.EQ,
                                value: version
                            }
                        ]
                    });
                }

                const [
                    workflowsResult,
                    preferredWorkflowResult
                ] = await Promise.all([
                    workflowsPromise,
                    preferredWorkflowPromise
                ]);
                const merged = [
                    ...workflowsResult.entities,
                    ...preferredWorkflowResult.entities
                ];
                return uniqBy(merged, ({ id: { version } }) => version);
            }
        },
        { workflowName, preferredVersion }
    );
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
        launchPlan: preferredLaunchPlanId,
        workflow: preferredWorkflowId
    } = initialParameters;

    const { createWorkflowExecution } = useAPIContext();
    const [
        lastSelectedLaunchPlanName,
        setLastSelectedLaunchPlanName
    ] = useState<string>();
    const formInputsRef = useRef<LaunchWorkflowFormInputsRef>(null);
    const [showErrors, setShowErrors] = useState(false);
    const workflows = useWorkflowsWithPreferredVersion(
        workflowId,
        preferredWorkflowId
    );

    const workflowSelectorOptions = useWorkflowSelectorOptions(workflows.value);
    const [selectedWorkflow, setWorkflow] = useState<
        SearchableSelectorOption<WorkflowId>
    >();
    const selectedWorkflowId = selectedWorkflow ? selectedWorkflow.data : null;

    const [inputValueCache] = useState(createInputValueCache());

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

    const [parsedInputs, setParsedInputs] = useState<ParsedInput[]>([]);
    const inputsReady = !!(
        launchPlanData && workflow.state.matches(fetchStates.LOADED)
    );

    const unsupportedRequiredInputs = useMemo(
        () => getUnsupportedRequiredInputs(parsedInputs),
        [parsedInputs]
    );

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
        if (newWorkflow === selectedWorkflow) {
            return;
        }
        setLaunchPlan(undefined);
        setWorkflow(newWorkflow);
    };

    const onSelectLaunchPlan = (
        newLaunchPlan: SearchableSelectorOption<LaunchPlan>
    ) => {
        if (newLaunchPlan === selectedLaunchPlan) {
            return;
        }
        setLastSelectedLaunchPlanName(newLaunchPlan.name);
        setLaunchPlan(newLaunchPlan);
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

    // Once the selected workflow and launch plan have loaded, parse and set
    // the inputs so we can render the rest of the form
    useEffect(() => {
        const parsedInputs =
            launchPlanData && workflow.state.matches(fetchStates.LOADED)
                ? getInputs(
                      workflow.value,
                      launchPlanData,
                      initialParameters.values
                  )
                : [];
        setParsedInputs(parsedInputs);
    }, [workflow.state.value, workflow.value, launchPlanData]);

    // Once workflows have loaded, attempt to select the preferred workflow
    // plan, or fall back to selecting the first option
    useEffect(() => {
        if (workflowSelectorOptions.length > 0 && !selectedWorkflow) {
            if (preferredWorkflowId) {
                const preferred = workflowSelectorOptions.find(({ data }) =>
                    isEqual(data, preferredWorkflowId)
                );
                if (preferred) {
                    setWorkflow(preferred);
                    return;
                }
            }
            setWorkflow(workflowSelectorOptions[0]);
        }
    }, [workflows.value]);

    // Once launch plans have been loaded, attempt to keep the previously
    // selected launch plan, followed by the preferred launch plan, the one
    // matching the workflow name, or just the first option.
    useEffect(() => {
        if (!launchPlanSelectorOptions.length) {
            return;
        }

        if (lastSelectedLaunchPlanName) {
            const lastSelected = launchPlanSelectorOptions.find(
                ({ name }) => name === lastSelectedLaunchPlanName
            );
            if (lastSelected) {
                onSelectLaunchPlan(lastSelected);
                return;
            }
        }

        if (preferredLaunchPlanId) {
            const preferred = launchPlanSelectorOptions.find(
                ({ data: { id } }) => isEqual(id, preferredLaunchPlanId)
            );
            if (preferred) {
                onSelectLaunchPlan(preferred);
                return;
            }
        }

        const defaultLaunchPlan = launchPlanSelectorOptions.find(
            ({ id }) => id === workflowId.name
        );
        if (defaultLaunchPlan) {
            onSelectLaunchPlan(defaultLaunchPlan);
            return;
        }
        onSelectLaunchPlan(launchPlanSelectorOptions[0]);
    }, [launchPlanSelectorOptions]);

    return {
        formInputsRef,
        formKey,
        inputValueCache,
        inputsReady,
        launchPlans,
        launchPlanSelectorOptions,
        onCancel,
        onSelectLaunchPlan,
        onSelectWorkflow,
        onSubmit,
        selectedLaunchPlan,
        selectedWorkflow,
        showErrors,
        submissionState,
        unsupportedRequiredInputs,
        workflowName,
        workflows,
        workflowSelectorOptions,
        inputs: parsedInputs
    };
}
