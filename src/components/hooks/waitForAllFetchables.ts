import { every, find, some } from 'lodash';
import { FetchableData } from './types';

export type MultiFetchableState = Pick<
    FetchableData<any>,
    'hasLoaded' | 'lastError' | 'loading'
>;

/** Merges the state of multiple fetchables into a state compatible with
 * WaitForData.
 * `hasLoaded` is true only when all fetchables have loaded.
 * `loading` is true if any fetchable is still loading.
 * `lastError` will be set to the first error found in the list of fetchables
 * or `null` if no fetchable has an error.
 */
export function waitForAllFetchables(
    fetchables: FetchableData<any>[]
): MultiFetchableState {
    const firstErrorItem = find(fetchables, f => f.lastError !== null);

    return {
        hasLoaded: every(fetchables, f => f.hasLoaded),
        loading: some(fetchables, f => f.loading),
        lastError: firstErrorItem ? firstErrorItem.lastError : null
    };
}
