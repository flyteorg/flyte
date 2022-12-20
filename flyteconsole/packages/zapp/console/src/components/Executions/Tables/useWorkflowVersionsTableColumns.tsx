import { Typography } from '@material-ui/core';
import { formatDateUTC } from 'common/formatters';
import { padExecutionPaths, padExecutions, timestampToDate } from 'common/utils';
import { WaitForData } from 'components/common/WaitForData';
import ProjectStatusBar from 'components/Project/ProjectStatusBar';
import * as moment from 'moment';
import * as React from 'react';
import { useWorkflowVersionsColumnStyles } from './styles';
import { WorkflowVersionColumnDefinition } from './types';

/**
 * Returns a memoized list of column definitions to use when rendering a
 * `WorkflowVersionRow`. Memoization is based on common/column style objects
 * and any fields in the incoming `WorkflowExecutionColumnOptions` object.
 */
export function useWorkflowVersionsTableColumns(): WorkflowVersionColumnDefinition[] {
  const styles = useWorkflowVersionsColumnStyles();
  return React.useMemo(
    () => [
      {
        cellRenderer: ({
          workflow: {
            id: { version },
          },
        }) => <Typography variant="body1">{version}</Typography>,
        className: styles.columnName,
        key: 'name',
        label: 'version id',
      },
      {
        cellRenderer: ({ workflow: { closure } }) => {
          if (!closure?.createdAt) {
            return '';
          }
          const createdAtDate = timestampToDate(closure.createdAt);
          return <Typography variant="body1">{formatDateUTC(createdAtDate)}</Typography>;
        },
        className: styles.columnCreatedAt,
        key: 'createdAt',
        label: 'time created',
      },
      {
        cellRenderer: ({ executions }) => {
          return (
            <WaitForData {...executions}>
              <Typography variant="body1">
                {executions.value.length
                  ? moment(timestampToDate(executions.value[0].closure.createdAt)).fromNow()
                  : ''}
              </Typography>
            </WaitForData>
          );
        },
        className: styles.columnLastRun,
        key: 'lastExecution',
        label: 'last execution',
      },
      {
        cellRenderer: ({ executions }) => {
          return (
            <WaitForData {...executions}>
              <ProjectStatusBar
                items={padExecutions(
                  executions.value.map((execution) => execution.closure.phase) || [],
                )}
                paths={padExecutionPaths(executions.value.map((execution) => execution.id) || [])}
              />
            </WaitForData>
          );
        },
        className: styles.columnRecentRun,
        key: 'recentRun',
        label: 'recent run',
      },
    ],
    [styles],
  );
}
