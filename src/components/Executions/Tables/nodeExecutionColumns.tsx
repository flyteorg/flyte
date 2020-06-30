import { Typography } from '@material-ui/core';
import {
    formatDateLocalTimezone,
    formatDateUTC,
    millisecondsToHMS
} from 'common/formatters';
import { timestampToDate } from 'common/utils';
import { useCommonStyles } from 'components/common/styles';
import { NodeExecutionPhase } from 'models/Execution/enums';
import * as React from 'react';
import { ExecutionStatusBadge, getNodeExecutionTimingMS } from '..';
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
            cellRenderer: props => (
                <>
                    <NodeExecutionName {...props} />
                    <Typography variant="subtitle1" color="textSecondary">
                        {props.execution.displayId}
                    </Typography>
                </>
            ),
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
                const timing = getNodeExecutionTimingMS(execution);
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
