import * as React from 'react';
import { useEffect, useState } from 'react';
import { Typography } from '@material-ui/core';
import { makeStyles, Theme } from '@material-ui/core/styles';
import { RowExpander } from 'components/Executions/Tables/RowExpander';
import { TaskExecutionPhase } from 'models/Execution/enums';
import { getTaskExecutionPhaseConstants } from 'components/Executions/utils';
import { Core } from 'flyteidl';
import { MapTaskExecution, TaskExecution } from 'models/Execution/types';
import { TaskNameList } from './TaskNameList';

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
  taskExecution: TaskExecution;
  taskLogs: Core.ITaskLog[];
  phase: TaskExecutionPhase;
  selectedPhase?: TaskExecutionPhase;
  onTaskSelected: (val: MapTaskExecution) => void;
}

export const MapTaskStatusInfo = ({
  taskExecution,
  taskLogs,
  phase,
  selectedPhase,
  onTaskSelected,
}: MapTaskStatusInfoProps) => {
  const [expanded, setExpanded] = useState<boolean>(selectedPhase === phase);
  const styles = useStyles();

  const phaseData = getTaskExecutionPhaseConstants(phase);

  useEffect(() => {
    setExpanded(selectedPhase === phase);
  }, [selectedPhase, phase]);

  const toggleExpanded = () => {
    setExpanded(!expanded);
  };

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
          <TaskNameList
            taskExecution={taskExecution}
            logs={taskLogs}
            onTaskSelected={onTaskSelected}
          />
        </div>
      )}
    </div>
  );
};
