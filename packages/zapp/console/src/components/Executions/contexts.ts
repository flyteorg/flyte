import { RequestConfig } from 'models/AdminEntity/types';
import { Execution, NodeExecution } from 'models/Execution/types';
import * as React from 'react';

export interface ExecutionContextData {
  execution: Execution;
}

export const ExecutionContext = React.createContext<ExecutionContextData>(
  {} as ExecutionContextData,
);

export const NodeExecutionsByIdContext = React.createContext<Dictionary<NodeExecution>>({});

export const NodeExecutionsRequestConfigContext = React.createContext<RequestConfig>({});
