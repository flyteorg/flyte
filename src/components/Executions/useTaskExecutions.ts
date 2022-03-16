import { APIContextValue, useAPIContext } from 'components/data/apiContext';
import { useDataRefresher } from 'components/hooks/useDataRefresher';
import { every } from 'lodash';
import { limits } from 'models/AdminEntity/constants';
import { SortDirection } from 'models/AdminEntity/types';
import {
  ExecutionData,
  NodeExecution,
  NodeExecutionIdentifier,
  TaskExecution,
  TaskExecutionIdentifier,
} from 'models/Execution/types';
import { taskSortFields } from 'models/Task/constants';
import { FetchableData } from '../hooks/types';
import { useFetchableData } from '../hooks/useFetchableData';
import { executionRefreshIntervalMs } from './constants';
import { nodeExecutionIsTerminal, taskExecutionIsTerminal } from './utils';

/** Fetches a list of `TaskExecution`s which are children of the given `NodeExecution`.
 * This function is meant to be consumed by hooks which are composing data.
 * If you're calling it from a component, consider using `useTaskExecutions` instead.
 */
export const fetchTaskExecutions = async (
  id: NodeExecutionIdentifier,
  apiContext: APIContextValue,
) => {
  const { listTaskExecutions } = apiContext;
  const { entities } = await listTaskExecutions(id, {
    limit: limits.NONE,
    sort: {
      key: taskSortFields.createdAt,
      direction: SortDirection.ASCENDING,
    },
  });
  return entities;
};

/** A hook for fetching the list of TaskExecutions associated with a
 * NodeExecution
 */
export function useTaskExecutions(id: NodeExecutionIdentifier): FetchableData<TaskExecution[]> {
  const apiContext = useAPIContext();
  return useFetchableData<TaskExecution[], NodeExecutionIdentifier>(
    {
      debugName: 'TaskExecutions',
      defaultValue: [],
      doFetch: async (id: NodeExecutionIdentifier) => fetchTaskExecutions(id, apiContext),
    },
    id,
  );
}

/** Fetches the signed URLs for TaskExecution data (inputs/outputs) */
export function useTaskExecutionData(id: TaskExecutionIdentifier): FetchableData<ExecutionData> {
  const { getTaskExecutionData } = useAPIContext();
  return useFetchableData<ExecutionData, TaskExecutionIdentifier>(
    {
      debugName: 'TaskExecutionData',
      defaultValue: {} as ExecutionData,
      doFetch: (id) => getTaskExecutionData(id),
    },
    id,
  );
}

/** Wraps the result of `useTaskExecutions` and will refresh the data as long
 * as the given `NodeExecution` is still in a non-final state.
 */
export function useTaskExecutionsRefresher(
  nodeExecution: NodeExecution,
  taskExecutionsFetchable: ReturnType<typeof useTaskExecutions>,
) {
  return useDataRefresher(nodeExecution.id, taskExecutionsFetchable, {
    interval: executionRefreshIntervalMs,
    valueIsFinal: (taskExecutions) =>
      every(taskExecutions, taskExecutionIsTerminal) && nodeExecutionIsTerminal(nodeExecution),
  });
}
