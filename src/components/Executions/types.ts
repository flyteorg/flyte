import {
    BranchNode,
    CompiledNode,
    TaskNode,
    WorkflowId,
    WorkflowNode
} from 'models';
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
    WaitableTask = 'Waitable Task'
}

export interface UniqueNodeId {
    workflowId: WorkflowId;
    nodeId: string;
}
export interface NodeInformation {
    id: UniqueNodeId;
    node: CompiledNode;
}

export interface CompiledTaskNode extends CompiledNode {
    taskNode: TaskNode;
}

export interface CompiledWorkflowNode extends CompiledNode {
    workflowNode: WorkflowNode;
}

export interface CompiledBranchNode extends CompiledNode {
    branchNode: BranchNode;
}

export interface ParentNodeExecution extends NodeExecution {
    metadata: NodeExecutionMetadata & {
        isParentNode: true;
    };
}

export interface WorkflowNodeExecutionClosure extends NodeExecutionClosure {
    workflowNodeMetadata: WorkflowNodeMetadata;
}

export interface WorkflowNodeExecution extends NodeExecution {
    closure: WorkflowNodeExecutionClosure;
}

export interface NodeExecutionDetails {
    displayType: NodeExecutionDisplayType;
    displayId: string;
    taskTemplate?: TaskTemplate;
}

export interface NodeExecutionGroup {
    name: string;
    nodeExecutions: NodeExecution[];
}
