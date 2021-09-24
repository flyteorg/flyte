import Typography from '@material-ui/core/Typography';
import { makeStyles, Theme } from '@material-ui/core/styles';
import { WaitForData } from 'components/common/WaitForData';
import { useWorkflowExecutionFiltersState } from 'components/Executions/filters/useExecutionFiltersState';
import { WorkflowVersionsTable } from 'components/Executions/Tables/WorkflowVersionsTable';
import { isLoadingState } from 'components/hooks/fetchMachine';
import { useWorkflowVersions } from 'components/hooks/useWorkflowVersions';
import { interactiveTextColor } from 'components/Theme/constants';
import { SortDirection } from 'models/AdminEntity/types';
import { ResourceIdentifier } from 'models/Common/types';
import { executionSortFields } from 'models/Execution/constants';
import { Routes } from 'routes/routes';
import { history } from 'routes/history';
import * as React from 'react';
import { executionFilterGenerator } from './generators';
import { WorkflowVersionsTablePageSize } from './constants';

const useStyles = makeStyles((theme: Theme) => ({
    headerContainer: {
        display: 'flex',
        justifyContent: 'space-between',
        marginLeft: theme.spacing(1),
        marginRight: theme.spacing(1)
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
    versionView?: boolean;
}

/**
 * The tab/page content for viewing a workflow's versions.
 * @param id
 * @param versionView
 */
export const EntityVersions: React.FC<EntityVersionsProps> = ({
    id,
    versionView = false
}) => {
    const { domain, project, resourceType, name } = id;
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
            limit: versionView ? 100 : WorkflowVersionsTablePageSize
        }
    );

    const handleViewAll = React.useCallback(() => {
        history.push(
            Routes.WorkflowVersionDetails.makeUrl(
                project,
                domain,
                name,
                versions.value[0].id.version ?? ''
            )
        );
    }, [project, domain, name, versions]);

    /** Don't render component until finish fetching user profile */
    if (filtersState.filters[4].status !== 'LOADED') {
        return null;
    }

    return (
        <>
            {!versionView && (
                <div className={styles.headerContainer}>
                    <Typography className={styles.header} variant="h6">
                        Recent Workflow Versions
                    </Typography>
                    <Typography
                        className={styles.viewAll}
                        variant="body1"
                        onClick={handleViewAll}
                    >
                        View All
                    </Typography>
                </div>
            )}
            <WaitForData {...versions}>
                <WorkflowVersionsTable
                    {...versions}
                    isFetching={isLoadingState(versions.state)}
                    versionView={versionView}
                />
            </WaitForData>
        </>
    );
};
