import { Typography } from '@material-ui/core';
import { formatDateUTC } from 'common/formatters';
import { timestampToDate } from 'common/utils';
import { useCommonStyles } from 'components/common/styles';
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
    const commonStyles = useCommonStyles();
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
            // {
            //     cellRenderer: ({ workflow: { closure } }) => {
            //         return (
            //             <Typography variant="body1">
            //                 1.8.1
            //             </Typography>
            //         );
            //     },
            //     className: styles.columnRelease,
            //     key: 'release',
            //     label: 'Release'
            // },
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
            // {
            //     cellRenderer: ({ workflow }) => {
            //         // const timing = getWorkflowExecutionTimingMS(workflow);
            //         // return (
            //         //     <Typography variant="body1">
            //         //         {timing !== null
            //         //             ? millisecondsToHMS(timing.duration)
            //         //             : ''}
            //         //     </Typography>
            //         // );
            //         return <Typography variant="body1">NA hours ago</Typography>
            //     },
            //     className: styles.columnLastRun,
            //     key: 'duration',
            //     label: 'last execution'
            // },
            // {
            //     cellRenderer: ({ workflow, state }) => {
            //         return (
            //             <Link
            //                 component="button"
            //                 variant="body1"
            //             >
            //                 View Inputs &amp; Outputs
            //             </Link>
            //         );
            //     },
            //     className: styles.columnRecentRun,
            //     key: 'inputsOutputs',
            //     label: 'recent run'
            // }
        ],
        [styles, commonStyles]
    );
}
