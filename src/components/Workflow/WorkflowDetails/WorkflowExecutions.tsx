import { Typography } from '@material-ui/core';
import { makeStyles, Theme } from '@material-ui/core/styles';
import { contentMarginGridUnits } from 'common/layout';
import { WaitForData } from 'components/common';
import { ExecutionFilters } from 'components/Executions/ExecutionFilters';
import { useWorkflowExecutionFiltersState } from 'components/Executions/filters/useExecutionFiltersState';
import { WorkflowExecutionsTable } from 'components/Executions/Tables/WorkflowExecutionsTable';
import { useWorkflowExecutions } from 'components/hooks';
import { NamedEntityIdentifier } from 'models';
import { FilterOperationName, SortDirection } from 'models/AdminEntity';
import { executionSortFields } from 'models/Execution';
import * as React from 'react';

const useStyles = makeStyles((theme: Theme) => ({
    filtersContainer: {
        borderTop: `1px solid ${theme.palette.divider}`
    },
    header: {
        marginBottom: theme.spacing(1),
        marginLeft: theme.spacing(contentMarginGridUnits)
    }
}));

export interface WorkflowExecutionsProps {
    workflowId: NamedEntityIdentifier;
}

/** The tab/page content for viewing a workflow's executions */
export const WorkflowExecutions: React.FC<WorkflowExecutionsProps> = ({
    workflowId: { project, domain, name }
}) => {
    const styles = useStyles();
    const filtersState = useWorkflowExecutionFiltersState();
    const sort = {
        key: executionSortFields.createdAt,
        direction: SortDirection.DESCENDING
    };

    const executions = useWorkflowExecutions(
        { domain, project },
        {
            sort,
            filter: [
                {
                    key: 'workflow.name',
                    operation: FilterOperationName.EQ,
                    value: name
                },
                ...filtersState.appliedFilters
            ]
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
                <WorkflowExecutionsTable {...executions} />
            </WaitForData>
        </>
    );
};
