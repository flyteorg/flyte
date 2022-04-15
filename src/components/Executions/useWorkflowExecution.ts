import { APIContextValue, useAPIContext } from 'components/data/apiContext';
import { QueryInput, QueryType } from 'components/data/types';
import { useConditionalQuery } from 'components/hooks/useConditionalQuery';
import { LiteralMap } from 'models/Common/types';
import { getExecution } from 'models/Execution/api';
import { Execution, ExecutionData, WorkflowExecutionIdentifier } from 'models/Execution/types';
import { QueryClient } from 'react-query';
import { FetchableData } from '../hooks/types';
import { useFetchableData } from '../hooks/useFetchableData';
import { executionRefreshIntervalMs } from './constants';
import { executionIsTerminal } from './utils';

function shouldRefreshExecution(execution: Execution): boolean {
  const result = !executionIsTerminal(execution);
  return result;
}

export function makeWorkflowExecutionQuery(id: WorkflowExecutionIdentifier): QueryInput<Execution> {
  return {
    queryKey: [QueryType.WorkflowExecution, id],
    queryFn: () => getExecution(id),
  };
}

export function fetchWorkflowExecution(queryClient: QueryClient, id: WorkflowExecutionIdentifier) {
  return queryClient.fetchQuery(makeWorkflowExecutionQuery(id));
}

export function useWorkflowExecutionQuery(id: WorkflowExecutionIdentifier) {
  return useConditionalQuery<Execution>(
    {
      ...makeWorkflowExecutionQuery(id),
      refetchInterval: executionRefreshIntervalMs,
    },
    shouldRefreshExecution,
  );
}

/** Fetches the signed URLs for NodeExecution data (inputs/outputs) */
export function useWorkflowExecutionData(
  id: WorkflowExecutionIdentifier,
): FetchableData<ExecutionData> {
  const { getExecutionData } = useAPIContext();
  return useFetchableData<ExecutionData, WorkflowExecutionIdentifier>(
    {
      debugName: 'ExecutionData',
      defaultValue: {} as ExecutionData,
      doFetch: (id) => getExecutionData(id),
    },
    id,
  );
}

/** Fetches the inputs object for a given WorkflowExecution.
 * This function is meant to be consumed by hooks which are composing data.
 * If you're calling it from a component, consider using `useTaskExecutions` instead.
 */
export const fetchWorkflowExecutionInputs = async (
  execution: Execution,
  apiContext: APIContextValue,
) => {
  const { getExecutionData } = apiContext;
  if (execution.closure.computedInputs) {
    return execution.closure.computedInputs;
  }
  /** Note:
   * getExecutionData will retun raw values (`fullInputs`) if input payload isn't too large.
   */
  const { fullInputs } = await getExecutionData(execution.id);
  if (fullInputs) {
    return LiteralMap.create(fullInputs);
  }

  return { literals: {} };
};

/** A hook for fetching the inputs object associated with an Execution. Will
 * handle both the legacy (`computedInputs`) and current (externally stored) formats
 */
export function useWorkflowExecutionInputs(execution: Execution) {
  const apiContext = useAPIContext();
  return useFetchableData<LiteralMap, WorkflowExecutionIdentifier>(
    {
      debugName: 'ExecutionInputs',
      defaultValue: { literals: {} } as LiteralMap,
      doFetch: async () => fetchWorkflowExecutionInputs(execution, apiContext),
    },
    execution.id,
  );
}
