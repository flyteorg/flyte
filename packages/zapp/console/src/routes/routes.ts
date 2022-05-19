import { ensureSlashPrefixed } from 'common/utils';
import { WorkflowExecutionIdentifier } from 'models/Execution/types';
import { projectBasePath, projectDomainBasePath } from './constants';
import { makeRoute } from './utils';

/** Creates a path relative to a particular project */
export const makeProjectBoundPath = (projectId: string, path = '') =>
  makeRoute(`/projects/${projectId}${path.length ? ensureSlashPrefixed(path) : path}`);

/** Creates a path relative to a particular project and domain. Paths should begin with a slash (/) */
export const makeProjectDomainBoundPath = (projectId: string, domainId: string, path = '') =>
  makeRoute(`/projects/${projectId}/domains/${domainId}${path}`);

export class Routes {
  static NotFound = {};
  // Projects
  static ProjectDetails = {
    makeUrl: (project: string, section?: string) =>
      makeProjectBoundPath(project, section ? `/${section}` : ''),
    path: projectBasePath,
    sections: {
      dashboard: {
        makeUrl: (project: string, domain?: string) =>
          makeProjectBoundPath(project, `/executions${domain ? `?domain=${domain}` : ''}`),
        path: `${projectBasePath}/executions`,
      },
      tasks: {
        makeUrl: (project: string, domain?: string) =>
          makeProjectBoundPath(project, `/tasks${domain ? `?domain=${domain}` : ''}`),
        path: `${projectBasePath}/tasks`,
      },
      workflows: {
        makeUrl: (project: string, domain?: string) =>
          makeProjectBoundPath(project, `/workflows${domain ? `?domain=${domain}` : ''}`),
        path: `${projectBasePath}/workflows`,
      },
    },
  };

  static ProjectDashboard = {
    makeUrl: (project: string, domain: string) =>
      makeProjectDomainBoundPath(project, domain, '/executions'),
    path: `${projectDomainBasePath}/executions`,
  };

  static ProjectTasks = {
    makeUrl: (project: string, domain: string) =>
      makeProjectDomainBoundPath(project, domain, '/tasks'),
    path: `${projectDomainBasePath}/tasks`,
  };

  static ProjectWorkflows = {
    makeUrl: (project: string, domain: string) =>
      makeProjectDomainBoundPath(project, domain, '/workflows'),
    path: `${projectDomainBasePath}/workflows`,
  };

  // Workflows
  static WorkflowDetails = {
    makeUrl: (project: string, domain: string, workflowName: string) =>
      makeProjectDomainBoundPath(project, domain, `/workflows/${workflowName}`),
    path: `${projectDomainBasePath}/workflows/:workflowName`,
  };

  // Entity Version Details
  static EntityVersionDetails = {
    makeUrl: (
      project: string,
      domain: string,
      entityName: string,
      entityType: string,
      version: string,
    ) =>
      makeProjectDomainBoundPath(
        project,
        domain,
        `/${entityType}/${entityName}/version/${version}`,
      ),
    path: `${projectDomainBasePath}/:entityType/:entityName/version/:entityVersion`,
  };

  // Tasks
  static TaskDetails = {
    makeUrl: (project: string, domain: string, taskName: string) =>
      makeProjectDomainBoundPath(project, domain, `/tasks/${taskName}`),
    path: `${projectDomainBasePath}/tasks/:taskName`,
  };

  // Executions
  static ExecutionDetails = {
    makeUrl: ({ domain, name, project }: WorkflowExecutionIdentifier) =>
      makeProjectDomainBoundPath(project, domain, `/executions/${name}`),
    path: `${projectDomainBasePath}/executions/:executionId`,
  };

  // Landing page
  static SelectProject = {
    path: makeRoute('/'),
  };
}
