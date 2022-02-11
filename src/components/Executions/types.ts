import {
    NodeExecution,
    NodeExecutionClosure,
    NodeExecutionMetadata,
    WorkflowNodeMetadata
} from 'models/Execution/types';
import { TaskTemplate } from 'models/Task/types';

export interface ExecutionPhaseConstants {
    badgeColor: string;
    text: string;
    textColor: string;
}

export enum NodeExecutionDisplayType {
    ArrayTask = 'Array Task',
    BatchHiveTask = 'Hive Batch Task',
    BranchNode = 'Branch Node',
    DynamicTask = 'Dynamic Task',
    HiveTask = 'Hive Task',
    PythonTask = 'Python Task',
    SidecarTask = 'Sidecar Task',
    SparkTask = 'Spark Task',
    Workflow = 'Sub-Workflow',
    Unknown = 'Unknown',
    UnknownTask = 'Unknown Task',
    WaitableTask = 'Waitable Task',
    MpiTask = 'MPI Task'
}

export interface ParentNodeExecution extends NodeExecution {
    metadata: NodeExecutionMetadata & {
        isParentNode: true;
    };
}

export interface WorkflowNodeExecutionClosure extends NodeExecutionClosure {
    workflowNodeMetadata: WorkflowNodeMetadata;
}

export interface NodeExecutionDetails {
    displayId?: string;
    displayName?: string;
    displayType: string;
    taskTemplate?: TaskTemplate;
}

export interface NodeExecutionGroup {
    name: string;
    nodeExecutions: NodeExecution[];
}
