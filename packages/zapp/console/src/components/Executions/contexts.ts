import { Execution, NodeExecution } from 'models/Execution/types';
import { createContext } from 'react';

export interface ExecutionContextData {
  execution: Execution;
}

export const ExecutionContext = createContext<ExecutionContextData>({} as ExecutionContextData);

export const NodeExecutionsByIdContext = createContext<Dictionary<NodeExecution>>({});
