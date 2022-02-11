import { render, waitFor } from '@testing-library/react';
import {
    cacheStatusMessages,
    viewSourceExecutionString
} from 'components/Executions/constants';
import { NodeExecutionDetailsContextProvider } from 'components/Executions/contextProvider/NodeExecutionDetails';
import { Core } from 'flyteidl';
import { basicPythonWorkflow } from 'mocks/data/fixtures/basicPythonWorkflow';
import { mockWorkflowId } from 'mocks/data/fixtures/types';
import { insertFixture } from 'mocks/data/insertFixture';
import { mockServer } from 'mocks/server';
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
        execution =
            fixture.workflowExecutions.top.nodeExecutions.pythonNode.data;
        insertFixture(mockServer, fixture);
        fetchWorkflow.mockImplementation(() =>
            Promise.resolve(fixture.workflows.top)
        );
        queryClient = createTestQueryClient();
    });

    const renderComponent = () =>
        render(
            <MemoryRouter>
                <QueryClientProvider client={queryClient}>
                    <NodeExecutionDetailsContextProvider
                        workflowId={mockWorkflowId}
                    >
                        <NodeExecutionDetailsPanelContent
                            nodeExecutionId={execution.id}
                        />
                    </NodeExecutionDetailsContextProvider>
                </QueryClientProvider>
            </MemoryRouter>
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
                cacheStatus: Core.CatalogCacheStatus.CACHE_MISS,
                catalogKey: {
                    datasetId: makeIdentifier({
                        resourceType: Core.ResourceType.DATASET
                    }),
                    sourceTaskExecution: { ...mockTaskExecution.id }
                }
            };
            execution.closure.taskNodeMetadata = taskNodeMetadata;
            mockServer.insertNodeExecution(execution);
        });

        [
            Core.CatalogCacheStatus.CACHE_DISABLED,
            Core.CatalogCacheStatus.CACHE_HIT,
            Core.CatalogCacheStatus.CACHE_LOOKUP_FAILURE,
            Core.CatalogCacheStatus.CACHE_MISS,
            Core.CatalogCacheStatus.CACHE_POPULATED,
            Core.CatalogCacheStatus.CACHE_PUT_FAILURE
        ].forEach(cacheStatusValue =>
            it(`renders correct status for ${Core.CatalogCacheStatus[cacheStatusValue]}`, async () => {
                taskNodeMetadata.cacheStatus = cacheStatusValue;
                mockServer.insertNodeExecution(execution);
                const { getByText } = renderComponent();
                await waitFor(() =>
                    expect(getByText(cacheStatusMessages[cacheStatusValue]))
                );
            })
        );

        it('renders source execution link for cache hits', async () => {
            taskNodeMetadata.cacheStatus = Core.CatalogCacheStatus.CACHE_HIT;
            const sourceWorkflowExecutionId = taskNodeMetadata.catalogKey!
                .sourceTaskExecution.nodeExecutionId.executionId;
            const { getByText } = renderComponent();
            const linkEl = await waitFor(() =>
                getByText(viewSourceExecutionString)
            );
            expect(linkEl.getAttribute('href')).toBe(
                Routes.ExecutionDetails.makeUrl(sourceWorkflowExecutionId)
            );
        });
    });
});
