import { Typography } from '@material-ui/core';
import { formatDateUTC, protobufDurationToHMS } from 'common/formatters';
import { timestampToDate } from 'common/utils';
import { useCommonStyles } from 'components/common/styles';
import { NodeExecutionPhase } from 'models/Execution/enums';
import * as React from 'react';
import { ExecutionStatusBadge } from '..';
import { SelectNodeExecutionLink } from './SelectNodeExecutionLink';
import { useColumnStyles } from './styles';
import {
    NodeExecutionCellRendererData,
    NodeExecutionColumnDefinition
} from './types';

const NodeExecutionName: React.FC<NodeExecutionCellRendererData> = ({
    execution,
    state
}) => {
    const commonStyles = useCommonStyles();
    const styles = useColumnStyles();
    const name = execution.id.nodeId;

    if (execution === state.selectedExecution) {
        return (
            <Typography
                variant="body1"
                className={styles.selectedExecutionName}
            >
                {name}
            </Typography>
        );
    }
    return (
        <SelectNodeExecutionLink
            className={commonStyles.primaryLink}
            execution={execution}
            linkText={name}
            state={state}
        />
    );
};

export function generateColumns(
    styles: ReturnType<typeof useColumnStyles>
): NodeExecutionColumnDefinition[] {
    return [
        {
            cellRenderer: props => <NodeExecutionName {...props} />,
            className: styles.columnName,
            key: 'name',
            label: 'node'
        },
        {
            cellRenderer: ({
                execution: {
                    closure: { phase = NodeExecutionPhase.UNDEFINED }
                }
            }) => <ExecutionStatusBadge phase={phase} type="node" />,
            className: styles.columnStatus,
            key: 'phase',
            label: 'status'
        },
        {
            cellRenderer: ({ execution: { displayType } }) => displayType,
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
            cellRenderer: ({ execution, state }) => (
                <SelectNodeExecutionLink
                    execution={execution}
                    linkText="View Logs"
                    state={state}
                />
            ),
            className: styles.columnLogs,
            key: 'logs',
            label: 'logs'
        }
    ];
}
