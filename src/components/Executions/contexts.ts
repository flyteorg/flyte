import * as React from 'react';

import { Execution, NodeExecution, RequestConfig } from 'models';

export interface ExecutionContextData {
    execution: Execution;
}

export const ExecutionContext = React.createContext<ExecutionContextData>(
    {} as ExecutionContextData
);
export const NodeExecutionsContext = React.createContext<
    Dictionary<NodeExecution>
>({});

export const NodeExecutionsRequestConfigContext = React.createContext<
    RequestConfig
>({});
