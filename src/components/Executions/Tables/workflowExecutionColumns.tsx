import { Typography } from '@material-ui/core';
import {
    formatDateLocalTimezone,
    formatDateUTC,
    millisecondsToHMS
} from 'common/formatters';
import { timestampToDate } from 'common/utils';
import { WorkflowExecutionPhase } from 'models/Execution/enums';
import * as React from 'react';
import { ExecutionStatusBadge, getWorkflowExecutionTimingMS } from '..';
import { useColumnStyles } from './styles';
import { WorkflowExecutionColumnDefinition } from './types';

export function generateColumns(
    styles: ReturnType<typeof useColumnStyles>
): WorkflowExecutionColumnDefinition[] {
    return [
        {
            cellRenderer: ({ execution }) => execution.id.name,
            className: styles.columnName,
            key: 'name',
            label: 'node'
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
            cellRenderer: () => 'Workflow Execution',
            className: styles.columnType,
            key: 'type',
            label: 'type'
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
                        <Typography variant="subtitle1" color="textSecondary">
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
                if (timing === null) {
                    return '';
                }
                return (
                    <>
                        <Typography variant="body1">
                            {millisecondsToHMS(timing.duration)}
                        </Typography>
                    </>
                );
            },
            className: styles.columnDuration,
            key: 'duration',
            label: () => (
                <>
                    <Typography component="div" variant="overline">
                        duration
                    </Typography>
                    <Typography
                        component="div"
                        variant="subtitle1"
                        color="textSecondary"
                    >
                        Queued Time
                    </Typography>
                </>
            )
        },
        {
            // TODO: Implement this content
            cellRenderer: () => null,
            className: styles.columnLogs,
            key: 'logs',
            label: 'logs'
        }
    ];
}
