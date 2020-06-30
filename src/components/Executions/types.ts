import {
    CompiledNode,
    GloballyUniqueNode,
    Identifier,
    NodeId,
    RequestConfig,
    Workflow,
    WorkflowId
} from 'models';
import {
    Execution,
    NodeExecution,
    NodeExecutionIdentifier,
    TaskExecution,
    TaskExecutionIdentifier,
    WorkflowExecutionIdentifier
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

/** An interface combining a NodeExecution with data pulled from the
 * corresponding Workflow Node structure.
 */
export interface DetailedNodeExecution extends NodeExecution {
    displayType: NodeExecutionDisplayType;
    displayId: string;
    cacheKey: string;
    taskTemplate?: TaskTemplate;
}

export interface DetailedTaskExecution extends TaskExecution {
    cacheKey: string;
}

export interface NodeExecutionGroup {
    name: string;
    nodeExecutions: NodeExecution[];
}

export interface DetailedNodeExecutionGroup extends NodeExecutionGroup {
    nodeExecutions: DetailedNodeExecution[];
}

export interface ExecutionDataCache {
    getNode(id: NodeId): GloballyUniqueNode | undefined;
    getNodeForNodeExecution(
        nodeExecutionId: NodeExecutionIdentifier
    ): GloballyUniqueNode | null | undefined;
    getNodeExecutions(
        workflowExecutionId: WorkflowExecutionIdentifier,
        config: RequestConfig
    ): Promise<NodeExecution[]>;
    getTaskExecutions(
        nodeExecutionId: NodeExecutionIdentifier
    ): Promise<TaskExecution[]>;
    getTaskExecutionChildren: (
        taskExecutionId: TaskExecutionIdentifier,
        config: RequestConfig
    ) => Promise<NodeExecution[]>;
    getTaskTemplate: (taskId: Identifier) => TaskTemplate | undefined;
    getWorkflow: (workflowId: Identifier) => Promise<Workflow>;
    getWorkflowExecution: (
        executionId: WorkflowExecutionIdentifier
    ) => Promise<Execution>;
    getWorkflowIdForWorkflowExecution: (
        executionId: WorkflowExecutionIdentifier
    ) => Promise<WorkflowId>;
    insertExecution(execution: Execution): void;
    insertNodes(nodes: GloballyUniqueNode[]): void;
    insertTaskTemplates(templates: TaskTemplate[]): void;
    insertWorkflow(workflow: Workflow): void;
    insertWorkflowExecutionReference(
        executionId: WorkflowExecutionIdentifier,
        workflowId: WorkflowId
    ): void;
}
