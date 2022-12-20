import { Admin } from 'flyteidl';
import { limits } from 'models/AdminEntity/constants';
import { EncodableType } from 'models/AdminEntity/types';
import { adminApiUrl, encodeProtoPayload } from 'models/AdminEntity/utils';
import { DomainIdentifierScope, Identifier, ResourceType } from 'models/Common/types';
import {
  Execution,
  NodeExecution,
  NodeExecutionIdentifier,
  TaskExecution,
  TaskExecutionIdentifier,
  WorkflowExecutionIdentifier,
} from 'models/Execution/types';
import { LaunchPlan } from 'models/Launch/types';
import { Project } from 'models/Project/types';
import { Task } from 'models/Task/types';
import { Workflow } from 'models/Workflow/types';
import { MockedRequest, RequestParams, ResponseResolver, rest } from 'msw';
import { RestContext } from 'msw/lib/types/rest';
import { RequestHandlersList } from 'msw/lib/types/setupWorker/glossary';
import { DefaultRequestBodyType } from 'msw/lib/types/utils/handlers/requestHandler';
import { obj } from 'test/utils';
import { notFoundError, RequestError, unexpectedError } from './errors';
import { stableStringify } from './utils';

const taskExecutionPath =
  '/task_executions/:project/:domain/:name/:nodeId/:taskProject/:taskDomain/:taskName/:taskVersion/:retryAttempt';

function isValidIdentifier(id: Partial<Identifier>): id is Identifier {
  return !!id.project && !!id.domain && !!id.name && !!id.version;
}

function nodeExecutionListQueryParams(params: QueryParamsMap): QueryParamsMap {
  return { limit: `${limits.NONE}`, ...params };
}

function workflowExecutionListQueryParams(params: QueryParamsMap): QueryParamsMap {
  const finalParams: QueryParamsMap = { limit: `${limits.DEFAULT}` };
  if (params.token && params.token.length) {
    finalParams.token = params.token;
  }
  if (params.filters && params.filters.length) {
    finalParams.filters = params.filters;
  }
  return finalParams;
}

function makeFullIdentifier(
  id: Partial<Identifier>,
  resourceType: ResourceType,
): Required<Identifier> {
  if (!isValidIdentifier(id)) {
    throw new Error(`Received incomplete Identifier: ${id}`);
  }
  return { ...id, resourceType };
}

function launchPlanIdentifier(id: Partial<Identifier>) {
  return makeFullIdentifier(id, ResourceType.LAUNCH_PLAN);
}

function workflowIdentifier(id: Partial<Identifier>) {
  return makeFullIdentifier(id, ResourceType.WORKFLOW);
}

function taskIdentifier(id: Partial<Identifier>) {
  return makeFullIdentifier(id, ResourceType.TASK);
}

function normalizeTaskExecutionIdentifier(id: TaskExecutionIdentifier): TaskExecutionIdentifier {
  return { ...id, taskId: taskIdentifier(id.taskId) };
}

function protobufResponse<T>(ctx: RestContext, data: unknown, encodeType: EncodableType<T>) {
  const buffer = encodeProtoPayload(data, encodeType);
  const contentLength = buffer.byteLength.toString();
  return [
    ctx.set('Content-Type', 'application/octet-stream'),
    ctx.set('Content-Length', contentLength),
    ctx.body(buffer),
  ];
}

function getItemKey(id: unknown) {
  return stableStringify(id);
}

function getItem<ItemType>(store: Map<string, unknown>, id: unknown): ItemType {
  const item = store.get(getItemKey(id));
  if (!item) {
    throw notFoundError(id);
  }
  if (item instanceof RequestError) {
    throw item;
  }
  return item as ItemType;
}

function insertItem(store: Map<string, unknown>, id: unknown, value: unknown) {
  return store.set(getItemKey(id), value);
}

type RestRequest = MockedRequest<DefaultRequestBodyType, RequestParams>;
type RestResolver = ResponseResolver<
  MockedRequest<DefaultRequestBodyType, RequestParams>,
  RestContext,
  any
>;
type QueryParamsMap = Record<string, string>;

function getQueryParams(req: RestRequest): QueryParamsMap {
  return Array.from(req.url.searchParams.entries()).reduce(
    (out, [key, value]) => ({ ...out, [key]: value }),
    {},
  );
}

function catchResponseErrors(resolver: RestResolver): RestResolver {
  return (req, res, ctx) => {
    try {
      return resolver(req, res, ctx);
    } catch (e) {
      if (e instanceof RequestError) {
        return res(ctx.status(e.type), ctx.text(e.message));
      } else return res(ctx.status(500), ctx.text(`Unexpected error: ${e}`));
    }
  };
}

interface AdminEntityHandlerConfig<DataType> {
  path: string;
  getDataForRequest(req: RestRequest): DataType;
  responseEncoder: EncodableType<DataType>;
}
function adminEntityHandler<DataType>({
  path,
  getDataForRequest,
  responseEncoder,
}: AdminEntityHandlerConfig<DataType>) {
  return rest.get(
    adminApiUrl(path),
    catchResponseErrors((req, res, ctx) =>
      res(...protobufResponse(ctx, getDataForRequest(req), responseEncoder)),
    ),
  );
}

type RequireIdField<T extends { id?: unknown }> = Omit<T, 'id'> & Pick<Required<T>, 'id'>;

/** A mock implementation of the Admin API server. Contains functions for inserting
 * mock data of various types. Paths for inserted entities are automatically determined
 * based on the parameters provided (usually the `id` field).
 */
export interface AdminServer {
  /** Insert a single `LaunchPlan`. */
  insertLaunchPlan(data: RequireIdField<Partial<LaunchPlan>>): void;
  /** Insert a single `NodeExecution`. */
  insertNodeExecution(data: RequireIdField<Partial<NodeExecution>>): void;
  /** Insert a list of `NodeExecution`s to be returned for an `Execution`.
   * Note: Does not add handlers for single `NodeExecution`s. Those must be inserted
   * separately.
   */
  insertNodeExecutionList(
    id: WorkflowExecutionIdentifier,
    data: RequireIdField<Partial<NodeExecution>>[] | RequestError,
    query?: Record<string, string>,
  ): void;
  /** Insert the global list of `Project`s. Overwrites the existing list. */
  insertProjects(data: RequireIdField<Partial<Project>>[]): void;
  /** Insert a single `Task` record. */
  insertTask(data: RequireIdField<Partial<Task>>): void;
  /** Insert a single `TaskExecution` record. */
  insertTaskExecution(data: RequireIdField<Partial<TaskExecution>>): void;
  /** Insert a list of `TaskExecution` records to be returned for a parent
   * `NodeExecution`.
   * Note: Does not insert single `TaskExecution` records. Those must be inserted
   * separately.
   */
  insertTaskExecutionList(
    id: NodeExecutionIdentifier,
    data: RequireIdField<Partial<TaskExecution>>[],
  ): void;
  /** Inserts a list of `NodeExecution` records to be returned for a parent
   * `TaskExecution`.
   * Note: Does not insert single `NodeExecution` records. Those must be inserted
   * separately.
   */
  insertTaskExecutionChildList(
    id: TaskExecutionIdentifier,
    data: RequireIdField<Partial<NodeExecution>>[],
  ): void;
  /** Inserts a single `Workflow` record. */
  insertWorkflow(data: RequireIdField<Partial<Workflow>>): void;
  /** Inserts a single `Execution` record. */
  insertWorkflowExecution(data: RequireIdField<Partial<Execution>>): void;
  /** Inserts a list of `Execution` records.
   * Note: Does not insert single `Execution` records. Those must be inserted
   * separately.
   */
  insertWorkflowExecutionList(
    scope: DomainIdentifierScope,
    data: RequireIdField<Partial<Execution>>[] | RequestError,
    query?: Record<string, string>,
  ): void;
  /** Debug utility which dumps the contents of the backing store. */
  printEntities(): void;
}

export interface CreateAdminServerResult {
  /** handlers to be inserted into mock-service-worker's `setupServer` function. */
  handlers: RequestHandlersList;
  /** The resulting `AdminServer` object, used for inserting mock data. */
  server: AdminServer;
}

enum EntityType {
  LaunchPlan = 'launchPlan',
  LaunchPlanList = 'launchPlanList',
  ProjectList = 'projectList',
  Workflow = 'workflow',
  Task = 'task',
  WorkflowExecution = 'workflowExecution',
  WorkflowExecutionList = 'workflowExecutionList',
  NodeExecution = 'nodeExecution',
  NodeExecutionList = 'nodeExecutionList',
  TaskExecution = 'taskExecution',
  TaskExecutionList = 'taskExecutionList',
  TaskExecutionChildList = 'taskExecutionChildList',
}

function taskExecutionIdFromParams({
  project,
  domain,
  name,
  nodeId,
  taskProject,
  taskDomain,
  taskName,
  taskVersion,
  retryAttempt,
}: RequestParams): TaskExecutionIdentifier {
  return {
    retryAttempt: Number.parseInt(retryAttempt, 10),
    nodeExecutionId: {
      nodeId,
      executionId: { project, domain, name },
    },
    taskId: taskIdentifier({
      project: taskProject,
      domain: taskDomain,
      name: taskName,
      version: taskVersion,
    }),
  };
}

/** Creates an instance of an Admin API mock server with an empty backing store. */
export function createAdminServer(): CreateAdminServerResult {
  const entityMap: Map<string, unknown> = new Map();
  const getWorkflowHandler = adminEntityHandler({
    path: '/workflows/:project/:domain/:name/:version',
    getDataForRequest: (req) => {
      const { project, domain, name, version } = req.params;
      const id = workflowIdentifier({
        project,
        domain,
        name,
        version,
      });
      return getItem(entityMap, [EntityType.Workflow, id]);
    },
    responseEncoder: Admin.Workflow,
  });

  const getTaskHandler = adminEntityHandler({
    path: '/tasks/:project/:domain/:name/:version',
    getDataForRequest: (req) => {
      const { project, domain, name, version } = req.params;
      const id = taskIdentifier({
        project,
        domain,
        name,
        version,
      });
      return getItem<Task>(entityMap, [EntityType.Task, id]);
    },
    responseEncoder: Admin.Task,
  });

  const getLaunchPlanHandler = adminEntityHandler({
    path: '/launch_plans/:project/:domain/:name/:version',
    getDataForRequest: (req) => {
      const { project, domain, name, version } = req.params;
      const id = launchPlanIdentifier({
        project,
        domain,
        name,
        version,
      });
      return getItem(entityMap, [EntityType.LaunchPlan, id]);
    },
    responseEncoder: Admin.LaunchPlan,
  });

  const getProjectListHandler = adminEntityHandler<Admin.IProjects>({
    path: '/projects',
    getDataForRequest: () => {
      const data = getItem<Project[]>(entityMap, EntityType.ProjectList);
      return { projects: data.map(Admin.Project.create) };
    },
    responseEncoder: Admin.Projects,
  });

  const getWorkflowExecutionListHandler = adminEntityHandler({
    path: '/executions/:project/:domain',
    getDataForRequest: (req) => {
      const { domain, project } = req.params;
      const scope: DomainIdentifierScope = { domain, project };
      const data = getItem<WorkflowExecutionIdentifier[]>(entityMap, [
        EntityType.WorkflowExecutionList,
        scope,
        workflowExecutionListQueryParams(getQueryParams(req)),
      ]);
      return {
        executions: data.map((executionId) =>
          Admin.Execution.create(getItem(entityMap, [EntityType.WorkflowExecution, executionId])),
        ),
      };
    },
    responseEncoder: Admin.ExecutionList,
  });

  const getWorkflowExecutionHandler = adminEntityHandler({
    path: '/executions/:project/:domain/:name',
    getDataForRequest: (req) => {
      const { project, domain, name } = req.params;
      const id: WorkflowExecutionIdentifier = { project, domain, name };
      return getItem(entityMap, [EntityType.WorkflowExecution, id]);
    },
    responseEncoder: Admin.Execution,
  });

  const getNodeExecutionListHandler = adminEntityHandler({
    path: '/node_executions/:project/:domain/:name',
    getDataForRequest: (req) => {
      const { project, domain, name } = req.params;
      const id: WorkflowExecutionIdentifier = { project, domain, name };
      const data = getItem<NodeExecutionIdentifier[]>(entityMap, [
        EntityType.NodeExecutionList,
        id,
        nodeExecutionListQueryParams(getQueryParams(req)),
      ]);
      return {
        nodeExecutions: data.map((nodeExecutionId) =>
          Admin.NodeExecution.create(
            getItem(entityMap, [EntityType.NodeExecution, nodeExecutionId]),
          ),
        ),
      };
    },
    responseEncoder: Admin.NodeExecutionList,
  });

  const getNodeExecutionHandler = adminEntityHandler({
    path: '/node_executions/:project/:domain/:name/:nodeId',
    getDataForRequest: (req) => {
      const { project, domain, name, nodeId } = req.params;
      const id: NodeExecutionIdentifier = {
        nodeId,
        executionId: { project, domain, name },
      };
      return getItem(entityMap, [EntityType.NodeExecution, id]);
    },
    responseEncoder: Admin.NodeExecution,
  });

  const getTaskExecutionHandler = adminEntityHandler({
    path: taskExecutionPath,
    getDataForRequest: (req) => {
      const id = taskExecutionIdFromParams(req.params);
      return getItem(entityMap, [EntityType.TaskExecution, id]);
    },
    responseEncoder: Admin.TaskExecution,
  });

  const getTaskExecutionListHandler = adminEntityHandler({
    path: '/task_executions/:project/:domain/:name/:nodeId',
    getDataForRequest: (req) => {
      const { project, domain, name, nodeId } = req.params;
      const id: NodeExecutionIdentifier = {
        executionId: { project, domain, name },
        nodeId,
      };
      const data = getItem<TaskExecutionIdentifier[]>(entityMap, [
        EntityType.TaskExecutionList,
        id,
      ]);
      return {
        taskExecutions: data.map((taskExecutionId) => {
          try {
            const execution = getItem<Admin.ITaskExecution>(entityMap, [
              EntityType.TaskExecution,
              normalizeTaskExecutionIdentifier(taskExecutionId),
            ]);
            return Admin.TaskExecution.create(execution);
          } catch (e) {
            throw unexpectedError(`Unexpected missing child item: ${obj(taskExecutionId)}`);
          }
        }),
      };
    },
    responseEncoder: Admin.TaskExecutionList,
  });

  const getTaskExecutionChildListHandler = adminEntityHandler({
    path: `/children${taskExecutionPath}`,
    getDataForRequest: (req) => {
      const id = taskExecutionIdFromParams(req.params);
      const data = getItem<NodeExecutionIdentifier[]>(entityMap, [
        EntityType.TaskExecutionChildList,
        id,
      ]);
      return {
        nodeExecutions: data.map((nodeExecutionId) => {
          try {
            const execution = getItem<Admin.INodeExecution>(entityMap, [
              EntityType.NodeExecution,
              nodeExecutionId,
            ]);
            return Admin.NodeExecution.create(execution);
          } catch (e) {
            throw unexpectedError(`Unexpected missing child item: ${obj(nodeExecutionId)}`);
          }
        }),
      };
    },
    responseEncoder: Admin.NodeExecutionList,
  });

  return {
    handlers: [
      getLaunchPlanHandler,
      getProjectListHandler,
      getWorkflowHandler,
      getTaskHandler,
      getWorkflowExecutionListHandler,
      getWorkflowExecutionHandler,
      getNodeExecutionHandler,
      getNodeExecutionListHandler,
      getTaskExecutionHandler,
      getTaskExecutionListHandler,
      getTaskExecutionChildListHandler,
    ],
    server: {
      insertLaunchPlan: (launchPlan) =>
        insertItem(
          entityMap,
          [EntityType.LaunchPlan, launchPlanIdentifier(launchPlan.id)],
          launchPlan,
        ),
      insertProjects: (projects) => insertItem(entityMap, EntityType.ProjectList, projects),
      insertTask: (task) => insertItem(entityMap, [EntityType.Task, taskIdentifier(task.id)], task),
      insertWorkflow: (workflow) =>
        insertItem(entityMap, [EntityType.Workflow, workflowIdentifier(workflow.id)], workflow),
      insertWorkflowExecution: (execution) =>
        insertItem(entityMap, [EntityType.WorkflowExecution, execution.id], execution),
      insertWorkflowExecutionList: (scope, data, query: QueryParamsMap = {}) =>
        insertItem(
          entityMap,
          [EntityType.WorkflowExecutionList, scope, workflowExecutionListQueryParams(query)],
          data instanceof RequestError ? data : data.map(({ id }) => id),
        ),
      insertNodeExecution: (execution) =>
        insertItem(entityMap, [EntityType.NodeExecution, execution.id], execution),
      insertNodeExecutionList: (parentExecutionId, data, query: QueryParamsMap = {}) =>
        insertItem(
          entityMap,
          [EntityType.NodeExecutionList, parentExecutionId, nodeExecutionListQueryParams(query)],
          data instanceof RequestError ? data : data.map(({ id }) => id),
        ),
      insertTaskExecution: (execution) =>
        insertItem(
          entityMap,
          [EntityType.TaskExecution, normalizeTaskExecutionIdentifier(execution.id)],
          execution,
        ),
      insertTaskExecutionList: (parentExecutionId, executions) =>
        insertItem(
          entityMap,
          [EntityType.TaskExecutionList, parentExecutionId],
          executions.map(({ id }) => normalizeTaskExecutionIdentifier(id)),
        ),
      insertTaskExecutionChildList: (parentExecutionId, executions) =>
        insertItem(
          entityMap,
          [EntityType.TaskExecutionChildList, normalizeTaskExecutionIdentifier(parentExecutionId)],
          executions.map(({ id }) => id),
        ),
      printEntities: () => console.log(Array.from(entityMap.entries())),
    },
  };
}
