import Typography from '@material-ui/core/Typography';
import { makeStyles, Theme } from '@material-ui/core/styles';
import { getCacheKey } from 'components/Cache/utils';
import { fetchStates } from 'components/hooks/types';
import { ErrorBoundary } from 'components/common/ErrorBoundary';
import { LargeLoadingSpinner } from 'components/common/LoadingSpinner';
import { DataError } from 'components/Errors/DataError';
import { ExecutionFilters } from 'components/Executions/ExecutionFilters';
import { useWorkflowExecutionFiltersState } from 'components/Executions/filters/useExecutionFiltersState';
import { WorkflowExecutionsTable } from 'components/Executions/Tables/WorkflowExecutionsTable';
import { makeWorkflowExecutionListQuery } from 'components/Executions/workflowExecutionQueries';
import { SortDirection } from 'models/AdminEntity/types';
import { executionSortFields } from 'models/Execution/constants';
import { Execution } from 'models/Execution/types';
import * as React from 'react';
import { useInfiniteQuery } from 'react-query';
import { BarChart } from 'components/common/BarChart';
import {
  getExecutionTimeData,
  getStartExecutionTime,
} from 'components/Entities/EntityExecutionsBarChart';
import classNames from 'classnames';
import { useWorkflowExecutions } from 'components/hooks/useWorkflowExecutions';
import { useExecutionShowArchivedState } from 'components/Executions/filters/useExecutionArchiveState';
import { useOnlyMyExecutionsFilterState } from 'components/Executions/filters/useOnlyMyExecutionsFilterState';
import { WaitForData } from 'components/common/WaitForData';
import { history } from 'routes/history';
import { Routes } from 'routes/routes';
import { compact } from 'lodash';
import { failedToLoadExecutionsString } from './constants';

const useStyles = makeStyles((theme: Theme) => ({
  container: {
    display: 'flex',
    flex: '1 1 auto',
    flexDirection: 'column',
  },
  header: {
    paddingBottom: theme.spacing(1),
    paddingLeft: theme.spacing(1),
    borderBottom: `1px solid ${theme.palette.divider}`,
  },
  marginTop: {
    marginTop: theme.spacing(2),
  },
  chartContainer: {
    paddingLeft: theme.spacing(1),
    paddingRight: theme.spacing(3),
    paddingTop: theme.spacing(1),
  },
}));
export interface ProjectExecutionsProps {
  projectId: string;
  domainId: string;
}

const defaultSort = {
  key: executionSortFields.createdAt,
  direction: SortDirection.DESCENDING,
};

/** A listing of all executions across a project/domain combination. */
export const ProjectExecutions: React.FC<ProjectExecutionsProps> = ({
  domainId: domain,
  projectId: project,
}) => {
  const styles = useStyles();
  const archivedFilter = useExecutionShowArchivedState();
  const filtersState = useWorkflowExecutionFiltersState();
  const onlyMyExecutionsFilterState = useOnlyMyExecutionsFilterState({});

  const allFilters = compact([
    ...filtersState.appliedFilters,
    archivedFilter.getFilter(),
    onlyMyExecutionsFilterState.getFilter(),
  ]);
  const config = {
    sort: defaultSort,
    filter: allFilters,
  };

  // Remount the table whenever we change project/domain/filters to ensure
  // things are virtualized correctly.
  const tableKey = React.useMemo(
    () =>
      getCacheKey({
        domain,
        project,
        filters: allFilters,
      }),
    [domain, project, allFilters],
  );

  const query = useInfiniteQuery({
    ...makeWorkflowExecutionListQuery({ domain, project }, config),
  });

  // useInfiniteQuery returns pages of items, but the table would like a single
  // flat list.
  const executions = React.useMemo(
    () =>
      query.data?.pages
        ? query.data.pages.reduce<Execution[]>((acc, { data }) => acc.concat(data), [])
        : [],
    [query.data?.pages],
  );

  const handleBarChartItemClick = React.useCallback((item) => {
    history.push(Routes.ExecutionDetails.makeUrl(item.metadata));
  }, []);

  // to show only in bar chart view
  const last100Executions = useWorkflowExecutions(
    { domain, project },
    {
      sort: defaultSort,
      filter: allFilters,
      limit: 100,
    },
  );

  const fetch = React.useCallback(() => query.fetchNextPage(), [query]);

  const content = query.isLoadingError ? (
    <DataError error={query.error} errorTitle={failedToLoadExecutionsString} retry={fetch} />
  ) : query.isLoading ? (
    <LargeLoadingSpinner />
  ) : (
    <WorkflowExecutionsTable
      key={tableKey}
      fetch={fetch}
      value={executions}
      lastError={query.error}
      moreItemsAvailable={!!query.hasNextPage}
      showWorkflowName={true}
      isFetching={query.isFetching}
      data-testid="workflow-table"
    />
  );

  return (
    <div className={styles.container}>
      <Typography className={classNames(styles.header, styles.marginTop)} variant="h6">
        Last 100 Executions in the Project
      </Typography>
      <div className={styles.chartContainer}>
        <WaitForData {...last100Executions}>
          <BarChart
            chartIds={[]}
            data={getExecutionTimeData(last100Executions.value)}
            startDate={getStartExecutionTime(last100Executions.value)}
            onClickItem={handleBarChartItemClick}
          />
        </WaitForData>
      </div>
      <Typography className={styles.header} variant="h6">
        All Executions in the Project
      </Typography>
      <ExecutionFilters
        {...filtersState}
        showArchived={archivedFilter.showArchived}
        onArchiveFilterChange={archivedFilter.setShowArchived}
        onlyMyExecutionsFilterState={onlyMyExecutionsFilterState}
      />
      <ErrorBoundary>{content}</ErrorBoundary>
    </div>
  );
};
