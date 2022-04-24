import { getTaskDisplayType } from 'components/Executions/utils';
import { fetchTaskExecutionList } from 'components/Executions/taskExecutionQueries';
import { NodeExecutionDetails } from 'components/Executions/types';
import { fetchTaskTemplate } from 'components/Task/taskQueries';
import { NodeExecution } from 'models/Execution/types';
import { TaskTemplate } from 'models/Task/types';
import { QueryClient } from 'react-query/types/core/queryClient';

export const getTaskThroughExecution = async (
  queryClient: QueryClient,
  nodeExecution: NodeExecution,
): Promise<NodeExecutionDetails> => {
  const taskExecutions = await fetchTaskExecutionList(queryClient, nodeExecution.id);

  let taskTemplate: TaskTemplate | undefined = undefined;
  if (taskExecutions && taskExecutions.length > 0) {
    taskTemplate = await fetchTaskTemplate(queryClient, taskExecutions[0].id.taskId);
    if (!taskTemplate) {
      // eslint-disable-next-line no-console
      console.error(
        `ERROR: Unexpected missing task template while fetching NodeExecution details: ${JSON.stringify(
          taskExecutions[0].id.taskId,
        )}`,
      );
    }
  }

  const taskDetails: NodeExecutionDetails = {
    displayId: nodeExecution.id.nodeId,
    displayName: taskExecutions?.[0]?.id.taskId.name,
    displayType: getTaskDisplayType(taskTemplate?.type),
    taskTemplate: taskTemplate,
  };

  return taskDetails;
};
