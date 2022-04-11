import { WorkflowListStructureItem } from 'components/Workflow/types';
import { WorkflowExecutionState } from 'models/Workflow/enums';
import { WorkflowId } from 'models/Workflow/types';

export const sampleWorkflowIds: WorkflowId[] = [
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
  'workflows.python.PythonTasksWorkflow',
].map((name) => ({
  name,
  project: 'flytekit',
  domain: 'development',
  version: '',
}));

export const sampleWorkflowNames: WorkflowListStructureItem[] = sampleWorkflowIds.map((id) => ({
  id,
  description: '',
  state: WorkflowExecutionState.NAMED_ENTITY_ACTIVE,
}));
