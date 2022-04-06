/**
 * Default Feature Flag config - used for features in developement.
 */

export enum FeatureFlag {
  // Test flag is created only for unit-tests
  TestFlagUndefined = 'test-flag-undefined',
  TestFlagTrue = 'test-flag-true',

  // Production flags
  LaunchPlan = 'launch-plan',
}

export type FeatureFlagConfig = { [k: string]: boolean };

export const defaultFlagConfig: FeatureFlagConfig = {
  // Test
  'test-flag-true': true,

  // Production - new code should be turned off by default
  // If you need to turn it on locally -> update runtimeConfig in ./index.tsx file
  'launch-plan': false,
};

export interface AdminVersion {
  major: number; // Int - major version
  minor: number; // Float - minor.patch version
}

export enum AdminFlag {
  // Test flag is created only for unit-tests
  TestAdminVersion = 'admin.test-version',
  TestFlagUndefined = 'admin.test-undefined',

  // Production flags
  MapTasks = 'admin.map-tasks', // 0.6.126',
}

export type AdminFlagConfig = { [k: string]: AdminVersion };
export const baseAdminConfig: AdminFlagConfig = {
  'admin.test-version': { major: 1, minor: 2.34 },

  // Production flags
  // Specified and upper versions are treated as ON, lower version turn feature OFF
  'admin.map-tasks': { major: 0, minor: 6.126 },
};
