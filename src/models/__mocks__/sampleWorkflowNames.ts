import { Admin, Core } from 'flyteidl';
import { NamedEntity, NamedEntityIdentifier } from 'models/Common/types';

export const sampleWorkflowIds: NamedEntityIdentifier[] = [
    'batch_workflow.BatchTasksWorkflow',
    'failing_workflows.DivideByZeroWf',
    'failing_workflows.RetrysWf',
    'failing_workflows.SleeperWf',
    'fancy.CustomContainerWorkflow',
    'fancy.WorkflowWithArbitraryFlyteTask',
    'hive_workflow.ExampleQueryWorkflow',
    'rich_workflow.RichWorkflow',
    'work.OnePythonTaskWF',
    'work.WorkflowWithIO',
    'simple_functional',
    'workflows-batch-batch-tasks-workflow',
    'workflows.batch.BatchTasksWorkflow',
    'workflows-gpu-simple-workflow',
    'workflows.gpu.SimpleWorkflow',
    'workflows-hive-example-query-workflow',
    'workflows.hive.ExampleQueryWorkflow',
    'workflows-nested-child',
    'workflows.nested.Child',
    'workflows-nested-parent',
    'workflows.nested.Parent',
    'workflows-notifications-basic-workflow',
    'workflows.notifications.BasicWorkflow',
    'workflows-python-python-tasks-workflow',
    'workflows.python.PythonTasksWorkflow'
].map(name => ({ name, project: 'flytekit', domain: 'development' }));

export const sampleWorkflowNames: NamedEntity[] = sampleWorkflowIds.map(id => ({
    id,
    resourceType: Core.ResourceType.WORKFLOW,
    metadata: {
        description: `A description for ${id.name}`,
        state: Admin.NamedEntityState.NAMED_ENTITY_ACTIVE
    }
}));
