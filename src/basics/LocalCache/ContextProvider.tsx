// More info on Local storage: https://developer.mozilla.org/en-US/docs/Web/API/Window/localStorage
import { log } from 'common/log';
import * as React from 'react';
import { useState, createContext, useCallback } from 'react';
import { defaultLocalCacheConfig, LocalCacheItem } from './defaultConfig';

export { LocalCacheItem } from './defaultConfig';

interface LocalCacheState {
  localCacheMap: Record<string, any>;
  setLocalCache: (localCacheItem: LocalCacheItem, newValue: any) => void;
  getLocalCache: (localCacheItem: LocalCacheItem) => any;
}

interface LocalCacheProviderProps {
  children?: React.ReactNode;
}

// get the local storage value, if local storage is null, use default value
function allStorage() {
  const defaultKeys = Object.keys(defaultLocalCacheConfig);

  const localCacheMap: Record<string, any> = {};
  let i = defaultKeys.length;
  while (i--) {
    const key = defaultKeys[i];
    const value = localStorage.getItem(key) || defaultLocalCacheConfig[key];
    localCacheMap[defaultKeys[i]] = JSON.parse(value);
  }
  return localCacheMap;
}

export const LocalCacheContext = createContext<LocalCacheState>({
  localCacheMap: { ...allStorage() },
  setLocalCache: () => {
    /* Provider is not initialized */
  },
  getLocalCache: () => false,
});

export const LocalCacheProvider = (props: LocalCacheProviderProps) => {
  const [localCacheMap, setLocalCacheMap] = useState({ ...allStorage() });

  const setLocalCache = useCallback((localCacheItem: LocalCacheItem, newValue: any) => {
    setLocalCacheMap((prev) => ({ ...prev, [localCacheItem]: newValue }));
  }, []);

  const getLocalCache = useCallback(
    (localCacheItem: LocalCacheItem) => {
      const localCache = localCacheMap[localCacheItem];
      if (localCache == null) {
        log.error(
          `ERROR: LocalCacheItem ${localCacheItem} doesn't have default value provided in defaultLocalCacheConfig`,
        );
        return null;
      }
      return localCache;
    },
    [localCacheMap],
  );

  return (
    <LocalCacheContext.Provider value={{ localCacheMap, setLocalCache, getLocalCache }}>
      {props.children}
    </LocalCacheContext.Provider>
  );
};
