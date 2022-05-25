import * as React from 'react';
import { createContext, useContext } from 'react';
import { getAdminApiUrl, getEndpointUrl } from '../utils';
import { AdminEndpoint, RawEndpoint } from '../utils/constants';
import { defaultLoginStatus, getLoginUrl, LoginStatus } from './login';

export interface FlyteApiContextState {
  loginStatus: LoginStatus;
  getLoginUrl: (redirect?: string) => string;
  getProfileUrl: () => string;
  getAdminApiUrl: (endpoint: AdminEndpoint | string) => string;
}

const FlyteApiContext = createContext<FlyteApiContextState>({
  // default values - used when Provider wrapper is not found
  loginStatus: defaultLoginStatus,
  getLoginUrl: () => '#',
  getProfileUrl: () => '#',
  getAdminApiUrl: () => '#',
});

interface FlyteApiProviderProps {
  flyteApiDomain?: string;
  children?: React.ReactNode;
}

export const useFlyteApi = () => useContext(FlyteApiContext);

export const FlyteApiProvider = (props: FlyteApiProviderProps) => {
  const { flyteApiDomain } = props;

  const [loginExpired, setLoginExpired] = React.useState(false);

  // Whenever we detect expired credentials, trigger a login redirect automatically
  React.useEffect(() => {
    if (loginExpired) {
      window.location.href = getLoginUrl(flyteApiDomain);
    }
  }, [loginExpired]);

  return (
    <FlyteApiContext.Provider
      value={{
        loginStatus: {
          expired: loginExpired,
          setExpired: setLoginExpired,
        },
        getLoginUrl: (redirect) => getLoginUrl(flyteApiDomain, redirect),
        getProfileUrl: () => getEndpointUrl(RawEndpoint.Profile, flyteApiDomain),
        getAdminApiUrl: (endpoint) => getAdminApiUrl(endpoint, flyteApiDomain),
      }}
    >
      {props.children}
    </FlyteApiContext.Provider>
  );
};
