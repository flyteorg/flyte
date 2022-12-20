import { render, waitFor } from '@testing-library/react';
import { NodeExecutionDetailsContextProvider } from 'components/Executions/contextProvider/NodeExecutionDetails';
import { NodeExecutionsByIdContext } from 'components/Executions/contexts';
import { basicPythonWorkflow } from 'mocks/data/fixtures/basicPythonWorkflow';
import { noExecutionsFoundString } from 'common/constants';
import { mockWorkflowId } from 'mocks/data/fixtures/types';
import { insertFixture } from 'mocks/data/insertFixture';
import { mockServer } from 'mocks/server';
import { NodeExecutionPhase } from 'models/Execution/enums';
import * as React from 'react';
import { dateToTimestamp } from 'common/utils';
import { QueryClient, QueryClientProvider } from 'react-query';
import { createTestQueryClient } from 'test/utils';
import { dNode } from 'models/Graph/types';
import { useNodeExecutionFiltersState } from 'components/Executions/filters/useExecutionFiltersState';
import { NodeExecutionsTable } from '../NodeExecutionsTable';

jest.mock('components/Workflow/workflowQueries');
const { fetchWorkflow } = require('components/Workflow/workflowQueries');

jest.mock('components/Executions/filters/useExecutionFiltersState');
const mockUseNodeExecutionFiltersState = useNodeExecutionFiltersState as jest.Mock<any>;
mockUseNodeExecutionFiltersState.mockReturnValue({ filters: [], appliedFilters: [] });

jest.mock('components/Executions/Tables/NodeExecutionRow', () => ({
  NodeExecutionRow: jest.fn(({ nodeExecution }) => (
    <div data-testid="node-execution-row">
      <div data-testid="node-execution-col-id">{nodeExecution?.id?.nodeId}</div>
      <div data-testid="node-execution-col-phase">{nodeExecution?.closure?.phase}</div>
    </div>
  )),
}));

const mockNodes = (n: number): dNode[] => {
  const nodes: dNode[] = [];
  for (let i = 1; i <= n; i++) {
    nodes.push({
      id: `node${i}`,
      scopedId: `n${i}`,
      type: 4,
      name: `Node ${i}`,
      nodes: [],
      edges: [],
    });
  }
  return nodes;
};

const mockExecutionsById = (n: number, phases: NodeExecutionPhase[]) => {
  const nodeExecutionsById = {};

  for (let i = 1; i <= n; i++) {
    nodeExecutionsById[`n${i}`] = {
      closure: {
        createdAt: dateToTimestamp(new Date()),
        outputUri: '',
        phase: phases[i - 1],
      },
      id: {
        executionId: { domain: 'domain', name: 'name', project: 'project' },
        nodeId: `node${i}`,
      },
      inputUri: '',
      scopedId: `n${i}`,
    };
  }
  return nodeExecutionsById;
};

describe('NodeExecutionsTableExecutions > Tables > NodeExecutionsTable', () => {
  let queryClient: QueryClient;
  let fixture: ReturnType<typeof basicPythonWorkflow.generate>;
  const initialNodes = mockNodes(2);

  beforeEach(() => {
    queryClient = createTestQueryClient();
    fixture = basicPythonWorkflow.generate();
    insertFixture(mockServer, fixture);
    fetchWorkflow.mockImplementation(() => Promise.resolve(fixture.workflows.top));
  });

  const renderTable = ({ nodeExecutionsById, initialNodes, filteredNodes }) =>
    render(
      <QueryClientProvider client={queryClient}>
        <NodeExecutionDetailsContextProvider workflowId={mockWorkflowId}>
          <NodeExecutionsByIdContext.Provider value={nodeExecutionsById}>
            <NodeExecutionsTable initialNodes={initialNodes} filteredNodes={filteredNodes} />
          </NodeExecutionsByIdContext.Provider>
        </NodeExecutionDetailsContextProvider>
      </QueryClientProvider>,
    );

  it('renders empty content when there are no nodes', async () => {
    const { queryByText, queryByTestId } = renderTable({
      initialNodes: [],
      nodeExecutionsById: {},
      filteredNodes: [],
    });

    await waitFor(() => queryByText(noExecutionsFoundString));

    expect(queryByText(noExecutionsFoundString)).toBeInTheDocument();
    expect(queryByTestId('node-execution-row')).not.toBeInTheDocument();
  });

  it('renders NodeExecutionRows with initialNodes when no filteredNodes were provided', async () => {
    const phases = [NodeExecutionPhase.FAILED, NodeExecutionPhase.SUCCEEDED];
    const nodeExecutionsById = mockExecutionsById(2, phases);

    const { queryAllByTestId } = renderTable({
      initialNodes,
      nodeExecutionsById,
      filteredNodes: undefined,
    });

    await waitFor(() => queryAllByTestId('node-execution-row'));

    expect(queryAllByTestId('node-execution-row')).toHaveLength(initialNodes.length);
    const ids = queryAllByTestId('node-execution-col-id');
    expect(ids).toHaveLength(initialNodes.length);
    const renderedPhases = queryAllByTestId('node-execution-col-phase');
    expect(renderedPhases).toHaveLength(initialNodes.length);
    for (const i in initialNodes) {
      expect(ids[i]).toHaveTextContent(initialNodes[i].id);
      expect(renderedPhases[i]).toHaveTextContent(phases[i].toString());
    }
  });

  it('renders NodeExecutionRows with initialNodes even when filterNodes were provided, if appliedFilters is empty', async () => {
    const phases = [NodeExecutionPhase.FAILED, NodeExecutionPhase.SUCCEEDED];
    const nodeExecutionsById = mockExecutionsById(2, phases);
    const filteredNodes = mockNodes(1);

    const { queryAllByTestId } = renderTable({
      initialNodes,
      nodeExecutionsById,
      filteredNodes,
    });

    await waitFor(() => queryAllByTestId('node-execution-row'));

    expect(queryAllByTestId('node-execution-row')).toHaveLength(initialNodes.length);
    const ids = queryAllByTestId('node-execution-col-id');
    expect(ids).toHaveLength(initialNodes.length);
    const renderedPhases = queryAllByTestId('node-execution-col-phase');
    expect(renderedPhases).toHaveLength(initialNodes.length);
    for (const i in initialNodes) {
      expect(ids[i]).toHaveTextContent(initialNodes[i].id);
      expect(renderedPhases[i]).toHaveTextContent(phases[i].toString());
    }
  });

  it('renders NodeExecutionRows with filterNodes if appliedFilters is not empty', async () => {
    mockUseNodeExecutionFiltersState.mockReturnValueOnce({
      filters: [],
      appliedFilters: [{ key: 'phase', operation: 'value_in', value: ['FAILED', 'SUCCEEDED'] }],
    });

    const phases = [NodeExecutionPhase.FAILED, NodeExecutionPhase.SUCCEEDED];
    const nodeExecutionsById = mockExecutionsById(2, phases);
    const filteredNodes = mockNodes(1);

    const { queryAllByTestId } = renderTable({
      initialNodes,
      nodeExecutionsById,
      filteredNodes,
    });

    await waitFor(() => queryAllByTestId('node-execution-row'));

    expect(queryAllByTestId('node-execution-row')).toHaveLength(filteredNodes.length);
    const ids = queryAllByTestId('node-execution-col-id');
    expect(ids).toHaveLength(filteredNodes.length);
    const renderedPhases = queryAllByTestId('node-execution-col-phase');
    expect(renderedPhases).toHaveLength(filteredNodes.length);
    for (const i in filteredNodes) {
      expect(ids[i]).toHaveTextContent(filteredNodes[i].id);
      expect(renderedPhases[i]).toHaveTextContent(phases[i].toString());
    }
  });
});
