import { Tooltip, Typography } from '@material-ui/core';
import { formatDateLocalTimezone, formatDateUTC, millisecondsToHMS } from 'common/formatters';
import { timestampToDate } from 'common/utils';
import { useCommonStyles } from 'components/common/styles';
import * as React from 'react';
import { useEffect, useState } from 'react';
import { CompiledNode } from 'models/Node/types';
import { NodeExecutionPhase } from 'models/Execution/enums';
import { getNodeTemplateName } from 'components/WorkflowGraph/utils';
import { useNodeExecutionContext } from '../contextProvider/NodeExecutionDetails';
import { ExecutionStatusBadge } from '../ExecutionStatusBadge';
import { NodeExecutionCacheStatus } from '../NodeExecutionCacheStatus';
import { getNodeExecutionTimingMS, getNodeFrontendPhase, isNodeGateNode } from '../utils';
import { NodeExecutionActions } from './NodeExecutionActions';
import { useColumnStyles } from './styles';
import { NodeExecutionCellRendererData, NodeExecutionColumnDefinition } from './types';
import t from '../strings';
import { NodeExecutionName } from '../ExecutionDetails/Timeline/NodeExecutionName';

const DisplayId: React.FC<NodeExecutionCellRendererData> = ({ execution }) => {
  const commonStyles = useCommonStyles();
  const { getNodeExecutionDetails } = useNodeExecutionContext();
  const [displayId, setDisplayId] = useState<string | undefined>();

  useEffect(() => {
    let isCurrent = true;
    getNodeExecutionDetails(execution).then((res) => {
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
  const { getNodeExecutionDetails } = useNodeExecutionContext();
  const [type, setType] = useState<string | undefined>();

  useEffect(() => {
    let isCurrent = true;
    getNodeExecutionDetails(execution).then((res) => {
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

export function generateColumns(
  styles: ReturnType<typeof useColumnStyles>,
  nodes: CompiledNode[],
): NodeExecutionColumnDefinition[] {
  return [
    {
      cellRenderer: ({ node }) => (
        <NodeExecutionName
          name={node.name}
          templateName={getNodeTemplateName(node)}
          execution={node.execution}
        />
      ),
      className: styles.columnName,
      key: 'name',
      label: t('nameLabel'),
    },
    {
      cellRenderer: (props) => <DisplayId {...props} />,
      className: styles.columnNodeId,
      key: 'nodeId',
      label: t('nodeIdLabel'),
    },
    {
      cellRenderer: (props) => <DisplayType {...props} />,
      className: styles.columnType,
      key: 'type',
      label: t('typeLabel'),
    },
    {
      cellRenderer: ({ execution }) => {
        const isGateNode = isNodeGateNode(nodes, execution.id);
        const phase = getNodeFrontendPhase(
          execution.closure?.phase ?? NodeExecutionPhase.UNDEFINED,
          isGateNode,
        );

        return (
          <>
            <ExecutionStatusBadge phase={phase} type="node" />
            <NodeExecutionCacheStatus execution={execution} variant="iconOnly" />
          </>
        );
      },
      className: styles.columnStatus,
      key: 'phase',
      label: t('phaseLabel'),
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
      label: t('startedAtLabel'),
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
            {t('durationLabel')}
          </Typography>
          <Typography component="div" variant="subtitle1" color="textSecondary">
            {t('queuedTimeLabel')}
          </Typography>
        </>
      ),
    },
    {
      cellRenderer: ({ execution }) =>
        execution.closure.phase === NodeExecutionPhase.UNDEFINED ? null : (
          <NodeExecutionActions execution={execution} />
        ),
      className: styles.columnLogs,
      key: 'actions',
      label: '',
    },
  ];
}
