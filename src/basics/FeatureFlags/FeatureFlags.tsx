/**
 * Feature Flag provider - allows a multi-stage development.
 */
import * as React from 'react';
import { createContext, useCallback, useContext, useEffect, useState } from 'react';
import { isDevEnv, isTestEnv } from 'common/env';
import { defaultFlagConfig, FeatureFlag, FeatureFlagConfig } from './defaultConfig';

export { FeatureFlag } from './defaultConfig';

// To turn on flag for local development only - update flag value here
// REMOVE change prior to commit
let runtimeConfig: FeatureFlagConfig = {
  ...defaultFlagConfig,
  // 'test-flag-true': true,  <== locally turns flag on
};

interface FeatureFlagState {
  flags: FeatureFlagConfig;
  setFeatureFlag: (flag: FeatureFlag, newValue: boolean) => void;
  getFeatureFlag: (flag: FeatureFlag) => boolean;
}

interface FeatureFlagProviderProps {
  children?: React.ReactNode;
}

/** FeatureFlagContext - used only if ContextProvider wasn't initialized */
const FeatureFlagContext = createContext<FeatureFlagState>({
  flags: defaultFlagConfig,
  setFeatureFlag: () => {
    /* Provider is not initialized */
  },
  getFeatureFlag: () => false,
});

/** useFeatureFlag - should be used to get flag value */
export const useFeatureFlag = (flag: FeatureFlag) =>
  useContext(FeatureFlagContext).getFeatureFlag(flag);

/** useFeatureFlagContext - could be used to set flags from code */
export const useFeatureFlagContext = () => useContext(FeatureFlagContext);

/** FeatureFlagsProvider - should wrap top level component for Production or feature flag related testing */
export const FeatureFlagsProvider = (props: FeatureFlagProviderProps) => {
  const [flags, setFlags] = useState<FeatureFlagConfig>({
    ...defaultFlagConfig,
    ...runtimeConfig,
  });

  const setFeatureFlag = useCallback((flag: FeatureFlag, newValue: boolean) => {
    runtimeConfig[flag] = newValue;
    setFlags({ ...defaultFlagConfig, ...runtimeConfig });
  }, []);

  const getFeatureFlag = useCallback(
    (flag: FeatureFlag) => {
      if (isDevEnv() && flags[flag] === undefined) {
        throw `Default config value is absent for ${flag}`;
      }
      return flags[flag] ?? false;
    },
    [flags],
  );

  const clearRuntimeConfig = useCallback(() => {
    runtimeConfig = { ...defaultFlagConfig };
  }, []);

  useEffect(() => {
    if (isDevEnv() || isTestEnv()) {
      // Allow manual change of feature flags from devtools
      window['setFeatureFlag'] = setFeatureFlag;
      window['getFeatureFlag'] = getFeatureFlag;
      if (isTestEnv()) {
        // allow reset flags to default - should be used in testing environment only
        window['clearRuntimeConfig'] = clearRuntimeConfig;
      }
    }
  }, [setFeatureFlag, getFeatureFlag, clearRuntimeConfig]);

  return (
    <FeatureFlagContext.Provider value={{ flags, setFeatureFlag, getFeatureFlag }}>
      {props.children}
    </FeatureFlagContext.Provider>
  );
};
