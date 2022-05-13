import * as React from 'react';
import { makeStyles, Theme } from '@material-ui/core/styles';
import Typography from '@material-ui/core/Typography';
import classnames from 'classnames';
import { PanelSection } from 'components/common/PanelSection';
import { useCommonStyles } from 'components/common/styles';
import { TaskExecutionPhase } from 'models/Execution/enums';
import { TaskExecution } from 'models/Execution/types';
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
  selectedPhase?: TaskExecutionPhase;
}

/** Renders an individual `TaskExecution` record as part of a list */
export const MapTaskExecutionsListItem: React.FC<MapTaskExecutionsListItemProps> = ({
  taskExecution,
  showAttempts,
  selectedPhase,
}) => {
  const commonStyles = useCommonStyles();
  const styles = useStyles();

  const { closure } = taskExecution;
  const taskHasStarted = closure.phase >= TaskExecutionPhase.QUEUED;
  const headerText = formatRetryAttempt(taskExecution.id.retryAttempt);
  const logsByPhase = getGroupedLogs(closure.metadata?.externalResources ?? []);

  return (
    <PanelSection>
      {/* Attempts header is ahown only if there is more than one attempt */}
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
      {closure.error ? (
        <section className={styles.section}>
          <TaskExecutionError error={closure.error} />
        </section>
      ) : null}

      {/* If main map task has log attached - show it here */}
      {closure.logs && closure.logs.length > 0 ? (
        <section className={styles.section}>
          <TaskExecutionLogs taskLogs={taskExecution.closure.logs || []} title="Task Log" />
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
            phase={phase}
            taskLogs={logs}
            isExpanded={selectedPhase === phase}
            key={key}
          />
        );
      })}

      {/* If map task is actively started - show 'started' and 'run time' details */}
      {taskHasStarted && (
        <section className={styles.section}>
          <TaskExecutionDetails taskExecution={taskExecution} />
        </section>
      )}
    </PanelSection>
  );
};
