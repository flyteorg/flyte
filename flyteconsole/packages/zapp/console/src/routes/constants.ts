import { makeRoute } from './utils';

const projectPrefix = '/projects/:projectId';

export const projectBasePath = makeRoute(projectPrefix);
export const projectDomainBasePath = makeRoute(`${projectPrefix}/domains/:domainId`);

export const taskExecutionPath = `${projectDomainBasePath}/task_executions/:executionName/:nodeId/:taskProject/:taskDomain/:taskName/:taskVersion/:retryAttempt`;
