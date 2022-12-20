import * as React from 'react';
import { render, waitFor } from '@testing-library/react';
import { NodeExecutionDetailsContextProvider } from 'components/Executions/contextProvider/NodeExecutionDetails';
import { mockWorkflowId } from 'mocks/data/fixtures/types';
import { QueryClient, QueryClientProvider } from 'react-query';
import { createTestQueryClient } from 'test/utils';
import { insertFixture } from 'mocks/data/insertFixture';
import { mockServer } from 'mocks/server';
import { basicPythonWorkflow } from 'mocks/data/fixtures/basicPythonWorkflow';
import { NodeExecution } from 'models/Execution/types';
import { dTypes } from 'models/Graph/types';
import { NodeExecutionRow } from '../NodeExecutionRow';

jest.mock('components/Workflow/workflowQueries');
const { fetchWorkflow } = require('components/Workflow/workflowQueries');

jest.mock('components/Executions/Tables/RowExpander', () => ({
  RowExpander: jest.fn(() => <div data-testid="expander"></div>),
}));

const columns = [];
const node = {
  id: 'n1',
  scopedId: 'n1',
  type: dTypes.start,
  name: 'node1',
  nodes: [],
  edges: [],
};
const onToggle = jest.fn();

describe('Executions > Tables > NodeExecutionRow', () => {
  let queryClient: QueryClient;
  let fixture: ReturnType<typeof basicPythonWorkflow.generate>;
  let execution: NodeExecution;

  beforeEach(() => {
    fixture = basicPythonWorkflow.generate();
    execution = fixture.workflowExecutions.top.nodeExecutions.pythonNode.data;
    queryClient = createTestQueryClient();
    insertFixture(mockServer, fixture);
    fetchWorkflow.mockImplementation(() => Promise.resolve(fixture.workflows.top));
  });

  const renderComponent = (props) =>
    render(
      <QueryClientProvider client={queryClient}>
        <NodeExecutionDetailsContextProvider workflowId={mockWorkflowId}>
          <NodeExecutionRow {...props} />
        </NodeExecutionDetailsContextProvider>
      </QueryClientProvider>,
    );

  it('should not render expander if node is a leaf', async () => {
    const { queryByRole, queryByTestId } = renderComponent({
      columns,
      node,
      nodeExecution: execution,
      onToggle,
    });
    await waitFor(() => queryByRole('listitem'));

    expect(queryByRole('listitem')).toBeInTheDocument();
    expect(queryByTestId('expander')).not.toBeInTheDocument();
  });

  it('should render expander if node contains list of nodes', async () => {
    const mockNode = { ...node, nodes: [node, node] };
    const { queryByRole, queryByTestId } = renderComponent({
      columns,
      node: mockNode,
      nodeExecution: execution,
      onToggle,
    });
    await waitFor(() => queryByRole('listitem'));

    expect(queryByRole('listitem')).toBeInTheDocument();
    expect(queryByTestId('expander')).toBeInTheDocument();
  });
});
