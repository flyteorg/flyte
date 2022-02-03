import { Typography } from '@material-ui/core';
import { formatDateUTC } from 'common/formatters';
import { timestampToDate } from 'common/utils';
import * as React from 'react';
import { useWorkflowVersionsColumnStyles } from './styles';
import { WorkflowVersionColumnDefinition } from './types';

/**
 * Returns a memoized list of column definitions to use when rendering a
 * `WorkflowVersionRow`. Memoization is based on common/column style objects
 * and any fields in the incoming `WorkflowExecutionColumnOptions` object.
 */
export function useWorkflowVersionsTableColumns(): WorkflowVersionColumnDefinition[] {
    const styles = useWorkflowVersionsColumnStyles();
    return React.useMemo(
        () => [
            {
                cellRenderer: ({
                    workflow: {
                        id: { version }
                    }
                }) => <Typography variant="body1">{version}</Typography>,
                className: styles.columnName,
                key: 'name',
                label: 'version id'
            },
            {
                cellRenderer: ({ workflow: { closure } }) => {
                    if (!closure?.createdAt) {
                        return '';
                    }
                    const createdAtDate = timestampToDate(closure.createdAt);
                    return (
                        <Typography variant="body1">
                            {formatDateUTC(createdAtDate)}
                        </Typography>
                    );
                },
                className: styles.columnCreatedAt,
                key: 'createdAt',
                label: 'time created'
            }
        ],
        [styles]
    );
}
