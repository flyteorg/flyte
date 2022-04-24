import * as React from 'react';
import { makeStyles, Theme } from '@material-ui/core/styles';
import Typography from '@material-ui/core/Typography';
import classnames from 'classnames';
import { PanelSection } from 'components/common/PanelSection';
import { useCommonStyles } from 'components/common/styles';
import { TaskExecutionPhase } from 'models/Execution/enums';
import { TaskExecution } from 'models/Execution/types';
import { ExecutionStatusBadge } from '../ExecutionStatusBadge';
import { TaskExecutionDetails } from './TaskExecutionDetails';
import { TaskExecutionError } from './TaskExecutionError';
import { TaskExecutionLogs } from './TaskExecutionLogs';
import { formatRetryAttempt } from './utils';

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

interface TaskExecutionsListItemProps {
  taskExecution: TaskExecution;
}

/** Renders an individual `TaskExecution` record as part of a list */
export const TaskExecutionsListItem: React.FC<TaskExecutionsListItemProps> = ({
  taskExecution,
}) => {
  const commonStyles = useCommonStyles();
  const styles = useStyles();
  const { closure } = taskExecution;
  const { error } = closure;
  const headerText = formatRetryAttempt(taskExecution.id.retryAttempt);
  const taskHasStarted = closure.phase >= TaskExecutionPhase.QUEUED;

  return (
    <PanelSection>
      <section className={styles.section}>
        <header className={styles.header}>
          <Typography variant="h6" className={classnames(styles.title, commonStyles.textWrapped)}>
            {headerText}
          </Typography>
        </header>
        <ExecutionStatusBadge phase={closure.phase} type="task" variant="text" />
      </section>
      {!!error && (
        <section className={styles.section}>
          <TaskExecutionError error={error} />
        </section>
      )}
      {taskHasStarted && (
        <>
          <section className={styles.section}>
            <TaskExecutionLogs taskLogs={taskExecution.closure.logs || []} />
          </section>
          <section className={styles.section}>
            <TaskExecutionDetails taskExecution={taskExecution} />
          </section>
        </>
      )}
    </PanelSection>
  );
};
