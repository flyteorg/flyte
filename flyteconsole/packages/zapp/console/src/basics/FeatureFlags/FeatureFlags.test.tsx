import * as React from 'react';
import { render, screen, waitFor, act } from '@testing-library/react';

import { useAdminVersion } from 'components/hooks/useVersion';
import { FeatureFlagsProvider, useFeatureFlag } from '.';
import { AdminFlag, FeatureFlag } from './defaultConfig';
import { useIsEnabledInAdmin } from './AdminFlag';

jest.mock('components/hooks/useVersion');

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

describe('AdminFlags', () => {
  const mockAdminVersion = useAdminVersion as jest.Mock<ReturnType<typeof useAdminVersion>>;
  it('useIsEnabledInAdmin returns FALSE if flag is not initialized', () => {
    mockAdminVersion.mockReturnValue({ adminVersion: '0.1.23' });
    const isAdminEnabled = useIsEnabledInAdmin(AdminFlag.TestFlagUndefined);

    expect(isAdminEnabled).toBeFalsy();
  });

  it('useIsEnabledInAdmin returns FALSE if current MINOR version is below required', () => {
    mockAdminVersion.mockReturnValue({ adminVersion: '1.2.3' });
    const isAdminEnabled = useIsEnabledInAdmin(AdminFlag.TestFlagUndefined);

    expect(isAdminEnabled).toBeFalsy();
  });

  it('useIsEnabledInAdmin returns FALSE if current MAJOR version is below required', () => {
    mockAdminVersion.mockReturnValue({ adminVersion: '0.3.45' });
    const isAdminEnabled = useIsEnabledInAdmin(AdminFlag.TestFlagUndefined);

    expect(isAdminEnabled).toBeFalsy();
  });

  it('useIsEnabledInAdmin return TRUE when current version equals required', () => {
    mockAdminVersion.mockReturnValue({ adminVersion: '1.2.34' });
    const isAdminEnabled = useIsEnabledInAdmin(AdminFlag.TestAdminVersion);

    expect(isAdminEnabled).toBeTruthy();
  });

  it('useIsEnabledInAdmin return TRUE when current MINOR version above required', () => {
    mockAdminVersion.mockReturnValue({ adminVersion: '1.2.37' });
    const isAdminEnabled = useIsEnabledInAdmin(AdminFlag.TestAdminVersion);

    expect(isAdminEnabled).toBeTruthy();
  });

  it('useIsEnabledInAdmin return TRUE when current MAJOR version above required', () => {
    mockAdminVersion.mockReturnValue({ adminVersion: '2.1.1' });
    const isAdminEnabled = useIsEnabledInAdmin(AdminFlag.TestAdminVersion);

    expect(isAdminEnabled).toBeTruthy();
  });
});
