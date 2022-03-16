import { NotAuthorizedError, NotFoundError } from 'errors/fetchErrors';
import {
  DefaultOptions,
  hashQueryKey,
  QueryCache,
  QueryClient,
  QueryKeyHashFunction,
} from 'react-query';
import { normalizeQueryKey } from './utils';

const allowedFailures = 3;

function isErrorRetryable(error: any) {
  // Fail immediately for auth errors, retries won't succeed
  return !(error instanceof NotAuthorizedError) && !(error instanceof NotFoundError);
}

const queryKeyHashFn: QueryKeyHashFunction = (queryKey) =>
  hashQueryKey(normalizeQueryKey(queryKey));

export function createQueryClient(options?: Partial<DefaultOptions>) {
  const queryCache = new QueryCache();
  return new QueryClient({
    queryCache,
    defaultOptions: {
      queries: {
        queryKeyHashFn,
        retry: (failureCount, error) => failureCount < allowedFailures && isErrorRetryable(error),
      },
      ...options,
    },
  });
}
