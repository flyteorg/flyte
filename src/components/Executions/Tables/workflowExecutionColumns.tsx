import { formatDateUTC, protobufDurationToHMS } from 'common/formatters';
import { timestampToDate } from 'common/utils';
import { WorkflowExecutionPhase } from 'models/Execution/enums';
import * as React from 'react';
import { ExecutionStatusBadge } from '..';
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
                return startedAt
                    ? formatDateUTC(timestampToDate(startedAt))
                    : '';
            },
            className: styles.columnStartedAt,
            key: 'startedAt',
            label: 'start time'
        },
        {
            cellRenderer: ({ execution: { closure } }) => {
                const { duration } = closure;
                return duration ? protobufDurationToHMS(duration!) : '';
            },
            className: styles.columnDuration,
            key: 'duration',
            label: 'duration'
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
