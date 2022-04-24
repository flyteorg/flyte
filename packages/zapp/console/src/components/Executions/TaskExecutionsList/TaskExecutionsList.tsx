import * as React from 'react';
import { makeStyles, Theme } from '@material-ui/core/styles';
import { noExecutionsFoundString } from 'common/constants';
import { NonIdealState } from 'components/common/NonIdealState';
import { WaitForData } from 'components/common/WaitForData';
import { NodeExecution, TaskExecution } from 'models/Execution/types';
import { isMapTaskType } from 'models/Task/utils';
import { useTaskExecutions, useTaskExecutionsRefresher } from '../useTaskExecutions';
import { MapTaskExecutionsListItem } from './MapTaskExecutionListItem';
import { TaskExecutionsListItem } from './TaskExecutionsListItem';
import { getUniqueTaskExecutionName } from './utils';

const useStyles = makeStyles((theme: Theme) => ({
  noExecutionsMessage: {
    paddingTop: theme.spacing(2),
  },
}));

interface TaskExecutionsListProps {
  nodeExecution: NodeExecution;
}

export const TaskExecutionsListContent: React.FC<{
  taskExecutions: TaskExecution[];
}> = ({ taskExecutions }) => {
  const styles = useStyles();
  if (!taskExecutions.length) {
    return (
      <NonIdealState
        className={styles.noExecutionsMessage}
        size="small"
        title={noExecutionsFoundString}
      />
    );
  }

  return (
    <>
      {taskExecutions.map((taskExecution) => {
        const taskType = taskExecution.closure.taskType ?? undefined;
        const useNewMapTaskView =
          isMapTaskType(taskType) && taskExecution.closure.metadata?.externalResources;
        return useNewMapTaskView ? (
          <MapTaskExecutionsListItem
            key={getUniqueTaskExecutionName(taskExecution)}
            taskExecution={taskExecution}
            showAttempts={taskExecutions.length > 1}
          />
        ) : (
          <TaskExecutionsListItem
            key={getUniqueTaskExecutionName(taskExecution)}
            taskExecution={taskExecution}
          />
        );
      })}
    </>
  );
};

/** Renders a vertical list of task execution records with horizontal separators
 */
export const TaskExecutionsList: React.FC<TaskExecutionsListProps> = ({ nodeExecution }) => {
  const taskExecutions = useTaskExecutions(nodeExecution.id);
  useTaskExecutionsRefresher(nodeExecution, taskExecutions);

  return (
    <WaitForData {...taskExecutions}>
      <TaskExecutionsListContent taskExecutions={taskExecutions.value} />
    </WaitForData>
  );
};
