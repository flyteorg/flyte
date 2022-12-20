import { NodeExecutionIdentifier } from 'models/Execution/types';
import { createContext } from 'react';

export interface DetailsPanelContextData {
  selectedExecution?: NodeExecutionIdentifier | null;
  setSelectedExecution: (selectedExecutionId: NodeExecutionIdentifier | null) => void;
}

export const DetailsPanelContext = createContext<DetailsPanelContextData>(
  {} as DetailsPanelContextData,
);
