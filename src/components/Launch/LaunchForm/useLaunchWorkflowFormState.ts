import { useMachine } from '@xstate/react';
import { defaultStateMachineConfig } from 'components/common/constants';
import { APIContextValue, useAPIContext } from 'components/data/apiContext';
import { isEqual, partial, uniqBy } from 'lodash';
import { FilterOperationName, SortDirection } from 'models/AdminEntity/types';
import { WorkflowExecutionIdentifier } from 'models/Execution/types';
import { LaunchPlan } from 'models/Launch/types';
import { workflowSortFields } from 'models/Workflow/constants';
import { Workflow, WorkflowId } from 'models/Workflow/types';
import { RefObject, useEffect, useMemo, useRef } from 'react';
import { getInputsForWorkflow } from './getInputs';
import {
    LaunchState,
    WorkflowLaunchContext,
    WorkflowLaunchEvent,
    workflowLaunchMachine,
    WorkflowLaunchTypestate
} from './launchMachine';
import { validate as baseValidate } from './services';
import {
    LaunchFormInputsRef,
    LaunchRoleInputRef,
    LaunchAdvancedOptionsRef,
    LaunchWorkflowFormProps,
    LaunchWorkflowFormState,
    ParsedInput
} from './types';
import { useWorkflowSourceSelectorState } from './useWorkflowSourceSelectorState';
import { getUnsupportedRequiredInputs } from './utils';
import { correctInputErrors } from './constants';

async function loadLaunchPlans(
    { listLaunchPlans }: APIContextValue,
    { preferredLaunchPlanId, workflowVersion }: WorkflowLaunchContext
) {
    if (workflowVersion == null) {
        return Promise.reject('No workflowVersion specified');
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

    const { project, domain, name, version } = workflowVersion;
    const launchPlansPromise = listLaunchPlans(
        { project, domain },
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

    const [launchPlansResult, preferredLaunchPlanResult] = await Promise.all([
        launchPlansPromise,
        preferredLaunchPlanPromise
    ]);
    const merged = [
        ...launchPlansResult.entities,
        ...preferredLaunchPlanResult.entities
    ];
    return uniqBy(merged, ({ id }) => id.name);
}

async function loadWorkflowVersions(
    { listWorkflows }: APIContextValue,
    { preferredWorkflowId, sourceId: sourceWorkflowName }: WorkflowLaunchContext
) {
    if (!sourceWorkflowName) {
        throw new Error('Cannot load workflows, missing workflowName');
    }
    const { project, domain, name } = sourceWorkflowName;
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
    if (preferredWorkflowId) {
        const { version, ...scope } = preferredWorkflowId;
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

    const [workflowsResult, preferredWorkflowResult] = await Promise.all([
        workflowsPromise,
        preferredWorkflowPromise
    ]);
    const merged = [
        ...workflowsResult.entities,
        ...preferredWorkflowResult.entities
    ];
    return uniqBy(merged, ({ id: { version } }) => version);
}

async function loadInputs(
    { getWorkflow }: APIContextValue,
    { defaultInputValues, workflowVersion, launchPlan }: WorkflowLaunchContext
) {
    if (!workflowVersion) {
        throw new Error('Failed to load inputs: missing workflowVersion');
    }
    if (!launchPlan) {
        throw new Error('Failed to load inputs: missing launchPlan');
    }
    const workflow = await getWorkflow(workflowVersion);
    const parsedInputs: ParsedInput[] = getInputsForWorkflow(
        workflow,
        launchPlan,
        defaultInputValues
    );

    return {
        parsedInputs,
        unsupportedRequiredInputs: getUnsupportedRequiredInputs(parsedInputs)
    };
}

async function submit(
    { createWorkflowExecution }: APIContextValue,
    formInputsRef: RefObject<LaunchFormInputsRef>,
    roleInputRef: RefObject<LaunchRoleInputRef>,
    advancedOptionsRef: RefObject<LaunchAdvancedOptionsRef>,
    { launchPlan, referenceExecutionId, workflowVersion }: WorkflowLaunchContext
) {
    if (!launchPlan) {
        throw new Error('Attempting to launch with no LaunchPlan');
    }
    if (!workflowVersion) {
        throw new Error('Attempting to launch with no Workflow version');
    }
    if (formInputsRef.current === null) {
        throw new Error('Unexpected empty form inputs ref');
    }
    // if (roleInputRef.current === null) {
    //     throw new Error('Unexpected empty role input ref');
    // }
    // if (advancedOptionsRef.current === null) {
    //     throw new Error('Unexpected empty advanced options ref');
    // }

    const authRole = roleInputRef.current?.getValue();
    const literals = formInputsRef.current.getValues();
    const { disableAll, qualityOfService, maxParallelism } =
        advancedOptionsRef.current?.getValues() || {};
    const launchPlanId = launchPlan.id;
    const { domain, project } = workflowVersion;

    const response = await createWorkflowExecution({
        authRole,
        disableAll,
        qualityOfServiceTier: qualityOfService?.tier,
        maxParallelism,
        domain,
        launchPlanId,
        project,
        referenceExecutionId,
        inputs: { literals }
    });
    const newExecutionId = response.id as WorkflowExecutionIdentifier;
    if (!newExecutionId) {
        throw new Error('API Response did not include new execution id');
    }

    return newExecutionId;
}

async function validate(
    formInputsRef: RefObject<LaunchFormInputsRef>,
    roleInputRef: RefObject<LaunchRoleInputRef>,
    advancedOptionsRef: RefObject<LaunchAdvancedOptionsRef>
) {
    if (roleInputRef.current === null) {
        throw new Error('Unexpected empty role input ref');
    }

    // if (!roleInputRef.current.validate()) {
    //     throw new Error(correctInputErrors);
    // }
    return baseValidate(formInputsRef);
}

function getServices(
    apiContext: APIContextValue,
    formInputsRef: RefObject<LaunchFormInputsRef>,
    roleInputRef: RefObject<LaunchRoleInputRef>,
    advancedOptionsRef: RefObject<LaunchAdvancedOptionsRef>
) {
    return {
        loadWorkflowVersions: partial(loadWorkflowVersions, apiContext),
        loadLaunchPlans: partial(loadLaunchPlans, apiContext),
        loadInputs: partial(loadInputs, apiContext),
        submit: partial(
            submit,
            apiContext,
            formInputsRef,
            roleInputRef,
            advancedOptionsRef
        ),
        validate: partial(
            validate,
            formInputsRef,
            roleInputRef,
            advancedOptionsRef
        )
    };
}

/** Contains all of the form state for a LaunchWorkflowForm, including input
 * definitions, current input values, and errors.
 */
export function useLaunchWorkflowFormState({
    initialParameters = {},
    workflowId: sourceId,
    referenceExecutionId
}: LaunchWorkflowFormProps): LaunchWorkflowFormState {
    // These values will be used to auto-select items from the workflow
    // version/launch plan drop downs.
    const {
        authRole: defaultAuthRole,
        launchPlan: preferredLaunchPlanId,
        workflowId: preferredWorkflowId,
        values: defaultInputValues,
        disableAll,
        maxParallelism,
        qualityOfService
    } = initialParameters;

    const apiContext = useAPIContext();
    const formInputsRef = useRef<LaunchFormInputsRef>(null);
    const roleInputRef = useRef<LaunchRoleInputRef>(null);
    const advancedOptionsRef = useRef<LaunchAdvancedOptionsRef>(null);

    const services = useMemo(
        () =>
            getServices(
                apiContext,
                formInputsRef,
                roleInputRef,
                advancedOptionsRef
            ),
        [apiContext, formInputsRef, roleInputRef, advancedOptionsRef]
    );

    const [state, sendEvent, service] = useMachine<
        WorkflowLaunchContext,
        WorkflowLaunchEvent,
        WorkflowLaunchTypestate
    >(workflowLaunchMachine, {
        ...defaultStateMachineConfig,
        services,
        context: {
            defaultAuthRole,
            defaultInputValues,
            preferredLaunchPlanId,
            preferredWorkflowId,
            referenceExecutionId,
            sourceId,
            disableAll,
            maxParallelism,
            qualityOfService
        }
    });

    const {
        launchPlanOptions = [],
        launchPlan,
        workflowVersionOptions = [],
        workflowVersion
    } = state.context;

    const selectWorkflowVersion = (newWorkflow: WorkflowId) => {
        if (newWorkflow === workflowVersion) {
            return;
        }
        sendEvent({
            type: 'SELECT_WORKFLOW_VERSION',
            workflowId: newWorkflow
        });
    };

    const selectLaunchPlan = (newLaunchPlan: LaunchPlan) => {
        if (newLaunchPlan === launchPlan) {
            return;
        }
        sendEvent({
            type: 'SELECT_LAUNCH_PLAN',
            launchPlan: newLaunchPlan
        });
    };

    const workflowSourceSelectorState = useWorkflowSourceSelectorState({
        launchPlan,
        launchPlanOptions,
        sourceId,
        selectLaunchPlan,
        selectWorkflowVersion,
        workflowVersion,
        workflowVersionOptions
    });

    useEffect(() => {
        const subscription = service.subscribe(newState => {
            if (newState.matches(LaunchState.SELECT_WORKFLOW_VERSION)) {
                const {
                    workflowVersionOptions,
                    preferredWorkflowId
                } = newState.context;
                if (workflowVersionOptions.length > 0) {
                    let workflowToSelect = workflowVersionOptions[0];
                    if (preferredWorkflowId) {
                        const preferred = workflowVersionOptions.find(
                            ({ id }) => isEqual(id, preferredWorkflowId)
                        );
                        if (preferred) {
                            workflowToSelect = preferred;
                        }
                    }
                    sendEvent({
                        type: 'SELECT_WORKFLOW_VERSION',
                        workflowId: workflowToSelect.id
                    });
                }
            }

            if (newState.matches(LaunchState.SELECT_LAUNCH_PLAN)) {
                const {
                    launchPlan,
                    launchPlanOptions,
                    sourceId
                } = newState.context;
                if (!launchPlanOptions.length) {
                    return;
                }

                let launchPlanToSelect = launchPlanOptions[0];
                /* Attempt to select, in order:
                 * 1. The last launch plan that was selected, matching by the name, to preserve
                 *    any user selection before switching workflow versions.
                 * 2. The launch plan that was specified when initializing the form, by full id
                 * 3. The default launch plan, which has the same `name` as the workflow
                 * 4. The first launch plan in the list
                 */
                if (launchPlan) {
                    const lastSelected = launchPlanOptions.find(
                        ({ id: { name } }) => name === launchPlan.id.name
                    );
                    if (lastSelected) {
                        launchPlanToSelect = lastSelected;
                    }
                } else if (preferredLaunchPlanId) {
                    const preferred = launchPlanOptions.find(({ id }) =>
                        isEqual(id, preferredLaunchPlanId)
                    );
                    if (preferred) {
                        launchPlanToSelect = preferred;
                    }
                } else {
                    const defaultLaunchPlan = launchPlanOptions.find(
                        ({ id: { name } }) => name === sourceId.name
                    );
                    if (defaultLaunchPlan) {
                        launchPlanToSelect = defaultLaunchPlan;
                    }
                }

                sendEvent({
                    type: 'SELECT_LAUNCH_PLAN',
                    launchPlan: launchPlanToSelect
                });
            }
        });

        return subscription.unsubscribe;
    }, [service, sendEvent]);

    return {
        advancedOptionsRef,
        formInputsRef,
        roleInputRef,
        state,
        service,
        workflowSourceSelectorState
    };
}
