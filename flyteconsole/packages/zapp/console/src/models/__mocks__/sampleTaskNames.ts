import { Core } from 'flyteidl';
import { NamedEntity, NamedEntityIdentifier } from 'models/Common/types';
import { NamedEntityState } from 'models/enums';

export const sampleTaskIds: NamedEntityIdentifier[] = [
  'app.complex_workflows.custom_image.task_object',
  'app.workflows.batch_workflow.int_sub_task',
  'app.workflows.batch_workflow.print_every_time',
  'app.workflows.batch_workflow.sample_batch_task_no_inputs',
  'app.workflows.batch_workflow.sample_batch_task_sq',
  'app.workflows.batch_workflow.sq_sub_task',
  'app.workflows.batch_workflow.sub_task',
  'app.workflows.dynamic_resources.override_cpu_task',
  'app.workflows.edge.canny_detection',
  'app.workflows.edge.demo.canny_detection',
  'app.workflows.failing_workflows.divider',
  'app.workflows.failing_workflows.oversleeper',
  'app.workflows.failing_workflows.retryer',
  'app.workflows.fancy.custom_image_task',
  'app.workflows.fancy.task_object',
  'app.workflows.hive_workflow.always_failiing_hive_task',
  'app.workflows.hive_workflow.dynamic_hive_task_producer',
  'app.workflows.hive_workflow.generate_queries',
  'app.workflows.hive_workflow.get_queries_cached',
  'app.workflows.hive_workflow.get_uncached_query_for_dynamic',
  'app.workflows.hive_workflow.print_schemas',
  'app.workflows.rich_workflow.add_one_and_print',
  'app.workflows.rich_workflow.print_every_time',
  'app.workflows.rich_workflow.print_int',
  'app.workflows.rich_workflow.simple_batch_task',
  'app.workflows.rich_workflow.sum_and_print',
  'app.workflows.rich_workflow.sum_ints',
  'app.workflows.rich_workflow.sum_non_none',
  'app.workflows.sample_tasks.add_one_and_print',
  'app.workflows.sample_tasks.identity_sub_task',
  'app.workflows.sample_tasks.print_every_time',
  'app.workflows.sample_tasks.print_int',
  'app.workflows.sample_tasks.simple_batch_task',
  'app.workflows.sample_tasks.sum_and_print',
  'app.workflows.sample_tasks.sum_ints',
  'app.workflows.sample_tasks.sum_non_none',
  'app.workflows.sidecar.a_sidecar_task',
  'app.workflows.sidecar.a_simple_sidecar_task',
  'app.workflows.sidecar_workflow.my_task',
  'app.workflows.sidecar_workflow.primary_sidecar_task',
  'app.workflows.sidecar_workflow.simulate_hours',
  'app.workflows.simple.add_one',
  'app.workflows.work.find_odd_numbers',
  'app.workflows.work.find_odd_numbers_with_string',
].map((name) => ({ name, project: 'flytekit', domain: 'development' }));

export const sampleTaskNames: NamedEntity[] = sampleTaskIds.map((id) => ({
  id,
  resourceType: Core.ResourceType.TASK,
  metadata: {
    description: `A description for ${id.name}`,
    state: NamedEntityState.NAMED_ENTITY_ACTIVE,
  },
}));
