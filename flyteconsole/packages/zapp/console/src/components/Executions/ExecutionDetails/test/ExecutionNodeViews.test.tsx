import { fireEvent, render, waitFor } from '@testing-library/react';
import { filterLabels } from 'components/Executions/filters/constants';
import { nodeExecutionStatusFilters } from 'components/Executions/filters/statusFilters';
import { oneFailedTaskWorkflow } from 'mocks/data/fixtures/oneFailedTaskWorkflow';
import { insertFixture } from 'mocks/data/insertFixture';
import { mockServer } from 'mocks/server';
import { Execution } from 'models/Execution/types';
import * as React from 'react';
import { QueryClient, QueryClientProvider } from 'react-query';
import { createTestQueryClient } from 'test/utils';
import { tabs } from '../constants';
import { ExecutionNodeViews } from '../ExecutionNodeViews';

jest.mock('components/Executions/Tables/NodeExecutionRow', () => ({
  NodeExecutionRow: jest.fn(({ nodeExecution }) => (
    <div data-testid="node-execution-row">
      <span id="node-execution-col-id">{nodeExecution?.id?.nodeId}</span>
    </div>
  )),
}));

jest.mock('components/Executions/ExecutionDetails/Timeline/ExecutionTimelineFooter', () => ({
  ExecutionTimelineFooter: jest.fn(() => <div></div>),
}));

jest.mock('components/Executions/ExecutionDetails/Timeline/TimelineChart/index', () => ({
  TimelineChart: jest.fn(() => <div></div>),
}));

jest.mock('components/Executions/ExecutionDetails/Timeline/NodeExecutionName', () => ({
  NodeExecutionName: jest.fn(({ name }) => <div>{name}</div>),
}));

// ExecutionNodeViews uses query params for NE list, so we must match them
// for the list to be returned properly
const baseQueryParams = {
  filters: '',
  'sort_by.direction': 'ASCENDING',
  'sort_by.key': 'created_at',
};

describe('ExecutionNodeViews', () => {
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
      <QueryClientProvider client={queryClient}>
        <ExecutionNodeViews execution={execution} />
      </QueryClientProvider>,
    );

  it('maintains filter when switching back to nodes tab', async () => {
    const { nodeExecutions } = fixture.workflowExecutions.top;
    const failedNodeName = nodeExecutions.failedNode.data.id.nodeId;
    const succeededNodeName = nodeExecutions.pythonNode.data.id.nodeId;

    const { getByText, queryByText, getByLabelText } = renderViews();

    await waitFor(() => getByText(tabs.nodes.label));

    const nodesTab = getByText(tabs.nodes.label);
    const timelineTab = getByText(tabs.timeline.label);

    // Ensure we are on Nodes tab
    fireEvent.click(nodesTab);
    await waitFor(() => queryByText(succeededNodeName));

    const statusButton = await waitFor(() => getByText(filterLabels.status));

    // Apply 'Failed' filter and wait for list to include only the failed item
    fireEvent.click(statusButton);
    const failedFilter = await waitFor(() =>
      getByLabelText(nodeExecutionStatusFilters.failed.label),
    );

    // Wait for succeeded task to disappear and ensure failed task remains
    fireEvent.click(failedFilter);
    await waitFor(() => queryByText(failedNodeName));

    expect(queryByText(succeededNodeName)).not.toBeInTheDocument();
    expect(queryByText(failedNodeName)).toBeInTheDocument();

    // Switch to the Graph tab
    fireEvent.click(statusButton);
    fireEvent.click(timelineTab);
    await waitFor(() => queryByText(succeededNodeName));

    expect(queryByText(succeededNodeName)).toBeInTheDocument();

    // Switch back to Nodes Tab and verify filter still applied
    fireEvent.click(nodesTab);
    await waitFor(() => queryByText(failedNodeName));
    expect(queryByText(succeededNodeName)).not.toBeInTheDocument();
    expect(queryByText(failedNodeName)).toBeInTheDocument();
  });
});
