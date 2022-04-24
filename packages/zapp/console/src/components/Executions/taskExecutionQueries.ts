import { QueryInput, QueryType } from 'components/data/types';
import { RequestConfig } from 'models/AdminEntity/types';
import { getTaskExecution, listTaskExecutions } from 'models/Execution/api';
import {
  NodeExecutionIdentifier,
  TaskExecution,
  TaskExecutionIdentifier,
} from 'models/Execution/types';
import { QueryClient } from 'react-query';

/** A query for fetching a single `TaskExecution` by id. */
export function makeTaskExecutionQuery(id: TaskExecutionIdentifier): QueryInput<TaskExecution> {
  return {
    queryKey: [QueryType.TaskExecution, id],
    queryFn: () => getTaskExecution(id),
  };
}

// On successful task execution list queries, extract and store all
// executions so they are individually fetchable from the cache.
function cacheTaskExecutions(queryClient: QueryClient, taskExecutions: TaskExecution[]) {
  taskExecutions.forEach((te) => queryClient.setQueryData([QueryType.TaskExecution, te.id], te));
}

/** A query for fetching a list of `TaskExecution`s which are children of a given
 * `NodeExecution`.
 */
export function makeTaskExecutionListQuery(
  queryClient: QueryClient,
  id: NodeExecutionIdentifier,
  config?: RequestConfig,
): QueryInput<TaskExecution[]> {
  return {
    queryKey: [QueryType.TaskExecutionList, id, config],
    queryFn: async () => {
      const taskExecutions = (await listTaskExecutions(id, config)).entities;
      cacheTaskExecutions(queryClient, taskExecutions);
      return taskExecutions;
    },
  };
}

/** Composable fetch function which wraps `makeTaskExecutionListQuery` */
export function fetchTaskExecutionList(
  queryClient: QueryClient,
  id: NodeExecutionIdentifier,
  config?: RequestConfig,
) {
  return queryClient.fetchQuery(makeTaskExecutionListQuery(queryClient, id, config));
}
