import { setupServer, SetupServerApi } from 'msw/node';
import { AdminServer, createAdminServer } from './createAdminServer';

const { handlers, server } = createAdminServer();
export type MockServer = SetupServerApi & AdminServer;
export const mockServer: MockServer = {
  ...setupServer(...handlers),
  ...server,
};
