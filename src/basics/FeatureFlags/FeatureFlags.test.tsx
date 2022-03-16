import * as React from 'react';
import { render, screen, waitFor, act } from '@testing-library/react';

import { FeatureFlagsProvider, useFeatureFlag } from '.';
import { FeatureFlag } from './defaultConfig';

function TestContent() {
  const enabledTestFlag = useFeatureFlag(FeatureFlag.TestFlagUndefined);
  return (
    <FeatureFlagsProvider>
      <button>{enabledTestFlag ? 'Enabled' : 'Disabled'}</button>
    </FeatureFlagsProvider>
  );
}

function TestPage() {
  return (
    <FeatureFlagsProvider>
      <TestContent />
    </FeatureFlagsProvider>
  );
}

declare global {
  interface Window {
    setFeatureFlag: (flag: FeatureFlag, newValue: boolean) => void;
    getFeatureFlag: (flag: FeatureFlag) => boolean;
    clearRuntimeConfig: () => void;
  }
}

describe('FeatureFlags', () => {
  beforeEach(() => {
    render(<TestPage />);
  });

  afterEach(() => {
    window.clearRuntimeConfig();
  });

  it('Feature flags can be read/set from dev tools', async () => {
    // flag defined and return proper value
    expect(window.getFeatureFlag(FeatureFlag.TestFlagTrue)).toBeTruthy();
    // flag undefined and returns false
    expect(window.getFeatureFlag(FeatureFlag.TestFlagUndefined)).toBeFalsy();

    act(() => window.setFeatureFlag(FeatureFlag.TestFlagUndefined, true));
    await waitFor(() => {
      // check that flag cghanged value
      expect(window.getFeatureFlag(FeatureFlag.TestFlagUndefined)).toBeTruthy();
    });
  });

  it('useFeatureFlags returns proper live value', async () => {
    // default value - flag is disabled
    expect(screen.getByText(/Disabled/i)).toBeTruthy();

    // Enable flag
    act(() => window.setFeatureFlag(FeatureFlag.TestFlagUndefined, true));
    await waitFor(() => {
      // check that component was updated accordingly
      expect(screen.getByText(/Enabled/i)).toBeTruthy();
    });
  });
});
