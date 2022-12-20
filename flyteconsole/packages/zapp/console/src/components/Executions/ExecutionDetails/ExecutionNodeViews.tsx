import * as React from 'react';
import { Tab, Tabs } from '@material-ui/core';
import { makeStyles, Theme } from '@material-ui/core/styles';
import { WaitForQuery } from 'components/common/WaitForQuery';
import { DataError } from 'components/Errors/DataError';
import { useTabState } from 'components/hooks/useTabState';
import { secondaryBackgroundColor } from 'components/Theme/constants';
import { Execution, ExternalResource, LogsByPhase, NodeExecution } from 'models/Execution/types';
import { useEffect, useMemo, useState } from 'react';
import { keyBy } from 'lodash';
import { isMapTaskV1 } from 'models/Task/utils';
import { useQueryClient } from 'react-query';
import { LargeLoadingSpinner } from 'components/common/LoadingSpinner';
import { FilterOperation } from 'models/AdminEntity/types';
import { NodeExecutionDetailsContextProvider } from '../contextProvider/NodeExecutionDetails';
import { NodeExecutionsByIdContext } from '../contexts';
import { ExecutionFilters } from '../ExecutionFilters';
import { useNodeExecutionFiltersState } from '../filters/useExecutionFiltersState';
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

const isPhaseFilter = (appliedFilters: FilterOperation[]) => {
  if (appliedFilters.length === 1 && appliedFilters[0].key === 'phase') {
    return true;
  }
  return false;
};

interface WorkflowNodeExecution extends NodeExecution {
  logsByPhase?: LogsByPhase;
}

interface ExecutionNodeViewsProps {
  execution: Execution;
}

/** Contains the available ways to visualize the nodes of a WorkflowExecution */
export const ExecutionNodeViews: React.FC<ExecutionNodeViewsProps> = ({ execution }) => {
  const defaultTab = tabs.nodes.id;
  const styles = useStyles();
  const filterState = useNodeExecutionFiltersState();
  const tabState = useTabState(tabs, defaultTab);
  const queryClient = useQueryClient();
  const [nodeExecutionsLoading, setNodeExecutionsLoading] = useState<boolean>(true);

  const {
    closure: { workflowId },
  } = execution;

  const [nodeExecutions, setNodeExecutions] = useState<NodeExecution[]>([]);
  const [nodeExecutionsWithResources, setNodeExecutionsWithResources] = useState<
    WorkflowNodeExecution[]
  >([]);

  const nodeExecutionsById = useMemo(() => {
    return keyBy(nodeExecutionsWithResources, 'scopedId');
  }, [nodeExecutionsWithResources]);

  // query to get all data to build Graph and Timeline
  const { nodeExecutionsQuery } = useExecutionNodeViewsState(execution);
  // query to get filtered data to narrow down Table outputs
  const {
    nodeExecutionsQuery: { data: filteredNodeExecutions },
  } = useExecutionNodeViewsState(execution, filterState.appliedFilters);

  useEffect(() => {
    let isCurrent = true;

    async function fetchData(baseNodeExecutions, queryClient) {
      setNodeExecutionsLoading(true);
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
        setNodeExecutionsLoading(false);
      }
    }

    if (nodeExecutions.length > 0) {
      fetchData(nodeExecutions, queryClient);
    } else {
      if (isCurrent) {
        setNodeExecutionsLoading(false);
      }
    }
    return () => {
      isCurrent = false;
    };
  }, [nodeExecutions]);

  const childGroupsQuery = useAllTreeNodeExecutionGroupsQuery(nodeExecutionsQuery.data ?? [], {});

  useEffect(() => {
    if (!childGroupsQuery.isLoading && childGroupsQuery.data) {
      setNodeExecutions(childGroupsQuery.data);
    }
  }, [childGroupsQuery.data]);

  const LoadingComponent = () => {
    return (
      <div className={styles.loading}>
        <LargeLoadingSpinner />
      </div>
    );
  };

  const renderTab = (tabType) => {
    if (nodeExecutionsLoading) {
      return <LoadingComponent />;
    }
    return (
      <WaitForQuery
        errorComponent={DataError}
        query={childGroupsQuery}
        loadingComponent={LoadingComponent}
      >
        {() => (
          <ExecutionTab
            tabType={tabType}
            // if only phase filter was applied, ignore request response, and filter out nodes via frontend filter
            filteredNodeExecutions={
              isPhaseFilter(filterState.appliedFilters) ? undefined : filteredNodeExecutions
            }
          />
        )}
      </WaitForQuery>
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
          <div className={styles.nodesContainer}>
            {tabState.value === tabs.nodes.id && (
              <div className={styles.filters}>
                <ExecutionFilters {...filterState} />
              </div>
            )}
            <WaitForQuery
              errorComponent={DataError}
              query={nodeExecutionsQuery}
              loadingComponent={LoadingComponent}
            >
              {() => renderTab(tabState.value)}
            </WaitForQuery>
          </div>
        </NodeExecutionsByIdContext.Provider>
      </NodeExecutionDetailsContextProvider>
    </>
  );
};
