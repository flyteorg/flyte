import { useAPIContext } from 'components/data/apiContext';
import { NotFoundError } from 'errors';
import {
    Identifier,
    IdentifierScope,
    RequestConfig,
    Task,
    TaskTemplate
} from 'models';
import { FetchableData } from './types';
import { useFetchableData } from './useFetchableData';
import { usePagination } from './usePagination';

/** A hook for fetching a Task template */
export function useTaskTemplate(id: Identifier): FetchableData<TaskTemplate> {
    return useFetchableData<TaskTemplate, Identifier>(
        {
            useCache: true,
            debugName: 'TaskTemplate',
            defaultValue: {} as TaskTemplate,
            // Fetching the parent workflow should insert these into the cache
            // for us. If we get to this point, something went wrong.
            doFetch: () =>
                Promise.reject(
                    new NotFoundError(
                        'Task template',
                        'No template has been loaded for this task'
                    )
                )
        },
        id
    );
}

/** A hook for fetching a paginated list of tasks */
export function useTaskList(scope: IdentifierScope, config: RequestConfig) {
    const { listTasks } = useAPIContext();
    return usePagination<Task, IdentifierScope>(
        { ...config, cacheItems: true, fetchArg: scope },
        listTasks
    );
}
