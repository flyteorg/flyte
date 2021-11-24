import { Typography } from '@material-ui/core';
import {
    formatDateLocalTimezone,
    formatDateUTC,
    millisecondsToHMS
} from 'common/formatters';
import { timestampToDate } from 'common/utils';
import { useCommonStyles } from 'components/common/styles';
import { WaitForQuery } from 'components/common/WaitForQuery';
import { Core } from 'flyteidl';
import { isEqual } from 'lodash';
import { NodeExecutionPhase } from 'models/Execution/enums';
import { TaskNodeMetadata } from 'models/Execution/types';
import * as React from 'react';
import { ExecutionStatusBadge } from '../ExecutionStatusBadge';
import { NodeExecutionCacheStatus } from '../NodeExecutionCacheStatus';
import { NodeExecutionDetails } from '../types';
import { useNodeExecutionDetails } from '../useNodeExecutionDetails';
import { getNodeExecutionTimingMS } from '../utils';
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
    const detailsQuery = useNodeExecutionDetails(execution);
    const commonStyles = useCommonStyles();
    const styles = useColumnStyles();
    const nodeId = execution.id.nodeId;

    const isSelected =
        state.selectedExecution != null &&
        isEqual(execution.id, state.selectedExecution);

    const renderReadableName = ({
        displayId,
        displayName
    }: NodeExecutionDetails) => {
        const readableName = isSelected ? (
            <Typography
                variant="body1"
                className={styles.selectedExecutionName}
            >
                {displayId || nodeId}
            </Typography>
        ) : (
            <SelectNodeExecutionLink
                className={commonStyles.primaryLink}
                execution={execution}
                linkText={displayId || nodeId}
                state={state}
            />
        );
        return (
            <>
                {readableName}
                <Typography variant="subtitle1" color="textSecondary">
                    {displayName}
                </Typography>
            </>
        );
    };

    return (
        <>
            <WaitForQuery query={detailsQuery}>
                {renderReadableName}
            </WaitForQuery>
        </>
    );
};

const NodeExecutionDisplayType: React.FC<NodeExecutionCellRendererData> = ({
    execution
}) => {
    const detailsQuery = useNodeExecutionDetails(execution);
    const extractDisplayType = ({ displayType }: NodeExecutionDetails) =>
        displayType;
    return (
        <WaitForQuery query={detailsQuery}>{extractDisplayType}</WaitForQuery>
    );
};

const hiddenCacheStatuses = [
    Core.CatalogCacheStatus.CACHE_MISS,
    Core.CatalogCacheStatus.CACHE_DISABLED
];
function hasCacheStatus(
    taskNodeMetadata?: TaskNodeMetadata
): taskNodeMetadata is TaskNodeMetadata {
    if (!taskNodeMetadata) {
        return false;
    }
    const { cacheStatus } = taskNodeMetadata;
    return !hiddenCacheStatuses.includes(cacheStatus);
}

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
                    closure: {
                        phase = NodeExecutionPhase.UNDEFINED,
                        taskNodeMetadata
                    }
                }
            }) => (
                <>
                    <ExecutionStatusBadge phase={phase} type="node" />
                    {hasCacheStatus(taskNodeMetadata) ? (
                        <NodeExecutionCacheStatus
                            taskNodeMetadata={taskNodeMetadata}
                            variant="iconOnly"
                        />
                    ) : null}
                </>
            ),
            className: styles.columnStatus,
            key: 'phase',
            label: 'status'
        },
        {
            cellRenderer: props => <NodeExecutionDisplayType {...props} />,
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
