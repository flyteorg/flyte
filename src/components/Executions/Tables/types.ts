import { Execution } from 'models/Execution';
import { DetailedNodeExecution, DetailedTaskExecution } from '../types';

export interface NodeExecutionsTableState {
    executions: DetailedNodeExecution[];
    selectedExecution?: DetailedNodeExecution | null;
    setSelectedExecution: (execution: DetailedNodeExecution | null) => void;
}

type LabelFn = () => JSX.Element;
export interface ColumnDefinition<CellRendererData> {
    cellRenderer(data: CellRendererData): React.ReactNode;
    className?: string;
    key: string;
    label: string | React.FC;
}

export interface NodeExecutionCellRendererData {
    execution: DetailedNodeExecution;
    state: NodeExecutionsTableState;
}
export type NodeExecutionColumnDefinition = ColumnDefinition<
    NodeExecutionCellRendererData
>;

export interface TaskExecutionCellRendererData {
    execution: DetailedTaskExecution;
    nodeExecution: DetailedNodeExecution;
    state: NodeExecutionsTableState;
}
export type TaskExecutionColumnDefinition = ColumnDefinition<
    TaskExecutionCellRendererData
>;

export interface WorkflowExecutionCellRendererData {
    execution: Execution;
    state: NodeExecutionsTableState;
}
export type WorkflowExecutionColumnDefinition = ColumnDefinition<
    WorkflowExecutionCellRendererData
>;
