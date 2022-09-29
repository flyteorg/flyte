import * as React from 'react';
import { Tab, Tabs } from '@material-ui/core';
import { makeStyles, Theme } from '@material-ui/core/styles';
import { WaitForQuery } from 'components/common/WaitForQuery';
import { DataError } from 'components/Errors/DataError';
import { useTabState } from 'components/hooks/useTabState';
import { secondaryBackgroundColor } from 'components/Theme/constants';
import { Execution, ExternalResource, LogsByPhase, NodeExecution } from 'models/Execution/types';
import { useContext, useEffect, useMemo, useState } from 'react';
import { keyBy } from 'lodash';
import { isMapTaskV1 } from 'models/Task/utils';
import { useQueryClient } from 'react-query';
import { LargeLoadingSpinner } from 'components/common/LoadingSpinner';
import { NodeExecutionDetailsContextProvider } from '../contextProvider/NodeExecutionDetails';
import { NodeExecutionsByIdContext, NodeExecutionsRequestConfigContext } from '../contexts';
import { ExecutionFilters } from '../ExecutionFilters';
import { useNodeExecutionFiltersState } from '../filters/useExecutionFiltersState';
import { NodeExecutionsTable } from '../Tables/NodeExecutionsTable';
import { tabs } from './constants';
import { useExecutionNodeViewsState } from './useExecutionNodeViewsState';
import { fetchTaskExecutionList } from '../taskExecutionQueries';
import { getGroupedLogs } from '../TaskExecutionsList/utils';
import { useAllTreeNodeExecutionGroupsQuery } from '../nodeExecutionQueries';
import { ExecutionTab } from './ExecutionTab';

const useStyles = makeStyles((theme: Theme) => ({
  filters: {
    paddingLeft: theme.spacing(3),
  },
  nodesContainer: {
    borderTop: `1px solid ${theme.palette.divider}`,
    display: 'flex',
    flex: '1 1 100%',
    flexDirection: 'column',
    minHeight: 0,
  },
  tabs: {
    background: secondaryBackgroundColor,
    paddingLeft: theme.spacing(3.5),
  },
  loading: {
    margin: 'auto',
  },
}));

interface WorkflowNodeExecution extends NodeExecution {
  logsByPhase?: LogsByPhase;
}

export interface ExecutionNodeViewsProps {
  execution: Execution;
}

/** Contains the available ways to visualize the nodes of a WorkflowExecution */
export const ExecutionNodeViews: React.FC<ExecutionNodeViewsProps> = ({ execution }) => {
  const defaultTab = tabs.nodes.id;
  const styles = useStyles();
  const filterState = useNodeExecutionFiltersState();
  const tabState = useTabState(tabs, defaultTab);
  const queryClient = useQueryClient();
  const requestConfig = useContext(NodeExecutionsRequestConfigContext);

  const {
    closure: { abortMetadata, workflowId },
  } = execution;

  const [nodeExecutions, setNodeExecutions] = useState<NodeExecution[]>([]);
  const [nodeExecutionsWithResources, setNodeExecutionsWithResources] = useState<
    WorkflowNodeExecution[]
  >([]);

  const nodeExecutionsById = useMemo(() => {
    return keyBy(nodeExecutionsWithResources, 'scopedId');
  }, [nodeExecutionsWithResources]);

  /* We want to maintain the filter selection when switching away from the Nodes
    tab and back, but do not want to filter the nodes when viewing the graph. So,
    we will only pass filters to the execution state when on the nodes tab. */
  const appliedFilters = tabState.value === tabs.nodes.id ? filterState.appliedFilters : [];

  const { nodeExecutionsQuery, nodeExecutionsRequestConfig } = useExecutionNodeViewsState(
    execution,
    appliedFilters,
  );

  useEffect(() => {
    let isCurrent = true;
    async function fetchData(baseNodeExecutions, queryClient) {
      const newValue = await Promise.all(
        baseNodeExecutions.map(async (baseNodeExecution) => {
          const taskExecutions = await fetchTaskExecutionList(queryClient, baseNodeExecution.id);

          const useNewMapTaskView = taskExecutions.every((taskExecution) => {
            const {
              closure: { taskType, metadata, eventVersion = 0 },
            } = taskExecution;
            return isMapTaskV1(
              eventVersion,
              metadata?.externalResources?.length ?? 0,
              taskType ?? undefined,
            );
          });
          const externalResources: ExternalResource[] = taskExecutions
            .map((taskExecution) => taskExecution.closure.metadata?.externalResources)
            .flat()
            .filter((resource): resource is ExternalResource => !!resource);

          const logsByPhase: LogsByPhase = getGroupedLogs(externalResources);

          return {
            ...baseNodeExecution,
            ...(useNewMapTaskView && logsByPhase.size > 0 && { logsByPhase }),
          };
        }),
      );

      if (isCurrent) {
        setNodeExecutionsWithResources(newValue);
      }
    }

    if (nodeExecutions.length > 0) {
      fetchData(nodeExecutions, queryClient);
    }
    return () => {
      isCurrent = false;
    };
  }, [nodeExecutions]);

  const childGroupsQuery = useAllTreeNodeExecutionGroupsQuery(
    nodeExecutionsQuery.data ?? [],
    requestConfig,
  );

  useEffect(() => {
    if (!childGroupsQuery.isLoading && childGroupsQuery.data) {
      setNodeExecutions(childGroupsQuery.data);
    }
  }, [childGroupsQuery.data]);

  const renderNodeExecutionsTable = (nodeExecutions: NodeExecution[]) => (
    <NodeExecutionsRequestConfigContext.Provider value={nodeExecutionsRequestConfig}>
      <NodeExecutionsTable
        abortMetadata={abortMetadata ?? undefined}
        nodeExecutions={nodeExecutions}
      />
    </NodeExecutionsRequestConfigContext.Provider>
  );

  const renderTab = (tabType) => <ExecutionTab tabType={tabType} />;

  const TimelineLoading = () => {
    return (
      <div className={styles.loading}>
        <LargeLoadingSpinner />
      </div>
    );
  };

  return (
    <>
      <Tabs className={styles.tabs} {...tabState}>
        <Tab value={tabs.nodes.id} label={tabs.nodes.label} />
        <Tab value={tabs.graph.id} label={tabs.graph.label} />
        <Tab value={tabs.timeline.id} label={tabs.timeline.label} />
      </Tabs>
      <NodeExecutionDetailsContextProvider workflowId={workflowId}>
        <NodeExecutionsByIdContext.Provider value={nodeExecutionsById}>
          {nodeExecutions.length > 0 ? (
            <div className={styles.nodesContainer}>
              {tabState.value === tabs.nodes.id ? (
                <>
                  <div className={styles.filters}>
                    <ExecutionFilters {...filterState} />
                  </div>
                  <WaitForQuery errorComponent={DataError} query={nodeExecutionsQuery}>
                    {renderNodeExecutionsTable}
                  </WaitForQuery>
                </>
              ) : (
                <WaitForQuery errorComponent={DataError} query={nodeExecutionsQuery}>
                  {() => renderTab(tabState.value)}
                </WaitForQuery>
              )}
            </div>
          ) : null}
        </NodeExecutionsByIdContext.Provider>
      </NodeExecutionDetailsContextProvider>
    </>
  );
};
