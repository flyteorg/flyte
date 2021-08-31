import { Typography } from '@material-ui/core';
import { makeStyles, Theme } from '@material-ui/core/styles';
import { contentMarginGridUnits } from 'common/layout';
import { WaitForData } from 'components/common/WaitForData';
import { useWorkflowExecutionFiltersState } from 'components/Executions/filters/useExecutionFiltersState';
import { WorkflowVersionsTable } from 'components/Executions/Tables/WorkflowVersionsTable';
import { isLoadingState } from 'components/hooks/fetchMachine';
import { useWorkflowVersions } from 'components/hooks/useWorkflowVersions';
import { interactiveTextColor } from 'components/Theme/constants';
import { SortDirection } from 'models/AdminEntity/types';
import { ResourceIdentifier } from 'models/Common/types';
import { executionSortFields } from 'models/Execution/constants';
import * as React from 'react';
import { executionFilterGenerator } from './generators';
import { WorkflowVersionsTablePageSize } from './constants';

const useStyles = makeStyles((theme: Theme) => ({
    headerContainer: {
        display: 'flex',
        justifyContent: 'space-between',
        marginLeft: theme.spacing(contentMarginGridUnits),
        marginRight: theme.spacing(contentMarginGridUnits)
    },
    header: {
        marginBottom: theme.spacing(1)
    },
    viewAll: {
        color: interactiveTextColor,
        cursor: 'pointer'
    }
}));

export interface EntityVersionsProps {
    id: ResourceIdentifier;
}

/**
 * The tab/page content for viewing a workflow's versions.
 * @param id
 */
export const EntityVersions: React.FC<EntityVersionsProps> = ({ id }) => {
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

    const versions = useWorkflowVersions(
        { domain, project },
        {
            sort,
            filter: [...baseFilters, ...filtersState.appliedFilters],
            limit: WorkflowVersionsTablePageSize
        }
    );

    /** Don't render component until finish fetching user profile */
    if (filtersState.filters[4].status !== 'LOADED') {
        return null;
    }

    return (
        <>
            <div className={styles.headerContainer}>
                <Typography className={styles.header} variant="h6">
                    Recent Workflow Versions
                </Typography>
                <Typography className={styles.viewAll} variant="body1">
                    View All
                </Typography>
            </div>
            <WaitForData {...versions}>
                <WorkflowVersionsTable
                    {...versions}
                    isFetching={isLoadingState(versions.state)}
                />
            </WaitForData>
        </>
    );
};
