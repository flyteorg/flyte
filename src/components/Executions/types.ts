import { NodeExecution, TaskExecution } from 'models/Execution/types';
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
