import { getTaskExecution } from 'models/Execution/api';
import { TaskExecution, TaskExecutionIdentifier } from 'models/Execution/types';
import { FetchableData } from './types';
import { useFetchableData } from './useFetchableData';

/** A hook for fetching a TaskExecution */
export function useTaskExecution(id: TaskExecutionIdentifier): FetchableData<TaskExecution> {
  return useFetchableData<TaskExecution, TaskExecutionIdentifier>(
    {
      debugName: 'TaskExecution',
      defaultValue: {} as TaskExecution,
      doFetch: (id) => getTaskExecution(id),
    },
    id,
  );
}
