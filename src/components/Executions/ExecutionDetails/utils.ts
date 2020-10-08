import { Execution, Identifier, ResourceType } from 'models';
import { Routes } from 'routes';

export function isSingleTaskExecution(execution: Execution) {
    return execution.spec.launchPlan.resourceType === ResourceType.TASK;
}

export function getExecutionSourceId(execution: Execution): Identifier {
    return isSingleTaskExecution(execution)
        ? execution.spec.launchPlan
        : execution.closure.workflowId;
}

export function getExecutionBackLink(execution: Execution): string {
    const { project, domain, name } = getExecutionSourceId(execution);
    return isSingleTaskExecution(execution)
        ? Routes.TaskDetails.makeUrl(project, domain, name)
        : Routes.WorkflowDetails.makeUrl(project, domain, name);
}
