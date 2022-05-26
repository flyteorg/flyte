import { Tooltip, Typography } from '@material-ui/core';
import { formatDateLocalTimezone, formatDateUTC, millisecondsToHMS } from 'common/formatters';
import { timestampToDate } from 'common/utils';
import { useCommonStyles } from 'components/common/styles';
import { isEqual } from 'lodash';
import { CatalogCacheStatus, NodeExecutionPhase } from 'models/Execution/enums';
import { TaskNodeMetadata } from 'models/Execution/types';
import * as React from 'react';
import { useNodeExecutionContext } from '../contextProvider/NodeExecutionDetails';
import { ExecutionStatusBadge } from '../ExecutionStatusBadge';
import { NodeExecutionCacheStatus } from '../NodeExecutionCacheStatus';
import { getNodeExecutionTimingMS } from '../utils';
import { NodeExecutionActions } from './NodeExecutionActions';
import { SelectNodeExecutionLink } from './SelectNodeExecutionLink';
import { useColumnStyles } from './styles';
import { NodeExecutionCellRendererData, NodeExecutionColumnDefinition } from './types';

const ExecutionName: React.FC<NodeExecutionCellRendererData> = ({ execution, state }) => {
  const detailsContext = useNodeExecutionContext();
  const [displayName, setDisplayName] = React.useState<string | undefined>();

  React.useEffect(() => {
    let isCurrent = true;
    detailsContext.getNodeExecutionDetails(execution).then((res) => {
      if (isCurrent) {
        setDisplayName(res.displayName);
      }
    });
    return () => {
      isCurrent = false;
    };
  });

  const commonStyles = useCommonStyles();
  const styles = useColumnStyles();

  const isSelected =
    state.selectedExecution != null && isEqual(execution.id, state.selectedExecution);

  const name = displayName ?? execution.id.nodeId;
  const truncatedName = name?.split('.').pop() || name;

  const readableName = isSelected ? (
    <Typography variant="body1" className={styles.selectedExecutionName}>
      {truncatedName}
    </Typography>
  ) : (
    <SelectNodeExecutionLink
      className={commonStyles.primaryLink}
      execution={execution}
      linkText={truncatedName || ''}
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

const DisplayId: React.FC<NodeExecutionCellRendererData> = ({ execution }) => {
  const commonStyles = useCommonStyles();
  const detailsContext = useNodeExecutionContext();
  const [displayId, setDisplayId] = React.useState<string | undefined>();

  React.useEffect(() => {
    let isCurrent = true;
    detailsContext.getNodeExecutionDetails(execution).then((res) => {
      if (isCurrent) {
        setDisplayId(res.displayId);
      }
    });
    return () => {
      isCurrent = false;
    };
  });

  const nodeId = displayId ?? execution.id.nodeId;
  return (
    <Tooltip arrow title={nodeId} placement="top-start">
      <div className={commonStyles.truncateText}>{nodeId}</div>
    </Tooltip>
  );
};

const DisplayType: React.FC<NodeExecutionCellRendererData> = ({ execution }) => {
  const detailsContext = useNodeExecutionContext();
  const [type, setType] = React.useState<string | undefined>();

  React.useEffect(() => {
    let isCurrent = true;
    detailsContext.getNodeExecutionDetails(execution).then((res) => {
      if (isCurrent) {
        setType(res.displayType);
      }
    });
    return () => {
      isCurrent = false;
    };
  });

  return <Typography color="textSecondary">{type}</Typography>;
};

const hiddenCacheStatuses = [CatalogCacheStatus.CACHE_MISS, CatalogCacheStatus.CACHE_DISABLED];
function hasCacheStatus(taskNodeMetadata?: TaskNodeMetadata): taskNodeMetadata is TaskNodeMetadata {
  if (!taskNodeMetadata) {
    return false;
  }
  const { cacheStatus } = taskNodeMetadata;
  return !hiddenCacheStatuses.includes(cacheStatus);
}

export function generateColumns(
  styles: ReturnType<typeof useColumnStyles>,
): NodeExecutionColumnDefinition[] {
  return [
    {
      cellRenderer: (props) => <ExecutionName {...props} />,
      className: styles.columnName,
      key: 'name',
      label: 'task name',
    },
    {
      cellRenderer: (props) => <DisplayId {...props} />,
      className: styles.columnNodeId,
      key: 'nodeId',
      label: 'node id',
    },
    {
      cellRenderer: (props) => <DisplayType {...props} />,
      className: styles.columnType,
      key: 'type',
      label: 'type',
    },
    {
      cellRenderer: ({
        execution: {
          closure: { phase = NodeExecutionPhase.UNDEFINED, taskNodeMetadata },
        },
      }) => (
        <>
          <ExecutionStatusBadge phase={phase} type="node" />
          {hasCacheStatus(taskNodeMetadata) ? (
            <NodeExecutionCacheStatus taskNodeMetadata={taskNodeMetadata} variant="iconOnly" />
          ) : null}
        </>
      ),
      className: styles.columnStatus,
      key: 'phase',
      label: 'status',
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
            <Typography variant="body1">{formatDateUTC(startedAtDate)}</Typography>
            <Typography variant="subtitle1" color="textSecondary">
              {formatDateLocalTimezone(startedAtDate)}
            </Typography>
          </>
        );
      },
      className: styles.columnStartedAt,
      key: 'startedAt',
      label: 'start time',
    },
    {
      cellRenderer: ({ execution }) => {
        const timing = getNodeExecutionTimingMS(execution);
        if (timing === null) {
          return '';
        }
        return (
          <>
            <Typography variant="body1">{millisecondsToHMS(timing.duration)}</Typography>
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
          <Typography component="div" variant="subtitle1" color="textSecondary">
            Queued Time
          </Typography>
        </>
      ),
    },
    {
      cellRenderer: ({ execution, state }) => (
        <NodeExecutionActions execution={execution} state={state} />
      ),
      className: styles.columnLogs,
      key: 'actions',
      label: '',
    },
  ];
}
