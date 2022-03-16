import ThemeProvider from '@material-ui/styles/ThemeProvider';
import { render, waitFor } from '@testing-library/react';
import { muiTheme } from 'components/Theme/muiTheme';
import { oneFailedTaskWorkflow } from 'mocks/data/fixtures/oneFailedTaskWorkflow';
import { insertFixture } from 'mocks/data/insertFixture';
import { mockServer } from 'mocks/server';
import { Execution } from 'models/Execution/types';
import * as React from 'react';
import { QueryClient, QueryClientProvider } from 'react-query';
import { createTestQueryClient } from 'test/utils';
import { ExecutionNodesTimeline } from '../Timeline';

// We don't need to verify the content of the graph component here and it is
// difficult to make it work correctly in a test environment.
jest.mock('../ExecutionWorkflowGraph.tsx', () => ({
  ExecutionWorkflowGraph: () => null,
}));

jest.mock('chart.js', () => ({
  Chart: { register: () => null },
  registerables: [],
}));

jest.mock('chartjs-plugin-datalabels', () => ({
  ChartDataLabels: null,
}));

const baseQueryParams = {
  filters: '',
  'sort_by.direction': 'ASCENDING',
  'sort_by.key': 'created_at',
};

describe('ExecutionDetails > Timeline', () => {
  let queryClient: QueryClient;
  let execution: Execution;
  let fixture: ReturnType<typeof oneFailedTaskWorkflow.generate>;

  beforeEach(() => {
    fixture = oneFailedTaskWorkflow.generate();
    execution = fixture.workflowExecutions.top.data;
    insertFixture(mockServer, fixture);
    const nodeExecutions = fixture.workflowExecutions.top.nodeExecutions;

    mockServer.insertNodeExecutionList(
      execution.id,
      Object.values(nodeExecutions).map(({ data }) => data),
      baseQueryParams,
    );
    mockServer.insertNodeExecutionList(execution.id, [nodeExecutions.failedNode.data], {
      ...baseQueryParams,
      filters: 'value_in(phase,FAILED)',
    });
    queryClient = createTestQueryClient();
  });

  const renderViews = () =>
    render(
      <ThemeProvider theme={muiTheme}>
        <QueryClientProvider client={queryClient}>
          <ExecutionNodesTimeline nodeExecutions={[]} />
        </QueryClientProvider>
      </ThemeProvider>,
    );

  it('Render Timeline with empty nodeExecution info will not crash', async () => {
    const { getByText } = renderViews();
    await waitFor(() => {
      expect(getByText('Local Time')).toBeTruthy();
    });
  });
});
