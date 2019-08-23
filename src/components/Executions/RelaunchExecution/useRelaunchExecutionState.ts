import { useFetchableData } from 'components/hooks';
import {
    Execution,
    relaunchWorkflowExecution,
    WorkflowExecutionIdentifier
} from 'models';
import { history, Routes } from 'routes';

const doRelaunchExecution = async (id: WorkflowExecutionIdentifier) => {
    const { id: newId } = await relaunchWorkflowExecution({ id });
    return newId as WorkflowExecutionIdentifier;
};

/** Holds state for `RelaunchExecutionForm` */
export function useRelaunchExecutionState(
    { id }: Execution,
    onClose: () => void
) {
    const fetchable = useFetchableData<
        WorkflowExecutionIdentifier,
        WorkflowExecutionIdentifier
    >(
        {
            autoFetch: false,
            debugName: 'RelaunchExecution',
            useCache: false,
            defaultValue: {} as WorkflowExecutionIdentifier,
            doFetch: id => doRelaunchExecution(id)
        },
        id
    );

    const relaunchWorkflowExecution = async () => {
        const id = await fetchable.fetch();
        onClose();
        history.push(Routes.ExecutionDetails.makeUrl(id));
        return id;
    };

    return {
        fetchable,
        relaunchWorkflowExecution
    };
}
