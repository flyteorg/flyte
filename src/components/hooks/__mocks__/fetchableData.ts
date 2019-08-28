import { FetchableData, FetchFn } from '../types';

export function createMockFetchable<T>(
    value: any,
    fetch: FetchFn<T, any>
): FetchableData<T> {
    return {
        fetch,
        value,
        debugName: '',
        hasLoaded: false,
        lastError: null,
        loading: false
    };
}

export function loadingFetchable<T>(
    value: any,
    fetch: FetchFn<T, any>
): FetchableData<T> {
    return { ...createMockFetchable(value, fetch), loading: true };
}

export function loadedFetchable<T>(
    value: any,
    fetch: FetchFn<T, any>
): FetchableData<T> {
    return { ...createMockFetchable(value, fetch), hasLoaded: true };
}
