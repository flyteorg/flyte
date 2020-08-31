import { render, waitFor } from '@testing-library/react';
import { mockAPIContextValue } from 'components/data/__mocks__/apiContext';
import { APIContext, APIContextValue } from 'components/data/apiContext';
import { createMockExecutionEntities } from 'components/Executions/__mocks__/createMockExecutionEntities';
import { cacheStatusMessages } from 'components/Executions/constants';
import {
    ExecutionContext,
    ExecutionContextData,
    ExecutionDataCacheContext,
    NodeExecutionsRequestConfigContext
} from 'components/Executions/contexts';
import { ExecutionDataCache } from 'components/Executions/types';
import { createExecutionDataCache } from 'components/Executions/useExecutionDataCache';
import { fetchStates } from 'components/hooks/types';
import { Core } from 'flyteidl';
import {
    FilterOperationName,
    getTask,
    NodeExecution,
    RequestConfig,
    TaskNodeMetadata,
    WorkflowExecutionIdentifier
} from 'models';
import { createMockExecution } from 'models/__mocks__/executionsData';
import {
    createMockTaskExecutionsListResponse,
    mockExecution as mockTaskExecution
} from 'models/Execution/__mocks__/mockTaskExecutionsData';
import {
    getExecution,
    listTaskExecutionChildren,
    listTaskExecutions
} from 'models/Execution/api';
import { mockTasks } from 'models/Task/__mocks__/mockTaskData';
import * as React from 'react';
import { makeIdentifier } from 'test/modelUtils';
import { Identifier } from 'typescript';
import { State } from 'xstate';
import {
    NodeExecutionsTable,
    NodeExecutionsTableProps
} from '../NodeExecutionsTable';

describe('NodeExecutionsTable', () => {
    let props: NodeExecutionsTableProps;
    let apiContext: APIContextValue;
    let executionContext: ExecutionContextData;
    let dataCache: ExecutionDataCache;
    let requestConfig: RequestConfig;
    let mockNodeExecutions: NodeExecution[];
    let mockGetExecution: jest.Mock<ReturnType<typeof getExecution>>;
    let mockGetTask: jest.Mock<ReturnType<typeof getTask>>;
    let mockListTaskExecutions: jest.Mock<ReturnType<
        typeof listTaskExecutions
    >>;
    let mockListTaskExecutionChildren: jest.Mock<ReturnType<
        typeof listTaskExecutionChildren
    >>;

    beforeEach(() => {
        const {
            nodeExecutions,
            workflow,
            workflowExecution
        } = createMockExecutionEntities({
            workflowName: 'SampleWorkflow',
            nodeExecutionCount: 2
        });

        mockNodeExecutions = nodeExecutions;

        mockListTaskExecutions = jest.fn().mockResolvedValue({ entities: [] });
        mockListTaskExecutionChildren = jest
            .fn()
            .mockResolvedValue({ entities: [] });
        mockGetExecution = jest
            .fn()
            .mockImplementation(async (id: WorkflowExecutionIdentifier) => {
                return { ...createMockExecution(id.name), id };
            });
        mockGetTask = jest.fn().mockImplementation(async (id: Identifier) => {
            return { template: { ...mockTasks[0].template, id } };
        });

        apiContext = mockAPIContextValue({
            getExecution: mockGetExecution,
            getTask: mockGetTask,
            listTaskExecutions: mockListTaskExecutions,
            listTaskExecutionChildren: mockListTaskExecutionChildren
        });

        dataCache = createExecutionDataCache(apiContext);
        dataCache.insertWorkflow(workflow);
        dataCache.insertWorkflowExecutionReference(
            workflowExecution.id,
            workflow.id
        );

        requestConfig = {};
        executionContext = {
            execution: workflowExecution,
            terminateExecution: jest.fn().mockRejectedValue('Not Implemented')
        };

        props = {
            value: nodeExecutions,
            lastError: null,
            state: State.from(fetchStates.LOADED),
            moreItemsAvailable: false,
            fetch: jest.fn()
        };
    });

    const renderTable = () =>
        render(
            <APIContext.Provider value={apiContext}>
                <NodeExecutionsRequestConfigContext.Provider
                    value={requestConfig}
                >
                    <ExecutionContext.Provider value={executionContext}>
                        <ExecutionDataCacheContext.Provider value={dataCache}>
                            <NodeExecutionsTable {...props} />
                        </ExecutionDataCacheContext.Provider>
                    </ExecutionContext.Provider>
                </NodeExecutionsRequestConfigContext.Provider>
            </APIContext.Provider>
        );

    it('renders task name for task nodes', async () => {
        const { queryAllByText } = renderTable();
        await waitFor(() => {});

        const node = dataCache.getNodeForNodeExecution(
            mockNodeExecutions[0].id
        );
        const taskId = node?.node.taskNode?.referenceId;
        expect(taskId).toBeDefined();
        const task = dataCache.getTaskTemplate(taskId!);
        expect(task).toBeDefined();
        expect(queryAllByText(task!.id.name)[0]).toBeInTheDocument();
    });

    it('requests child node executions using configuration from context', async () => {
        const { taskExecutions } = createMockTaskExecutionsListResponse(1);
        taskExecutions[0].isParent = true;
        mockListTaskExecutions.mockResolvedValue({ entities: taskExecutions });
        requestConfig.filter = [
            { key: 'test', operation: FilterOperationName.EQ, value: 'test' }
        ];

        renderTable();
        await waitFor(() =>
            expect(mockListTaskExecutionChildren).toHaveBeenCalled()
        );

        expect(mockListTaskExecutionChildren).toHaveBeenCalledWith(
            taskExecutions[0].id,
            expect.objectContaining(requestConfig)
        );
    });

    describe('for task nodes with cache status', () => {
        let taskNodeMetadata: TaskNodeMetadata;
        let cachedNodeExecution: NodeExecution;
        beforeEach(() => {
            cachedNodeExecution = mockNodeExecutions[0];
            taskNodeMetadata = {
                cacheStatus: Core.CatalogCacheStatus.CACHE_MISS,
                catalogKey: {
                    datasetId: makeIdentifier({
                        resourceType: Core.ResourceType.DATASET
                    }),
                    sourceTaskExecution: { ...mockTaskExecution.id }
                }
            };
            cachedNodeExecution.closure.taskNodeMetadata = taskNodeMetadata;
        });

        [
            Core.CatalogCacheStatus.CACHE_HIT,
            Core.CatalogCacheStatus.CACHE_LOOKUP_FAILURE,
            Core.CatalogCacheStatus.CACHE_POPULATED,
            Core.CatalogCacheStatus.CACHE_PUT_FAILURE
        ].forEach(cacheStatusValue =>
            it(`renders correct icon for ${Core.CatalogCacheStatus[cacheStatusValue]}`, async () => {
                taskNodeMetadata.cacheStatus = cacheStatusValue;
                const { getByTitle } = renderTable();
                await waitFor(() =>
                    getByTitle(cacheStatusMessages[cacheStatusValue])
                );
            })
        );

        [
            Core.CatalogCacheStatus.CACHE_DISABLED,
            Core.CatalogCacheStatus.CACHE_MISS
        ].forEach(cacheStatusValue =>
            it(`renders no icon for ${Core.CatalogCacheStatus[cacheStatusValue]}`, async () => {
                taskNodeMetadata.cacheStatus = cacheStatusValue;
                const { getByText, queryByTitle } = renderTable();
                await waitFor(() => getByText(cachedNodeExecution.id.nodeId));
                expect(
                    queryByTitle(cacheStatusMessages[cacheStatusValue])
                ).toBeNull();
            })
        );
    });
});
