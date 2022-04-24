import { render, waitFor } from '@testing-library/react';
import { cacheStatusMessages, viewSourceExecutionString } from 'components/Executions/constants';
import { NodeExecutionDetailsContextProvider } from 'components/Executions/contextProvider/NodeExecutionDetails';
import { basicPythonWorkflow } from 'mocks/data/fixtures/basicPythonWorkflow';
import { mockWorkflowId } from 'mocks/data/fixtures/types';
import { insertFixture } from 'mocks/data/insertFixture';
import { mockServer } from 'mocks/server';
import { ResourceType } from 'models/Common/types';
import { CatalogCacheStatus } from 'models/Execution/enums';
import { NodeExecution, TaskNodeMetadata } from 'models/Execution/types';
import { mockExecution as mockTaskExecution } from 'models/Execution/__mocks__/mockTaskExecutionsData';
import * as React from 'react';
import { QueryClient, QueryClientProvider } from 'react-query';
import { MemoryRouter } from 'react-router';
import { Routes } from 'routes/routes';
import { makeIdentifier } from 'test/modelUtils';
import { createTestQueryClient } from 'test/utils';
import { NodeExecutionDetailsPanelContent } from '../NodeExecutionDetailsPanelContent';

jest.mock('components/Workflow/workflowQueries');
const { fetchWorkflow } = require('components/Workflow/workflowQueries');

describe('NodeExecutionDetails', () => {
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
            <NodeExecutionDetailsPanelContent nodeExecutionId={execution.id} />
          </NodeExecutionDetailsContextProvider>
        </QueryClientProvider>
      </MemoryRouter>,
    );

  it('renders name for task nodes', async () => {
    const { name } = fixture.tasks.python.id;
    const { getByText } = renderComponent();
    await waitFor(() => expect(getByText(name)));
  });

  describe('with cache information', () => {
    let taskNodeMetadata: TaskNodeMetadata;
    beforeEach(() => {
      taskNodeMetadata = {
        cacheStatus: CatalogCacheStatus.CACHE_MISS,
        catalogKey: {
          datasetId: makeIdentifier({
            resourceType: ResourceType.DATASET,
          }),
          sourceTaskExecution: { ...mockTaskExecution.id },
        },
      };
      execution.closure.taskNodeMetadata = taskNodeMetadata;
      mockServer.insertNodeExecution(execution);
    });

    [
      CatalogCacheStatus.CACHE_DISABLED,
      CatalogCacheStatus.CACHE_HIT,
      CatalogCacheStatus.CACHE_LOOKUP_FAILURE,
      CatalogCacheStatus.CACHE_MISS,
      CatalogCacheStatus.CACHE_POPULATED,
      CatalogCacheStatus.CACHE_PUT_FAILURE,
    ].forEach((cacheStatusValue) =>
      it(`renders correct status for ${CatalogCacheStatus[cacheStatusValue]}`, async () => {
        taskNodeMetadata.cacheStatus = cacheStatusValue;
        mockServer.insertNodeExecution(execution);
        const { getByText } = renderComponent();
        await waitFor(() => expect(getByText(cacheStatusMessages[cacheStatusValue])));
      }),
    );

    it('renders source execution link for cache hits', async () => {
      taskNodeMetadata.cacheStatus = CatalogCacheStatus.CACHE_HIT;
      const sourceWorkflowExecutionId =
        taskNodeMetadata.catalogKey!.sourceTaskExecution.nodeExecutionId.executionId;
      const { getByText } = renderComponent();
      const linkEl = await waitFor(() => getByText(viewSourceExecutionString));
      expect(linkEl.getAttribute('href')).toBe(
        Routes.ExecutionDetails.makeUrl(sourceWorkflowExecutionId),
      );
    });
  });
});
