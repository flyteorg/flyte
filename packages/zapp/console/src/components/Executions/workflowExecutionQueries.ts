import { createPaginationQuery } from 'components/data/queryUtils';
import { InfiniteQueryInput, QueryType } from 'components/data/types';
import { RequestConfig } from 'models/AdminEntity/types';
import { DomainIdentifierScope } from 'models/Common/types';
import { listExecutions } from 'models/Execution/api';
import { Execution } from 'models/Execution/types';

/** A query for fetching a list of workflow executions belonging to a project/domain */
export function makeWorkflowExecutionListQuery(
  { domain, project }: DomainIdentifierScope,
  config?: RequestConfig,
): InfiniteQueryInput<Execution> {
  return createPaginationQuery({
    queryKey: [QueryType.WorkflowExecutionList, { domain, project }, config],
    queryFn: async ({ pageParam }) => {
      const finalConfig = pageParam ? { ...config, token: pageParam } : config;
      const { entities: data, token } = await listExecutions({ domain, project }, finalConfig);
      return { data, token };
    },
  });
}
