## @flyteconsole/flyte-api

This package provides ability to do FlyteAdmin API calls from JS/TS code.

At this point it allows to get though authentication steps, request user profile and FlyteAdmin version.
In future releases we will add ability to do all types of FlyteAdmin API calls.

### Installation

To install the package please run:
```bash
yarn add @flyteconsole/flyte-api
```

### Usage

To use in your application

- Wrap parent component with <FlyteApiProvider flyteApiDomain={ADMIN_API_URL ?? ''}>
 
`ADMIN_API_URL` is a flyte admin domain URL to which `/api/v1/_endpoint` part would be added, to perform REST API call.
  `
Then from any child component

```js
import useAxios from 'axios-hooks';
import { useFlyteApi, defaultAxiosConfig } from '@flyteconsole/flyte-api';

...
/** Get profile information */
const apiContext = useFlyteApi();

const profilePath = apiContext.getProfileUrl();
const [{ data: profile, loading }] = useAxios({
  url: profilePath,
  method: 'GET',
  ...defaultAxiosConfig,
});
```
