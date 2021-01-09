import { createPaginationQuery } from 'components/data/queryUtils';
import { InfiniteQueryInput, InfiniteQueryPage, QueryType } from 'components/data/types';
import {
    DomainIdentifierScope,
    Execution,
    listExecutions,
    RequestConfig
} from 'models';
import { InfiniteData } from 'react-query';

export type ExecutionListQueryData = InfiniteData<InfiniteQueryPage<Execution>>;

/** A query for fetching a list of workflow executions belonging to a project/domain */
export function makeWorkflowExecutionListQuery(
    { domain, project }: DomainIdentifierScope,
    config?: RequestConfig
): InfiniteQueryInput<Execution> {
    return createPaginationQuery({
        queryKey: [
            QueryType.WorkflowExecutionList,
            { domain, project },
            config
        ],
        queryFn: async ({ pageParam }) => {
            const finalConfig = pageParam
                ? { ...config, token: pageParam }
                : config;
            const { entities: data, token } = await listExecutions(
                { domain, project },
                finalConfig
            );
            return { data, token };
        },
    });
}
