import { Link, Typography } from '@material-ui/core';
import {
    dateFromNow,
    formatDateLocalTimezone,
    formatDateUTC,
    millisecondsToHMS
} from 'common/formatters';
import { timestampToDate } from 'common/utils';
import { useCommonStyles } from 'components/common/styles';
import { WorkflowExecutionPhase } from 'models/Execution/enums';
import * as React from 'react';
import { ExecutionStatusBadge } from '../ExecutionStatusBadge';
import { getWorkflowExecutionTimingMS } from '../utils';
import { useWorkflowExecutionsColumnStyles } from './styles';
import { WorkflowExecutionColumnDefinition } from './types';
import { WorkflowExecutionLink } from './WorkflowExecutionLink';

interface WorkflowExecutionColumnOptions {
    showWorkflowName: boolean;
}
/** Returns a memoized list of column definitions to use when rendering a
 * `WorkflowExecutionRow`. Memoization is based on common/column style objects
 * and any fields in the incoming `WorkflowExecutionColumnOptions` object.
 */
export function useWorkflowExecutionsTableColumns({
    showWorkflowName
}: WorkflowExecutionColumnOptions): WorkflowExecutionColumnDefinition[] {
    const styles = useWorkflowExecutionsColumnStyles();
    const commonStyles = useCommonStyles();
    return React.useMemo(
        () => [
            {
                cellRenderer: ({
                    execution: {
                        id,
                        closure: { startedAt, workflowId }
                    }
                }) => (
                    <>
                        <WorkflowExecutionLink id={id} />
                        <Typography
                            className={commonStyles.textWrapped}
                            variant="subtitle1"
                            color="textSecondary"
                        >
                            {showWorkflowName
                                ? workflowId.name
                                : startedAt
                                ? `Last run ${dateFromNow(
                                      timestampToDate(startedAt)
                                  )}`
                                : ''}
                        </Typography>
                    </>
                ),
                className: styles.columnName,
                key: 'name',
                label: 'execution id'
            },
            {
                cellRenderer: ({
                    execution: {
                        closure: { phase = WorkflowExecutionPhase.UNDEFINED }
                    }
                }) => <ExecutionStatusBadge phase={phase} type="workflow" />,
                className: styles.columnStatus,
                key: 'phase',
                label: 'status'
            },
            {
                cellRenderer: ({ execution: { closure } }) => {
                    const { startedAt } = closure;
                    if (!startedAt) {
                        return '';
                    }
                    const startedAtDate = timestampToDate(startedAt);
                    return (
                        <>
                            <Typography variant="body1">
                                {formatDateUTC(startedAtDate)}
                            </Typography>
                            <Typography
                                variant="subtitle1"
                                color="textSecondary"
                            >
                                {formatDateLocalTimezone(startedAtDate)}
                            </Typography>
                        </>
                    );
                },
                className: styles.columnStartedAt,
                key: 'startedAt',
                label: 'start time'
            },
            {
                cellRenderer: ({ execution }) => {
                    const timing = getWorkflowExecutionTimingMS(execution);
                    return (
                        <Typography variant="body1">
                            {timing !== null
                                ? millisecondsToHMS(timing.duration)
                                : ''}
                        </Typography>
                    );
                },
                className: styles.columnDuration,
                key: 'duration',
                label: 'duration'
            },
            {
                cellRenderer: ({ execution, state }) => {
                    const onClick = () =>
                        state.setSelectedIOExecution(execution);
                    return (
                        <Link
                            component="button"
                            onClick={onClick}
                            variant="body1"
                        >
                            View Inputs &amp; Outputs
                        </Link>
                    );
                },
                className: styles.columnInputsOutputs,
                key: 'inputsOutputs',
                label: ''
            }
        ],
        [styles, commonStyles, showWorkflowName]
    );
}
