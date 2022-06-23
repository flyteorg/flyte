# Flyte Console Contribution Guide

First off all, thank you for thinking about contributing!
Below you’ll find instructions that will hopefully guide you through how to contribute to, fix, and improve Flyte Console.

## Protobuf and Debug Output

Communication with the Flyte Admin API is done using Protobuf as the
request/response format. Protobuf is a binary format, which means looking at
responses in the Network tab won't be very helpful. To make debugging easier,
each network request is logged to the console with it's URL followed by the
decoded Protobuf payload. You must have debug output enabled (on by default in
development) to see these messages.

This application makes use of the [debug](https://github.com/visionmedia/debug)
libary to provide namespaced **debug output** in the browser console. In
development, all debug output is enabled. For other environments, the debug
output must be enabled manually. You can do this by setting a flag in
localStorage using the console: `localStorage.debug = 'flyte:*'`. Each module in
the application sets its own namespace. So if you'd like to only view output for
a single module, you can specify that one specifically
(ex. `localStorage.debug = 'flyte:adminEntity'` to only see decoded Flyte
Admin API requests).

## Generate new package

To add a new package use a script

```bash
yarn generate:package
```

After new package is generated, you will need to update some values to be able to use it with other packages.
For example in case if package plan to be used in `console` app

Ensure to add proper webpack alias path resolutions into:
* ./storybook/main.js -  as `'@flyteconsole/flyte-api': path.resolve(__dirname, '../packages/plugins/flyte-api/src’),`
* packages/zapp/console/webpack.common.config.ts to alias section -  as `'@flyteconsole/flyte-api': path.resolve(__dirname, '../packages/plugins/flyte-api/src’),`

To add child package usage to other package, in parent package ->
* Add `{ "path": “../../${type}/${package-name}" }` to tsconfig.json
* Add `{ "path": “../../${type}/${package-name}/tsconfig.build.json" }` to tsconfig.build.json (if exists)
- Then you can import your changes as `import { getLoginUrl } from '@flyteconsole/flyte-api’;`

> If you see `yarn lint` package not defined issues update `.\eslintrc.js` by adding your package to 
    'import/core-modules': ['@clients/locale', '@clients/primitives', '@clients/theme'],


## Storybook

This project has support for [Storybook](https://storybook.js.org/).
Component stories live next to the components they test, in a `__stories__`
directory, with the filename pattern `{Component}.stories.tsx`.

You can run storybook with `yarn run storybook`, and view the stories at http://localhost:9001.

## Feature flags

We are using our internal feature flag solution to allow continuos integration,
while features are in development. There are two types of flags:

-   **FeatureFlag**: boolean flags which indicate if feature is enabled.
-   **AdminFlag**: the minimal version of flyteadmin in which feature supported.

All flags currently available could be found in [/FeatureFlags/defaultConfig.ts](./src/basics/FeatureFlags/defaultConfig.ts)
file. Most of them under active development, which means we don't guarantee it will work as you expect.

If you want to add your own flag, you need to add it to both `enum FeatureFlag` and `defaultFlagConfig`
under production section.
Initally all flags must be disabled, meaning you code path should not be executed by default.

**Example - flag usage**:

```javascript
import { FeatureFlag, useFeatureFlag } from 'basics/FeatureFlags';

export function MyComponent(props: Props): React.ReactNode {
    ...
    const isFlagEnabled = useFeatureFlag(FeatureFlag.AddNewPage);

    return isFlagEnabled ? <NewPage ...props/> : null;
}
```

More info in [FEATURE_FLAGS.md](src/basics/FeatureFlags/FEATURE_FLAGS.md)

## Local storage

We allow to save user "settings" choice to the browser Local Storage, to persist specific values between sessions. The local storage entry is stored as a JSON string, so can represent any object. However, it is a good practise to minimize your object fields prior to storing.

All available LocalCacheItems could be found in [/LocalCache/defaultConfig.ts](./src/basics/LocalCache/defaultConfig.ts). We are using `flyte.` prefix in items which are storing user settings.

**Example - flag usage**:

```javascript
import { LocalCacheItem, useLocalCache } from 'basics/LocalCache';

export function MyComponent(props: Props): React.ReactNode {
    ...
    const [showTable, setShowTable] = useLocalCache(LocalCacheItem.ShowWorkflowVersions);

    return showTable ? <SomeComponent ...props onClick={() => setShowTable(!showTable)}/> : null;
}
```

## Unit tests

You can run unit tests locally, for both of the command listed below `NODE_ENV=test` is set-up, so if you need a specific error/log/mock treatment for these cases feel free to use isTestEnv() check

```
import { isTestEnv } from 'common/env';
...
if (isTestEnv()) {...}
```

To run unit tests locally: `yarn test`
To check coverage `yarn test-coverage`

## Google Analytics

This application makes use of the `react-ga4 <https://github.com/PriceRunner/react-ga4>` libary to include Google Analytics tracking code in a website or app. For all the environments, it is configured using ENABLE_GA environment variable.
By default, it's enabled like this: `ENABLE_GA=true`. If you want to disable it, just set it false. (ex. `ENABLE_GA=false`).

## 📦 Install Dependencies

Running flyteconsole locally requires [NodeJS](https://nodejs.org) and
[yarn](https://yarnpkg.com). We recommend for you to use **asdf** to manage NodeJS version.
You can find currently used versions in `.tool-versions` file.

-   Install asdf through homebrew

```bash
brew install asdf
```

-   (Optional) If you are using **M1 MacBook**, you will need to install `vips` for proper build experience

```bash
brew install vips
```

-   Add Yarn and NodeJs plugins to asdf, to manage versions

```bash
asdf plugin add nodejs https://github.com/asdf-vm/asdf-nodejs.git
asdf plugin-add yarn https://github.com/twuni/asdf-yarn.git
brew install gpg
```

-   From flyteconsole directory - install proper NodeJS and yarn versions:

```bash
asdf install
```

-   Install nodepackages

```bash
yarn install
```

1. Set `ADMIN_API_URL` and `ADMIN_API_USE_SSL`

    ```
      export ADMIN_API_URL=https://different.admin.service.com
      export ADMIN_API_USE_SSL="https"
      export LOCAL_DEV_HOST=localhost.different.admin.service.com
    ```

    _NOTE:_ Add these to your local profile (e.g., `./profile`) to prevent having to do this step each time

2. Generate SSL certificate

    Run the following command from your `flyteconsole` directory

    ```bash
      make generate_ssl
    ```

3. Add new record to hosts file

    ```bash
      sudo vim /etc/hosts
    ```

    Add the following record

    ```
      127.0.0.1 localhost.different.admin.service.com
    ```

4. Install Chrome plugin: [Moesif Origin & CORS Changer](https://chrome.google.com/webstore/detail/moesif-origin-cors-change/digfbfaphojjndkpccljibejjbppifbc)

    > _NOTE:_
    > 1. Activate plugin (toggle to "on")
    > 1. Open 'Advanced Settings':
    > - set `Access-Control-Allow-Credentials`: `true`
    > - set `Domain List`: `your.localhost.com`

5. Start `flyteconsole`

    ```bash
      yarn start
    ```

    Your new localhost is [localhost.different.admin.service.com](http://localhost.different.admin.service.com)

> Ensure you don't have `ADMIN_API_URL` or `DISABLE_AUTH` set (e.g., in your `/.profile`.)
