import { render, waitFor } from '@testing-library/react';
import { cacheStatusMessages } from 'components/Executions/constants';
import { NodeExecutionDetailsContextProvider } from 'components/Executions/contextProvider/NodeExecutionDetails';
import { basicPythonWorkflow } from 'mocks/data/fixtures/basicPythonWorkflow';
import { mockWorkflowId } from 'mocks/data/fixtures/types';
import { insertFixture } from 'mocks/data/insertFixture';
import { mockServer } from 'mocks/server';
import { CatalogCacheStatus } from 'models/Execution/enums';
import { NodeExecution } from 'models/Execution/types';
import * as React from 'react';
import { QueryClient, QueryClientProvider } from 'react-query';
import { MemoryRouter } from 'react-router';
import { createTestQueryClient } from 'test/utils';
import { NodeExecutionCacheStatus } from '../NodeExecutionCacheStatus';

jest.mock('models/Task/utils');
jest.mock('components/Workflow/workflowQueries');
const { fetchWorkflow } = require('components/Workflow/workflowQueries');

// TODO add test to cover mapTask branch
describe('Executions > NodeExecutionCacheStatus', () => {
  let fixture: ReturnType<typeof basicPythonWorkflow.generate>;
  let execution: NodeExecution;
  let queryClient: QueryClient;

  beforeEach(() => {
    fixture = basicPythonWorkflow.generate();
    insertFixture(mockServer, fixture);
    execution = fixture.workflowExecutions.top.nodeExecutions.pythonNode.data;
    fetchWorkflow.mockImplementation(() => Promise.resolve(fixture.workflows.top));
    queryClient = createTestQueryClient();
  });

  const renderComponent = (props) =>
    render(
      <MemoryRouter>
        <QueryClientProvider client={queryClient}>
          <NodeExecutionDetailsContextProvider workflowId={mockWorkflowId}>
            <NodeExecutionCacheStatus {...props} />
          </NodeExecutionDetailsContextProvider>
        </QueryClientProvider>
      </MemoryRouter>,
    );

  it('should not render anything, if cacheStatus is undefined', async () => {
    const { container } = renderComponent({ execution });
    await waitFor(() => container);

    expect(container).toBeEmptyDOMElement();
  });

  it('should not render anything, if cacheStatus is null', async () => {
    const mockExecution = {
      ...execution,
      closure: { taskNodeMetadata: { cacheStatus: null } },
    };
    const { container } = renderComponent({ execution: mockExecution });
    await waitFor(() => container);

    expect(container).toBeEmptyDOMElement();
  });

  it('should render cache hit status text, if execution has cacheStatus CACHE_HIT', async () => {
    const cacheStatus = CatalogCacheStatus.CACHE_HIT;
    const mockExecution = { ...execution, closure: { taskNodeMetadata: { cacheStatus } } };
    const { queryByText } = renderComponent({ execution: mockExecution });
    await waitFor(() => queryByText(cacheStatusMessages[cacheStatus]));

    expect(queryByText(cacheStatusMessages[cacheStatus])).toBeInTheDocument();
  });
});
