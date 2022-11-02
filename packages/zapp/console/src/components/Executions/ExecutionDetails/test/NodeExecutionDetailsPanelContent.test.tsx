import { render, waitFor } from '@testing-library/react';
import { NodeExecutionDetailsContextProvider } from 'components/Executions/contextProvider/NodeExecutionDetails';
import { basicPythonWorkflow } from 'mocks/data/fixtures/basicPythonWorkflow';
import { mockWorkflowId } from 'mocks/data/fixtures/types';
import { insertFixture } from 'mocks/data/insertFixture';
import { mockServer } from 'mocks/server';
import { TaskExecutionPhase } from 'models/Execution/enums';
import { NodeExecution } from 'models/Execution/types';
import * as React from 'react';
import { QueryClient, QueryClientProvider } from 'react-query';
import { MemoryRouter } from 'react-router';
import { createTestQueryClient } from 'test/utils';
import { NodeExecutionDetailsPanelContent } from '../NodeExecutionDetailsPanelContent';

jest.mock('components/Workflow/workflowQueries');
const { fetchWorkflow } = require('components/Workflow/workflowQueries');

describe('NodeExecutionDetailsPanelContent', () => {
  let fixture: ReturnType<typeof basicPythonWorkflow.generate>;
  let execution: NodeExecution;
  let queryClient: QueryClient;

  beforeEach(() => {
    fixture = basicPythonWorkflow.generate();
    execution = fixture.workflowExecutions.top.nodeExecutions.pythonNode.data;
    insertFixture(mockServer, fixture);
    fetchWorkflow.mockImplementation(() => Promise.resolve(fixture.workflows.top));
    queryClient = createTestQueryClient();
  });

  const renderComponent = () =>
    render(
      <MemoryRouter>
        <QueryClientProvider client={queryClient}>
          <NodeExecutionDetailsContextProvider workflowId={mockWorkflowId}>
            <NodeExecutionDetailsPanelContent
              nodeExecutionId={execution.id}
              taskPhase={TaskExecutionPhase.UNDEFINED}
            />
          </NodeExecutionDetailsContextProvider>
        </QueryClientProvider>
      </MemoryRouter>,
    );

  it('renders name for task nodes', async () => {
    const { name } = fixture.tasks.python.id;
    const { getByText } = renderComponent();
    await waitFor(() => expect(getByText(name)).toBeInTheDocument());
  });
});
