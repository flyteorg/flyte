import { log } from 'common/log';
import { QueryClient, QueryFunction, QueryKey } from 'react-query';
import { InfiniteQueryInput, InfiniteQueryPage } from './types';

const defaultRefetchInterval = 1000;
const defaultTimeout = 30000;

export interface WaitForQueryValueConfig<TResult> {
  queryClient: QueryClient;
  queryKey: QueryKey;
  queryFn: QueryFunction<TResult>;
  refetchInterval?: number;
  timeout?: number;
  valueCheckFn: (value: TResult) => boolean;
}

/** Executes a query against the given QueryClient on an interval until the data
 * returned by the query satisfies the `valueCheckFn` predicate or the timeout is reached.
 */
export async function waitForQueryState<TResult>({
  queryClient,
  queryKey,
  queryFn,
  refetchInterval = defaultRefetchInterval,
  timeout = defaultTimeout,
  valueCheckFn,
}: WaitForQueryValueConfig<TResult>): Promise<TResult> {
  const queryWaitPromise = new Promise<TResult>((resolve) => {
    const doFetch = async () => {
      try {
        const result = await queryClient.fetchQuery<TResult, Error>({
          queryKey,
          queryFn,
        });
        if (valueCheckFn(result)) {
          return resolve(result);
        }
      } catch (e) {
        log.warn(`Unexpected failure while waiting for query: ${e}`, queryKey);
      }
      setTimeout(doFetch, refetchInterval);
    };
    doFetch();
  });
  const timeoutPromise = new Promise<TResult>((_, reject) => {
    setTimeout(() => reject(new Error('Timed Out')), timeout);
  });
  return Promise.race([queryWaitPromise, timeoutPromise]);
}

function getNextPageParam<T>({ token }: InfiniteQueryPage<T>) {
  // An empty token will cause pagination code to think there are more results.
  // Only return a defined value if it is a non-zero-length string.
  return token != null && token.length > 0 ? token : undefined;
}

/** Composes a `queryOptions` object with generic options which make our API responses
 * compatible with `useInfiniteQuery`
 */
export function createPaginationQuery<T>(queryOptions: InfiniteQueryInput<T>) {
  return { ...queryOptions, getNextPageParam };
}
