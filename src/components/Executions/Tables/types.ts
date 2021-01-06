import {
    Execution,
    NodeExecution,
    NodeExecutionIdentifier
} from 'models/Execution';

export interface NodeExecutionsTableState {
    selectedExecution?: NodeExecutionIdentifier | null;
    setSelectedExecution: (
        selectedExecutionId: NodeExecutionIdentifier | null
    ) => void;
}

type LabelFn = () => JSX.Element;
export interface ColumnDefinition<CellRendererData> {
    cellRenderer(data: CellRendererData): React.ReactNode;
    className?: string;
    key: string;
    label: string | React.FC;
}

export interface NodeExecutionCellRendererData {
    execution: NodeExecution;
    state: NodeExecutionsTableState;
}
export type NodeExecutionColumnDefinition = ColumnDefinition<
    NodeExecutionCellRendererData
>;

export interface WorkflowExecutionCellRendererData {
    execution: Execution;
    state: NodeExecutionsTableState;
}
export type WorkflowExecutionColumnDefinition = ColumnDefinition<
    WorkflowExecutionCellRendererData
>;
