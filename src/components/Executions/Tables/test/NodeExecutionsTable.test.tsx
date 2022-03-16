import {
  fireEvent,
  getAllByRole,
  getAllByText,
  getByText,
  getByTitle,
  render,
  screen,
  waitFor,
} from '@testing-library/react';
import { cacheStatusMessages } from 'components/Executions/constants';
import { NodeExecutionDetailsContextProvider } from 'components/Executions/contextProvider/NodeExecutionDetails';
import { UNKNOWN_DETAILS } from 'components/Executions/contextProvider/NodeExecutionDetails/types';
import {
  ExecutionContext,
  ExecutionContextData,
  NodeExecutionsRequestConfigContext,
} from 'components/Executions/contexts';
import { makeNodeExecutionListQuery } from 'components/Executions/nodeExecutionQueries';
import { NodeExecutionDisplayType } from 'components/Executions/types';
import { nodeExecutionIsTerminal } from 'components/Executions/utils';
import { useConditionalQuery } from 'components/hooks/useConditionalQuery';
import { Core } from 'flyteidl';
import { cloneDeep } from 'lodash';
import { basicPythonWorkflow } from 'mocks/data/fixtures/basicPythonWorkflow';
import { dynamicExternalSubWorkflow } from 'mocks/data/fixtures/dynamicExternalSubworkflow';
import {
  dynamicPythonNodeExecutionWorkflow,
  dynamicPythonTaskWorkflow,
} from 'mocks/data/fixtures/dynamicPythonWorkflow';
import { oneFailedTaskWorkflow } from 'mocks/data/fixtures/oneFailedTaskWorkflow';
import { mockWorkflowId } from 'mocks/data/fixtures/types';
import { insertFixture } from 'mocks/data/insertFixture';
import { notFoundError } from 'mocks/errors';
import { mockServer } from 'mocks/server';
import { FilterOperationName, RequestConfig } from 'models/AdminEntity/types';
import { nodeExecutionQueryParams } from 'models/Execution/constants';
import { NodeExecutionPhase } from 'models/Execution/enums';
import { Execution, NodeExecution, TaskNodeMetadata } from 'models/Execution/types';
import * as React from 'react';
import { QueryClient, QueryClientProvider, useQueryClient } from 'react-query';
import { makeIdentifier } from 'test/modelUtils';
import {
  createTestQueryClient,
  disableQueryLogger,
  enableQueryLogger,
  findNearestAncestorByRole,
} from 'test/utils';
import * as moduleApi from 'components/Executions/contextProvider/NodeExecutionDetails/getTaskThroughExecution';
import { titleStrings } from '../constants';
import { NodeExecutionsTable } from '../NodeExecutionsTable';

jest.mock('components/Workflow/workflowQueries');
const { fetchWorkflow } = require('components/Workflow/workflowQueries');

describe('NodeExecutionsTable', () => {
  let workflowExecution: Execution;
  let queryClient: QueryClient;
  let executionContext: ExecutionContextData;
  let requestConfig: RequestConfig;

  beforeEach(() => {
    requestConfig = {};
    queryClient = createTestQueryClient();
  });

  const shouldUpdateFn = (nodeExecutions: NodeExecution[]) =>
    nodeExecutions.some((ne) => !nodeExecutionIsTerminal(ne));

  const selectNode = async (container: HTMLElement, truncatedName: string, nodeId: string) => {
    const nodeNameAnchor = await waitFor(() => getByText(container, truncatedName));
    fireEvent.click(nodeNameAnchor);
    // Wait for Details Panel to render and then for the nodeId header
    const detailsPanel = await waitFor(() => screen.getByTestId('details-panel'));
    await waitFor(() => getByText(detailsPanel, nodeId));
    return detailsPanel;
  };

  const expandParentNode = async (rowContainer: HTMLElement) => {
    const expander = await waitFor(() => getByTitle(rowContainer, titleStrings.expandRow));
    fireEvent.click(expander);
    return await waitFor(() => getAllByRole(rowContainer, 'list'));
  };

  const TestTable = () => {
    const query = useConditionalQuery(
      {
        ...makeNodeExecutionListQuery(useQueryClient(), workflowExecution.id, requestConfig),
        // During tests, we only want to wait for the next tick to refresh
        refetchInterval: 1,
      },
      shouldUpdateFn,
    );
    return query.data ? <NodeExecutionsTable nodeExecutions={query.data} /> : null;
  };

  const renderTable = () =>
    render(
      <QueryClientProvider client={queryClient}>
        <NodeExecutionsRequestConfigContext.Provider value={requestConfig}>
          <ExecutionContext.Provider value={executionContext}>
            <NodeExecutionDetailsContextProvider workflowId={mockWorkflowId}>
              <TestTable />
            </NodeExecutionDetailsContextProvider>
          </ExecutionContext.Provider>
        </NodeExecutionsRequestConfigContext.Provider>
      </QueryClientProvider>,
    );

  describe('when rendering the DetailsPanel', () => {
    let nodeExecution: NodeExecution;
    let fixture: ReturnType<typeof basicPythonWorkflow.generate>;
    beforeEach(() => {
      fixture = basicPythonWorkflow.generate();
      workflowExecution = fixture.workflowExecutions.top.data;
      insertFixture(mockServer, fixture);
      fetchWorkflow.mockImplementation(() => Promise.resolve(fixture.workflows.top));

      executionContext = {
        execution: workflowExecution,
      };
      nodeExecution = fixture.workflowExecutions.top.nodeExecutions.pythonNode.data;
    });

    const updateNodeExecutions = (executions: NodeExecution[]) => {
      executions.forEach(mockServer.insertNodeExecution);
      mockServer.insertNodeExecutionList(fixture.workflowExecutions.top.data.id, executions);
    };

    it('should render updated state if selected nodeExecution object changes', async () => {
      nodeExecution.closure.phase = NodeExecutionPhase.RUNNING;
      updateNodeExecutions([nodeExecution]);
      const truncatedName = fixture.tasks.python.id.name.split('.').pop() || '';
      // Render table, click first node
      const { container } = renderTable();
      const detailsPanel = await selectNode(container, truncatedName, nodeExecution.id.nodeId);
      expect(getByText(detailsPanel, 'Running')).toBeInTheDocument();

      const updatedExecution = cloneDeep(nodeExecution);
      updatedExecution.closure.phase = NodeExecutionPhase.FAILED;
      updateNodeExecutions([updatedExecution]);
      await waitFor(() => expect(getByText(detailsPanel, 'Failed')));
    });

    describe('with nested children', () => {
      let fixture: ReturnType<typeof dynamicPythonNodeExecutionWorkflow.generate>;
      beforeEach(() => {
        fixture = dynamicPythonNodeExecutionWorkflow.generate();
        workflowExecution = fixture.workflowExecutions.top.data;
        insertFixture(mockServer, fixture);
        fetchWorkflow.mockImplementation(() => Promise.resolve(fixture.workflows.top));
        executionContext = { execution: workflowExecution };
      });

      it('should correctly render details for nested executions', async () => {
        const childNodeExecution =
          fixture.workflowExecutions.top.nodeExecutions.dynamicNode.nodeExecutions.firstChild.data;
        const { container } = renderTable();
        const dynamicTaskNameEl = await waitFor(() =>
          getByText(container, fixture.tasks.dynamic.id.name),
        );
        const dynamicRowEl = findNearestAncestorByRole(dynamicTaskNameEl, 'listitem');
        const parentNodeEl = await expandParentNode(dynamicRowEl);
        const truncatedName = fixture.tasks.python.id.name.split('.').pop() || '';
        await selectNode(parentNodeEl[0], truncatedName, childNodeExecution.id.nodeId);

        // Wait for Details Panel to render and then for the nodeId header
        const detailsPanel = await waitFor(() => screen.getByTestId('details-panel'));
        await waitFor(() => expect(getByText(detailsPanel, childNodeExecution.id.nodeId)));
        expect(getByText(detailsPanel, fixture.tasks.python.id.name)).toBeInTheDocument();
      });
    });
  });

  describe('for basic executions', () => {
    let fixture: ReturnType<typeof basicPythonWorkflow.generate>;

    beforeEach(() => {
      fixture = basicPythonWorkflow.generate();
      workflowExecution = fixture.workflowExecutions.top.data;
      insertFixture(mockServer, fixture);
      fetchWorkflow.mockImplementation(() => Promise.resolve(fixture.workflows.top));

      executionContext = {
        execution: workflowExecution,
      };
    });

    const updateNodeExecutions = (executions: NodeExecution[]) => {
      executions.forEach(mockServer.insertNodeExecution);
      mockServer.insertNodeExecutionList(fixture.workflowExecutions.top.data.id, executions);
    };

    it('renders task name for task nodes', async () => {
      const { getByText } = renderTable();
      await waitFor(() => expect(getByText(fixture.tasks.python.id.name)).toBeInTheDocument());
    });

    it('renders NodeExecutions with no associated spec information as Unknown', async () => {
      const workflowExecution = fixture.workflowExecutions.top.data;
      // For a NodeExecution which has no node in the associated workflow spec and
      // no task executions, we don't have a way to identify its type.
      // We'll change the python NodeExecution to reference a node id which doesn't exist
      // in the spec and remove its TaskExecutions.
      const nodeExecution = fixture.workflowExecutions.top.nodeExecutions.pythonNode.data;
      nodeExecution.id.nodeId = 'unknownNode';
      nodeExecution.metadata = {};
      mockServer.insertNodeExecution(nodeExecution);
      mockServer.insertNodeExecutionList(workflowExecution.id, [nodeExecution]);
      mockServer.insertTaskExecutionList(nodeExecution.id, []);

      const { container } = renderTable();
      const pythonNodeNameEl = await waitFor(() =>
        getAllByText(container, nodeExecution.id.nodeId),
      );
      const rowEl = findNearestAncestorByRole(pythonNodeNameEl?.[0], 'listitem');
      await waitFor(() => expect(getByText(rowEl, NodeExecutionDisplayType.Unknown)));
    });

    describe('for task nodes with cache status', () => {
      let taskNodeMetadata: TaskNodeMetadata;
      let cachedNodeExecution: NodeExecution;
      beforeEach(() => {
        const { nodeExecutions } = fixture.workflowExecutions.top;
        const { taskExecutions } = nodeExecutions.pythonNode;
        cachedNodeExecution = nodeExecutions.pythonNode.data;
        taskNodeMetadata = {
          cacheStatus: Core.CatalogCacheStatus.CACHE_MISS,
          catalogKey: {
            datasetId: makeIdentifier({
              resourceType: Core.ResourceType.DATASET,
            }),
            sourceTaskExecution: {
              ...taskExecutions.firstAttempt.data.id,
            },
          },
        };
        cachedNodeExecution.closure.taskNodeMetadata = taskNodeMetadata;
      });

      [
        Core.CatalogCacheStatus.CACHE_HIT,
        Core.CatalogCacheStatus.CACHE_LOOKUP_FAILURE,
        Core.CatalogCacheStatus.CACHE_POPULATED,
        Core.CatalogCacheStatus.CACHE_PUT_FAILURE,
      ].forEach((cacheStatusValue) =>
        it(`renders correct icon for ${Core.CatalogCacheStatus[cacheStatusValue]}`, async () => {
          taskNodeMetadata.cacheStatus = cacheStatusValue;
          updateNodeExecutions([cachedNodeExecution]);
          const { getByTitle } = renderTable();

          await waitFor(() =>
            expect(getByTitle(cacheStatusMessages[cacheStatusValue])).toBeDefined(),
          );
        }),
      );

      [Core.CatalogCacheStatus.CACHE_DISABLED, Core.CatalogCacheStatus.CACHE_MISS].forEach(
        (cacheStatusValue) =>
          it(`renders no icon for ${Core.CatalogCacheStatus[cacheStatusValue]}`, async () => {
            taskNodeMetadata.cacheStatus = cacheStatusValue;
            updateNodeExecutions([cachedNodeExecution]);
            const { getByText, queryByTitle } = renderTable();
            await waitFor(() => {
              getByText(cachedNodeExecution.id.nodeId);
            });
            expect(queryByTitle(cacheStatusMessages[cacheStatusValue])).toBeNull();
          }),
      );
    });
  });

  describe('for nodes with children', () => {
    describe('with isParentNode flag', () => {
      let fixture: ReturnType<typeof dynamicPythonNodeExecutionWorkflow.generate>;
      beforeEach(() => {
        fixture = dynamicPythonNodeExecutionWorkflow.generate();
        workflowExecution = fixture.workflowExecutions.top.data;
        insertFixture(mockServer, fixture);
        fetchWorkflow.mockImplementation(() => Promise.resolve(fixture.workflows.top));
        executionContext = { execution: workflowExecution };
      });

      it('correctly renders children', async () => {
        const { container } = renderTable();
        const dynamicTaskNameEl = await waitFor(() =>
          getByText(container, fixture.tasks.dynamic.id.name),
        );
        const dynamicRowEl = findNearestAncestorByRole(dynamicTaskNameEl, 'listitem');
        const childContainerList = await expandParentNode(dynamicRowEl);
        await waitFor(() => expect(getByText(childContainerList[0], fixture.tasks.python.id.name)));
      });

      it('correctly renders groups', async () => {
        const { nodeExecutions } = fixture.workflowExecutions.top;
        // We returned two task execution attempts, each with children
        const { container } = renderTable();
        const nodeNameEl = await waitFor(() =>
          getByText(container, nodeExecutions.dynamicNode.data.id.nodeId),
        );
        const rowEl = findNearestAncestorByRole(nodeNameEl, 'listitem');
        const childGroups = await expandParentNode(rowEl);
        expect(childGroups).toHaveLength(2);
      });

      describe('with initial failure to fetch children', () => {
        // Disable react-query logger output to avoid a console.error
        // when the request fails.
        beforeEach(() => {
          disableQueryLogger();
        });
        afterEach(() => {
          enableQueryLogger();
        });
        it('renders error icon with retry', async () => {
          const {
            data: { id: workflowExecutionId },
            nodeExecutions,
          } = fixture.workflowExecutions.top;
          const parentNodeExecution = nodeExecutions.dynamicNode.data;
          // Simulate an error when attempting to list children of first NE.
          mockServer.insertNodeExecutionList(
            workflowExecutionId,
            notFoundError(parentNodeExecution.id.nodeId),
            {
              [nodeExecutionQueryParams.parentNodeId]: parentNodeExecution.id.nodeId,
            },
          );

          const { container, getByTitle } = renderTable();
          // We expect to find an error icon in place of the child expander
          const errorIconButton = await waitFor(() =>
            getByTitle(titleStrings.childGroupFetchFailed),
          );
          // restore proper handler for node execution children
          insertFixture(mockServer, fixture);
          // click error icon
          await fireEvent.click(errorIconButton);

          // wait for expander and open it to verify children loaded correctly
          const nodeNameEl = await waitFor(() =>
            getByText(container, nodeExecutions.dynamicNode.data.id.nodeId),
          );
          const rowEl = findNearestAncestorByRole(nodeNameEl, 'listitem');
          const childGroups = await expandParentNode(rowEl);
          expect(childGroups.length).toBeGreaterThan(0);
        });
      });
    });

    describe('without isParentNode flag, using taskNodeMetadata', () => {
      let fixture: ReturnType<typeof dynamicPythonTaskWorkflow.generate>;
      beforeEach(() => {
        fixture = dynamicPythonTaskWorkflow.generate();
        workflowExecution = fixture.workflowExecutions.top.data;
        fetchWorkflow.mockImplementation(() => Promise.resolve(fixture.workflows.top));
        executionContext = {
          execution: workflowExecution,
        };
      });

      it('correctly renders children', async () => {
        // The dynamic task node should have a single child node
        // which runs the basic python task. Expand it and then
        // look for the python task name to verify it was rendered.
        const { container } = renderTable();
        const dynamicTaskNameEl = await waitFor(() =>
          getByText(container, fixture.tasks.dynamic.id.name),
        );
        const dynamicRowEl = findNearestAncestorByRole(dynamicTaskNameEl, 'listitem');
        const childContainerList = await expandParentNode(dynamicRowEl);
        await waitFor(() => expect(getByText(childContainerList[0], fixture.tasks.python.id.name)));
      });

      it('correctly renders groups', async () => {
        // We returned two task execution attempts, each with children
        const { container } = renderTable();
        const nodeNameEl = await waitFor(() =>
          getByText(
            container,
            fixture.workflowExecutions.top.nodeExecutions.dynamicNode.data.id.nodeId,
          ),
        );
        const rowEl = findNearestAncestorByRole(nodeNameEl, 'listitem');
        const childGroups = await expandParentNode(rowEl);
        expect(childGroups).toHaveLength(2);
      });
    });

    describe('without isParentNode flag, using workflowNodeMetadata', () => {
      let fixture: ReturnType<typeof dynamicExternalSubWorkflow.generate>;
      let mockGetTaskThroughExecution: any;

      beforeEach(() => {
        fixture = dynamicExternalSubWorkflow.generate();
        insertFixture(mockServer, fixture);
        fetchWorkflow.mockImplementation(() => Promise.resolve(fixture.workflows.top));
        workflowExecution = fixture.workflowExecutions.top.data;
        executionContext = {
          execution: workflowExecution,
        };

        mockGetTaskThroughExecution = jest.spyOn(moduleApi, 'getTaskThroughExecution');
        mockGetTaskThroughExecution.mockImplementation(() => {
          return Promise.resolve({
            ...UNKNOWN_DETAILS,
            displayName: fixture.workflows.sub.id.name,
          });
        });
      });

      afterEach(() => {
        mockGetTaskThroughExecution.mockReset();
      });

      it('correctly renders children', async () => {
        const { container } = renderTable();
        const dynamicTaskNameEl = await waitFor(() =>
          getByText(container, fixture.tasks.generateSubWorkflow.id.name),
        );
        const dynamicRowEl = findNearestAncestorByRole(dynamicTaskNameEl, 'listitem');
        const childContainerList = await expandParentNode(dynamicRowEl);
        await waitFor(() =>
          expect(getByText(childContainerList[0], fixture.workflows.sub.id.name)),
        );
      });

      it('correctly renders groups', async () => {
        const parentNodeId =
          fixture.workflowExecutions.top.nodeExecutions.dynamicWorkflowGenerator.data.metadata
            ?.specNodeId || 'not found';
        // We returned a single WF execution child, so there should only
        // be one child group
        const { container } = renderTable();
        const nodeNameEl = await waitFor(() => getByText(container, parentNodeId));
        const rowEl = findNearestAncestorByRole(nodeNameEl, 'listitem');
        const childGroups = await expandParentNode(rowEl);
        expect(childGroups).toHaveLength(1);
      });
    });
  });

  describe('with a request filter', () => {
    let fixture: ReturnType<typeof oneFailedTaskWorkflow.generate>;

    beforeEach(() => {
      fixture = oneFailedTaskWorkflow.generate();
      workflowExecution = fixture.workflowExecutions.top.data;
      insertFixture(mockServer, fixture);
      fetchWorkflow.mockImplementation(() => Promise.resolve(fixture.workflows.top));
      // Adding a request filter to only show failed NodeExecutions
      requestConfig = {
        filter: [
          {
            key: 'phase',
            operation: FilterOperationName.EQ,
            value: NodeExecutionPhase[NodeExecutionPhase.FAILED],
          },
        ],
      };
      const nodeExecutions = fixture.workflowExecutions.top.nodeExecutions;
      mockServer.insertNodeExecutionList(workflowExecution.id, [nodeExecutions.failedNode.data], {
        filters: 'eq(phase,FAILED)',
      });
      executionContext = {
        execution: workflowExecution,
      };
    });

    it('requests child node executions using configuration from context', async () => {
      const { getByText, queryByText } = renderTable();
      const { nodeExecutions } = fixture.workflowExecutions.top;

      await waitFor(() => expect(getByText(nodeExecutions.failedNode.data.id.nodeId)));

      expect(queryByText(nodeExecutions.pythonNode.data.id.nodeId)).toBeNull();
    });
  });
});
