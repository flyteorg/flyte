import { getLoginUrl } from 'models/AdminEntity/utils';
import * as CommonAPI from 'models/Common/api';
import * as ExecutionAPI from 'models/Execution/api';
import * as LaunchAPI from 'models/Launch/api';
import * as ProjectAPI from 'models/Project/api';
import * as TaskAPI from 'models/Task/api';
import * as WorkflowAPI from 'models/Workflow/api';
import * as React from 'react';

type APIFunctions = typeof CommonAPI &
  typeof ExecutionAPI &
  typeof LaunchAPI &
  typeof ProjectAPI &
  typeof TaskAPI &
  typeof WorkflowAPI;

export interface LoginStatus {
  expired: boolean;
  setExpired(expired: boolean): void;
}
export interface APIContextValue extends APIFunctions {
  loginStatus: LoginStatus;
}

export const defaultAPIContextValue = {
  ...CommonAPI,
  ...ExecutionAPI,
  ...LaunchAPI,
  ...ProjectAPI,
  ...TaskAPI,
  ...WorkflowAPI,
  loginStatus: {
    expired: false,
    setExpired: () => {
      // do nothing
    },
  },
};

/** Exposes all of the model layer api functions for use by data fetching
 * hooks. Using this context is preferred over directly importing the api functions,
 * as this will allow mocking of the API in stories/tests.
 */
export const APIContext = React.createContext<APIContextValue>(defaultAPIContextValue);

function useLoginStatus(): LoginStatus {
  const [expired, setExpired] = React.useState(false);

  // Whenever we detect expired credentials, trigger a login redirect automatically
  React.useEffect(() => {
    if (expired) {
      window.location.href = getLoginUrl();
    }
  }, [expired]);
  return {
    expired,
    setExpired,
  };
}

/** Creates a state object that can be used as the value for APIContext.Provider */
export function useAPIState(): APIContextValue {
  return {
    ...defaultAPIContextValue,
    loginStatus: useLoginStatus(),
  };
}

/** Convenience hook for consuming the `APIContext` */
export function useAPIContext() {
  return React.useContext(APIContext);
}
