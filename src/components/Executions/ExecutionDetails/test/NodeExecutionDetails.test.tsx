import { render, waitFor } from '@testing-library/react';
import { mockAPIContextValue } from 'components/data/__mocks__/apiContext';
import { APIContext } from 'components/data/apiContext';
import {
    cacheStatusMessages,
    viewSourceExecutionString
} from 'components/Executions/constants';
import {
    DetailedNodeExecution,
    NodeExecutionDisplayType
} from 'components/Executions/types';
import { Core } from 'flyteidl';
import { TaskNodeMetadata } from 'models';
import { createMockNodeExecutions } from 'models/Execution/__mocks__/mockNodeExecutionsData';
import { mockExecution as mockTaskExecution } from 'models/Execution/__mocks__/mockTaskExecutionsData';
import { listTaskExecutions } from 'models/Execution/api';
import * as React from 'react';
import { MemoryRouter } from 'react-router';
import { Routes } from 'routes';
import { makeIdentifier } from 'test/modelUtils';
import { NodeExecutionDetails } from '../NodeExecutionDetails';

describe('NodeExecutionDetails', () => {
    let execution: DetailedNodeExecution;
    let mockListTaskExecutions: jest.Mock<ReturnType<
        typeof listTaskExecutions
    >>;
    beforeEach(() => {
        const { executions } = createMockNodeExecutions(1);
        execution = {
            ...executions[0],
            displayType: NodeExecutionDisplayType.PythonTask,
            displayId: 'com.flyte.testTask',
            cacheKey: 'abcdefg'
        };
        mockListTaskExecutions = jest.fn().mockResolvedValue({ entities: [] });
    });

    const renderComponent = () =>
        render(
            <MemoryRouter>
                <APIContext.Provider
                    value={mockAPIContextValue({
                        listTaskExecutions: mockListTaskExecutions
                    })}
                >
                    <NodeExecutionDetails execution={execution} />
                </APIContext.Provider>
            </MemoryRouter>
        );

    it('renders displayId', async () => {
        const { queryByText } = renderComponent();
        await waitFor(() => {});
        expect(queryByText(execution.displayId)).toBeInTheDocument();
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
                const { getByText } = renderComponent();
                await waitFor(() =>
                    getByText(cacheStatusMessages[cacheStatusValue])
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
