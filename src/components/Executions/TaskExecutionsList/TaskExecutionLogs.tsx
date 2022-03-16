import { Typography } from '@material-ui/core';
import { makeStyles, Theme } from '@material-ui/core/styles';
import { NewTargetLink } from 'components/common/NewTargetLink';
import { useCommonStyles } from 'components/common/styles';
import { TaskLog } from 'models/Common/types';
import { TaskExecution } from 'models/Execution/types';
import * as React from 'react';
import { noLogsFoundString } from '../constants';

const useStyles = makeStyles((theme: Theme) => ({
  logLink: {
    margin: `${theme.spacing(0.5)} 0`,
  },
  sectionHeader: {
    marginTop: theme.spacing(1),
  },
}));

const TaskLogList: React.FC<{ logs: TaskLog[] }> = ({ logs }) => {
  const styles = useStyles();
  const commonStyles = useCommonStyles();
  if (!(logs && logs.length > 0)) {
    return <span className={commonStyles.hintText}>{noLogsFoundString}</span>;
  }
  return (
    <>
      {logs.map(({ name, uri }) => (
        <NewTargetLink className={styles.logLink} key={name} external={true} href={uri}>
          {name}
        </NewTargetLink>
      ))}
    </>
  );
};

/** Renders log links from a `TaskExecution`, if they exist. Otherwise renders
 * a message indicating that no logs are available.
 */
export const TaskExecutionLogs: React.FC<{ taskExecution: TaskExecution }> = ({
  taskExecution,
}) => {
  const styles = useStyles();
  return (
    <section>
      <header className={styles.sectionHeader}>
        <Typography variant="h6">Logs</Typography>
      </header>
      <TaskLogList logs={taskExecution.closure.logs || []} />
    </section>
  );
};
