import { State } from 'xstate';
import { FetchableData, fetchStates } from '../types';

export function createMockFetchable<T>(value: any, fetch: () => void): FetchableData<T> {
  return {
    fetch,
    value,
    debugName: '',
    lastError: null,
    state: State.from(fetchStates.IDLE),
  };
}

export function loadingFetchable<T>(value: any, fetch: () => void): FetchableData<T> {
  return {
    ...createMockFetchable(value, fetch),
    state: State.from(fetchStates.LOADING),
  };
}

export function loadedFetchable<T>(value: any, fetch: () => void): FetchableData<T> {
  return {
    ...createMockFetchable(value, fetch),
    state: State.from(fetchStates.LOADED),
  };
}
