import { QueryKey, useQuery, useQueryClient, UseQueryOptions } from 'react-query';

interface ConditionalQueryOptions<TData, TError, TQueryFnData>
  extends UseQueryOptions<TData, TError, TQueryFnData> {
  queryKey: QueryKey;
}

/** Returns an instance of useQuery that is conditionally enabled based on
 * a `shouldEnable` function. The passed function will be given the current
 * data value and must return whether the query should fetch an updated value.
 * NOTE: The query will *always* run while the value is `undefined`, so this
 * hook is not safe for use with queries which may return `undefined` value
 * after a fetch.
 */
export function useConditionalQuery<TData = unknown, TError = Error, TQueryFnData = TData>(
  options: ConditionalQueryOptions<TData, TError, TQueryFnData>,
  shouldEnableFn: (value: TData) => boolean,
) {
  const queryClient = useQueryClient();
  const queryData = queryClient.getQueryData<TData>(options.queryKey);
  const enabled = queryData === undefined || shouldEnableFn(queryData);
  const staleTime = enabled ? options.staleTime : Infinity;
  const refetchInterval = enabled ? options.refetchInterval : false;
  return useQuery<TData, TError, TQueryFnData>({ ...options, refetchInterval, staleTime });
}
