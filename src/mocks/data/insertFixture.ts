import { AdminServer } from 'mocks/createAdminServer';
import { nodeExecutionQueryParams } from 'models/Execution/constants';
import {
  MockDataFixture,
  MockNodeExecutionData,
  MockTaskExecutionData,
  MockWorkflowExecutionData,
} from './fixtures/types';

function insertTaskExecutionData(server: AdminServer, mock: MockTaskExecutionData): void {
  server.insertTaskExecution(mock.data);
  if (mock.nodeExecutions) {
    const nodeExecutions = Object.values(mock.nodeExecutions);
    nodeExecutions.forEach((ne) => insertNodeExecutionData(server, ne));
    server.insertTaskExecutionChildList(
      mock.data.id,
      nodeExecutions.map(({ data }) => data),
    );
  }
}

function insertNodeExecutionData(server: AdminServer, mock: MockNodeExecutionData): void {
  server.insertNodeExecution(mock.data);
  if (mock.taskExecutions) {
    const taskExecutions = Object.values(mock.taskExecutions);
    taskExecutions.forEach((te) => insertTaskExecutionData(server, te));
    server.insertTaskExecutionList(
      mock.data.id,
      taskExecutions.map(({ data }) => data),
    );
  }

  if (mock.nodeExecutions) {
    const nodeExecutions = Object.values(mock.nodeExecutions);
    nodeExecutions.forEach((ne) => insertNodeExecutionData(server, ne));
    server.insertNodeExecutionList(
      mock.data.id.executionId,
      nodeExecutions.map(({ data }) => data),
      { [nodeExecutionQueryParams.parentNodeId]: mock.data.id.nodeId },
    );
  }
}

function insertWorkflowExecutionData(server: AdminServer, mock: MockWorkflowExecutionData): void {
  server.insertWorkflowExecution(mock.data);
  const nodeExecutions = Object.values(mock.nodeExecutions);
  nodeExecutions.forEach((ne) => insertNodeExecutionData(server, ne));
  server.insertNodeExecutionList(
    mock.data.id,
    nodeExecutions.map(({ data }) => data),
  );
}

/** Deep-inserts all entities from a generated `MockDataFixture`. */
export function insertFixture(
  server: AdminServer,
  { launchPlans, tasks, workflowExecutions, workflows }: MockDataFixture,
): void {
  if (launchPlans) {
    Object.values(launchPlans).forEach(server.insertLaunchPlan);
  }
  if (tasks) {
    Object.values(tasks).forEach(server.insertTask);
  }
  if (workflows) {
    Object.values(workflows).forEach(server.insertWorkflow);
  }
  if (workflowExecutions) {
    Object.values(workflowExecutions).forEach((execution) =>
      insertWorkflowExecutionData(server, execution),
    );
  }
}
