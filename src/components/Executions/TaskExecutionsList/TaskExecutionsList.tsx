import { makeStyles, Theme } from '@material-ui/core/styles';
import { compareTimestampsAscending } from 'common/utils';
import { NonIdealState, WaitForData } from 'components/common';
import { NodeExecution, TaskExecution } from 'models';
import * as React from 'react';
import {
    useTaskExecutions,
    useTaskExecutionsRefresher
} from '../useTaskExecutions';
import { TaskExecutionsListItem } from './TaskExecutionsListItem';
import { getUniqueTaskExecutionName } from './utils';

const useStyles = makeStyles((theme: Theme) => ({
    noExecutionsMessage: {
        paddingTop: theme.spacing(2)
    }
}));

interface TaskExecutionsListProps {
    nodeExecution: NodeExecution;
}

function compareTaskExecutionCreatedAtAscending(
    a: TaskExecution,
    b: TaskExecution
) {
    return compareTimestampsAscending(a.closure.createdAt, b.closure.createdAt);
}

const TaskExecutionsListContent: React.FC<{
    taskExecutions: TaskExecution[];
}> = ({ taskExecutions }) => {
    const styles = useStyles();
    if (!taskExecutions.length) {
        return (
            <NonIdealState
                className={styles.noExecutionsMessage}
                size="small"
                title="No executions found"
            />
        );
    }
    return (
        <>
            {taskExecutions
                .sort(compareTaskExecutionCreatedAtAscending)
                .map(taskExecution => (
                    <TaskExecutionsListItem
                        key={getUniqueTaskExecutionName(taskExecution)}
                        taskExecution={taskExecution}
                    />
                ))}
        </>
    );
};

/** Renders a vertical list of task execution records with horizontal separators
 */
export const TaskExecutionsList: React.FC<TaskExecutionsListProps> = ({
    nodeExecution
}) => {
    const taskExecutions = useTaskExecutions(nodeExecution.id);
    useTaskExecutionsRefresher(nodeExecution, taskExecutions);

    return (
        <WaitForData {...taskExecutions}>
            <TaskExecutionsListContent taskExecutions={taskExecutions.value} />
        </WaitForData>
    );
};
