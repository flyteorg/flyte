import { GloballyUniqueNode } from 'models/Node/types';
import { TaskTemplate } from 'models/Task/types';
import { Workflow } from 'models/Workflow/types';

export function extractTaskTemplates(workflow: Workflow): TaskTemplate[] {
  if (!workflow.closure || !workflow.closure.compiledWorkflow) {
    return [];
  }
  return workflow.closure.compiledWorkflow.tasks.map((t) => t.template);
}

export function extractAndIdentifyNodes(workflow: Workflow): GloballyUniqueNode[] {
  if (!workflow.closure || !workflow.closure.compiledWorkflow) {
    return [];
  }
  const { primary, subWorkflows = [] } = workflow.closure.compiledWorkflow;
  const nodes = subWorkflows.reduce(
    (out, subWorkflow) => [...out, ...subWorkflow.template.nodes],
    primary.template.nodes,
  );

  return nodes.map((node) => ({
    node,
    id: {
      nodeId: node.id,
      // TODO: This is technically incorrect, as sub-workflow nodes
      // will use the wrong parent workflow id. This is done intentionally
      // to make sure that looking up the node information for a NodeExecution
      // finds the entry successfully.
      // When we are rendering sub-workflow nodes correctly, this should
      // be updated to use the proper parent workflow id
      // (subWorkflow.template.id)
      // See https://github.com/flyteorg/flyte/issues/357
      workflowId: workflow.id,
    },
  }));
}
