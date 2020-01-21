import { createContext, useContext } from 'react';
import { InputValue } from './types';

export type InputValueCache = Map<string, InputValue>;

export function createInputValueCache(): InputValueCache {
    return new Map();
}

export const InputValueCacheContext = createContext<InputValueCache>(
    createInputValueCache()
);

export function useInputValueCacheContext() {
    return useContext(InputValueCacheContext);
}
