import * as React from 'react';

import { Execution, NodeExecution } from 'models';

export interface ExecutionContextData {
    execution: Execution;
    terminateExecution(cause: string): Promise<void>;
}
export const ExecutionContext = React.createContext<ExecutionContextData>(
    {} as ExecutionContextData
);
export const NodeExecutionsContext = React.createContext<
    Dictionary<NodeExecution>
>({});
