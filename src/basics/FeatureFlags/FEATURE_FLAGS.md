## Feature Flags

We are using our internal feature flag solution to allow continuos integration, while features are in development. There are two types of flags:

FeatureFlag: boolean flags which indicate if feature is enabled.
AdminFlag: the minimal version of flyteadmin in which feature supported.
All flags currently available could be found in /FeatureFlags/defaultConfig.ts file. Most of them under active development, which means we don't guarantee it will work as you expect.

If you want to add your own flag, you need to add it to both enum FeatureFlag and defaultFlagConfig under production section. Initally all flags must be disabled, meaning you code path should not be executed by default

**Example - adding flags:**

```javascript
enum FeatureFlags {
    ...
    AddNewPage: 'add-new-page'
    UseCommonPath: 'use-common-path'
}

export const defaultFlagConfig: FeatureFlagConfig = {
    ...
    'add-new-page': false, // default/prior behavior doesn't include new page
    'use-common-path': true, // default/prior behavior uses common path
};
```

To use flags in code you need to ensure that the most top level component is wrapped by `FeatureFlagsProvider`.
By default we are wrapping top component in Apps file, so if you do not plan to include
feature flags checks in the `\*.tests.tsx` - you should be good to go.
To check flag's value use `useFeatureFlag` hook.

**Example - flag usage**:

```javascript
import { FeatureFlag, useFeatureFlag } from 'basics/FeatureFlags';

export function MyComponent(props: Props): React.ReactNode {
    ...
    const isFlagEnabled = useFeatureFlag(FeatureFlag.AddNewPage);

    return isFlagEnabled ? <NewPage ...props/> : null;
}
```

During your local development you can either:

-   temporarily switch flags value in runtimeConfig as:
    ```javascript
    let runtimeConfig = {
        ...defaultFlagConfig,
        'add-new-page': true,
    };
    ```
-   turn flag on/off from the devTools console in Chrome
    ![SetFeatureFlagFromConsole](https://user-images.githubusercontent.com/55718143/150002962-f12bbe57-f221-4bbd-85e3-717aa0221e89.gif)

#### Unit tests

If you plan to test non-default flag value in your unit tests, make sure to wrap your component with `FeatureFlagsProvider`.
Use `window.setFeatureFlag(flag, newValue)` function to set needed value and `window.clearRuntimeConfig()`
to return to defaults. Beware to comment out/remove any changes in `runtimeConfig` during testing;

```javascript
function TestWrapper() {
    return <FeatureFlagsProvider> <TestContent /> </FeatureFlagsProvider>
}

describe('FeatureFlags', () => {
    afterEach(() => {
        window.clearRuntimeConfig(); // clean up flags
    });

   it('Test', async () => {
        render(<TestWrapper />);

        window.setFeatureFlag(FeatureFlag.FlagInQuestion, true);
        await waitFor(() => {
            // check after flag changed value
        });
    });
```
