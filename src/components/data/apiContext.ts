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
        setExpired: () => {}
    }
};

/** Exposes all of the model layer api functions for use by data fetching
 * hooks. Using this context is preferred over directly importing the api functions,
 * as this will allow mocking of the API in stories/tests.
 */
export const APIContext = React.createContext<APIContextValue>(
    defaultAPIContextValue
);

export function useLoginStatus(): LoginStatus {
    const [expired, setExpired] = React.useState(false);
    return {
        expired,
        setExpired
    };
}

export function useAPIState(): APIContextValue {
    return {
        ...defaultAPIContextValue,
        loginStatus: useLoginStatus()
    };
}

export function useAPIContext() {
    return React.useContext(APIContext);
}
