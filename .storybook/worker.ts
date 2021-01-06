import { setupWorker, SetupWorkerApi } from 'msw';
import { AdminServer, createAdminServer } from '../src/mocks/createAdminServer';
import { basicPythonWorkflow } from '../src/mocks/data/fixtures/basicPythonWorkflow';
import { dynamicExternalSubWorkflow } from '../src/mocks/data/fixtures/dynamicExternalSubworkflow';
import { dynamicPythonNodeExecutionWorkflow } from '../src/mocks/data/fixtures/dynamicPythonWorkflow';
import { oneFailedTaskWorkflow } from '../src/mocks/data/fixtures/oneFailedTaskWorkflow';
import { insertFixture } from '../src/mocks/data/insertFixture';
import { insertDefaultData } from '../src/mocks/insertDefaultData';

const { handlers, server } = createAdminServer();
export type MockServer = SetupWorkerApi & AdminServer;
const worker: MockServer = { ...setupWorker(...handlers), ...server };
insertDefaultData(worker);
insertFixture(worker, basicPythonWorkflow.generate());
insertFixture(worker, dynamicExternalSubWorkflow.generate());
insertFixture(worker, dynamicPythonNodeExecutionWorkflow.generate());
insertFixture(worker, oneFailedTaskWorkflow.generate());

export { worker };
