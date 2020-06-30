import { useContext, useState } from 'react';
import { ExecutionContext } from '../contexts';

/** Holds state for `TerminateExecutionForm` */
export function useTerminateExecutionState(onClose: () => void) {
    const [cause, setCause] = useState('');
    const [terminating, setTerminating] = useState(false);
    const [terminationError, setTerminationError] = useState('');
    const { terminateExecution: doTerminateExecution } = useContext(
        ExecutionContext
    );

    const terminateExecution = async (cause: string) => {
        setTerminating(true);
        try {
            await doTerminateExecution(cause);
            onClose();
        } catch (e) {
            setTerminationError(`Failed to terminate execution: ${e}`);
            setTerminating(false);
        }
    };

    return {
        cause,
        setCause,
        setTerminating,
        setTerminationError,
        terminateExecution,
        terminating,
        terminationError
    };
}
