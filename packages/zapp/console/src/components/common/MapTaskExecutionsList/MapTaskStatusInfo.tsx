import * as React from 'react';
import { useState } from 'react';
import { Typography } from '@material-ui/core';
import { makeStyles, Theme } from '@material-ui/core/styles';
import { RowExpander } from 'components/Executions/Tables/RowExpander';
import { TaskExecutionPhase } from 'models/Execution/enums';
import { getTaskExecutionPhaseConstants } from 'components/Executions/utils';
import { TaskLogList } from 'components/Executions/TaskExecutionsList/TaskExecutionLogs';
import { Core } from 'flyteidl';

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
  taskLogs: Core.ITaskLog[];
  phase: TaskExecutionPhase;
  isExpanded: boolean;
}

export const MapTaskStatusInfo = ({ taskLogs, phase, isExpanded }: MapTaskStatusInfoProps) => {
  const [expanded, setExpanded] = useState(isExpanded);
  const styles = useStyles();

  const toggleExpanded = () => {
    setExpanded(!expanded);
  };

  const phaseData = getTaskExecutionPhaseConstants(phase);
  return (
    <div className={styles.mainWrapper}>
      <div className={styles.headerWrapper}>
        <RowExpander expanded={expanded} onClick={toggleExpanded} />
        <div className={styles.colorBar} style={{ borderColor: phaseData.badgeColor }} />
        <Typography variant="body2" className={styles.title}>
          {phaseData.text}
        </Typography>
        <span>{`×${taskLogs.length}`}</span>
      </div>
      {expanded && (
        <div className={styles.logs}>
          <TaskLogList logs={taskLogs} />
        </div>
      )}
    </div>
  );
};
