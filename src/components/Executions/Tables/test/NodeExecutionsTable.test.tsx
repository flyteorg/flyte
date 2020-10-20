import {
    fireEvent,
    getAllByRole,
    getByTitle,
    render,
    waitFor
} from '@testing-library/react';
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
import { cloneDeep } from 'lodash';
import {
    CompiledNode,
    Execution,
    FilterOperationName,
    getTask,
    NodeExecution,
    nodeExecutionQueryParams,
    RequestConfig,
    TaskExecution,
    TaskNodeMetadata,
    Workflow,
    WorkflowExecutionIdentifier
} from 'models';
import { createMockExecution } from 'models/__mocks__/executionsData';
import {
    createMockTaskExecutionForNodeExecution,
    createMockTaskExecutionsListResponse,
    mockExecution as mockTaskExecution
} from 'models/Execution/__mocks__/mockTaskExecutionsData';
import {
    getExecution,
    listNodeExecutions,
    listTaskExecutionChildren,
    listTaskExecutions
} from 'models/Execution/api';
import { mockTasks } from 'models/Task/__mocks__/mockTaskData';
import * as React from 'react';
import { makeIdentifier } from 'test/modelUtils';
import { obj } from 'test/utils';
import { Identifier } from 'typescript';
import { State } from 'xstate';
import { titleStrings } from '../constants';
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
    let mockNodes: CompiledNode[];
    let mockWorkflow: Workflow;
    let mockGetExecution: jest.Mock<ReturnType<typeof getExecution>>;
    let mockGetTask: jest.Mock<ReturnType<typeof getTask>>;
    let mockListTaskExecutions: jest.Mock<ReturnType<
        typeof listTaskExecutions
    >>;
    let mockListTaskExecutionChildren: jest.Mock<ReturnType<
        typeof listTaskExecutionChildren
    >>;
    let mockListNodeExecutions: jest.Mock<ReturnType<
        typeof listNodeExecutions
    >>;

    beforeEach(() => {
        const {
            nodes,
            nodeExecutions,
            workflow,
            workflowExecution
        } = createMockExecutionEntities({
            workflowName: 'SampleWorkflow',
            nodeExecutionCount: 2
        });

        mockNodes = nodes;
        mockNodeExecutions = nodeExecutions;
        mockWorkflow = workflow;

        mockListNodeExecutions = jest.fn().mockResolvedValue({ entities: [] });
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
            listNodeExecutions: mockListNodeExecutions,
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
            value: mockNodeExecutions,
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

        const node = dataCache.getNodeForNodeExecution(mockNodeExecutions[0]);
        const taskId = node?.node.taskNode?.referenceId;
        expect(taskId).toBeDefined();
        const task = dataCache.getTaskTemplate(taskId!);
        expect(task).toBeDefined();
        expect(queryAllByText(task!.id.name)[0]).toBeInTheDocument();
    });

    describe('for nodes with children', () => {
        let parentNodeExecution: NodeExecution;
        let childNodeExecutions: NodeExecution[];
        beforeEach(() => {
            parentNodeExecution = mockNodeExecutions[0];
        });

        const expandParentNode = async (container: HTMLElement) => {
            const expander = await waitFor(() =>
                getByTitle(container, titleStrings.expandRow)
            );
            fireEvent.click(expander);
            return await waitFor(() => getAllByRole(container, 'list'));
        };

        describe('with isParentNode flag', () => {
            beforeEach(() => {
                const id = parentNodeExecution.id;
                const { nodeId } = id;
                childNodeExecutions = [
                    {
                        ...parentNodeExecution,
                        id: { ...id, nodeId: `${nodeId}-child1` },
                        metadata: { retryGroup: '0', specNodeId: nodeId }
                    },
                    {
                        ...parentNodeExecution,
                        id: { ...id, nodeId: `${nodeId}-child2` },
                        metadata: { retryGroup: '0', specNodeId: nodeId }
                    },
                    {
                        ...parentNodeExecution,
                        id: { ...id, nodeId: `${nodeId}-child1` },
                        metadata: { retryGroup: '1', specNodeId: nodeId }
                    },
                    {
                        ...parentNodeExecution,
                        id: { ...id, nodeId: `${nodeId}-child2` },
                        metadata: { retryGroup: '1', specNodeId: nodeId }
                    }
                ];
                mockNodeExecutions[0].metadata = { isParentNode: true };
                mockListNodeExecutions.mockResolvedValue({
                    entities: childNodeExecutions
                });
            });

            it('correctly fetches children', async () => {
                const { getByText } = renderTable();
                await waitFor(() => getByText(mockNodeExecutions[0].id.nodeId));
                expect(mockListNodeExecutions).toHaveBeenCalledWith(
                    expect.anything(),
                    expect.objectContaining({
                        params: {
                            [nodeExecutionQueryParams.parentNodeId]:
                                parentNodeExecution.id.nodeId
                        }
                    })
                );
                expect(mockListTaskExecutionChildren).not.toHaveBeenCalled();
            });

            it('does not fetch children if flag is false', async () => {
                mockNodeExecutions[0].metadata = { isParentNode: false };
                const { getByText } = renderTable();
                await waitFor(() => getByText(mockNodeExecutions[0].id.nodeId));
                expect(mockListNodeExecutions).not.toHaveBeenCalled();
                expect(mockListTaskExecutionChildren).not.toHaveBeenCalled();
            });

            it('correctly renders groups', async () => {
                const { container } = renderTable();
                const childGroups = await expandParentNode(container);
                expect(childGroups).toHaveLength(2);
            });
        });

        describe('without isParentNode flag, using taskNodeMetadata ', () => {
            let taskExecutions: TaskExecution[];
            beforeEach(() => {
                taskExecutions = [0, 1].map(retryAttempt =>
                    createMockTaskExecutionForNodeExecution(
                        parentNodeExecution.id,
                        mockNodes[0],
                        retryAttempt,
                        { isParent: true }
                    )
                );
                childNodeExecutions = [
                    {
                        ...parentNodeExecution
                    }
                ];
                mockNodeExecutions = mockNodeExecutions.slice(0, 1);
                // In the case of a non-isParent node execution, API should not
                // return anything from the list endpoint
                mockListNodeExecutions.mockResolvedValue({ entities: [] });
                mockListTaskExecutions.mockImplementation(async id => {
                    const entities =
                        id.nodeId === parentNodeExecution.id.nodeId
                            ? taskExecutions
                            : [];
                    return { entities };
                });
                mockListTaskExecutionChildren.mockResolvedValue({
                    entities: childNodeExecutions
                });
            });

            it('correctly fetches children', async () => {
                const { getByText } = renderTable();
                await waitFor(() => getByText(mockNodeExecutions[0].id.nodeId));
                expect(mockListNodeExecutions).not.toHaveBeenCalled();
                expect(mockListTaskExecutionChildren).toHaveBeenCalledWith(
                    expect.objectContaining(taskExecutions[0].id),
                    expect.anything()
                );
            });

            it('correctly renders groups', async () => {
                // We returned two task execution attempts, each with children
                const { container } = renderTable();
                const childGroups = await expandParentNode(container);
                expect(childGroups).toHaveLength(2);
            });
        });

        describe('without isParentNode flag, using workflowNodeMetadata', () => {
            let childExecution: Execution;
            let childNodeExecutions: NodeExecution[];
            beforeEach(() => {
                childExecution = cloneDeep(executionContext.execution);
                childExecution.id.name = 'childExecution';
                dataCache.insertExecution(childExecution);
                dataCache.insertWorkflowExecutionReference(
                    childExecution.id,
                    mockWorkflow.id
                );

                childNodeExecutions = cloneDeep(mockNodeExecutions);
                childNodeExecutions.forEach(
                    ne => (ne.id.executionId = childExecution.id)
                );
                mockNodeExecutions[0].closure.workflowNodeMetadata = {
                    executionId: childExecution.id
                };
                mockGetExecution.mockImplementation(async id => {
                    if (id.name !== childExecution.id.name) {
                        throw new Error(
                            `Unexpected call to getExecution with execution id: ${obj(
                                id
                            )}`
                        );
                    }
                    return childExecution;
                });
                mockListNodeExecutions.mockImplementation(async id => {
                    const entities =
                        id.name === childExecution.id.name
                            ? childNodeExecutions
                            : [];
                    return { entities };
                });
            });

            it('correctly fetches children', async () => {
                const { getByText } = renderTable();
                await waitFor(() => getByText(mockNodeExecutions[0].id.nodeId));
                expect(mockListNodeExecutions).toHaveBeenCalledWith(
                    expect.objectContaining({ name: childExecution.id.name }),
                    expect.anything()
                );
            });

            it('correctly renders groups', async () => {
                // We returned a single WF execution child, so there should only
                // be one child group
                const { container } = renderTable();
                const childGroups = await expandParentNode(container);
                expect(childGroups).toHaveLength(1);
            });
        });
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
