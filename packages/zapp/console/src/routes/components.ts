import { ExecutionDetails } from 'components/Executions/ExecutionDetails/ExecutionDetails';
import { NotFound } from 'components/NotFound/NotFound';
import { ProjectDetails } from 'components/Project/ProjectDetails';
import { SelectProject } from 'components/SelectProject/SelectProject';
import { TaskDetails } from 'components/Task/TaskDetails';
import { WorkflowDetails } from 'components/Workflow/WorkflowDetails';
import { EntityVersionsDetailsContainer } from 'components/Entities/VersionDetails/EntityVersionDetailsContainer';

/** Indexes the components for each defined route. These are done separately to avoid circular references
 * in components which include the Routes dictionary
 */
export const components = {
  executionDetails: ExecutionDetails,
  notFound: NotFound,
  projectDetails: ProjectDetails,
  selectProject: SelectProject,
  taskDetails: TaskDetails,
  workflowDetails: WorkflowDetails,
  entityVersionDetails: EntityVersionsDetailsContainer,
};
