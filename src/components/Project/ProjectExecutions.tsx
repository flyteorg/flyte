import { makeStyles } from '@material-ui/core/styles';
import { getCacheKey } from 'components/Cache';
import { ErrorBoundary, LargeLoadingSpinner } from 'components/common';
import { DataError } from 'components/Errors/DataError';
import { ExecutionFilters } from 'components/Executions/ExecutionFilters';
import { useWorkflowExecutionFiltersState } from 'components/Executions/filters/useExecutionFiltersState';
import { WorkflowExecutionsTable } from 'components/Executions/Tables/WorkflowExecutionsTable';
import { makeWorkflowExecutionListQuery } from 'components/Executions/workflowExecutionQueries';
import { Execution, executionSortFields, SortDirection } from 'models';
import * as React from 'react';
import { useInfiniteQuery } from 'react-query';
import { failedToLoadExecutionsString } from './constants';

const useStyles = makeStyles(() => ({
    container: {
        display: 'flex',
        flex: '1 1 auto',
        flexDirection: 'column'
    }
}));
export interface ProjectExecutionsProps {
    projectId: string;
    domainId: string;
}

const defaultSort = {
    key: executionSortFields.createdAt,
    direction: SortDirection.DESCENDING
};

/** A listing of all executions across a project/domain combination. */
export const ProjectExecutions: React.FC<ProjectExecutionsProps> = ({
    domainId: domain,
    projectId: project
}) => {
    const styles = useStyles();
    const filtersState = useWorkflowExecutionFiltersState();

    const config = {
        sort: defaultSort,
        filter: filtersState.appliedFilters
    };

    // Remount the table whenever we change project/domain/filters to ensure
    // things are virtualized correctly.
    const tableKey = React.useMemo(
        () =>
            getCacheKey({
                domain,
                project,
                filters: filtersState.appliedFilters
            }),
        [domain, project, filtersState.appliedFilters]
    );

    const query = useInfiniteQuery({
        ...makeWorkflowExecutionListQuery({ domain, project }, config)
    });

    // useInfiniteQuery returns pages of items, but the table would like a single
    // flat list.
    const executions = React.useMemo(
        () =>
            query.data?.pages
                ? query.data.pages.reduce<Execution[]>(
                      (acc, { data }) => acc.concat(data),
                      []
                  )
                : [],
        [query.data?.pages]
    );

    const fetch = React.useCallback(() => query.fetchNextPage(), [query]);

    const content = query.isLoadingError ? (
        <DataError
            error={query.error}
            errorTitle={failedToLoadExecutionsString}
            retry={fetch}
        />
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
        />
    );

    return (
        <div className={styles.container}>
            <ExecutionFilters {...filtersState} />
            <ErrorBoundary>{content}</ErrorBoundary>
        </div>
    );
};
