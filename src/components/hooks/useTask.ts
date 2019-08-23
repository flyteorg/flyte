import { NotFoundError } from 'errors';
import { Identifier, TaskTemplate } from 'models';

import { FetchableData } from './types';
import { useFetchableData } from './useFetchableData';

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
