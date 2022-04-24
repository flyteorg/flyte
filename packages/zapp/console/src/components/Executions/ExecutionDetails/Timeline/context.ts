import { NodeExecutionIdentifier } from 'models/Execution/types';
import * as React from 'react';

export interface NodeExecutionsTimelineContextData {
  selectedExecution?: NodeExecutionIdentifier | null;
  setSelectedExecution: (selectedExecutionId: NodeExecutionIdentifier | null) => void;
}

export const NodeExecutionsTimelineContext = React.createContext<NodeExecutionsTimelineContextData>(
  {} as NodeExecutionsTimelineContextData,
);
