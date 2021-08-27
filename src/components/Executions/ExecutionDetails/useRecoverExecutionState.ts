import { useAPIContext } from 'components/data/apiContext';
import { useContext } from 'react';
import { useMutation } from 'react-query';
import { ExecutionContext } from '../contexts';
import { WorkflowExecutionIdentifier } from 'models/Execution/types';

export function useRecoverExecutionState() {
    const { recoverWorkflowExecution } = useAPIContext();
    const {
        execution: { id }
    } = useContext(ExecutionContext);

    const { mutate, ...recoverState } = useMutation<
        WorkflowExecutionIdentifier,
        Error
    >(async () => {
        const { id: recoveredId } = await recoverWorkflowExecution({ id });
        if (!recoveredId) {
            throw new Error('API Response did not include new execution id');
        }
        return recoveredId as WorkflowExecutionIdentifier;
    });

    const recoverExecution = () => mutate();

    return {
        recoverState,
        recoverExecution
    };
}
