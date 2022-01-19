/**
 * Default Feature Flag config - used for features in developement.
 */

export enum FeatureFlag {
    // Test flag is created only for unit-tests
    TestFlagUndefined = 'test-flag-undefined',
    TestFlagTrue = 'test-flag-true',

    // Production flags
    LaunchPlan = 'launch-plan'
}

export type FeatureFlagConfig = { [k: string]: boolean };

export const defaultFlagConfig: FeatureFlagConfig = {
    // Test
    'test-flag-true': true,

    // Production - new code should be turned off by default
    // If you need to turn it on locally -> update runtimeConfig in ./index.tsx file
    'launch-plan': false
};
