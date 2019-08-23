import { FetchableData } from '../types';

export function createMockFetchable<T>(value: any): FetchableData<T> {
    return {
        value,
        debugName: '',
        hasLoaded: false,
        fetch: jest.fn(),
        lastError: null,
        loading: false
    };
}

export function loadingFetchable<T>(value: any): FetchableData<T> {
    return { ...createMockFetchable(value), loading: true };
}

export function loadedFetchable<T>(value: any): FetchableData<T> {
    return { ...createMockFetchable(value), hasLoaded: true };
}
