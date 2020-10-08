import { useMachine } from '@xstate/react';
import { defaultStateMachineConfig } from 'components/common/constants';
import { APIContextValue, useAPIContext } from 'components/data/apiContext';
import { isEqual, partial, uniqBy } from 'lodash';
import {
    FilterOperationName,
    Identifier,
    SortDirection,
    Task,
    taskSortFields,
    WorkflowExecutionIdentifier
} from 'models';
import { RefObject, useEffect, useMemo, useRef } from 'react';
import { getInputsForTask } from './getInputs';
import {
    LaunchState,
    TaskLaunchContext,
    TaskLaunchEvent,
    taskLaunchMachine,
    TaskLaunchTypestate
} from './launchMachine';
import { validate } from './services';
import {
    LaunchFormInputsRef,
    LaunchTaskFormProps,
    LaunchTaskFormState,
    ParsedInput
} from './types';
import { useTaskSourceSelectorState } from './useTaskSourceSelectorState';
import { getUnsupportedRequiredInputs } from './utils';

async function loadTaskVersions(
    { listTasks }: APIContextValue,
    { preferredTaskId, sourceId }: TaskLaunchContext
) {
    if (!sourceId) {
        throw new Error('Cannot load tasks, missing sourceId');
    }
    const { project, domain, name } = sourceId;
    const tasksPromise = listTasks(
        { project, domain, name },
        {
            limit: 10,
            sort: {
                key: taskSortFields.createdAt,
                direction: SortDirection.DESCENDING
            }
        }
    );

    let preferredTaskPromise = Promise.resolve({
        entities: [] as Task[]
    });
    if (preferredTaskId) {
        const { version, ...scope } = preferredTaskId;
        preferredTaskPromise = listTasks(scope, {
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

    const [tasksResult, preferredTaskResult] = await Promise.all([
        tasksPromise,
        preferredTaskPromise
    ]);
    const merged = [...tasksResult.entities, ...preferredTaskResult.entities];
    return uniqBy(merged, ({ id: { version } }) => version);
}

async function loadInputs(
    { getTask }: APIContextValue,
    { defaultInputValues, taskVersion }: TaskLaunchContext
) {
    if (!taskVersion) {
        throw new Error('Failed to load inputs: missing taskVersion');
    }

    const task = await getTask(taskVersion);
    const parsedInputs: ParsedInput[] = getInputsForTask(
        task,
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
    { referenceExecutionId, taskVersion }: TaskLaunchContext
) {
    if (!taskVersion) {
        throw new Error('Attempting to launch with no Task version');
    }
    if (formInputsRef.current === null) {
        throw new Error('Unexpected empty form inputs ref');
    }
    const literals = formInputsRef.current.getValues();
    const launchPlanId = taskVersion;
    const { domain, project } = taskVersion;

    const response = await createWorkflowExecution({
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

function getServices(
    apiContext: APIContextValue,
    formInputsRef: RefObject<LaunchFormInputsRef>
) {
    return {
        loadTaskVersions: partial(loadTaskVersions, apiContext),
        loadInputs: partial(loadInputs, apiContext),
        submit: partial(submit, apiContext, formInputsRef),
        validate: partial(validate, formInputsRef)
    };
}

/** Contains all of the form state for a LaunchTaskForm, including input
 * definitions, current input values, and errors.
 */
export function useLaunchTaskFormState({
    initialParameters = {},
    taskId: sourceId,
    referenceExecutionId
}: LaunchTaskFormProps): LaunchTaskFormState {
    // These values will be used to auto-select items from the task
    // version/launch plan drop downs.
    const {
        taskId: preferredTaskId,
        values: defaultInputValues
    } = initialParameters;

    const apiContext = useAPIContext();
    const formInputsRef = useRef<LaunchFormInputsRef>(null);

    const services = useMemo(() => getServices(apiContext, formInputsRef), [
        apiContext,
        formInputsRef
    ]);

    const [state, sendEvent, service] = useMachine<
        TaskLaunchContext,
        TaskLaunchEvent,
        TaskLaunchTypestate
    >(taskLaunchMachine, {
        ...defaultStateMachineConfig,
        services,
        context: {
            defaultInputValues,
            preferredTaskId,
            referenceExecutionId,
            sourceId
        }
    });

    const { taskVersionOptions = [], taskVersion } = state.context;

    const selectTaskVersion = (newTask: Identifier) => {
        if (newTask === taskVersion) {
            return;
        }
        sendEvent({
            type: 'SELECT_TASK_VERSION',
            taskId: newTask
        });
    };

    const taskSourceSelectorState = useTaskSourceSelectorState({
        sourceId,
        selectTaskVersion,
        taskVersion,
        taskVersionOptions
    });

    useEffect(() => {
        const subscription = service.subscribe(newState => {
            if (newState.matches(LaunchState.SELECT_TASK_VERSION)) {
                const {
                    taskVersionOptions,
                    preferredTaskId
                } = newState.context;
                if (taskVersionOptions.length > 0) {
                    let taskToSelect = taskVersionOptions[0];
                    if (preferredTaskId) {
                        const preferred = taskVersionOptions.find(({ id }) =>
                            isEqual(id, preferredTaskId)
                        );
                        if (preferred) {
                            taskToSelect = preferred;
                        }
                    }
                    sendEvent({
                        type: 'SELECT_TASK_VERSION',
                        taskId: taskToSelect.id
                    });
                }
            }
        });

        return subscription.unsubscribe;
    }, [service, sendEvent]);

    return {
        formInputsRef,
        state,
        service,
        taskSourceSelectorState
    };
}
