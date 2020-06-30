import * as React from 'react';

import { Execution, NodeExecution, RequestConfig } from 'models';
import { ExecutionDataCache } from './types';

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

export const NodeExecutionsRequestConfigContext = React.createContext<
    RequestConfig
>({});

export const ExecutionDataCacheContext = React.createContext<
    ExecutionDataCache
>({} as ExecutionDataCache);
