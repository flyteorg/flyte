import { ensureSlashPrefixed } from 'common/utils';
import {
    TaskExecutionIdentifier,
    WorkflowExecutionIdentifier
} from 'models/Execution/types';
import {
    projectBasePath,
    projectDomainBasePath,
    taskExecutionPath
} from './constants';
import { makeRoute } from './utils';

/** Creates a path relative to a particular project */
export const makeProjectBoundPath = (projectId: string, path = '') =>
    makeRoute(
        `/projects/${projectId}${
            path.length ? ensureSlashPrefixed(path) : path
        }`
    );

/** Creates a path relative to a particular project and domain. Paths should begin with a slash (/) */
export const makeProjectDomainBoundPath = (
    projectId: string,
    domainId: string,
    path = ''
) => makeRoute(`/projects/${projectId}/domains/${domainId}${path}`);

export class Routes {
    static NotFound = {};
    // Projects
    static ProjectDetails = {
        makeUrl: (project: string, section?: string) =>
            makeProjectBoundPath(project, section ? `/${section}` : ''),
        path: projectBasePath,
        sections: {
            executions: {
                makeUrl: (project: string, domain?: string) =>
                    makeProjectBoundPath(
                        project,
                        `/executions${domain ? `?domain=${domain}` : ''}`
                    ),
                path: `${projectBasePath}/executions`
            },
            tasks: {
                makeUrl: (project: string, domain?: string) =>
                    makeProjectBoundPath(
                        project,
                        `/tasks${domain ? `?domain=${domain}` : ''}`
                    ),
                path: `${projectBasePath}/tasks`
            },
            workflows: {
                makeUrl: (project: string, domain?: string) =>
                    makeProjectBoundPath(
                        project,
                        `/workflows${domain ? `?domain=${domain}` : ''}`
                    ),
                path: `${projectBasePath}/workflows`
            }
        }
    };
    static ProjectExecutions = {
        makeUrl: (project: string, domain: string) =>
            makeProjectDomainBoundPath(project, domain, '/executions'),
        path: `${projectDomainBasePath}/executions`
    };
    static ProjectTasks = {
        makeUrl: (project: string, domain: string) =>
            makeProjectDomainBoundPath(project, domain, '/tasks'),
        path: `${projectDomainBasePath}/tasks`
    };
    static ProjectWorkflows = {
        makeUrl: (project: string, domain: string) =>
            makeProjectDomainBoundPath(project, domain, '/workflows'),
        path: `${projectDomainBasePath}/workflows`
    };

    // Workflows
    static WorkflowDetails = {
        makeUrl: (project: string, domain: string, workflowName: string) =>
            makeProjectDomainBoundPath(
                project,
                domain,
                `/workflows/${workflowName}`
            ),
        path: `${projectDomainBasePath}/workflows/:workflowName`
    };

    // Workflow Version Details
    static WorkflowVersionDetails = {
        makeUrl: (
            project: string,
            domain: string,
            workflowName: string,
            version: string
        ) =>
            makeProjectDomainBoundPath(
                project,
                domain,
                `/workflows/${workflowName}/version/${version}`
            ),
        path: `${projectDomainBasePath}/workflows/:workflowName/version/:workflowVersion`
    };

    // Tasks
    static TaskDetails = {
        makeUrl: (project: string, domain: string, taskName: string) =>
            makeProjectDomainBoundPath(project, domain, `/tasks/${taskName}`),
        path: `${projectDomainBasePath}/tasks/:taskName`
    };

    // Executions
    static ExecutionDetails = {
        makeUrl: ({ domain, name, project }: WorkflowExecutionIdentifier) =>
            makeProjectDomainBoundPath(project, domain, `/executions/${name}`),
        path: `${projectDomainBasePath}/executions/:executionId`
    };
    static TaskExecutionDetails = {
        makeUrl: (taskExecutionId: TaskExecutionIdentifier) => {
            const {
                nodeExecutionId: {
                    executionId: { project, domain, name: executionName },
                    nodeId
                },
                taskId: {
                    project: taskProject,
                    domain: taskDomain,
                    name: taskName,
                    version: taskVersion
                },
                retryAttempt
            } = taskExecutionId;
            return makeProjectDomainBoundPath(
                project,
                domain,
                `/task_executions/${executionName}/${nodeId}/${taskProject}/${taskDomain}/${taskName}/${taskVersion}/${retryAttempt}`
            );
        },
        path: taskExecutionPath
    };

    // Landing page
    static SelectProject = {
        path: makeRoute('/')
    };
}
