// More info on Local storage: https://developer.mozilla.org/en-US/docs/Web/API/Window/localStorage
import { useState } from 'react';
import { defaultLocalCacheConfig, LocalCacheItem } from './defaultConfig';

export { LocalCacheItem } from './defaultConfig';

export function ClearLocalCache() {
    localStorage.clear();
}

const getDefault = (setting: LocalCacheItem) => {
    const result = defaultLocalCacheConfig[setting];
    if (!result) {
        console.error(
            `ERROR: LocalCacheItem ${setting} doesn't have default value provided in defaultLocalCacheConfig`
        );
        return null;
    }
    return JSON.parse(result);
};

export function useLocalCache<T>(setting: LocalCacheItem) {
    const defaultValue = getDefault(setting);
    const [value, setValue] = useState<T>(() => {
        const data = localStorage.getItem(setting);
        const value = data ? JSON.parse(data) : defaultValue;
        if (typeof value === typeof defaultValue) {
            return value;
        }

        return defaultValue;
    });

    const setLocalCache = (newValue: T) => {
        localStorage.setItem(setting, JSON.stringify(newValue));
        setValue(newValue);
    };

    const clearState = () => {
        localStorage.removeItem(setting);
        setValue(defaultValue);
    };

    return [value, setLocalCache, clearState];
}

export const onlyForTesting = {
    getDefault
};
