import * as React from 'react';
import { makeStyles, Theme } from '@material-ui/core/styles';
import Typography from '@material-ui/core/Typography';
import classnames from 'classnames';
import { PanelSection } from 'components/common/PanelSection';
import { useCommonStyles } from 'components/common/styles';
import { TaskExecutionPhase } from 'models/Execution/enums';
import { MapTaskExecution, TaskExecution } from 'models/Execution/types';
import { MapTaskStatusInfo } from 'components/common/MapTaskExecutionsList/MapTaskStatusInfo';
import { TaskExecutionDetails } from './TaskExecutionDetails';
import { TaskExecutionError } from './TaskExecutionError';
import { TaskExecutionLogs } from './TaskExecutionLogs';
import { formatRetryAttempt, getGroupedLogs } from './utils';
import { RENDER_ORDER } from './constants';

const useStyles = makeStyles((theme: Theme) => ({
  detailsLink: {
    fontWeight: 'normal',
  },
  header: {
    marginBottom: theme.spacing(1),
  },
  title: {
    marginBottom: theme.spacing(1),
  },
  showDetailsButton: {
    marginTop: theme.spacing(1),
  },
  section: {
    marginBottom: theme.spacing(2),
  },
}));

interface MapTaskExecutionsListItemProps {
  taskExecution: TaskExecution;
  showAttempts: boolean;
  onTaskSelected: (val: MapTaskExecution) => void;
  selectedPhase?: TaskExecutionPhase;
}

/** Renders an individual `TaskExecution` record as part of a list */
export const MapTaskExecutionsListItem: React.FC<MapTaskExecutionsListItemProps> = ({
  taskExecution,
  showAttempts,
  onTaskSelected,
  selectedPhase,
}) => {
  const commonStyles = useCommonStyles();
  const styles = useStyles();

  const {
    closure: { error, startedAt, updatedAt, duration, phase, logs, metadata },
    id: { retryAttempt },
  } = taskExecution;
  const taskHasStarted = phase >= TaskExecutionPhase.QUEUED;
  const headerText = formatRetryAttempt(retryAttempt);
  const logsByPhase = getGroupedLogs(metadata?.externalResources ?? []);

  return (
    <PanelSection>
      {/* Attempts header is shown only if there is more than one attempt */}
      {showAttempts ? (
        <section className={styles.section}>
          <header className={styles.header}>
            <Typography variant="h6" className={classnames(styles.title, commonStyles.textWrapped)}>
              {headerText}
            </Typography>
          </header>
        </section>
      ) : null}
      {/* Error info is shown only if there is an error present for this map task */}
      {error ? (
        <section className={styles.section}>
          <TaskExecutionError error={error} />
        </section>
      ) : null}

      {/* If main map task has log attached - show it here */}
      {logs && logs.length > 0 ? (
        <section className={styles.section}>
          <TaskExecutionLogs taskLogs={logs || []} title="Task Log" />
        </section>
      ) : null}
      {/* child/array logs separated by subtasks phase */}
      {RENDER_ORDER.map((phase, id) => {
        const logs = logsByPhase.get(phase);
        if (!logs) {
          return null;
        }
        const key = `${id}-${phase}`;
        return (
          <MapTaskStatusInfo
            taskExecution={taskExecution}
            phase={phase}
            taskLogs={logs}
            selectedPhase={selectedPhase}
            key={key}
            onTaskSelected={onTaskSelected}
          />
        );
      })}

      {/* If map task is actively started - show 'started' and 'run time' details */}
      {taskHasStarted && (
        <section className={styles.section}>
          <TaskExecutionDetails startedAt={startedAt} updatedAt={updatedAt} duration={duration} />
        </section>
      )}
    </PanelSection>
  );
};
