// More info on Local storage: https://developer.mozilla.org/en-US/docs/Web/API/Window/localStorage
import { log } from 'common/log';
import { useContext } from 'react';
import { defaultLocalCacheConfig, LocalCacheItem } from './defaultConfig';
import { LocalCacheContext } from './ContextProvider';

export { LocalCacheItem } from './defaultConfig';

export function ClearLocalCache() {
  localStorage.clear();
}

const getDefault = (setting: LocalCacheItem) => {
  const result = defaultLocalCacheConfig[setting];
  if (!result) {
    log.error(
      `ERROR: LocalCacheItem ${setting} doesn't have default value provided in defaultLocalCacheConfig`,
    );
    return null;
  }
  return JSON.parse(result);
};

export function useLocalCache<T>(setting: LocalCacheItem) {
  const localCache = useContext(LocalCacheContext);

  const value = localCache.getLocalCache(setting);

  const setLocalCache = (newValue: T) => {
    localCache.setLocalCache(setting, newValue);
    localStorage.setItem(setting, JSON.stringify(newValue));
  };

  const clearState = () => {
    localCache.setLocalCache(setting, JSON.parse(defaultLocalCacheConfig[setting]));
    localStorage.removeItem(setting);
  };

  return [value, setLocalCache, clearState];
}

export const onlyForTesting = {
  getDefault,
};
