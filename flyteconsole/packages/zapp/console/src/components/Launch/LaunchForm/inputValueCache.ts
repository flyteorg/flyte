import { createContext, useContext } from 'react';
import { InputValueMap } from './types';

/** Creates a Map of InputValues, optionally initializing it with the passed values */
export function createInputValueCache(values?: InputValueMap): InputValueMap {
  return values ? new Map(values.entries()) : new Map();
}

/** Provides an InputValue cache Map to child components */
export const InputValueCacheContext = createContext<InputValueMap>(createInputValueCache());

/** Convenience hook to retrieve the nearest ancestor InputValueCache context object. */
export function useInputValueCacheContext() {
  return useContext(InputValueCacheContext);
}
