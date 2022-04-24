import { createMockWorkflow, createMockWorkflowClosure } from 'models/__mocks__/workflowData';
import { CompiledWorkflow, Workflow } from 'models/Workflow/types';
import { extractAndIdentifyNodes, extractTaskTemplates } from '../utils';

describe('hooks/utils', () => {
  let workflow: Workflow;
  let subWorkflow: CompiledWorkflow;

  beforeEach(() => {
    workflow = {
      ...createMockWorkflow('SampleWorkflow'),
      closure: createMockWorkflowClosure(),
    };
    const { template } = workflow.closure!.compiledWorkflow!.primary;
    subWorkflow = {
      // We don't process connections in these functions, so it can be empty
      connections: { downstream: {}, upstream: {} },
      template: {
        id: { ...template.id, name: `${template.id.name}_inner` },
        nodes: template.nodes.map((node) => ({
          ...node,
          id: `${node.id}_inner`,
        })),
      },
    };
    workflow.closure!.compiledWorkflow!.subWorkflows = [subWorkflow];
  });

  describe('extractAndIdentifyNodes', () => {
    it('returns empty for missing closure', () => {
      delete workflow.closure;
      expect(extractAndIdentifyNodes(workflow)).toEqual([]);
    });

    it('returns empty for missing compiledWorkflow', () => {
      delete workflow.closure?.compiledWorkflow;
      expect(extractAndIdentifyNodes(workflow)).toEqual([]);
    });

    it('includes nodes from subWorkflows', () => {
      const nodes = extractAndIdentifyNodes(workflow);
      subWorkflow.template.nodes.forEach((node) =>
        expect(nodes).toContainEqual(
          expect.objectContaining({
            id: { nodeId: node.id, workflowId: expect.anything() },
          }),
        ),
      );
    });

    it('assigns parent workflow id to subworkflow nodes', () => {
      const nodes = extractAndIdentifyNodes(workflow);
      subWorkflow.template.nodes.forEach((_node) =>
        expect(nodes).toContainEqual(
          expect.objectContaining({
            id: {
              nodeId: expect.anything(),
              workflowId: workflow.id,
            },
          }),
        ),
      );
    });
  });

  describe('extractTaskTemplates', () => {
    it('returns empty for missing closure', () => {
      delete workflow.closure;
      expect(extractTaskTemplates(workflow)).toEqual([]);
    });
    it('returns empty for missing compiledWorkflow', () => {
      delete workflow.closure?.compiledWorkflow;
      expect(extractTaskTemplates(workflow)).toEqual([]);
    });
  });
});
