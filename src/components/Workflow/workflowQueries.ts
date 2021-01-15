import { QueryInput, QueryType } from 'components/data/types';
import { extractTaskTemplates } from 'components/hooks/utils';
import { getWorkflow } from 'models/Workflow/api';
import { Workflow, WorkflowId } from 'models/Workflow/types';
import { QueryClient } from 'react-query';

export function makeWorkflowQuery(
    queryClient: QueryClient,
    id: WorkflowId
): QueryInput<Workflow> {
    return {
        queryKey: [QueryType.Workflow, id],
        queryFn: async () => {
            const workflow = await getWorkflow(id);
            // On successful workflow fetch, extract and cache all task templates
            // stored on the workflow so that we don't need to fetch them separately
            // if future queries reference them.
            extractTaskTemplates(workflow).forEach(task =>
                queryClient.setQueryData(
                    [QueryType.TaskTemplate, task.id],
                    task
                )
            );
            return workflow;
        },
        // `Workflow` objects (individual versions) are immutable and safe to
        // cache indefinitely once retrieved in full
        staleTime: Infinity
    };
}

export async function fetchWorkflow(queryClient: QueryClient, id: WorkflowId) {
    return queryClient.fetchQuery(makeWorkflowQuery(queryClient, id));
}
