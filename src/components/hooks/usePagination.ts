import { CacheContext } from 'components/Cache/CacheContext';
import { RequestConfig } from 'models/AdminEntity/types';
import { useContext, useMemo } from 'react';
import { FetchFn, PaginatedFetchableData, PaginatedFetchFn, PaginationValue } from './types';
import { useFetchableData } from './useFetchableData';

export interface PaginationConfig<FetchArgType> extends RequestConfig {
  /** Whether to individually cache returned items. Should only be used
   * with item types containing an `id` property
   */
  cacheItems?: boolean;
  debugName?: string;
  /** An argument to be consumed by the fetch function. Any values which would
   * change the response (such as a `scope` or `id`) and are not part of a
   * standard `RequestConfig` should be passed here to ensure that the hook
   * fires whenever those values change.
   */
  fetchArg: FetchArgType;
}

/** A hook for fetching data from a paginated source
 * @param config An extended `RequestConfig` object to be used during requests
 * @param config.debugName An optional string to prepend to log messages
 * @returns An extended `FetchableData` object whose `value` is the complete
 * list of items across all pages fetched so far, with a `moreItemsAvailable`
 * boolean to indicate if another fetch would result in additional data.
 * Changing any values in `config` will reset the list.
 */
export function usePagination<T, FetchArgType>(
  config: PaginationConfig<FetchArgType>,
  doFetch: PaginatedFetchFn<T, FetchArgType>,
): PaginatedFetchableData<T> {
  const { cacheItems = false, debugName } = config;
  const cache = useContext(CacheContext);

  const fetch: FetchFn<PaginationValue<T>, PaginationConfig<FetchArgType>> = useMemo(
    () => async (params, currentValue) => {
      const { token: previousToken = '', items: previousItems = [] } = currentValue || {};
      const { fetchArg, ...requestConfig } = params;

      // If our last fetch call returned a token,
      // we have to pass that along in order to retrieve the next page
      const finalConfig = { ...requestConfig, token: previousToken };

      const { entities, token } = await doFetch(fetchArg, finalConfig);
      const result = cacheItems ? cache.mergeArray(entities) : entities;

      const items = previousItems.concat(result);
      return {
        items,
        token,
      };
    },
    [cache, cacheItems],
  );

  const fetchable = useFetchableData(
    {
      debugName,
      defaultValue: { token: '', items: [] },
      doFetch: fetch,
    },
    config,
  );

  const { items: value, token } = fetchable.value;
  return {
    ...fetchable,
    value,
    moreItemsAvailable: !!token,
  };
}
