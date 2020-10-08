import { ExecutionDetails } from 'components/Executions/ExecutionDetails';
import { TaskExecutionDetails } from 'components/Executions/TaskExecutionDetails';
import { NotFound } from 'components/NotFound';
import { ProjectDetails } from 'components/Project';
import { SelectProject } from 'components/SelectProject';
import { TaskDetails } from 'components/Task/TaskDetails';
import { WorkflowVersionDetails } from 'components/Workflow';
import { WorkflowDetails } from 'components/Workflow/WorkflowDetails';

/** Indexes the components for each defined route. These are done separately to avoid circular references
 * in components which include the Routes dictionary
 */
export const components = {
    executionDetails: ExecutionDetails,
    notFound: NotFound,
    projectDetails: ProjectDetails,
    selectProject: SelectProject,
    taskExecutionDetails: TaskExecutionDetails,
    taskDetails: TaskDetails,
    workflowDetails: WorkflowDetails,
    workflowVersionDetails: WorkflowVersionDetails
};
