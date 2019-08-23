import { TaskTemplate, Workflow } from 'models';

export function extractTaskTemplates(workflow: Workflow): TaskTemplate[] {
    if (!workflow.closure || !workflow.closure.compiledWorkflow) {
        return [];
    }
    return workflow.closure.compiledWorkflow.tasks.map(t => t.template);
}
