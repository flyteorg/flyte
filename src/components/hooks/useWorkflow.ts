import { useContext } from 'react';

import { CacheContext } from 'components/Cache';
import { getWorkflow, Workflow, WorkflowId } from 'models';

import { FetchableData } from './types';
import { useFetchableData } from './useFetchableData';
import { extractTaskTemplates } from './utils';

/** A hook for fetching a Workflow */
export function useWorkflow(id: WorkflowId): FetchableData<Workflow> {
    const cache = useContext(CacheContext);

    const doFetch = async (id: WorkflowId) => {
        const workflow = await getWorkflow(id);
        const templates = extractTaskTemplates(workflow);
        cache.mergeArray(templates);
        return workflow;
    };

    return useFetchableData<Workflow, WorkflowId>(
        {
            doFetch,
            useCache: false,
            debugName: 'Workflow',
            defaultValue: {} as Workflow
        },
        id
    );
}
