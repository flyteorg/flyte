import {
    Execution,
    NodeExecution,
    TaskExecution
} from 'models/Execution/types';
import { LaunchPlan } from 'models/Launch/types';
import { Task } from 'models/Task/types';
import { Workflow } from 'models/Workflow/types';
import { Identifier } from 'models/Common/types';

export const mockWorkflowId: Identifier = {
    project: 'project',
    domain: 'domain',
    name: 'specific.workflow.name_wf',
    version: '0.1'
};

/** Represents a TaskExecution and its associated children. */
export interface MockTaskExecutionData {
    data: TaskExecution;
    // Legacy Dynamic Task executions may yield additional NodeExecutions at runtime.
    nodeExecutions?: Record<string, MockNodeExecutionData>;
}

/** Represents a NodeExecution record and its associated children. */
export interface MockNodeExecutionData {
    data: NodeExecution;
    // NodeExecutions may be parents of other NodeExecutions
    nodeExecutions?: Record<string, MockNodeExecutionData>;
    // NodeExecutions which executed a Task will have child
    // TaskExecutions.
    taskExecutions?: Record<string, MockTaskExecutionData>;
}

/** Represents a WorkflowExecution record and its associated children. */
export interface MockWorkflowExecutionData {
    data: Execution;
    nodeExecutions: Record<string, MockNodeExecutionData>;
}

/** Contains maps of objects to be inserted into an `AdminServer` instance.
 * All are optional (i.e. you can create a fixture containing only a list
 * of tasks).
 */
export interface MockDataFixture {
    launchPlans?: Record<string, LaunchPlan>;
    tasks?: Record<string, Task>;
    workflows?: Record<string, Workflow>;
    workflowExecutions?: Record<string, MockWorkflowExecutionData>;
}
