import { useTerminateExecutionState } from '../useTerminateExecutionState';

export function createMockTerminateExecutionState(): ReturnType<
    typeof useTerminateExecutionState
> {
    return {
        cause: '',
        setCause: jest.fn(),
        setTerminating: jest.fn(),
        setTerminationError: jest.fn(),
        terminateExecution: jest.fn(),
        terminating: false,
        terminationError: ''
    };
}
