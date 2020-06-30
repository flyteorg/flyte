import { useAPIContext } from 'components/data/apiContext';
import { Workflow, WorkflowId } from 'models';
import { FetchableData } from './types';
import { useFetchableData } from './useFetchableData';

/** A hook for fetching a Workflow */
export function useWorkflow(
    id: WorkflowId | null = null
): FetchableData<Workflow> {
    const { getWorkflow } = useAPIContext();

    const doFetch = async (id: WorkflowId | null) => {
        if (id === null) {
            throw new Error('workflow id missing');
        }
        return getWorkflow(id);
    };

    return useFetchableData<Workflow, WorkflowId | null>(
        {
            doFetch,
            autoFetch: id !== null,
            debugName: 'Workflow',
            defaultValue: {} as Workflow
        },
        id
    );
}
