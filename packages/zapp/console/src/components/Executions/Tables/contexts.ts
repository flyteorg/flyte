import * as React from 'react';
import { NodeExecutionColumnDefinition, NodeExecutionsTableState } from './types';

export interface NodeExecutionsTableContextData {
  columns: NodeExecutionColumnDefinition[];
  state: NodeExecutionsTableState;
}

export const NodeExecutionsTableContext = React.createContext<NodeExecutionsTableContextData>(
  {} as NodeExecutionsTableContextData,
);
