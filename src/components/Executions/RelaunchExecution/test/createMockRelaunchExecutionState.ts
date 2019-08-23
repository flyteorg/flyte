import { createMockFetchable } from 'components/hooks/__mocks__/fetchableData';
import { WorkflowExecutionIdentifier } from 'models';
import { useRelaunchExecutionState } from '../useRelaunchExecutionState';

export function createMockRelaunchExecutionState(): ReturnType<
    typeof useRelaunchExecutionState
> {
    return {
        fetchable: createMockFetchable<WorkflowExecutionIdentifier>({}),
        relaunchWorkflowExecution: jest.fn()
    };
}
