import { createDebugLogger } from 'common/log';
import { getCacheKey } from 'components/Cache/utils';
import { useEffect } from 'react';
import { isLoadingState } from './fetchMachine';
import { FetchableData, fetchStates, RefreshConfig } from './types';

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
  refreshConfig: RefreshConfig<T>,
) {
  const { interval } = refreshConfig;
  const { debugName, state, value } = fetchable;

  const isFinal = refreshConfig.valueIsFinal(value);

  // Default refresh implementation is just to fetch the current entity
  const doRefresh = refreshConfig.doRefresh || defaultRefresh;
  const stateIsRefreshable = state.matches(fetchStates.LOADED) || isLoadingState(state);

  useEffect(() => {
    if (isFinal) {
      return;
    }

    let timerId = 0;

    const clear = () => {
      if (timerId === 0) {
        return;
      }

      log(`${debugName} detaching data refresher`);
      window.clearInterval(timerId);
    };

    if (stateIsRefreshable) {
      log(`${debugName} attaching data refresher`);
      timerId = window.setInterval(() => doRefresh(fetchable), interval);
    } else {
      log(`${debugName} not refreshing fetchable because it is in a failed/idle state`, fetchable);
    }

    // When this effect is cleaned up, we should stop refreshing
    return clear;
  }, [getCacheKey(id), stateIsRefreshable, isFinal, doRefresh]);
}
