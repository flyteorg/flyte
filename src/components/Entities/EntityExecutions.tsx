import { Typography } from '@material-ui/core';
import { makeStyles, Theme } from '@material-ui/core/styles';
import { contentMarginGridUnits } from 'common/layout';
import { WaitForData } from 'components/common';
import { ExecutionFilters } from 'components/Executions/ExecutionFilters';
import { useWorkflowExecutionFiltersState } from 'components/Executions/filters/useExecutionFiltersState';
import { WorkflowExecutionsTable as ExecutionsTable } from 'components/Executions/Tables/WorkflowExecutionsTable';
import { useWorkflowExecutions as useExecutions } from 'components/hooks';
import { isLoadingState } from 'components/hooks/fetchMachine';
import { ResourceIdentifier } from 'models';
import { SortDirection } from 'models/AdminEntity';
import { executionSortFields } from 'models/Execution';
import * as React from 'react';
import { executionFilterGenerator } from './generators';

const useStyles = makeStyles((theme: Theme) => ({
    filtersContainer: {
        borderTop: `1px solid ${theme.palette.divider}`
    },
    header: {
        marginBottom: theme.spacing(1),
        marginLeft: theme.spacing(contentMarginGridUnits)
    }
}));

export interface EntityExecutionsProps {
    id: ResourceIdentifier;
}

/** The tab/page content for viewing a workflow's executions */
export const EntityExecutions: React.FC<EntityExecutionsProps> = ({ id }) => {
    const { domain, project, resourceType } = id;
    const styles = useStyles();
    const filtersState = useWorkflowExecutionFiltersState();
    const sort = {
        key: executionSortFields.createdAt,
        direction: SortDirection.DESCENDING
    };

    const baseFilters = React.useMemo(
        () => executionFilterGenerator[resourceType](id),
        [id]
    );

    const executions = useExecutions(
        { domain, project },
        {
            sort,
            filter: [...baseFilters, ...filtersState.appliedFilters]
        }
    );

    return (
        <>
            <Typography className={styles.header} variant="h6">
                Executions
            </Typography>
            <div className={styles.filtersContainer}>
                <ExecutionFilters {...filtersState} />
            </div>
            <WaitForData {...executions}>
                <ExecutionsTable {...executions} isFetching={isLoadingState(executions.state)} />
            </WaitForData>
        </>
    );
};
