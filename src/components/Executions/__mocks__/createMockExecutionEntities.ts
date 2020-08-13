import { cloneDeep } from 'lodash';
import {
    createMockWorkflow,
    createMockWorkflowClosure
} from 'models/__mocks__/workflowData';
import { createMockNodeExecutions } from 'models/Execution/__mocks__/mockNodeExecutionsData';
import { mockExecution as mockWorkflowExecution } from 'models/Execution/__mocks__/mockWorkflowExecutionsData';
import { mockTasks } from 'models/Task/__mocks__/mockTaskData';

interface CreateExecutionEntitiesArgs {
    workflowName: string;
    nodeExecutionCount: number;
}

/** Creates the basic entities necessary to render the majority of the
 * ExecutionDetails page. These can be inserted into an ExecutionDataCache
 * for mocking.
 */
export function createMockExecutionEntities({
    workflowName,
    nodeExecutionCount
}: CreateExecutionEntitiesArgs) {
    const { executions: nodeExecutions, nodes } = createMockNodeExecutions(
        nodeExecutionCount
    );

    const workflow = createMockWorkflow(workflowName);
    const workflowClosure = createMockWorkflowClosure();
    const compiledWorkflow = workflowClosure.compiledWorkflow!;
    const {
        primary: { template },
        tasks
    } = compiledWorkflow;
    template.nodes = template.nodes.concat(nodes);
    compiledWorkflow.tasks = tasks.concat(cloneDeep(mockTasks));
    workflow.closure = workflowClosure;

    const workflowExecution = { ...mockWorkflowExecution };
    workflowExecution.closure.workflowId = workflow.id;

    return {
        nodes,
        nodeExecutions,
        tasks,
        workflow,
        workflowExecution
    };
}
