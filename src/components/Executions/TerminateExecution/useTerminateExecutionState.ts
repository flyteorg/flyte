import { useAPIContext } from 'components/data/apiContext';
import { QueryType } from 'components/data/types';
import { waitForQueryState } from 'components/data/queryUtils';
import { Execution } from 'models';
import { useContext, useState } from 'react';
import { useMutation, useQueryClient } from 'react-query';
import { ExecutionContext } from '../contexts';
import { executionIsTerminal } from '../utils';

interface TerminateExecutionVariables {
    cause: string;
}

/** Holds state for `TerminateExecutionForm` */
export function useTerminateExecutionState(onClose: () => void) {
    const { getExecution, terminateWorkflowExecution } = useAPIContext();
    const [cause, setCause] = useState('');
    const {
        execution: { id }
    } = useContext(ExecutionContext);
    const queryClient = useQueryClient();

    const { mutate, ...terminationState } = useMutation<
        Execution,
        Error,
        TerminateExecutionVariables
    >(
        async ({ cause }: TerminateExecutionVariables) => {
            await terminateWorkflowExecution(id, cause);
            return await waitForQueryState<Execution>({
                queryClient,
                queryKey: [QueryType.WorkflowExecution, id],
                queryFn: () => getExecution(id),
                valueCheckFn: executionIsTerminal
            });
        },
        {
            onSuccess: updatedExecution => {
                queryClient.setQueryData(
                    [QueryType.WorkflowExecution, id],
                    updatedExecution
                );
                onClose();
            }
        }
    );

    const terminateExecution = async (cause: string) => await mutate({ cause });

    return {
        cause,
        setCause,
        terminationState,
        terminateExecution
    };
}
