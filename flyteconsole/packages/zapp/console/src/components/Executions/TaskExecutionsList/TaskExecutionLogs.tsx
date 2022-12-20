import * as React from 'react';
import { Typography } from '@material-ui/core';
import { makeStyles, Theme } from '@material-ui/core/styles';
import { Core } from 'flyteidl';
import { NewTargetLink } from 'components/common/NewTargetLink';
import { useCommonStyles } from 'components/common/styles';
import { noLogsFoundString } from '../constants';

const useStyles = makeStyles((theme: Theme) => ({
  logLink: {
    margin: `${theme.spacing(0.5)} 0`,
  },
  sectionHeader: {
    marginTop: theme.spacing(1),
  },
  logName: {
    fontWeight: 'lighter',
  },
}));

export const TaskLogList: React.FC<{ logs: Core.ITaskLog[] }> = ({ logs }) => {
  const styles = useStyles();
  const commonStyles = useCommonStyles();
  if (!(logs && logs.length > 0)) {
    return <span className={commonStyles.hintText}>{noLogsFoundString}</span>;
  }
  return (
    <>
      {logs.map(({ name, uri }) =>
        uri ? (
          <NewTargetLink className={styles.logLink} key={name} external={true} href={uri}>
            {name}
          </NewTargetLink>
        ) : (
          // If there is no url, show item a a name string only, as it's not really clickable
          <div className={styles.logName}>{name}</div>
        ),
      )}
    </>
  );
};

/** Renders log links from a `taskLogs`(aka taskExecution.closure.logs), if they exist.
 *  Otherwise renders a message indicating that no logs are available.
 */
export const TaskExecutionLogs: React.FC<{ taskLogs: Core.ITaskLog[]; title?: string }> = ({
  taskLogs,
  title,
}) => {
  const styles = useStyles();
  return (
    <section>
      <header className={styles.sectionHeader}>
        <Typography variant="h6">{title ?? 'Logs'}</Typography>
      </header>
      <TaskLogList logs={taskLogs} />
    </section>
  );
};
