import { useEffect } from 'react';

import { createDebugLogger } from 'common/log';
import { getCacheKey } from 'components/Cache';

import { FetchableData, RefreshConfig } from './types';

const log = createDebugLogger('useDataRefresher');

const defaultRefresh = (fetchable: FetchableData<any>) => fetchable.fetch();

/** A hook which attaches auto-refresh behavior to a `FetchableData` object.
 * @param id A unique id value used to key this hook. This usually should match
 * the value provided to your data-fetching hook
 * @param fetchable An instance of `FetchableData`, usually provided by a data-
 * fetching hook
 * @param refreshConfig Configures the refresh behavior (interval, termination
 * logic, fetch function, etc)
 */
export function useDataRefresher<T, IDType extends object | string>(
    id: IDType,
    fetchable: FetchableData<T>,
    refreshConfig: RefreshConfig<T>
) {
    const { interval } = refreshConfig;
    const { debugName, hasLoaded, lastError, value } = fetchable;

    const isFinal = refreshConfig.valueIsFinal(value);

    // Default refresh implementation is just to fetch the current entity
    const doRefresh = refreshConfig.doRefresh || defaultRefresh;

    useEffect(() => {
        let timerId: number = 0;

        const clear = () => {
            if (timerId === 0) {
                return;
            }

            log(`${debugName} detaching data refresher`);
            window.clearInterval(timerId);
        };

        if (!hasLoaded || isFinal || lastError) {
            if (lastError) {
                log(
                    `${debugName} not refreshing fetchable because it is in an error state`,
                    fetchable
                );
            }
        } else {
            log(`${debugName} attaching data refresher`);
            timerId = window.setInterval(() => doRefresh(fetchable), interval);
        }

        // When this effect is cleaned up, we should stop refreshing
        return clear;
    }, [getCacheKey(id), hasLoaded, isFinal, lastError]);
}
