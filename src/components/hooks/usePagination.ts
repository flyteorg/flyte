import { useContext, useEffect, useState } from 'react';

import { CacheContext, getCacheKey } from 'components/Cache';
import { RequestConfig } from 'models';

import { FetchFn, PaginatedFetchableData, PaginatedFetchFn } from './types';
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
    doFetch: PaginatedFetchFn<T, FetchArgType>
): PaginatedFetchableData<T> {
    const { cacheItems = false, debugName } = config;
    const cacheKey = getCacheKey(config);
    const [token, setToken] = useState('');
    const [moreItemsAvailable, setMoreItemsAvailable] = useState(false);
    const cache = useContext(CacheContext);

    // Reset our state if the pagination config changes
    useEffect(() => {
        setToken('');
        setMoreItemsAvailable(false);
    }, [cacheKey]);

    const fetch: FetchFn<T[], PaginationConfig<FetchArgType>> = async (
        params,
        currentValue = []
    ) => {
        const { fetchArg, ...requestConfig } = params;

        // If our last fetch call returned a token,
        // we have to pass that along in order to retrieve the next page
        if (token) {
            requestConfig.token = token;
        }

        const { entities, token: newToken } = await doFetch(
            fetchArg,
            requestConfig
        );
        const values = cacheItems ? cache.mergeArray(entities) : entities;

        if (newToken) {
            setToken(newToken);
        }
        const newValue = currentValue.concat(values);
        setMoreItemsAvailable(!!newToken);
        return newValue;
    };

    return {
        ...useFetchableData(
            {
                debugName,
                defaultValue: [],
                doFetch: fetch
            },
            config
        ),
        moreItemsAvailable
    };
}
