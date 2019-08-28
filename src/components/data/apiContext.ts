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

export interface APIContextValue extends APIFunctions {}

export const APIContext = React.createContext<APIContextValue>({
    ...CommonAPI,
    ...ExecutionAPI,
    ...LaunchAPI,
    ...ProjectAPI,
    ...TaskAPI,
    ...WorkflowAPI
});

export function useAPIContext() {
    return React.useContext(APIContext);
}
