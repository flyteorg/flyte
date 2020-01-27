import { createContext, useContext } from 'react';
import { InputValueMap } from './types';

export function createInputValueCache(values?: InputValueMap): InputValueMap {
    return values ? new Map(values.entries()) : new Map();
}

export const InputValueCacheContext = createContext<InputValueMap>(
    createInputValueCache()
);

export function useInputValueCacheContext() {
    return useContext(InputValueCacheContext);
}
