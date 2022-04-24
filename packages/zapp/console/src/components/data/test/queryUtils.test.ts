import { createPaginationQuery } from '../queryUtils';
import { InfiniteQueryPage, QueryType } from '../types';

describe('queryUtils', () => {
  describe('createPaginationQuery', () => {
    it('should treat empty string token as undefined', () => {
      const response: InfiniteQueryPage<string> = { data: [], token: '' };
      const query = createPaginationQuery<string>({
        queryKey: [QueryType.WorkflowExecutionList, 'test'],
        queryFn: () => response,
      });
      expect(query.getNextPageParam(response)).toBeUndefined();
    });
  });
});
