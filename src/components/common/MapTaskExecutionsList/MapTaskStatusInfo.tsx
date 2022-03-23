import * as React from 'react';
import { useState } from 'react';
import { makeStyles, Theme } from '@material-ui/core/styles';
import { RowExpander } from 'components/Executions/Tables/RowExpander';
import { NodeExecutionPhase } from 'models/Execution/enums';
import { getNodeExecutionPhaseConstants } from 'components/Executions/utils';
import { Typography } from '@material-ui/core';
import { TaskExecutionLogs } from 'components/Executions/TaskExecutionsList/TaskExecutionLogs';
import { TaskLog } from 'models/Common/types';

const useStyles = makeStyles((_theme: Theme) => ({
  mainWrapper: {
    display: 'flex',
    flex: 1,
    flexDirection: 'column',
  },
  headerWrapper: {
    display: 'flex',
    flex: 'auto',
    alignItems: 'center',
  },
  colorBar: {
    height: '20px',
    borderLeft: 'solid 4px red',
    margin: '0 8px 0 4px',
  },
  title: {
    flex: 'auto',
  },
  logs: {
    marginLeft: '46px',
  },
  semiboldText: {
    fontWeight: 600,
  },
}));

interface MapTaskStatusInfoProps {
  taskLogs: TaskLog[];
  status: NodeExecutionPhase;
  expanded: boolean;
}

export const MapTaskStatusInfo = (props: MapTaskStatusInfoProps) => {
  const [expanded, setExpanded] = useState(props.expanded);
  const styles = useStyles();

  const toggleExpanded = () => {
    setExpanded(!expanded);
  };

  const phaseData = getNodeExecutionPhaseConstants(props.status);
  return (
    <div className={styles.mainWrapper}>
      <div className={styles.headerWrapper}>
        <RowExpander expanded={expanded} onClick={toggleExpanded} />
        <div className={styles.colorBar} style={{ borderColor: phaseData.badgeColor }} />
        <Typography variant="body2" className={styles.title}>
          {phaseData.text}
        </Typography>
        <span>{`x${props.taskLogs.length}`}</span>
      </div>
      {expanded && (
        <div className={styles.logs}>
          <TaskExecutionLogs taskLogs={props.taskLogs} />
        </div>
      )}
    </div>
  );
};
