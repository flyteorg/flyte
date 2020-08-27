import { render, waitFor } from '@testing-library/react';
import { mockAPIContextValue } from 'components/data/__mocks__/apiContext';
import { APIContext, APIContextValue } from 'components/data/apiContext';
import { createMockExecutionEntities } from 'components/Executions/__mocks__/createMockExecutionEntities';
import {
    ExecutionContext,
    ExecutionContextData,
    ExecutionDataCacheContext,
    NodeExecutionsRequestConfigContext
} from 'components/Executions/contexts';
import { ExecutionDataCache } from 'components/Executions/types';
import { createExecutionDataCache } from 'components/Executions/useExecutionDataCache';
import { fetchStates } from 'components/hooks/types';
import {
    FilterOperationName,
    getTask,
    NodeExecution,
    RequestConfig,
    WorkflowExecutionIdentifier
} from 'models';
import { createMockExecution } from 'models/__mocks__/executionsData';
import { createMockTaskExecutionsListResponse } from 'models/Execution/__mocks__/mockTaskExecutionsData';
import {
    getExecution,
    listTaskExecutionChildren,
    listTaskExecutions
} from 'models/Execution/api';
import { mockTasks } from 'models/Task/__mocks__/mockTaskData';
import * as React from 'react';
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
});
