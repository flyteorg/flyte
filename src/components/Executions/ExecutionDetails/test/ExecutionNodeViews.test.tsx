import {
    fireEvent,
    render,
    waitFor,
    waitForElementToBeRemoved
} from '@testing-library/react';
import { mockAPIContextValue } from 'components/data/__mocks__/apiContext';
import { APIContext, APIContextValue } from 'components/data/apiContext';
import { createMockExecutionEntities } from 'components/Executions/__mocks__/createMockExecutionEntities';
import {
    ExecutionContextData,
    ExecutionDataCacheContext
} from 'components/Executions/contexts';
import { filterLabels } from 'components/Executions/filters/constants';
import { nodeExecutionStatusFilters } from 'components/Executions/filters/statusFilters';
import { ExecutionDataCache } from 'components/Executions/types';
import { createExecutionDataCache } from 'components/Executions/useExecutionDataCache';
import {
    getExecution,
    Identifier,
    listNodeExecutions,
    WorkflowExecutionIdentifier
} from 'models';
import { createMockExecution } from 'models/__mocks__/executionsData';
import { mockTasks } from 'models/Task/__mocks__/mockTaskData';
import * as React from 'react';
import { tabs } from '../constants';
import {
    ExecutionNodeViews,
    ExecutionNodeViewsProps
} from '../ExecutionNodeViews';

// We don't need to verify the content of the graph component here and it is
// difficult to make it work correctly in a test environment.
jest.mock('../ExecutionWorkflowGraph.tsx', () => ({
    ExecutionWorkflowGraph: () => null
}));

describe('ExecutionNodeViews', () => {
    let props: ExecutionNodeViewsProps;
    let apiContext: APIContextValue;
    let executionContext: ExecutionContextData;
    let dataCache: ExecutionDataCache;
    let mockListNodeExecutions: jest.Mock<ReturnType<
        typeof listNodeExecutions
    >>;
    let mockGetExecution: jest.Mock<ReturnType<typeof getExecution>>;

    beforeEach(() => {
        const {
            nodeExecutions,
            workflow,
            workflowExecution
        } = createMockExecutionEntities({
            workflowName: 'SampleWorkflow',
            nodeExecutionCount: 2
        });

        mockGetExecution = jest
            .fn()
            .mockImplementation(async (id: WorkflowExecutionIdentifier) => {
                return { ...createMockExecution(id.name), id };
            });

        mockListNodeExecutions = jest
            .fn()
            .mockResolvedValue({ entities: nodeExecutions });
        apiContext = mockAPIContextValue({
            getExecution: mockGetExecution,
            getTask: jest.fn().mockImplementation(async (id: Identifier) => {
                return { template: { ...mockTasks[0].template, id } };
            }),
            listNodeExecutions: mockListNodeExecutions,
            listTaskExecutions: jest.fn().mockResolvedValue({ entities: [] }),
            listTaskExecutionChildren: jest
                .fn()
                .mockResolvedValue({ entities: [] })
        });

        dataCache = createExecutionDataCache(apiContext);
        dataCache.insertWorkflow(workflow);
        dataCache.insertWorkflowExecutionReference(
            workflowExecution.id,
            workflow.id
        );

        executionContext = {
            execution: workflowExecution,
            terminateExecution: jest.fn().mockRejectedValue('Not Implemented')
        };

        props = { execution: workflowExecution };
    });

    const renderViews = () =>
        render(
            <APIContext.Provider value={apiContext}>
                <ExecutionDataCacheContext.Provider value={dataCache}>
                    <ExecutionNodeViews {...props} />
                </ExecutionDataCacheContext.Provider>
            </APIContext.Provider>
        );

    it('only applies filter when viewing the nodes tab', async () => {
        const { getByText } = renderViews();
        const nodesTab = await waitFor(() => getByText(tabs.nodes.label));
        const graphTab = await waitFor(() => getByText(tabs.graph.label));

        fireEvent.click(nodesTab);
        const statusButton = await waitFor(() =>
            getByText(filterLabels.status)
        );
        fireEvent.click(statusButton);
        const successFilter = await waitFor(() =>
            getByText(nodeExecutionStatusFilters.succeeded.label)
        );

        mockListNodeExecutions.mockClear();
        fireEvent.click(successFilter);
        await waitFor(() => mockListNodeExecutions.mock.calls.length > 0);
        // Verify at least one filter is passed
        expect(mockListNodeExecutions).toHaveBeenCalledWith(
            expect.anything(),
            expect.objectContaining({
                filter: expect.arrayContaining([
                    expect.objectContaining({ key: expect.any(String) })
                ])
            })
        );

        fireEvent.click(statusButton);
        await waitForElementToBeRemoved(successFilter);
        mockListNodeExecutions.mockClear();
        fireEvent.click(graphTab);
        await waitFor(() => mockListNodeExecutions.mock.calls.length > 0);
        // No filter expected on the graph tab
        expect(mockListNodeExecutions).toHaveBeenCalledWith(
            expect.anything(),
            expect.objectContaining({ filter: [] })
        );

        mockListNodeExecutions.mockClear();
        fireEvent.click(nodesTab);
        await waitFor(() => mockListNodeExecutions.mock.calls.length > 0);
        // Verify (again) at least one filter is passed, after changing back to
        // nodes tab.
        expect(mockListNodeExecutions).toHaveBeenCalledWith(
            expect.anything(),
            expect.objectContaining({
                filter: expect.arrayContaining([
                    expect.objectContaining({ key: expect.any(String) })
                ])
            })
        );
    });
});
