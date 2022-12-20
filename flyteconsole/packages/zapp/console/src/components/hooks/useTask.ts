import { useAPIContext } from 'components/data/apiContext';
import { RequestConfig } from 'models/AdminEntity/types';
import { Identifier, IdentifierScope } from 'models/Common/types';
import { Task, TaskTemplate } from 'models/Task/types';
import { FetchableData } from './types';
import { useFetchableData } from './useFetchableData';
import { usePagination } from './usePagination';

/** A hook for fetching a Task template. TaskTemplates may have already been
 * fetched as part of retrieving a Workflow. If not, we can retrieve the Task
 * directly and read the template from there.
 */
export function useTaskTemplate(id: Identifier): FetchableData<TaskTemplate> {
  const { getTask } = useAPIContext();
  return useFetchableData<TaskTemplate, Identifier>(
    {
      // Tasks are immutable
      useCache: true,
      debugName: 'TaskTemplate',
      defaultValue: {} as TaskTemplate,
      doFetch: async (taskId) => (await getTask(taskId)) as TaskTemplate,
    },
    id,
  );
}

/** A hook for fetching a paginated list of tasks */
export function useTaskList(scope: IdentifierScope, config: RequestConfig) {
  const { listTasks } = useAPIContext();
  return usePagination<Task, IdentifierScope>(
    { ...config, cacheItems: true, fetchArg: scope },
    listTasks,
  );
}
