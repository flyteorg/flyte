import { formatDateUTC, protobufDurationToHMS } from 'common/formatters';
import { timestampToDate } from 'common/utils';
import { NewTargetLink } from 'components/common';
import { useCommonStyles } from 'components/common/styles';
import { useTheme } from 'components/Theme/useTheme';
import { TaskExecutionPhase } from 'models/Execution/enums';
import * as React from 'react';
import { ExecutionStatusBadge } from '..';
import { noLogsFoundString } from '../constants';
import { getUniqueTaskExecutionName } from '../TaskExecutionsList/utils';
import { nodeExecutionsTableColumnWidths } from './constants';
import { SelectNodeExecutionLink } from './SelectNodeExecutionLink';
import { useColumnStyles, useExecutionTableStyles } from './styles';
import {
    TaskExecutionCellRendererData,
    TaskExecutionColumnDefinition
} from './types';
import { splitLogLinksAtWidth } from './utils';

const TaskExecutionLogLinks: React.FC<TaskExecutionCellRendererData> = ({
    execution,
    nodeExecution,
    state
}) => {
    const tableStyles = useExecutionTableStyles();
    const commonStyles = useCommonStyles();
    const { logs = [] } = execution.closure;
    const { measureTextWidth } = useTheme();
    const measuredLogs = React.useMemo(
        () =>
            logs.map(log => ({
                ...log,
                width: measureTextWidth('body1', log.name)
            })),
        [logs, measureTextWidth]
    );

    if (measuredLogs.length === 0) {
        return (
            <span className={commonStyles.hintText}>{noLogsFoundString}</span>
        );
    }

    // Leaving room at the end to render the "xxx More" string
    const [taken, left] = splitLogLinksAtWidth(
        measuredLogs,
        nodeExecutionsTableColumnWidths.logs - 56
    );

    // If we don't have enough room to render any individual links, just
    // show the selection link to open the details panel
    if (!taken.length) {
        return (
            <SelectNodeExecutionLink
                execution={nodeExecution}
                state={state}
                linkText={`View Logs`}
            />
        );
    }

    return (
        <div className={tableStyles.logLinksContainer}>
            {taken.map(({ name, uri }, index) => (
                <NewTargetLink
                    className={tableStyles.logLink}
                    key={index}
                    href={uri}
                >
                    {name}
                </NewTargetLink>
            ))}
            {left.length === 0 ? null : (
                <SelectNodeExecutionLink
                    className={tableStyles.logLink}
                    execution={nodeExecution}
                    state={state}
                    linkText={`${left.length} More`}
                />
            )}
        </div>
    );
};

export function generateColumns(
    styles: ReturnType<typeof useColumnStyles>
): TaskExecutionColumnDefinition[] {
    return [
        {
            cellRenderer: ({ execution }) =>
                getUniqueTaskExecutionName(execution),
            className: styles.columnName,
            key: 'name',
            label: 'node'
        },
        {
            cellRenderer: ({
                execution: {
                    closure: { phase = TaskExecutionPhase.UNDEFINED }
                }
            }) => <ExecutionStatusBadge phase={phase} type="task" />,
            className: styles.columnStatus,
            key: 'phase',
            label: 'status'
        },
        {
            cellRenderer: () => 'Task Execution',
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
            cellRenderer: props => <TaskExecutionLogLinks {...props} />,
            className: styles.columnLogs,
            key: 'logs',
            label: 'logs'
        }
    ];
}
