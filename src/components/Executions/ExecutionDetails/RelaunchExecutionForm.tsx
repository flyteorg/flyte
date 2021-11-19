import { WaitForData } from 'components/common/WaitForData';
import { useAPIContext } from 'components/data/apiContext';
import { fetchStates } from 'components/hooks/types';
import { useFetchableData } from 'components/hooks/useFetchableData';
import { LaunchForm } from 'components/Launch/LaunchForm/LaunchForm';
import {
    TaskInitialLaunchParameters,
    WorkflowInitialLaunchParameters
} from 'components/Launch/LaunchForm/types';
import { fetchAndMapExecutionInputValues } from 'components/Launch/LaunchForm/useMappedExecutionInputValues';
import {
    getTaskInputs,
    getWorkflowInputs
} from 'components/Launch/LaunchForm/utils';
import { Execution } from 'models/Execution/types';
import * as React from 'react';
import { isSingleTaskExecution } from './utils';

export interface RelaunchExecutionFormProps {
    execution: Execution;
    onClose(): void;
}

function useRelaunchWorkflowFormState({
    execution
}: RelaunchExecutionFormProps) {
    const apiContext = useAPIContext();
    const initialParameters = useFetchableData<
        WorkflowInitialLaunchParameters,
        Execution
    >(
        {
            defaultValue: {} as WorkflowInitialLaunchParameters,
            doFetch: async execution => {
                const {
                    closure: { workflowId },
                    spec: {
                        launchPlan,
                        disableAll,
                        maxParallelism,
                        labels,
                        annotations
                    }
                } = execution;
                const workflow = await apiContext.getWorkflow(workflowId);
                const inputDefinitions = getWorkflowInputs(workflow);
                const values = await fetchAndMapExecutionInputValues(
                    {
                        execution,
                        inputDefinitions
                    },
                    apiContext
                );
                return {
                    values,
                    launchPlan,
                    workflowId,
                    disableAll,
                    maxParallelism,
                    labels,
                    annotations
                };
            }
        },
        execution
    );
    return { initialParameters };
}

function useRelaunchTaskFormState({ execution }: RelaunchExecutionFormProps) {
    const apiContext = useAPIContext();
    const initialParameters = useFetchableData<
        TaskInitialLaunchParameters,
        Execution
    >(
        {
            defaultValue: {} as TaskInitialLaunchParameters,
            doFetch: async execution => {
                const {
                    spec: { authRole, launchPlan: taskId }
                } = execution;
                const task = await apiContext.getTask(taskId);
                const inputDefinitions = getTaskInputs(task);
                const values = await fetchAndMapExecutionInputValues(
                    {
                        execution,
                        inputDefinitions
                    },
                    apiContext
                );
                return { authRole, values, taskId };
            }
        },
        execution
    );
    return { initialParameters };
}

const RelaunchTaskForm: React.FC<RelaunchExecutionFormProps> = props => {
    const { initialParameters } = useRelaunchTaskFormState(props);
    const {
        spec: { launchPlan: taskId }
    } = props.execution;
    return (
        <WaitForData {...initialParameters}>
            {initialParameters.state.matches(fetchStates.LOADED) ? (
                <LaunchForm
                    initialParameters={initialParameters.value}
                    onClose={props.onClose}
                    referenceExecutionId={props.execution.id}
                    taskId={taskId}
                />
            ) : null}
        </WaitForData>
    );
};
const RelaunchWorkflowForm: React.FC<RelaunchExecutionFormProps> = props => {
    const { initialParameters } = useRelaunchWorkflowFormState(props);
    const {
        closure: { workflowId }
    } = props.execution;
    return (
        <WaitForData {...initialParameters}>
            {initialParameters.state.matches(fetchStates.LOADED) ? (
                <LaunchForm
                    initialParameters={initialParameters.value}
                    onClose={props.onClose}
                    referenceExecutionId={props.execution.id}
                    workflowId={workflowId}
                />
            ) : null}
        </WaitForData>
    );
};

/** For a given execution, fetches the associated Workflow/Task and renders a
 * `LaunchForm` based on the same source with input values taken from the execution. */
export const RelaunchExecutionForm: React.FC<RelaunchExecutionFormProps> = props => {
    return isSingleTaskExecution(props.execution) ? (
        <RelaunchTaskForm {...props} />
    ) : (
        <RelaunchWorkflowForm {...props} />
    );
};
