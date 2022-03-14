import { QueryInput, QueryType } from 'components/data/types';
import { DataError } from 'components/Errors/DataError';
import { extractTaskTemplates } from 'components/hooks/utils';
import { getNodeExecutionData } from 'models/Execution/api';
import { getWorkflow } from 'models/Workflow/api';
import { Workflow, WorkflowId } from 'models/Workflow/types';
import { QueryClient, QueryObserverResult } from 'react-query';

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

export function makeNodeExecutionDynamicWorkflowQuery(
    queryClient: QueryClient,
    parentsToFetch
): QueryInput<{ [key: string]: any }> {
    return {
        queryKey: [QueryType.DynamicWorkflowFromNodeExecution, parentsToFetch],
        queryFn: async () => {
            return await Promise.all(
                Object.keys(parentsToFetch).map(id => {
                    const executionId = parentsToFetch[id];
                    const data = getNodeExecutionData(executionId.id).then(
                        value => {
                            return { key: id, value: value };
                        }
                    );
                    return data;
                })
            ).then(values => {
                const output: { [key: string]: any } = {};
                for (let i = 0; i < values.length; i++) {
                    /* Filter to only include dynamicWorkflow */
                    if (values[i].value.dynamicWorkflow) {
                        output[values[i].key] = values[i].value;
                    }
                }
                return output;
            });
        }
    };
}

export async function fetchWorkflow(queryClient: QueryClient, id: WorkflowId) {
    return queryClient.fetchQuery(makeWorkflowQuery(queryClient, id));
}
